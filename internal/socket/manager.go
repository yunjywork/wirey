package socket

import (
	"errors"
	"sync"
	"github.com/yunjywork/wirey/internal/framing"
	"github.com/yunjywork/wirey/internal/models"
)

// Manager manages multiple socket connections
type Manager struct {
	connections map[string]Connection
	mu          sync.RWMutex
	onData      func(sessionID string, rawBytes []byte, decodedContent string, meta models.FramingMeta, localAddr, remoteAddr string, timestamp int64)
	onSent      func(sessionID string, rawBytes []byte, originalContent string, meta models.FramingMeta, localAddr, remoteAddr string, timestamp int64)
	onError     func(sessionID, err string)
	onStatus    func(sessionID string, info models.ConnectionStatusInfo)
}

// NewManager creates a new connection manager
func NewManager(
	onData func(sessionID string, rawBytes []byte, decodedContent string, meta models.FramingMeta, localAddr, remoteAddr string, timestamp int64),
	onSent func(sessionID string, rawBytes []byte, originalContent string, meta models.FramingMeta, localAddr, remoteAddr string, timestamp int64),
	onError func(sessionID, err string),
	onStatus func(sessionID string, info models.ConnectionStatusInfo),
) *Manager {
	return &Manager{
		connections: make(map[string]Connection),
		onData:      onData,
		onSent:      onSent,
		onError:     onError,
		onStatus:    onStatus,
	}
}

// Connect creates and connects a new session
func (m *Manager) Connect(sessionID string, protocol models.Protocol, host string, port int, framingCfg framing.Config, charsetName string, connectTimeoutMs, readTimeoutMs int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already exists
	if existing, ok := m.connections[sessionID]; ok {
		if existing.IsConnected() {
			return nil // Already connected
		}
		// Remove old connection
		existing.Disconnect()
		delete(m.connections, sessionID)
	}

	// Create new connection
	conn := NewConnection(sessionID, protocol, host, port, m.onData, m.onSent, m.onError, m.onStatus, framingCfg, charsetName, connectTimeoutMs, readTimeoutMs)
	if err := conn.Connect(); err != nil {
		return err
	}

	m.connections[sessionID] = conn
	return nil
}

// UpdateFraming updates framing configuration for a session
func (m *Manager) UpdateFraming(sessionID string, cfg framing.Config) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok {
		return errors.New("session not found")
	}

	conn.SetFraming(cfg)
	return nil
}

// UpdateCharset updates charset for a session
func (m *Manager) UpdateCharset(sessionID string, charsetName string) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok {
		return errors.New("session not found")
	}

	conn.SetCharset(charsetName)
	return nil
}

// Disconnect closes a specific session
func (m *Manager) Disconnect(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, ok := m.connections[sessionID]
	if !ok {
		return errors.New("session not found")
	}

	err := conn.Disconnect()
	delete(m.connections, sessionID)
	return err
}

// Send sends data to a specific session
func (m *Manager) Send(sessionID string, data string, format models.MessageFormat) error {
	m.mu.RLock()
	conn, ok := m.connections[sessionID]
	m.mu.RUnlock()

	if !ok {
		return errors.New("session not found")
	}

	if !conn.IsConnected() {
		return errors.New("not connected")
	}

	return conn.Send(data, format)
}

// DisconnectAll closes all connections
func (m *Manager) DisconnectAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, conn := range m.connections {
		conn.Disconnect()
		delete(m.connections, id)
	}
}

// GetConnection returns a connection by session ID
func (m *Manager) GetConnection(sessionID string) (Connection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.connections[sessionID]
	return conn, ok
}

// IsConnected checks if a session is connected
func (m *Manager) IsConnected(sessionID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn, ok := m.connections[sessionID]
	if !ok {
		return false
	}
	return conn.IsConnected()
}
