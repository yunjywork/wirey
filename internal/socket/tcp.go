package socket

import (
	"bufio"
	"net"
	"time"
	"github.com/yunjywork/wirey/internal/charset"
	"github.com/yunjywork/wirey/internal/framing"
	"github.com/yunjywork/wirey/internal/models"
)

// TCPConnection handles TCP socket connections
type TCPConnection struct {
	*BaseConnection
}

// Connect establishes a TCP connection
func (c *TCPConnection) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return nil // Already connected
	}

	// Use DialTimeout with configured connect timeout
	conn, err := net.DialTimeout("tcp", c.Address(), c.connectTimeout)
	if err != nil {
		if c.onError != nil {
			c.onError(c.SessionID, err.Error())
		}
		return err
	}

	c.conn = conn
	c.stopChan = make(chan struct{})

	// Emit connected status with full info (get local addr while holding lock)
	localAddr := conn.LocalAddr().String()
	c.emitConnected(localAddr)

	// Start reading data
	go c.readLoop()

	return nil
}

// Disconnect closes the TCP connection
func (c *TCPConnection) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	// Signal stop
	close(c.stopChan)

	err := c.conn.Close()
	c.conn = nil

	// Emit disconnected status with stats (user initiated)
	c.emitDisconnected(models.ReasonUser)

	return err
}

// Send sends data over TCP connection
func (c *TCPConnection) Send(data string, format models.MessageFormat) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return net.ErrClosed
	}

	bytes, err := encodeMessage(data, format, c.charsetName)
	if err != nil {
		return err
	}

	// Apply framing
	framedData, err := framing.Frame(bytes, c.framingCfg)
	if err != nil {
		return err
	}

	n, err := c.conn.Write(framedData)
	if err != nil {
		return err
	}

	// Track bytes sent
	c.addBytesSent(n)

	// Mark that we're waiting for a response (enables read timeout)
	c.setWaitingForResponse()

	// Notify sent data
	if c.onSent != nil {
		cfg := c.framingCfg
		meta := models.FramingMeta{
			Mode:        cfg.Mode,
			PayloadSize: len(bytes),
			TotalSize:   len(framedData),
			Settings:    cfg.SettingsDescription(),
		}
		// Calculate frame header/footer based on mode
		if len(framedData) > len(bytes) {
			extraSize := len(framedData) - len(bytes)
			switch cfg.Mode {
			case "delimiter":
				// Delimiter is appended at the end (footer)
				meta.FrameFooter = framing.BytesToHex(framedData[len(bytes):])
			case "length-prefix":
				// Length prefix is at the beginning (header)
				meta.FrameHeader = framing.BytesToHex(framedData[:extraSize])
			}
		}
		// Decode payload bytes to get display content
		decodedContent, _ := charset.Decode(bytes, c.charsetName)
		c.onSent(c.SessionID, framedData, decodedContent, meta, c.LocalAddress(), c.Address(), time.Now().UnixMilli())
	}

	return nil
}

// readLoop continuously reads from the connection
func (c *TCPConnection) readLoop() {
	reader := bufio.NewReader(c.conn)
	buffer := make([]byte, 4096)

	for {
		select {
		case <-c.stopChan:
			return
		default:
			// Apply read timeout only when waiting for response after send
			// When idle (not waiting), no timeout - connection stays open
			if c.readTimeout > 0 && c.isWaitingForResponse() {
				c.conn.SetReadDeadline(time.Now().Add(c.readTimeout))
			} else {
				// Clear any previous deadline
				c.conn.SetReadDeadline(time.Time{})
			}
			n, err := reader.Read(buffer)
			if err != nil {
				// Check if intentionally closed
				select {
				case <-c.stopChan:
					return
				default:
					if c.onError != nil {
						c.onError(c.SessionID, err.Error())
					}
					c.mu.Lock()
					if c.conn != nil {
						c.conn.Close()
						c.conn = nil
					}
					c.mu.Unlock()
					// Server closed connection or error
					c.emitDisconnected(models.ReasonServer)
					return
				}
			}

			if n > 0 {
				// Track bytes received
				c.addBytesRecv(n)

				// Clear waiting flag - response received, no more timeout
				c.clearWaitingForResponse()

				if c.onData != nil {
					// Record timestamp BEFORE delay
					timestamp := time.Now().UnixMilli()

					// Small delay to ensure sent event is processed before receive
					time.Sleep(5 * time.Millisecond)

					// Apply framing to parse messages with metadata
					messages := c.framer.FeedWithMeta(buffer[:n])
					cfg := c.framer.GetConfig()
					localAddr := c.LocalAddress()
					remoteAddr := c.Address()
					charsetName := c.GetCharset()

					for _, msg := range messages {
						meta := models.FramingMeta{
							Mode:        cfg.Mode,
							PayloadSize: len(msg.Payload),
							TotalSize:   len(msg.RawFrame),
							Settings:    cfg.SettingsDescription(),
						}
						if len(msg.FrameHeader) > 0 {
							meta.FrameHeader = framing.BytesToHex(msg.FrameHeader)
						}
						if len(msg.FrameFooter) > 0 {
							meta.FrameFooter = framing.BytesToHex(msg.FrameFooter)
						}
						// Decode payload only (without frame header/footer) for display
						// This keeps consistency with SEND which shows payload content
						decodedContent, _ := charset.Decode(msg.Payload, charsetName)
						c.onData(c.SessionID, msg.RawFrame, decodedContent, meta, localAddr, remoteAddr, timestamp)
					}
				}
			}
		}
	}
}
