package socket

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"github.com/yunjywork/wirey/internal/charset"
	"github.com/yunjywork/wirey/internal/framing"
	"github.com/yunjywork/wirey/internal/models"
)

// Connection interface for TCP and UDP connections
type Connection interface {
	Connect() error
	Disconnect() error
	Send(data string, format models.MessageFormat) error
	IsConnected() bool
	GetSessionID() string
	SetFraming(cfg framing.Config)
	SetCharset(charsetName string)
	GetCharset() string
}

// BaseConnection holds common connection fields
type BaseConnection struct {
	SessionID      string
	Protocol       models.Protocol
	Host           string
	Port           int
	conn           net.Conn
	mu             sync.RWMutex
	onData         func(sessionID string, rawBytes []byte, decodedContent string, meta models.FramingMeta, localAddr, remoteAddr string, timestamp int64)
	onSent         func(sessionID string, rawBytes []byte, originalContent string, meta models.FramingMeta, localAddr, remoteAddr string, timestamp int64)
	onError        func(sessionID, err string)
	onStatus       func(sessionID string, info models.ConnectionStatusInfo)
	stopChan       chan struct{}
	framingCfg     framing.Config
	framer         *framing.Framer
	charsetName    string // Character encoding for text conversion
	connectTimeout time.Duration
	readTimeout    time.Duration
	// Connection stats
	connectedAt time.Time
	bytesSent   int64
	bytesRecv   int64
	// Response timeout tracking: only apply timeout after send, clear after receive
	waitingForResponse int32 // atomic: 1 = waiting, 0 = idle
}

// NewConnection creates a new connection based on protocol
func NewConnection(
	sessionID string,
	protocol models.Protocol,
	host string,
	port int,
	onData func(sessionID string, rawBytes []byte, decodedContent string, meta models.FramingMeta, localAddr, remoteAddr string, timestamp int64),
	onSent func(sessionID string, rawBytes []byte, originalContent string, meta models.FramingMeta, localAddr, remoteAddr string, timestamp int64),
	onError func(sessionID, err string),
	onStatus func(sessionID string, info models.ConnectionStatusInfo),
	framingCfg framing.Config,
	charsetName string,
	connectTimeoutMs int,
	readTimeoutMs int,
) Connection {
	if charsetName == "" {
		charsetName = "utf-8"
	}

	// Convert ms to Duration, use defaults if 0
	connectTimeout := time.Duration(connectTimeoutMs) * time.Millisecond
	if connectTimeoutMs <= 0 {
		connectTimeout = 10 * time.Second // Default 10s
	}
	readTimeout := time.Duration(readTimeoutMs) * time.Millisecond
	// readTimeout 0 means no timeout (unlimited)

	base := &BaseConnection{
		SessionID:      sessionID,
		Protocol:       protocol,
		Host:           host,
		Port:           port,
		onData:         onData,
		onSent:         onSent,
		onError:        onError,
		onStatus:       onStatus,
		stopChan:       make(chan struct{}),
		framingCfg:     framingCfg,
		framer:         framing.NewFramer(framingCfg),
		charsetName:    charsetName,
		connectTimeout: connectTimeout,
		readTimeout:    readTimeout,
	}

	switch protocol {
	case models.ProtocolTCP:
		return &TCPConnection{BaseConnection: base}
	case models.ProtocolUDP:
		return &UDPConnection{BaseConnection: base}
	default:
		return &TCPConnection{BaseConnection: base}
	}
}

// SetFraming updates framing configuration
func (c *BaseConnection) SetFraming(cfg framing.Config) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.framingCfg = cfg
	c.framer = framing.NewFramer(cfg)
}

// SetCharset updates the charset for encoding/decoding
func (c *BaseConnection) SetCharset(charsetName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if charsetName == "" {
		charsetName = "utf-8"
	}
	c.charsetName = charsetName
}

// GetCharset returns the current charset
func (c *BaseConnection) GetCharset() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.charsetName
}

// GetSessionID returns the session ID
func (c *BaseConnection) GetSessionID() string {
	return c.SessionID
}

// IsConnected checks if connected
func (c *BaseConnection) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil
}

// encodeMessage converts message based on format and charset
func encodeMessage(data string, format models.MessageFormat, charsetName string) ([]byte, error) {
	switch format {
	case models.FormatHex:
		// Remove spaces and newlines from hex string
		cleanHex := strings.ReplaceAll(data, " ", "")
		cleanHex = strings.ReplaceAll(cleanHex, "\n", "")
		cleanHex = strings.ReplaceAll(cleanHex, "\r", "")
		cleanHex = strings.ReplaceAll(cleanHex, "\t", "")
		return hex.DecodeString(cleanHex)
	default:
		// Text format: encode with specified charset
		return charset.Encode(data, charsetName)
	}
}

// Address returns the remote connection address
func (c *BaseConnection) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// LocalAddress returns the local connection address
func (c *BaseConnection) LocalAddress() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.conn != nil {
		return c.conn.LocalAddr().String()
	}
	return ""
}

// emitConnected sends connected status with full info
// Must be called with lock held, localAddr should be passed from caller
func (c *BaseConnection) emitConnected(localAddr string) {
	c.connectedAt = time.Now()
	c.bytesSent = 0
	c.bytesRecv = 0

	if c.onStatus != nil {
		c.onStatus(c.SessionID, models.ConnectionStatusInfo{
			Status:     "connected",
			LocalAddr:  localAddr,
			RemoteAddr: c.Address(),
			Protocol:   string(c.Protocol),
		})
	}
}

// emitDisconnected sends disconnected status with stats
func (c *BaseConnection) emitDisconnected(reason models.DisconnectReason) {
	if c.onStatus != nil {
		duration := int64(0)
		if !c.connectedAt.IsZero() {
			duration = time.Since(c.connectedAt).Milliseconds()
		}
		c.onStatus(c.SessionID, models.ConnectionStatusInfo{
			Status:    "disconnected",
			Reason:    reason,
			Duration:  duration,
			BytesSent: atomic.LoadInt64(&c.bytesSent),
			BytesRecv: atomic.LoadInt64(&c.bytesRecv),
		})
	}
}

// addBytesSent tracks sent bytes (atomic, no lock needed)
func (c *BaseConnection) addBytesSent(n int) {
	atomic.AddInt64(&c.bytesSent, int64(n))
}

// addBytesRecv tracks received bytes (atomic, no lock needed)
func (c *BaseConnection) addBytesRecv(n int) {
	atomic.AddInt64(&c.bytesRecv, int64(n))
}

// setWaitingForResponse marks that we're waiting for a response (timeout applies)
func (c *BaseConnection) setWaitingForResponse() {
	atomic.StoreInt32(&c.waitingForResponse, 1)
}

// clearWaitingForResponse marks that we received a response (no timeout)
func (c *BaseConnection) clearWaitingForResponse() {
	atomic.StoreInt32(&c.waitingForResponse, 0)
}

// isWaitingForResponse checks if we're waiting for a response
func (c *BaseConnection) isWaitingForResponse() bool {
	return atomic.LoadInt32(&c.waitingForResponse) == 1
}
