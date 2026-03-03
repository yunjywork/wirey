package script

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SetGlobalVar sets a global variable
func (e *Engine) SetGlobalVar(key string, value interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.globalVars[key] = value
}

// GetGlobalVar gets a global variable
func (e *Engine) GetGlobalVar(key string) interface{} {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.globalVars[key]
}

// SetCollectionVar sets a collection variable and persists to storage
func (e *Engine) SetCollectionVar(collectionName, key string, value interface{}) {
	e.mu.Lock()
	if e.collectionVars[collectionName] == nil {
		e.collectionVars[collectionName] = make(map[string]interface{})
	}
	e.collectionVars[collectionName][key] = value

	// Copy vars for callback (to avoid holding lock during callback)
	var varsCopy map[string]interface{}
	if e.collectionVarSaveCallback != nil {
		varsCopy = make(map[string]interface{})
		for k, v := range e.collectionVars[collectionName] {
			varsCopy[k] = v
		}
	}
	cb := e.collectionVarSaveCallback
	e.mu.Unlock()

	// Call save callback outside of lock
	if cb != nil {
		cb(collectionName, varsCopy)
	}
}

// GetCollectionVar gets a collection variable
func (e *Engine) GetCollectionVar(collectionName, key string) interface{} {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.collectionVars[collectionName] == nil {
		return nil
	}
	return e.collectionVars[collectionName][key]
}

// SetCaseVar sets a case variable
func (e *Engine) SetCaseVar(caseID, key string, value interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.caseVars[caseID] == nil {
		e.caseVars[caseID] = make(map[string]interface{})
	}
	e.caseVars[caseID][key] = value
}

// GetCaseVar gets a case variable
func (e *Engine) GetCaseVar(caseID, key string) interface{} {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.caseVars[caseID] == nil {
		return nil
	}
	return e.caseVars[caseID][key]
}

// GetVar gets a variable with automatic scope resolution (case -> collection -> global)
func (e *Engine) GetVar(caseID, collectionName, key string) interface{} {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check case scope first
	if e.caseVars[caseID] != nil {
		if v, ok := e.caseVars[caseID][key]; ok {
			return v
		}
	}

	// Check collection scope
	if e.collectionVars[collectionName] != nil {
		if v, ok := e.collectionVars[collectionName][key]; ok {
			return v
		}
	}

	// Check global scope
	return e.globalVars[key]
}

// LoadVariables loads variables from stored config into the engine
// Collection variables are only loaded if the collection doesn't exist yet (to preserve script changes)
func (e *Engine) LoadVariables(caseID, collectionName string, caseVars, collectionVars, globalVars map[string]interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Load case variables (always merge - case vars are session-specific)
	if caseVars != nil {
		if e.caseVars[caseID] == nil {
			e.caseVars[caseID] = make(map[string]interface{})
		}
		for k, v := range caseVars {
			e.caseVars[caseID][k] = v
		}
	}

	// Load collection variables only if not already loaded
	// (to preserve runtime changes from wirey.collection.set)
	if collectionVars != nil && e.collectionVars[collectionName] == nil {
		e.collectionVars[collectionName] = make(map[string]interface{})
		for k, v := range collectionVars {
			e.collectionVars[collectionName][k] = v
		}
	}

	// Load global variables
	if globalVars != nil {
		for k, v := range globalVars {
			e.globalVars[k] = v
		}
	}
}

// ClearCaseVariables clears all variables for a specific case
func (e *Engine) ClearCaseVariables(caseID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.caseVars, caseID)
}

// ClearCollectionVariables clears all variables for a specific collection
func (e *Engine) ClearCollectionVariables(collectionName string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.collectionVars, collectionName)
}

// SyncCollectionVariables replaces all variables for a collection
// (used when UI updates variables to sync with engine memory)
func (e *Engine) SyncCollectionVariables(collectionName string, vars map[string]interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if vars == nil {
		delete(e.collectionVars, collectionName)
	} else {
		e.collectionVars[collectionName] = make(map[string]interface{})
		for k, v := range vars {
			e.collectionVars[collectionName][k] = v
		}
	}
}

// GetBuiltinVariable returns a built-in variable value
// Returns (value as string, exists)
func (e *Engine) GetBuiltinVariable(key, caseID string) (string, bool) {
	now := time.Now()
	switch key {
	case "timestamp":
		return strconv.FormatInt(now.Unix(), 10), true
	case "timestamp_ms":
		return strconv.FormatInt(now.UnixMilli(), 10), true
	case "uuid":
		return uuid.New().String(), true
	case "random":
		return strconv.Itoa(rand.Intn(1000000)), true
	case "counter":
		return strconv.FormatInt(e.GetAndIncrementCounter(caseID, "default"), 10), true
	case "date":
		return now.UTC().Format("2006-01-02"), true
	case "time":
		return now.UTC().Format("15:04:05"), true
	case "datetime":
		return now.UTC().Format(time.RFC3339), true
	}
	return "", false
}

// GetAndIncrementCounter gets the current counter value and increments it
func (e *Engine) GetAndIncrementCounter(caseID, name string) int64 {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := caseID + ":{{counter}}"
	if name != "default" {
		key = caseID + ":{{counter:" + name + "}}"
	}

	if _, exists := e.counters[key]; !exists {
		e.counters[key] = 1
	}

	value := e.counters[key]
	e.counters[key]++
	return value
}

// ResetCounter resets the counter for a specific case
func (e *Engine) ResetCounter(caseID string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Remove all counters for this case
	for key := range e.counters {
		if strings.HasPrefix(key, caseID+":") {
			delete(e.counters, key)
		}
	}
}
