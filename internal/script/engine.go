package script

import (
	"bytes"
	cryptoRand "crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/yunjywork/wirey/internal/preprocess"
	"github.com/dop251/goja"
	"github.com/google/uuid"
)

// HTTP client with 10 second timeout
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

var (
	ErrScriptCancelled = errors.New("script cancelled send")
	ErrScriptTimeout   = errors.New("script execution timeout")
)

// LogCallback is called when wirey.log() is invoked in a script
type LogCallback func(caseID string, args []interface{})

// CollectionVarSaveCallback is called when wirey.collection.set() is invoked
// to persist the change to storage
type CollectionVarSaveCallback func(collectionName string, vars map[string]interface{})

// Engine represents a JavaScript execution engine
type Engine struct {
	mu sync.Mutex

	// Variable stores (hierarchical: case -> collection -> global)
	globalVars     map[string]interface{}
	collectionVars map[string]map[string]interface{} // collectionName -> vars
	caseVars       map[string]map[string]interface{} // caseID -> vars

	// Counter state (per session)
	counters map[string]int64

	// Log callback
	logCallback LogCallback

	// Collection variable save callback
	collectionVarSaveCallback CollectionVarSaveCallback

	// Log capture mode (for dry run)
	captureMode  bool
	capturedLogs []string
}

// NewEngine creates a new script engine
func NewEngine() *Engine {
	return &Engine{
		globalVars:     make(map[string]interface{}),
		collectionVars: make(map[string]map[string]interface{}),
		caseVars:       make(map[string]map[string]interface{}),
		counters:       make(map[string]int64),
	}
}

// Global engine instance
var defaultEngine = NewEngine()

// GetEngine returns the global engine instance
func GetEngine() *Engine {
	return defaultEngine
}

// SetLogCallback sets the callback for wirey.log()
func (e *Engine) SetLogCallback(cb LogCallback) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.logCallback = cb
}

// SetCollectionVarSaveCallback sets the callback for persisting collection variables
func (e *Engine) SetCollectionVarSaveCallback(cb CollectionVarSaveCallback) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.collectionVarSaveCallback = cb
}

// StartCapture enables log capture mode for dry run
func (e *Engine) StartCapture() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.captureMode = true
	e.capturedLogs = nil
}

// StopCapture disables log capture mode and returns captured logs
func (e *Engine) StopCapture() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.captureMode = false
	logs := e.capturedLogs
	e.capturedLogs = nil
	return logs
}

// appendCapturedLog adds a log entry during capture mode (must hold lock)
func (e *Engine) appendCapturedLog(message string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.captureMode {
		e.capturedLogs = append(e.capturedLogs, message)
	}
}

// RunSetupScript executes a setup script (no return value expected)
func (e *Engine) RunSetupScript(script, caseID, collectionName string) error {
	vm := goja.New()

	// Inject wirey API
	e.injectWireyAPI(vm, caseID, collectionName, "SET")

	// Execute script
	_, err := vm.RunString(script)
	if err != nil {
		return fmt.Errorf("setup script error: %w", err)
	}

	return nil
}

// RunPreSendScript executes a pre-send script (returns processed message or nil to cancel)
func (e *Engine) RunPreSendScript(script, message, caseID, collectionName string) (string, error) {
	vm := goja.New()

	// Inject wirey API
	e.injectWireyAPI(vm, caseID, collectionName, "PRE")

	// Set the message variable
	vm.Set("msg", message)

	// Wrap script in a function to handle return
	wrappedScript := fmt.Sprintf(`(function() { %s })()`, script)

	// Execute script
	result, err := vm.RunString(wrappedScript)
	if err != nil {
		return "", fmt.Errorf("pre-send script error: %w", err)
	}

	// Handle return value
	if goja.IsNull(result) || goja.IsUndefined(result) {
		// null means cancel
		if goja.IsNull(result) {
			return "", ErrScriptCancelled
		}
		// undefined means no return, send empty
		return "", nil
	}

	// Convert result to string
	return result.String(), nil
}

// RunPostRecvScript executes a post-receive script (side effects only, no return value)
func (e *Engine) RunPostRecvScript(script, message, caseID, collectionName string) error {
	vm := goja.New()

	// Inject wirey API
	e.injectWireyAPI(vm, caseID, collectionName, "POST")

	// Set the received message variable
	vm.Set("msg", message)

	// Execute script
	_, err := vm.RunString(script)
	if err != nil {
		return fmt.Errorf("post-recv script error: %w", err)
	}

	return nil
}

// injectWireyAPI injects the wirey object into the VM
func (e *Engine) injectWireyAPI(vm *goja.Runtime, caseID, collectionName, scriptType string) {
	wirey := vm.NewObject()

	// wirey.get(key) - builtin + case scope only
	wirey.Set("get", func(key string) interface{} {
		// Check built-in variables first
		if val, ok := e.GetBuiltinVariable(key, caseID); ok {
			return val
		}
		// Case scope only (no collection/global fallback)
		return e.GetCaseVar(caseID, key)
	})

	// wirey.set(key, value) - sets in case scope
	wirey.Set("set", func(key string, value interface{}) {
		e.SetCaseVar(caseID, key, value)
	})

	// wirey.collection.get/set - collection scope (persisted)
	collectionObj := vm.NewObject()
	collectionObj.Set("get", func(key string) interface{} {
		return e.GetCollectionVar(collectionName, key)
	})
	collectionObj.Set("set", func(key string, value interface{}) {
		e.SetCollectionVar(collectionName, key, value)
	})
	wirey.Set("collection", collectionObj)

	// Utility functions
	wirey.Set("randomHex", func(n int) string {
		if n <= 0 || n > 256 {
			n = 4
		}
		bytes := make([]byte, n)
		cryptoRand.Read(bytes)
		return strings.ToUpper(hex.EncodeToString(bytes))
	})

	wirey.Set("toHex", func(s string) string {
		return strings.ToUpper(hex.EncodeToString([]byte(s)))
	})

	wirey.Set("fromHex", func(s string) string {
		// Remove spaces from hex string
		s = strings.ReplaceAll(s, " ", "")
		bytes, err := hex.DecodeString(s)
		if err != nil {
			return ""
		}
		return string(bytes)
	})

	// wirey.toBytes(str) - string → byte array (JS array)
	wirey.Set("toBytes", func(s string) []byte {
		return []byte(s)
	})

	// wirey.fromBytes(bytes) - byte array → string
	wirey.Set("fromBytes", func(bytes []byte) string {
		return string(bytes)
	})

	// wirey.subBytes(str, start, end?) - extract bytes as string
	wirey.Set("subBytes", func(s string, start int, end ...int) string {
		data := []byte(s)
		if start < 0 {
			start = 0
		}
		if start >= len(data) {
			return ""
		}
		endIdx := len(data)
		if len(end) > 0 && end[0] < len(data) {
			endIdx = end[0]
		}
		if endIdx <= start {
			return ""
		}
		return string(data[start:endIdx])
	})

	// wirey.appendBytes(str, bytes) - append bytes to string
	wirey.Set("appendBytes", func(s string, toAppend interface{}) string {
		data := []byte(s)
		switch v := toAppend.(type) {
		case []byte:
			data = append(data, v...)
		case []interface{}:
			for _, b := range v {
				if n, ok := b.(int64); ok {
					data = append(data, byte(n))
				} else if n, ok := b.(float64); ok {
					data = append(data, byte(int(n)))
				}
			}
		case string:
			data = append(data, []byte(v)...)
		}
		return string(data)
	})

	// wirey.replaceBytes(str, start, bytes) - replace bytes at position
	wirey.Set("replaceBytes", func(s string, start int, replacement interface{}) string {
		data := []byte(s)
		if start < 0 || start >= len(data) {
			return s
		}
		var replaceData []byte
		switch v := replacement.(type) {
		case []byte:
			replaceData = v
		case []interface{}:
			for _, b := range v {
				if n, ok := b.(int64); ok {
					replaceData = append(replaceData, byte(n))
				} else if n, ok := b.(float64); ok {
					replaceData = append(replaceData, byte(int(n)))
				}
			}
		case string:
			replaceData = []byte(v)
		}
		// Replace bytes starting at position
		for i, b := range replaceData {
			if start+i < len(data) {
				data[start+i] = b
			}
		}
		return string(data)
	})

	// wirey.byteAt(str, index) - get byte value at position
	wirey.Set("byteAt", func(s string, index int) int {
		data := []byte(s)
		if index < 0 || index >= len(data) {
			return -1
		}
		return int(data[index])
	})

	// wirey.setByteAt(str, index, value) - set byte value at position
	wirey.Set("setByteAt", func(s string, index int, value int) string {
		data := []byte(s)
		if index < 0 || index >= len(data) {
			return s
		}
		data[index] = byte(value)
		return string(data)
	})

	wirey.Set("uuid", func() string {
		return uuid.New().String()
	})

	// wirey.httpGet(url) - HTTP GET request
	wirey.Set("httpGet", func(url string) map[string]interface{} {
		resp, err := httpClient.Get(url)
		if err != nil {
			return map[string]interface{}{
				"error":  err.Error(),
				"status": 0,
				"body":   "",
			}
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		return map[string]interface{}{
			"error":  nil,
			"status": resp.StatusCode,
			"body":   string(body),
		}
	})

	// wirey.httpPost(url, body, contentType?) - HTTP POST request
	wirey.Set("httpPost", func(url string, body string, contentType ...string) map[string]interface{} {
		ct := "application/json"
		if len(contentType) > 0 && contentType[0] != "" {
			ct = contentType[0]
		}

		resp, err := httpClient.Post(url, ct, bytes.NewBufferString(body))
		if err != nil {
			return map[string]interface{}{
				"error":  err.Error(),
				"status": 0,
				"body":   "",
			}
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		return map[string]interface{}{
			"error":  nil,
			"status": resp.StatusCode,
			"body":   string(respBody),
		}
	})

	// wirey.log() - output to message log or capture
	wirey.Set("log", func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.Export()
		}

		// Format message
		var parts []string
		for _, arg := range args {
			parts = append(parts, fmt.Sprintf("%v", arg))
		}
		message := strings.Join(parts, " ")

		// Get line number from call stack
		stack := vm.CaptureCallStack(10, nil)
		lineInfo := ""
		for _, frame := range stack {
			pos := frame.Position()
			if pos.Line > 0 {
				lineInfo = fmt.Sprintf("[%s:%d] ", scriptType, pos.Line)
				break
			}
		}
		messageWithLine := lineInfo + message

		// Capture mode: store logs instead of emitting
		e.mu.Lock()
		captureMode := e.captureMode
		if captureMode {
			e.capturedLogs = append(e.capturedLogs, messageWithLine)
		}
		e.mu.Unlock()

		// Normal mode: emit via callback (with line info)
		if !captureMode && e.logCallback != nil {
			e.logCallback(caseID, []interface{}{messageWithLine})
		}

		return goja.Undefined()
	})

	vm.Set("wirey", wirey)
}

// ProcessMessage runs the full preprocessing pipeline
func (e *Engine) ProcessMessage(
	message string,
	caseID, collectionName string,
	collectionSetupScript, collectionPreSendScript string,
	collectionSetupEnabled, collectionPreSendEnabled bool,
	caseSetupScript, casePreSendScript string,
	caseSetupEnabled, casePreSendEnabled bool,
) (string, error) {
	// 1. Run Collection Setup script
	if collectionSetupEnabled && collectionSetupScript != "" {
		if err := e.RunSetupScript(collectionSetupScript, caseID, collectionName); err != nil {
			return "", err
		}
	}

	// 2. Run Case Setup script
	if caseSetupEnabled && caseSetupScript != "" {
		if err := e.RunSetupScript(caseSetupScript, caseID, collectionName); err != nil {
			return "", err
		}
	}

	// 3. Replace variables (built-in + custom)
	message = e.ReplaceVariables(message, caseID, collectionName)

	// 4. Process escape sequences (\n → 0x0a, etc.)
	message = preprocess.ProcessEscapeSequences(message)

	// 5. Run Collection Pre-send script
	if collectionPreSendEnabled && collectionPreSendScript != "" {
		var err error
		message, err = e.RunPreSendScript(collectionPreSendScript, message, caseID, collectionName)
		if err != nil {
			return "", err
		}
	}

	// 6. Run Case Pre-send script
	if casePreSendEnabled && casePreSendScript != "" {
		var err error
		message, err = e.RunPreSendScript(casePreSendScript, message, caseID, collectionName)
		if err != nil {
			return "", err
		}
	}

	return message, nil
}
