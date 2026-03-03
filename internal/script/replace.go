package script

import (
	cryptoRand "crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Regex patterns for variable replacement
var (
	randomRegex    = regexp.MustCompile(`\{\{random(?::(\d+))?\}\}`)
	counterRegex   = regexp.MustCompile(`\{\{counter(?::(\d+))?\}\}`)
	customVarRegex = regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)
)

// ReplaceVariables replaces {{variable}} patterns with actual values
func (e *Engine) ReplaceVariables(message, caseID, collectionName string) string {
	// Simple built-in variables (same value for all occurrences)
	simpleVars := []string{"timestamp", "timestamp_ms", "datetime", "date", "time"}
	for _, varName := range simpleVars {
		if val, ok := e.GetBuiltinVariable(varName, caseID); ok {
			message = strings.ReplaceAll(message, "{{"+varName+"}}", val)
		}
	}

	// UUID - each occurrence gets a NEW value
	message = replaceUUID(message)

	// Random - each occurrence gets a NEW value (supports {{random:N}})
	message = replaceRandom(message)

	// Counter - each occurrence increments (supports {{counter:N}})
	message = e.replaceCounter(message, caseID)

	// Custom variables from scripts
	message = e.replaceCustomVariables(message, caseID, collectionName)

	return message
}

// replaceUUID replaces {{uuid}} with a new UUID v4
func replaceUUID(message string) string {
	for strings.Contains(message, "{{uuid}}") {
		message = strings.Replace(message, "{{uuid}}", uuid.New().String(), 1)
	}
	return message
}

// replaceRandom replaces {{random}} or {{random:N}} with N bytes of random hex
func replaceRandom(message string) string {
	return randomRegex.ReplaceAllStringFunc(message, func(match string) string {
		matches := randomRegex.FindStringSubmatch(match)

		n := 4
		if len(matches) >= 2 && matches[1] != "" {
			parsed, err := strconv.Atoi(matches[1])
			if err != nil || parsed <= 0 || parsed > 256 {
				return match
			}
			n = parsed
		}

		bytes := make([]byte, n)
		if _, err := cryptoRand.Read(bytes); err != nil {
			return match
		}

		return strings.ToUpper(hex.EncodeToString(bytes))
	})
}

// replaceCounter replaces {{counter}} and {{counter:N}} with auto-incrementing values
func (e *Engine) replaceCounter(message, caseID string) string {
	e.mu.Lock()
	defer e.mu.Unlock()

	return counterRegex.ReplaceAllStringFunc(message, func(match string) string {
		matches := counterRegex.FindStringSubmatch(match)

		startValue := int64(1)
		if len(matches) >= 2 && matches[1] != "" {
			if v, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
				startValue = v
			}
		}

		// Use caseID + match pattern as key
		key := caseID + ":" + match
		if _, exists := e.counters[key]; !exists {
			e.counters[key] = startValue
		}

		value := e.counters[key]
		e.counters[key]++

		return strconv.FormatInt(value, 10)
	})
}

// replaceCustomVariables replaces {{varName}} with custom variable values
func (e *Engine) replaceCustomVariables(message, caseID, collectionName string) string {
	return customVarRegex.ReplaceAllStringFunc(message, func(match string) string {
		// Skip built-in variables
		builtins := []string{"{{timestamp}}", "{{timestamp_ms}}", "{{datetime}}", "{{date}}", "{{time}}", "{{uuid}}"}
		for _, b := range builtins {
			if match == b {
				return match
			}
		}

		// Extract variable name
		matches := customVarRegex.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}

		varName := matches[1]

		// Skip counter and random (already handled)
		if varName == "counter" || varName == "random" {
			return match
		}

		// Look up variable value
		value := e.GetVar(caseID, collectionName, varName)
		if value == nil {
			return match // Keep original if not found
		}

		// Convert to string
		switch v := value.(type) {
		case string:
			return v
		case int, int64, int32:
			return fmt.Sprintf("%d", v)
		case float64, float32:
			return fmt.Sprintf("%v", v)
		case bool:
			return fmt.Sprintf("%t", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	})
}
