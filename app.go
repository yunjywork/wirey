package main

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"github.com/yunjywork/wirey/internal/charset"
	"github.com/yunjywork/wirey/internal/config"
	"github.com/yunjywork/wirey/internal/echo"
	"github.com/yunjywork/wirey/internal/framing"
	"github.com/yunjywork/wirey/internal/models"
	"github.com/yunjywork/wirey/internal/preprocess"
	"github.com/yunjywork/wirey/internal/script"
	"github.com/yunjywork/wirey/internal/socket"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx          context.Context
	manager      *socket.Manager
	storage      *config.Storage
	echoServer   *echo.Server
	scriptEngine *script.Engine
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize storage
	storage, err := config.NewStorage()
	if err != nil {
		runtime.LogError(ctx, "Failed to initialize storage: "+err.Error())
	}
	a.storage = storage

	// Initialize script engine with log callback
	a.scriptEngine = script.NewEngine()
	a.scriptEngine.SetLogCallback(func(caseID string, args []interface{}) {
		// Format args as string
		var parts []string
		for _, arg := range args {
			parts = append(parts, fmt.Sprintf("%v", arg))
		}
		message := strings.Join(parts, " ")

		// Emit script log event
		runtime.EventsEmit(ctx, "script:log", map[string]interface{}{
			"caseId":    caseID,
			"message":   message,
			"timestamp": time.Now().UnixMilli(),
		})
	})

	// Set collection variable save callback to persist changes
	a.scriptEngine.SetCollectionVarSaveCallback(func(collectionName string, vars map[string]interface{}) {
		if a.storage == nil {
			return
		}

		// Load current collection, update sharedVariables, and save
		collections, err := a.storage.LoadCollections()
		if err != nil {
			runtime.LogError(ctx, "Failed to load collections for variable save: "+err.Error())
			return
		}

		for _, colWithCases := range collections {
			if colWithCases.Collection.Name == collectionName {
				// Update sharedVariables
				colWithCases.Collection.SharedVariables = vars
				if err := a.storage.UpdateCollection(collectionName, colWithCases.Collection); err != nil {
					runtime.LogError(ctx, "Failed to save collection variables: "+err.Error())
					return
				}

				// Emit event to frontend to update UI
				runtime.EventsEmit(ctx, "collection:varsUpdated", map[string]interface{}{
					"collectionName": collectionName,
					"variables":      vars,
				})
				break
			}
		}
	})

	// Initialize sample data on first run
	if a.storage != nil && a.storage.IsFirstRun() {
		if err := a.storage.InitializeSamples(); err != nil {
			runtime.LogWarning(ctx, "Failed to create sample data: "+err.Error())
		}
	}

	// Initialize echo server with log callback
	a.echoServer = echo.NewServer(func(entry echo.LogEntry) {
		runtime.EventsEmit(ctx, "echo:log", entry)
	})

	// Initialize socket manager with event callbacks
	a.manager = socket.NewManager(
		// onData callback (received data)
		func(caseID string, rawBytes []byte, decodedContent string, meta models.FramingMeta, localAddr, remoteAddr string, timestamp int64) {
			// Run post-recv scripts if enabled
			a.runPostRecvScripts(caseID, decodedContent)

			runtime.EventsEmit(ctx, "socket:data", map[string]interface{}{
				"caseId":      caseID,
				"content":     decodedContent,
				"rawBytes":    base64.StdEncoding.EncodeToString(rawBytes),
				"size":        len(rawBytes),
				"localAddr":   localAddr,
				"remoteAddr":  remoteAddr,
				"framingInfo": meta,
				"timestamp":   timestamp,
			})
		},
		// onSent callback (sent data)
		func(caseID string, rawBytes []byte, originalContent string, meta models.FramingMeta, localAddr, remoteAddr string, timestamp int64) {
			runtime.EventsEmit(ctx, "socket:sent", map[string]interface{}{
				"caseId":      caseID,
				"content":     originalContent,
				"rawBytes":    base64.StdEncoding.EncodeToString(rawBytes),
				"size":        len(rawBytes),
				"localAddr":   localAddr,
				"remoteAddr":  remoteAddr,
				"framingInfo": meta,
				"timestamp":   timestamp,
			})
		},
		// onError callback
		func(caseID, err string) {
			runtime.EventsEmit(ctx, "socket:error", map[string]string{
				"caseId": caseID,
				"error":  err,
			})
		},
		// onStatus callback
		func(caseID string, info models.ConnectionStatusInfo) {
			runtime.EventsEmit(ctx, "socket:status", map[string]interface{}{
				"caseId":     caseID,
				"status":     info.Status,
				"localAddr":  info.LocalAddr,
				"remoteAddr": info.RemoteAddr,
				"protocol":   info.Protocol,
				"reason":     info.Reason,
				"duration":   info.Duration,
				"bytesSent":  info.BytesSent,
				"bytesRecv":  info.BytesRecv,
			})
		},
	)
}

// runPostRecvScripts executes post-recv scripts for a received message
func (a *App) runPostRecvScripts(caseID, message string) {
	if a.scriptEngine == nil || a.storage == nil {
		return
	}

	// Load collections to find the case and its script config
	collections, err := a.storage.LoadCollections()
	if err != nil {
		return
	}

	var collectionName string
	var collectionScriptConfig *config.ScriptConfig
	var caseScriptConfig *config.ScriptConfig
	var collectionVars config.Variables
	var caseVars config.Variables

	// Find the case in collections
	for _, col := range collections {
		for _, c := range col.Cases {
			if c.ID == caseID {
				collectionName = col.Collection.Name
				collectionScriptConfig = col.Collection.SharedScriptConfig
				caseScriptConfig = c.ScriptConfig
				collectionVars = col.Collection.SharedVariables
				caseVars = c.LocalVariables
				break
			}
		}
		if collectionName != "" {
			break
		}
	}

	// If case not found or no post-recv scripts enabled, return
	if collectionName == "" {
		return
	}

	collPostRecvEnabled := collectionScriptConfig != nil && collectionScriptConfig.PostRecvEnabled && collectionScriptConfig.PostRecvScript != ""
	casePostRecvEnabled := caseScriptConfig != nil && caseScriptConfig.PostRecvEnabled && caseScriptConfig.PostRecvScript != ""

	if !collPostRecvEnabled && !casePostRecvEnabled {
		return
	}

	// Load variables into engine
	a.scriptEngine.LoadVariables(caseID, collectionName, caseVars, collectionVars, nil)

	// Run collection post-recv script first
	if collPostRecvEnabled {
		if err := a.scriptEngine.RunPostRecvScript(collectionScriptConfig.PostRecvScript, message, caseID, collectionName); err != nil {
			// Log error but don't interrupt message display
			runtime.LogWarning(a.ctx, "Collection post-recv script error: "+err.Error())
		}
	}

	// Run case post-recv script
	if casePostRecvEnabled {
		if err := a.scriptEngine.RunPostRecvScript(caseScriptConfig.PostRecvScript, message, caseID, collectionName); err != nil {
			runtime.LogWarning(a.ctx, "Case post-recv script error: "+err.Error())
		}
	}
}

// shutdown is called when the app is closing
func (a *App) shutdown(ctx context.Context) {
	if a.manager != nil {
		a.manager.DisconnectAll()
	}
	if a.echoServer != nil {
		a.echoServer.Stop()
	}
}

// Connect establishes a socket connection
func (a *App) Connect(caseID string, protocol string, host string, port int, framingCfg *config.FramingConfig, charset string, connectionSettings *config.ConnectionSettings) error {
	var proto models.Protocol
	switch protocol {
	case "tcp":
		proto = models.ProtocolTCP
	case "udp":
		proto = models.ProtocolUDP
	default:
		proto = models.ProtocolTCP
	}

	// Convert config.FramingConfig to framing.Config
	fc := toFramingConfig(framingCfg)

	// Default charset
	if charset == "" {
		charset = "utf-8"
	}

	// Extract timeout values (use 0 for defaults)
	connectTimeoutMs := 0
	readTimeoutMs := 0
	if connectionSettings != nil {
		connectTimeoutMs = connectionSettings.ConnectTimeout
		readTimeoutMs = connectionSettings.ReadTimeout
	}

	return a.manager.Connect(caseID, proto, host, port, fc, charset, connectTimeoutMs, readTimeoutMs)
}

// UpdateFraming updates framing configuration for a connected session
func (a *App) UpdateFraming(caseID string, framingCfg *config.FramingConfig) error {
	fc := toFramingConfig(framingCfg)
	return a.manager.UpdateFraming(caseID, fc)
}

// UpdateCharset updates charset for a connected session
func (a *App) UpdateCharset(caseID string, charset string) error {
	return a.manager.UpdateCharset(caseID, charset)
}

// toFramingConfig converts config.FramingConfig to framing.Config with defaults
func toFramingConfig(cfg *config.FramingConfig) framing.Config {
	fc := framing.DefaultConfig()
	if cfg == nil {
		return fc
	}

	if cfg.Mode != "" {
		fc.Mode = cfg.Mode
	}
	if cfg.Delimiter != "" {
		fc.Delimiter = cfg.Delimiter
	}
	if cfg.LengthEncoding != "" {
		fc.LengthEncoding = cfg.LengthEncoding
	}
	// LengthOffset can be 0, so always set it
	fc.LengthOffset = cfg.LengthOffset
	if cfg.LengthBytes > 0 {
		fc.LengthBytes = cfg.LengthBytes
	}
	if cfg.Endian != "" {
		fc.Endian = cfg.Endian
	}
	fc.IncludeHeader = cfg.IncludeHeader
	if cfg.LengthMode != "" {
		fc.LengthMode = cfg.LengthMode
	}
	if cfg.FixedSize > 0 {
		fc.FixedSize = cfg.FixedSize
	}

	return fc
}

// Disconnect closes a socket connection
func (a *App) Disconnect(caseID string) error {
	return a.manager.Disconnect(caseID)
}

// Send sends data through a socket connection
func (a *App) Send(caseID string, data string, format string, useVariables bool) error {
	var fmt models.MessageFormat
	switch format {
	case "hex":
		fmt = models.FormatHex
	default:
		fmt = models.FormatText
	}

	// Apply preprocessing (variables + escape sequences) for text format
	if fmt == models.FormatText {
		processed, err := preprocess.Process(data, useVariables)
		if err != nil {
			return err
		}
		data = processed
	}

	return a.manager.Send(caseID, data, fmt)
}

// PreprocessMessage applies variable substitution and escape sequences for preview
func (a *App) PreprocessMessage(message string, useVariables bool) (string, error) {
	return preprocess.Process(message, useVariables)
}

// PreviewMessage returns a preview of the preprocessed message (without side effects)
func (a *App) PreviewMessage(message string) string {
	return preprocess.Preview(message)
}

// GetSupportedVariables returns the list of supported built-in variables
func (a *App) GetSupportedVariables() []string {
	return preprocess.SupportedVariables()
}

// GetSupportedEscapeSequences returns the supported escape sequences
func (a *App) GetSupportedEscapeSequences() map[string]string {
	return preprocess.SupportedEscapeSequences()
}

// ScriptProcessRequest contains all parameters for script processing
type ScriptProcessRequest struct {
	Message                  string             `json:"message"`
	CaseID                   string             `json:"caseId"`
	CollectionName           string             `json:"collectionName"`
	CollectionScriptConfig   *config.ScriptConfig `json:"collectionScriptConfig"`
	CaseScriptConfig         *config.ScriptConfig `json:"caseScriptConfig"`
	CollectionVariables      config.Variables   `json:"collectionVariables"`
	CaseVariables            config.Variables   `json:"caseVariables"`
}

// ProcessMessageWithScripts runs the full script preprocessing pipeline
func (a *App) ProcessMessageWithScripts(req ScriptProcessRequest) (string, error) {
	if a.scriptEngine == nil {
		return req.Message, nil
	}

	// Load collection variables from storage (source of truth)
	var collectionVars config.Variables
	if a.storage != nil {
		collections, err := a.storage.LoadCollections()
		if err == nil {
			for _, c := range collections {
				if c.Collection.Name == req.CollectionName {
					collectionVars = c.Collection.SharedVariables
					break
				}
			}
		}
	}

	// Load variables into engine
	a.scriptEngine.LoadVariables(
		req.CaseID,
		req.CollectionName,
		req.CaseVariables,
		collectionVars,
		nil, // global variables not used for now
	)

	// Extract script configs
	var collSetup, collPreSend, caseSetup, casePreSend string
	var collSetupEnabled, collPreSendEnabled, caseSetupEnabled, casePreSendEnabled bool

	if req.CollectionScriptConfig != nil {
		collSetup = req.CollectionScriptConfig.SetupScript
		collSetupEnabled = req.CollectionScriptConfig.SetupEnabled
		collPreSend = req.CollectionScriptConfig.PreSendScript
		collPreSendEnabled = req.CollectionScriptConfig.PreSendEnabled
	}

	if req.CaseScriptConfig != nil {
		caseSetup = req.CaseScriptConfig.SetupScript
		caseSetupEnabled = req.CaseScriptConfig.SetupEnabled
		casePreSend = req.CaseScriptConfig.PreSendScript
		casePreSendEnabled = req.CaseScriptConfig.PreSendEnabled
	}

	// Run the full pipeline
	return a.scriptEngine.ProcessMessage(
		req.Message,
		req.CaseID,
		req.CollectionName,
		collSetup, collPreSend,
		collSetupEnabled, collPreSendEnabled,
		caseSetup, casePreSend,
		caseSetupEnabled, casePreSendEnabled,
	)
}

// ValidateScript checks if a script has valid JavaScript syntax
func (a *App) ValidateScript(scriptContent string) error {
	// Simple validation by trying to parse
	vm := script.NewEngine()
	return vm.RunSetupScript(scriptContent, "validate", "validate")
}

// ScriptDryRunRequest contains parameters for script dry run
type ScriptDryRunRequest struct {
	Message             string               `json:"message"`
	CaseID              string               `json:"caseId"`
	CollectionName      string               `json:"collectionName"`
	CaseScriptConfig    *config.ScriptConfig `json:"caseScriptConfig"`
	CollectionVariables config.Variables     `json:"collectionVariables"`
	CaseVariables       config.Variables     `json:"caseVariables"`
}

// ScriptDryRunResponse contains the dry run results
type ScriptDryRunResponse struct {
	Result string   `json:"result"`
	Logs   []string `json:"logs"`
	Error  string   `json:"error,omitempty"`
}

// DryRunScript runs a script without sending, returns result and captured logs
func (a *App) DryRunScript(req ScriptDryRunRequest) ScriptDryRunResponse {
	if a.scriptEngine == nil {
		return ScriptDryRunResponse{Result: req.Message, Logs: []string{}}
	}

	// Use temporary case ID for isolation
	tempCaseID := "dryrun-" + req.CaseID

	// Load collection variables from storage (source of truth)
	var collectionVars config.Variables
	if a.storage != nil {
		collections, err := a.storage.LoadCollections()
		if err == nil {
			for _, c := range collections {
				if c.Collection.Name == req.CollectionName {
					collectionVars = c.Collection.SharedVariables
					break
				}
			}
		}
	}

	// Load variables into engine (use temp case ID but real collection name)
	a.scriptEngine.LoadVariables(
		tempCaseID,
		req.CollectionName,
		req.CaseVariables,
		collectionVars,
		nil,
	)

	// Start log capture
	a.scriptEngine.StartCapture()

	// Extract script config
	var caseSetup, casePreSend string
	var caseSetupEnabled, casePreSendEnabled bool

	if req.CaseScriptConfig != nil {
		caseSetup = req.CaseScriptConfig.SetupScript
		caseSetupEnabled = req.CaseScriptConfig.SetupEnabled
		casePreSend = req.CaseScriptConfig.PreSendScript
		casePreSendEnabled = req.CaseScriptConfig.PreSendEnabled
	}

	// Run the pipeline (no collection scripts for now)
	result, err := a.scriptEngine.ProcessMessage(
		req.Message,
		tempCaseID,
		req.CollectionName,
		"", "",    // no collection setup/presend
		false, false,
		caseSetup, casePreSend,
		caseSetupEnabled, casePreSendEnabled,
	)

	// Stop capture and get logs
	logs := a.scriptEngine.StopCapture()

	// Clean up temporary variables
	a.scriptEngine.ClearCaseVariables(tempCaseID)
	a.scriptEngine.ResetCounter(tempCaseID)

	// Build response
	resp := ScriptDryRunResponse{
		Result: result,
		Logs:   logs,
	}

	if err != nil {
		if err == script.ErrScriptCancelled {
			resp.Error = "Script cancelled (returned null)"
			resp.Result = ""
		} else {
			resp.Error = err.Error()
		}
	}

	return resp
}

// PostRecvDryRunRequest contains parameters for post-recv script dry run
type PostRecvDryRunRequest struct {
	Message             string               `json:"message"`
	CaseID              string               `json:"caseId"`
	CollectionName      string               `json:"collectionName"`
	CaseScriptConfig    *config.ScriptConfig `json:"caseScriptConfig"`
	CollectionVariables config.Variables     `json:"collectionVariables"`
	CaseVariables       config.Variables     `json:"caseVariables"`
}

// DryRunPostRecvScript runs a post-recv script without actual reception
func (a *App) DryRunPostRecvScript(req PostRecvDryRunRequest) ScriptDryRunResponse {
	if a.scriptEngine == nil {
		return ScriptDryRunResponse{Result: req.Message, Logs: []string{}}
	}

	// Use temporary case ID for isolation
	tempCaseID := "dryrun-postrecv-" + req.CaseID

	// Load collection variables from storage (source of truth)
	var collectionVars config.Variables
	if a.storage != nil {
		collections, err := a.storage.LoadCollections()
		if err == nil {
			for _, c := range collections {
				if c.Collection.Name == req.CollectionName {
					collectionVars = c.Collection.SharedVariables
					break
				}
			}
		}
	}

	// Load variables into engine
	a.scriptEngine.LoadVariables(
		tempCaseID,
		req.CollectionName,
		req.CaseVariables,
		collectionVars,
		nil,
	)

	// Start log capture
	a.scriptEngine.StartCapture()

	// Run post-recv script if enabled
	var err error
	if req.CaseScriptConfig != nil && req.CaseScriptConfig.PostRecvEnabled && req.CaseScriptConfig.PostRecvScript != "" {
		err = a.scriptEngine.RunPostRecvScript(
			req.CaseScriptConfig.PostRecvScript,
			req.Message,
			tempCaseID,
			req.CollectionName,
		)
	}

	// Stop capture and get logs
	logs := a.scriptEngine.StopCapture()

	// Clean up temporary variables
	a.scriptEngine.ClearCaseVariables(tempCaseID)

	// Build response
	resp := ScriptDryRunResponse{
		Result: req.Message, // Post-recv doesn't modify message
		Logs:   logs,
	}

	if err != nil {
		resp.Error = err.Error()
	}

	return resp
}

// SetScriptVariable sets a variable in the script engine
func (a *App) SetScriptVariable(scope string, scopeID string, key string, value interface{}) {
	if a.scriptEngine == nil {
		return
	}

	switch scope {
	case "global":
		a.scriptEngine.SetGlobalVar(key, value)
	case "collection":
		a.scriptEngine.SetCollectionVar(scopeID, key, value)
	case "case":
		a.scriptEngine.SetCaseVar(scopeID, key, value)
	}
}

// GetScriptVariable gets a variable from the script engine
func (a *App) GetScriptVariable(scope string, scopeID string, key string) interface{} {
	if a.scriptEngine == nil {
		return nil
	}

	switch scope {
	case "global":
		return a.scriptEngine.GetGlobalVar(key)
	case "collection":
		return a.scriptEngine.GetCollectionVar(scopeID, key)
	case "case":
		return a.scriptEngine.GetCaseVar(scopeID, key)
	}
	return nil
}

// ResetScriptCounter resets the script counter for a case
func (a *App) ResetScriptCounter(caseID string) {
	if a.scriptEngine != nil {
		a.scriptEngine.ResetCounter(caseID)
	}
}

// ResetCounter resets the counter for a specific session
func (a *App) ResetCounter(sessionID string) {
	preprocess.NewPreprocessor().ResetCounter(sessionID)
}

// IsConnected checks if a case is connected
func (a *App) IsConnected(caseID string) bool {
	return a.manager.IsConnected(caseID)
}

// LoadCollections loads all collections with their cases
func (a *App) LoadCollections() ([]config.CollectionWithCases, error) {
	if a.storage == nil {
		return []config.CollectionWithCases{}, nil
	}
	return a.storage.LoadCollections()
}

// CreateCollection creates a new collection
func (a *App) CreateCollection(name string) error {
	if a.storage == nil {
		return nil
	}
	return a.storage.CreateCollection(name)
}

// UpdateCollection updates a collection's metadata
func (a *App) UpdateCollection(name string, col config.Collection) error {
	if a.storage == nil {
		return nil
	}

	// Sync engine memory with new variables
	a.scriptEngine.SyncCollectionVariables(name, col.SharedVariables)

	return a.storage.UpdateCollection(name, col)
}

// DeleteCollection deletes a collection and all its cases
func (a *App) DeleteCollection(name string) error {
	if a.storage == nil {
		return nil
	}
	return a.storage.DeleteCollection(name)
}

// RenameCollection renames a collection
func (a *App) RenameCollection(oldName, newName string) error {
	if a.storage == nil {
		return nil
	}
	return a.storage.RenameCollection(oldName, newName)
}

// SaveCase saves a case to a collection
func (a *App) SaveCase(collectionName string, c config.SavedCase) error {
	if a.storage == nil {
		return nil
	}
	return a.storage.SaveCase(collectionName, c)
}

// LoadCase loads a single case from a collection
func (a *App) LoadCase(collectionName string, caseID string) (config.SavedCase, error) {
	if a.storage == nil {
		return config.SavedCase{}, nil
	}
	return a.storage.LoadCase(collectionName, caseID)
}

// DeleteCase deletes a case from a collection
func (a *App) DeleteCase(collectionName string, caseID string) error {
	if a.storage == nil {
		return nil
	}
	return a.storage.DeleteCase(collectionName, caseID)
}

// MoveCase moves a case from one collection to another
func (a *App) MoveCase(fromCollection, toCollection, caseID string) error {
	if a.storage == nil {
		return nil
	}
	return a.storage.MoveCase(fromCollection, toCollection, caseID)
}

// ReorderCollections updates the display order of collections
func (a *App) ReorderCollections(orders []config.CollectionOrder) error {
	if a.storage == nil {
		return nil
	}
	return a.storage.ReorderCollections(orders)
}

// ReorderCases updates the display order of cases in a collection
func (a *App) ReorderCases(collectionName string, orders []config.CaseOrder) error {
	if a.storage == nil {
		return nil
	}
	return a.storage.ReorderCases(collectionName, orders)
}

// LoadSettings loads global application settings
func (a *App) LoadSettings() (config.AppSettings, error) {
	if a.storage == nil {
		return config.AppSettings{DefaultCharset: "utf-8"}, nil
	}
	return a.storage.LoadSettings()
}

// SaveSettings saves global application settings
func (a *App) SaveSettings(settings config.AppSettings) error {
	if a.storage == nil {
		return nil
	}
	return a.storage.SaveSettings(settings)
}

// EncodeToHex encodes text to hex string using specified charset
func (a *App) EncodeToHex(text string, charsetName string) (string, error) {
	if charsetName == "" {
		charsetName = "utf-8"
	}
	bytes, err := charset.Encode(text, charsetName)
	if err != nil {
		return "", err
	}
	// Convert bytes to hex string with spaces
	var result string
	for i, b := range bytes {
		if i > 0 {
			result += " "
		}
		result += fmt.Sprintf("%02X", b)
	}
	return result, nil
}

// DecodeFromHex decodes hex string to text using specified charset
func (a *App) DecodeFromHex(hexStr string, charsetName string) (string, error) {
	if charsetName == "" {
		charsetName = "utf-8"
	}
	// Remove spaces and convert hex to bytes
	cleanHex := strings.ReplaceAll(hexStr, " ", "")
	cleanHex = strings.ReplaceAll(cleanHex, "\n", "")
	cleanHex = strings.ReplaceAll(cleanHex, "\r", "")
	cleanHex = strings.ReplaceAll(cleanHex, "\t", "")

	bytes, err := hex.DecodeString(cleanHex)
	if err != nil {
		return hexStr, err // Return original on error
	}
	return charset.Decode(bytes, charsetName)
}

// ========== Echo Server APIs ==========

// StartEchoServer starts the built-in echo server
func (a *App) StartEchoServer(port int, protocol string) error {
	if a.echoServer == nil {
		return fmt.Errorf("echo server not initialized")
	}
	err := a.echoServer.Start(port, protocol)
	if err != nil {
		return err
	}
	// Emit status change
	runtime.EventsEmit(a.ctx, "echo:status", a.echoServer.GetStatus())
	return nil
}

// StopEchoServer stops the built-in echo server
func (a *App) StopEchoServer() error {
	if a.echoServer == nil {
		return nil
	}
	err := a.echoServer.Stop()
	if err != nil {
		return err
	}
	// Emit status change
	runtime.EventsEmit(a.ctx, "echo:status", a.echoServer.GetStatus())
	return nil
}

// GetEchoStatus returns the current echo server status
func (a *App) GetEchoStatus() echo.Status {
	if a.echoServer == nil {
		return echo.Status{Running: false}
	}
	return a.echoServer.GetStatus()
}

// GetEchoLogs returns all echo server logs
func (a *App) GetEchoLogs() []echo.LogEntry {
	if a.echoServer == nil {
		return []echo.LogEntry{}
	}
	return a.echoServer.GetLogs()
}

// ClearEchoLogs clears all echo server logs
func (a *App) ClearEchoLogs() {
	if a.echoServer != nil {
		a.echoServer.ClearLogs()
	}
}
