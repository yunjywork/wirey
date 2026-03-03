package models

// Protocol type for socket connections
type Protocol string

const (
	ProtocolTCP Protocol = "tcp"
	ProtocolUDP Protocol = "udp"
)

// MessageFormat for encoding/decoding
type MessageFormat string

const (
	FormatText MessageFormat = "text"
	FormatHex  MessageFormat = "hex"
)

// SocketEvent represents events emitted to frontend
type SocketEvent struct {
	SessionID string `json:"sessionId"`
	Type      string `json:"type"`
	Content   string `json:"content,omitempty"`
	Error     string `json:"error,omitempty"`
}

// ConnectionInfo holds connection details
type ConnectionInfo struct {
	SessionID string   `json:"sessionId"`
	Protocol  Protocol `json:"protocol"`
	Host      string   `json:"host"`
	Port      int      `json:"port"`
}

// FramingMeta holds framing metadata for a received message
type FramingMeta struct {
	Mode        string `json:"mode"`
	FrameHeader string `json:"frameHeader,omitempty"` // hex representation
	FrameFooter string `json:"frameFooter,omitempty"` // hex representation
	PayloadSize int    `json:"payloadSize"`
	TotalSize   int    `json:"totalSize"`
	Settings    string `json:"settings,omitempty"` // e.g., "4 bytes, Big Endian"
}

// MessageData holds complete message data for frontend
type MessageData struct {
	CaseID      string      `json:"caseId"`
	Content     string      `json:"content"`    // ASCII representation
	RawBytes    string      `json:"rawBytes"`   // base64 encoded
	Size        int         `json:"size"`       // byte count
	RemoteAddr  string      `json:"remoteAddr"` // "host:port"
	FramingInfo FramingMeta `json:"framingInfo"`
}

// DisconnectReason represents why connection was closed
type DisconnectReason string

const (
	ReasonUser   DisconnectReason = "user"   // User initiated disconnect
	ReasonServer DisconnectReason = "server" // Server closed connection
	ReasonError  DisconnectReason = "error"  // Connection error
)

// ConnectionStatusInfo holds detailed connection status information
type ConnectionStatusInfo struct {
	Status       string           `json:"status"`                 // "connected" or "disconnected"
	LocalAddr    string           `json:"localAddr,omitempty"`    // Local IP:Port
	RemoteAddr   string           `json:"remoteAddr,omitempty"`   // Remote IP:Port
	Protocol     string           `json:"protocol,omitempty"`     // "tcp" or "udp"
	Reason       DisconnectReason `json:"reason,omitempty"`       // Disconnect reason
	Duration     int64            `json:"duration,omitempty"`     // Connection duration in milliseconds
	BytesSent    int64            `json:"bytesSent,omitempty"`    // Total bytes sent
	BytesRecv    int64            `json:"bytesRecv,omitempty"`    // Total bytes received
}
