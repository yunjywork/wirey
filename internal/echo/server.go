package echo

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	DefaultPort = 39876
	MaxLogSize  = 1000 // Maximum number of log entries to keep
)

// LogEntry represents a single echo log entry
type LogEntry struct {
	ID         string `json:"id"`
	Timestamp  int64  `json:"timestamp"`
	Direction  string `json:"direction"` // "recv" or "echo"
	RemoteAddr string `json:"remoteAddr"`
	Data       string `json:"data"` // hex encoded
	Size       int    `json:"size"`
}

// Status represents the echo server status
type Status struct {
	Running  bool   `json:"running"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
}

// Server is an echo server that echoes back received data
type Server struct {
	mu           sync.RWMutex
	port         int
	protocol     string
	running      bool
	logs         []LogEntry
	logCounter   int64
	tcpListener  net.Listener
	udpConn      *net.UDPConn
	ctx          context.Context
	cancel       context.CancelFunc
	onLog        func(entry LogEntry) // callback for new log entries
	tcpConns     map[net.Conn]struct{}
	tcpConnMu    sync.Mutex
}

// NewServer creates a new echo server
func NewServer(onLog func(entry LogEntry)) *Server {
	return &Server{
		logs:     make([]LogEntry, 0),
		onLog:    onLog,
		tcpConns: make(map[net.Conn]struct{}),
	}
}

// Start starts the echo server
func (s *Server) Start(port int, protocol string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server already running")
	}

	protocol = strings.ToLower(protocol)
	if protocol != "tcp" && protocol != "udp" {
		return fmt.Errorf("invalid protocol: %s (must be tcp or udp)", protocol)
	}

	s.port = port
	s.protocol = protocol
	s.ctx, s.cancel = context.WithCancel(context.Background())

	var err error
	if protocol == "tcp" {
		err = s.startTCP()
	} else {
		err = s.startUDP()
	}

	if err != nil {
		s.cancel()
		return err
	}

	s.running = true
	return nil
}

// startTCP starts a TCP echo server
func (s *Server) startTCP() error {
	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start TCP listener: %w", err)
	}
	s.tcpListener = listener

	go s.acceptTCPConnections()
	return nil
}

// acceptTCPConnections accepts incoming TCP connections
func (s *Server) acceptTCPConnections() {
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		// Set accept deadline to allow checking context
		if tcpListener, ok := s.tcpListener.(*net.TCPListener); ok {
			tcpListener.SetDeadline(time.Now().Add(1 * time.Second))
		}

		conn, err := s.tcpListener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			select {
			case <-s.ctx.Done():
				return
			default:
				continue
			}
		}

		// Track the connection
		s.tcpConnMu.Lock()
		s.tcpConns[conn] = struct{}{}
		s.tcpConnMu.Unlock()

		go s.handleTCPConnection(conn)
	}
}

// handleTCPConnection handles a single TCP connection
func (s *Server) handleTCPConnection(conn net.Conn) {
	defer func() {
		s.tcpConnMu.Lock()
		delete(s.tcpConns, conn)
		s.tcpConnMu.Unlock()
		conn.Close()
	}()

	buf := make([]byte, 65536)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		// Set read deadline
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return
		}

		data := buf[:n]
		remoteAddr := conn.RemoteAddr().String()

		// Log received data
		s.addLog("recv", remoteAddr, data)

		// Echo back
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_, err = conn.Write(data)
		if err != nil {
			return
		}

		// Log echoed data
		s.addLog("echo", remoteAddr, data)
	}
}

// startUDP starts a UDP echo server
func (s *Server) startUDP() error {
	addr := fmt.Sprintf(":%d", s.port)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to start UDP listener: %w", err)
	}
	s.udpConn = conn

	go s.handleUDP()
	return nil
}

// handleUDP handles UDP echo
func (s *Server) handleUDP() {
	buf := make([]byte, 65536)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		// Set read deadline
		s.udpConn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, remoteAddr, err := s.udpConn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			select {
			case <-s.ctx.Done():
				return
			default:
				continue
			}
		}

		data := buf[:n]
		addrStr := remoteAddr.String()

		// Log received data
		s.addLog("recv", addrStr, data)

		// Echo back
		s.udpConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_, err = s.udpConn.WriteToUDP(data, remoteAddr)
		if err != nil {
			continue
		}

		// Log echoed data
		s.addLog("echo", addrStr, data)
	}
}

// addLog adds a log entry
func (s *Server) addLog(direction, remoteAddr string, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logCounter++
	entry := LogEntry{
		ID:         fmt.Sprintf("echo-%d", s.logCounter),
		Timestamp:  time.Now().UnixMilli(),
		Direction:  direction,
		RemoteAddr: remoteAddr,
		Data:       strings.ToUpper(hex.EncodeToString(data)),
		Size:       len(data),
	}

	s.logs = append(s.logs, entry)

	// Trim if too many logs
	if len(s.logs) > MaxLogSize {
		s.logs = s.logs[len(s.logs)-MaxLogSize:]
	}

	// Notify callback
	if s.onLog != nil {
		go s.onLog(entry)
	}
}

// Stop stops the echo server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	// Cancel context to signal goroutines to stop
	if s.cancel != nil {
		s.cancel()
	}

	// Close TCP listener
	if s.tcpListener != nil {
		s.tcpListener.Close()
		s.tcpListener = nil
	}

	// Close all TCP connections
	s.tcpConnMu.Lock()
	for conn := range s.tcpConns {
		conn.Close()
	}
	s.tcpConns = make(map[net.Conn]struct{})
	s.tcpConnMu.Unlock()

	// Close UDP connection
	if s.udpConn != nil {
		s.udpConn.Close()
		s.udpConn = nil
	}

	s.running = false
	return nil
}

// GetStatus returns the current server status
func (s *Server) GetStatus() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	address := ""
	if s.running {
		address = fmt.Sprintf(":%d", s.port)
	}

	return Status{
		Running:  s.running,
		Port:     s.port,
		Protocol: s.protocol,
		Address:  address,
	}
}

// GetLogs returns all log entries
func (s *Server) GetLogs() []LogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy
	logs := make([]LogEntry, len(s.logs))
	copy(logs, s.logs)
	return logs
}

// ClearLogs clears all log entries
func (s *Server) ClearLogs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logs = make([]LogEntry, 0)
	s.logCounter = 0
}

// IsRunning returns whether the server is running
func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}
