package framing

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// Config represents framing configuration
type Config struct {
	Mode           string // none, delimiter, length-prefix, fixed-length
	Delimiter      string // \n, \r\n, \0, custom
	LengthEncoding string // binary, ascii, hex, bcd
	LengthOffset   int
	LengthBytes    int    // 1, 2, 4
	Endian         string // big, little
	IncludeHeader  bool
	LengthMode     string // append, rewrite
	FixedSize      int
}

// DefaultConfig returns default framing config (no framing)
func DefaultConfig() Config {
	return Config{
		Mode:           "none",
		Delimiter:      "\\n",
		LengthEncoding: "binary",
		LengthBytes:    4,
		Endian:         "big",
		LengthMode:     "append",
		FixedSize:      1024,
	}
}

// Frame applies framing to outgoing data
func Frame(data []byte, cfg Config) ([]byte, error) {
	switch cfg.Mode {
	case "delimiter":
		return frameDelimiter(data, cfg)
	case "length-prefix":
		return frameLengthPrefix(data, cfg)
	case "fixed-length":
		return frameFixedLength(data, cfg)
	default:
		return data, nil
	}
}

// frameDelimiter adds delimiter to the end of data
func frameDelimiter(data []byte, cfg Config) ([]byte, error) {
	delimiter := parseDelimiter(cfg.Delimiter)
	return append(data, delimiter...), nil
}

// frameLengthPrefix adds length prefix to data
func frameLengthPrefix(data []byte, cfg Config) ([]byte, error) {
	if cfg.LengthMode == "rewrite" {
		// Rewrite mode: write length at offset position
		offset := cfg.LengthOffset
		endPos := offset + cfg.LengthBytes

		if len(data) < endPos {
			return nil, fmt.Errorf("data too short: need at least %d bytes, got %d", endPos, len(data))
		}

		// Calculate length (payload size after the length field)
		dataLen := len(data) - endPos
		if cfg.IncludeHeader {
			// Include everything from start
			dataLen = len(data)
		}

		// Encode length
		var lengthBytes []byte
		var err error

		switch cfg.LengthEncoding {
		case "binary":
			lengthBytes, err = encodeBinaryLength(dataLen, cfg.LengthBytes, cfg.Endian)
		case "ascii":
			lengthBytes, err = encodeASCIILength(dataLen, cfg.LengthBytes)
		case "hex":
			lengthBytes, err = encodeHexLength(dataLen, cfg.LengthBytes)
		case "bcd":
			lengthBytes, err = encodeBCDLength(dataLen, cfg.LengthBytes)
		default:
			lengthBytes, err = encodeBinaryLength(dataLen, cfg.LengthBytes, cfg.Endian)
		}

		if err != nil {
			return nil, err
		}

		// Write length at offset position
		copy(data[offset:endPos], lengthBytes)
		return data, nil
	}

	// Append mode: prepend length to data (offset ignored)
	dataLen := len(data)
	if cfg.IncludeHeader {
		dataLen += cfg.LengthBytes
	}

	var lengthBytes []byte
	var err error

	switch cfg.LengthEncoding {
	case "binary":
		lengthBytes, err = encodeBinaryLength(dataLen, cfg.LengthBytes, cfg.Endian)
	case "ascii":
		lengthBytes, err = encodeASCIILength(dataLen, cfg.LengthBytes)
	case "hex":
		lengthBytes, err = encodeHexLength(dataLen, cfg.LengthBytes)
	case "bcd":
		lengthBytes, err = encodeBCDLength(dataLen, cfg.LengthBytes)
	default:
		lengthBytes, err = encodeBinaryLength(dataLen, cfg.LengthBytes, cfg.Endian)
	}

	if err != nil {
		return nil, err
	}

	return append(lengthBytes, data...), nil
}

// frameFixedLength pads or truncates data to fixed size
func frameFixedLength(data []byte, cfg Config) ([]byte, error) {
	if cfg.FixedSize <= 0 {
		return data, nil
	}

	result := make([]byte, cfg.FixedSize)
	if len(data) >= cfg.FixedSize {
		copy(result, data[:cfg.FixedSize])
	} else {
		copy(result, data)
		// Remaining bytes are already zero-filled
	}
	return result, nil
}

// parseDelimiter converts escaped delimiter string to bytes
func parseDelimiter(delim string) []byte {
	switch delim {
	case "\\n":
		return []byte{'\n'}
	case "\\r\\n":
		return []byte{'\r', '\n'}
	case "\\0":
		return []byte{0}
	default:
		// Handle other escape sequences or return as-is
		result := strings.ReplaceAll(delim, "\\n", "\n")
		result = strings.ReplaceAll(result, "\\r", "\r")
		result = strings.ReplaceAll(result, "\\0", "\x00")
		result = strings.ReplaceAll(result, "\\t", "\t")
		return []byte(result)
	}
}

// encodeBinaryLength encodes length as binary
func encodeBinaryLength(length int, numBytes int, endian string) ([]byte, error) {
	buf := make([]byte, numBytes)
	var order binary.ByteOrder
	if endian == "little" {
		order = binary.LittleEndian
	} else {
		order = binary.BigEndian
	}

	switch numBytes {
	case 1:
		if length > 255 {
			return nil, fmt.Errorf("length %d exceeds 1-byte limit", length)
		}
		buf[0] = byte(length)
	case 2:
		if length > 65535 {
			return nil, fmt.Errorf("length %d exceeds 2-byte limit", length)
		}
		order.PutUint16(buf, uint16(length))
	case 4:
		order.PutUint32(buf, uint32(length))
	case 8:
		order.PutUint64(buf, uint64(length))
	default:
		// For custom lengths, encode as big-endian bytes
		if numBytes < 1 || numBytes > 16 {
			return nil, fmt.Errorf("unsupported length bytes: %d (must be 1-16)", numBytes)
		}
		val := uint64(length)
		if endian == "little" {
			for i := 0; i < numBytes; i++ {
				buf[i] = byte(val & 0xFF)
				val >>= 8
			}
		} else {
			for i := numBytes - 1; i >= 0; i-- {
				buf[i] = byte(val & 0xFF)
				val >>= 8
			}
		}
	}

	return buf, nil
}

// encodeASCIILength encodes length as ASCII decimal string
func encodeASCIILength(length int, numBytes int) ([]byte, error) {
	format := fmt.Sprintf("%%0%dd", numBytes)
	str := fmt.Sprintf(format, length)
	if len(str) > numBytes {
		return nil, fmt.Errorf("length %d requires more than %d ASCII digits", length, numBytes)
	}
	return []byte(str), nil
}

// encodeHexLength encodes length as hex string
func encodeHexLength(length int, numBytes int) ([]byte, error) {
	format := fmt.Sprintf("%%0%dX", numBytes)
	str := fmt.Sprintf(format, length)
	if len(str) > numBytes {
		return nil, fmt.Errorf("length %d requires more than %d hex digits", length, numBytes)
	}
	return []byte(str), nil
}

// encodeBCDLength encodes length as BCD (Binary Coded Decimal)
func encodeBCDLength(length int, numBytes int) ([]byte, error) {
	// BCD packs two decimal digits per byte
	maxDigits := numBytes * 2
	str := strconv.Itoa(length)
	if len(str) > maxDigits {
		return nil, fmt.Errorf("length %d requires more than %d BCD digits", length, maxDigits)
	}

	// Pad with leading zeros
	str = fmt.Sprintf("%0*s", maxDigits, str)

	result := make([]byte, numBytes)
	for i := 0; i < numBytes; i++ {
		high := str[i*2] - '0'
		low := str[i*2+1] - '0'
		result[i] = (high << 4) | low
	}

	return result, nil
}

// ParsedMessage represents a parsed message with metadata
type ParsedMessage struct {
	Payload     []byte // The actual message payload
	RawFrame    []byte // Complete raw frame including header/footer
	FrameHeader []byte // Frame header bytes (e.g., length prefix)
	FrameFooter []byte // Frame footer bytes (e.g., delimiter)
}

// Framer handles receiving framed data
type Framer struct {
	cfg    Config
	buffer []byte
}

// NewFramer creates a new framer for receiving
func NewFramer(cfg Config) *Framer {
	return &Framer{
		cfg:    cfg,
		buffer: make([]byte, 0),
	}
}

// GetConfig returns the current framing configuration
func (f *Framer) GetConfig() Config {
	return f.cfg
}

// Feed adds data to the buffer and returns complete messages
func (f *Framer) Feed(data []byte) [][]byte {
	messages := f.FeedWithMeta(data)
	result := make([][]byte, len(messages))
	for i, msg := range messages {
		result[i] = msg.Payload
	}
	return result
}

// FeedWithMeta adds data to the buffer and returns complete messages with metadata
func (f *Framer) FeedWithMeta(data []byte) []ParsedMessage {
	f.buffer = append(f.buffer, data...)

	switch f.cfg.Mode {
	case "delimiter":
		return f.parseDelimiterWithMeta()
	case "length-prefix":
		return f.parseLengthPrefixWithMeta()
	case "fixed-length":
		return f.parseFixedLengthWithMeta()
	default:
		// No framing - return data as-is
		result := make([]ParsedMessage, 0)
		if len(f.buffer) > 0 {
			rawCopy := make([]byte, len(f.buffer))
			copy(rawCopy, f.buffer)
			result = append(result, ParsedMessage{
				Payload:  rawCopy,
				RawFrame: rawCopy,
			})
			f.buffer = make([]byte, 0)
		}
		return result
	}
}

// parseDelimiter extracts messages separated by delimiter
func (f *Framer) parseDelimiter() [][]byte {
	delimiter := parseDelimiter(f.cfg.Delimiter)
	var messages [][]byte

	for {
		idx := bytes.Index(f.buffer, delimiter)
		if idx == -1 {
			break
		}

		// Extract message (without delimiter)
		msg := make([]byte, idx)
		copy(msg, f.buffer[:idx])
		messages = append(messages, msg)

		// Remove message and delimiter from buffer
		f.buffer = f.buffer[idx+len(delimiter):]
	}

	return messages
}

// parseLengthPrefix extracts messages with length prefix
func (f *Framer) parseLengthPrefix() [][]byte {
	var messages [][]byte

	for {
		if len(f.buffer) < f.cfg.LengthBytes {
			break
		}

		// Read length
		length, err := f.decodeLength(f.buffer[:f.cfg.LengthBytes])
		if err != nil {
			break
		}

		// Calculate total message size
		totalSize := f.cfg.LengthBytes + length - f.cfg.LengthOffset
		if f.cfg.IncludeHeader {
			totalSize = length - f.cfg.LengthOffset
		}

		if len(f.buffer) < totalSize {
			break
		}

		// Extract data (excluding length prefix)
		dataStart := f.cfg.LengthBytes
		dataEnd := totalSize
		if f.cfg.IncludeHeader {
			dataEnd = length - f.cfg.LengthOffset
		}

		msg := make([]byte, dataEnd-dataStart)
		copy(msg, f.buffer[dataStart:dataEnd])
		messages = append(messages, msg)

		// Remove from buffer
		f.buffer = f.buffer[totalSize:]
	}

	return messages
}

// parseFixedLength extracts fixed-size messages
func (f *Framer) parseFixedLength() [][]byte {
	var messages [][]byte

	for len(f.buffer) >= f.cfg.FixedSize {
		msg := make([]byte, f.cfg.FixedSize)
		copy(msg, f.buffer[:f.cfg.FixedSize])
		messages = append(messages, msg)
		f.buffer = f.buffer[f.cfg.FixedSize:]
	}

	return messages
}

// decodeLength decodes length from bytes based on encoding
func (f *Framer) decodeLength(data []byte) (int, error) {
	switch f.cfg.LengthEncoding {
	case "binary":
		return f.decodeBinaryLength(data)
	case "ascii":
		return f.decodeASCIILength(data)
	case "hex":
		return f.decodeHexLength(data)
	case "bcd":
		return f.decodeBCDLength(data)
	default:
		return f.decodeBinaryLength(data)
	}
}

func (f *Framer) decodeBinaryLength(data []byte) (int, error) {
	var order binary.ByteOrder
	if f.cfg.Endian == "little" {
		order = binary.LittleEndian
	} else {
		order = binary.BigEndian
	}

	switch f.cfg.LengthBytes {
	case 1:
		return int(data[0]), nil
	case 2:
		return int(order.Uint16(data)), nil
	case 4:
		return int(order.Uint32(data)), nil
	case 8:
		return int(order.Uint64(data)), nil
	default:
		// For custom lengths, decode as big-endian bytes
		if f.cfg.LengthBytes < 1 || f.cfg.LengthBytes > 16 {
			return 0, fmt.Errorf("unsupported length bytes: %d (must be 1-16)", f.cfg.LengthBytes)
		}
		var val uint64
		if f.cfg.Endian == "little" {
			for i := f.cfg.LengthBytes - 1; i >= 0; i-- {
				val = (val << 8) | uint64(data[i])
			}
		} else {
			for i := 0; i < f.cfg.LengthBytes; i++ {
				val = (val << 8) | uint64(data[i])
			}
		}
		return int(val), nil
	}
}

func (f *Framer) decodeASCIILength(data []byte) (int, error) {
	str := strings.TrimLeft(string(data), "0")
	if str == "" {
		return 0, nil
	}
	return strconv.Atoi(str)
}

func (f *Framer) decodeHexLength(data []byte) (int, error) {
	val, err := strconv.ParseInt(string(data), 16, 64)
	return int(val), err
}

func (f *Framer) decodeBCDLength(data []byte) (int, error) {
	result := 0
	for _, b := range data {
		high := int(b >> 4)
		low := int(b & 0x0F)
		result = result*100 + high*10 + low
	}
	return result, nil
}

// Reset clears the buffer
func (f *Framer) Reset() {
	f.buffer = make([]byte, 0)
}

// parseDelimiterWithMeta extracts messages separated by delimiter with metadata
func (f *Framer) parseDelimiterWithMeta() []ParsedMessage {
	delimiter := parseDelimiter(f.cfg.Delimiter)
	var messages []ParsedMessage

	for {
		idx := bytes.Index(f.buffer, delimiter)
		if idx == -1 {
			break
		}

		// Extract payload (without delimiter)
		payload := make([]byte, idx)
		copy(payload, f.buffer[:idx])

		// Raw frame includes delimiter
		rawFrame := make([]byte, idx+len(delimiter))
		copy(rawFrame, f.buffer[:idx+len(delimiter)])

		// Footer is the delimiter
		footer := make([]byte, len(delimiter))
		copy(footer, delimiter)

		messages = append(messages, ParsedMessage{
			Payload:     payload,
			RawFrame:    rawFrame,
			FrameFooter: footer,
		})

		// Remove message and delimiter from buffer
		f.buffer = f.buffer[idx+len(delimiter):]
	}

	return messages
}

// parseLengthPrefixWithMeta extracts messages with length prefix with metadata
func (f *Framer) parseLengthPrefixWithMeta() []ParsedMessage {
	var messages []ParsedMessage

	for {
		if len(f.buffer) < f.cfg.LengthBytes {
			break
		}

		// Read length
		length, err := f.decodeLength(f.buffer[:f.cfg.LengthBytes])
		if err != nil {
			break
		}

		// Calculate total message size
		totalSize := f.cfg.LengthBytes + length - f.cfg.LengthOffset
		if f.cfg.IncludeHeader {
			totalSize = length - f.cfg.LengthOffset
		}

		if len(f.buffer) < totalSize {
			break
		}

		// Extract header (length prefix)
		header := make([]byte, f.cfg.LengthBytes)
		copy(header, f.buffer[:f.cfg.LengthBytes])

		// Extract data (excluding length prefix)
		dataStart := f.cfg.LengthBytes
		dataEnd := totalSize
		if f.cfg.IncludeHeader {
			dataEnd = length - f.cfg.LengthOffset
		}

		payload := make([]byte, dataEnd-dataStart)
		copy(payload, f.buffer[dataStart:dataEnd])

		// Raw frame is the complete frame
		rawFrame := make([]byte, totalSize)
		copy(rawFrame, f.buffer[:totalSize])

		messages = append(messages, ParsedMessage{
			Payload:     payload,
			RawFrame:    rawFrame,
			FrameHeader: header,
		})

		// Remove from buffer
		f.buffer = f.buffer[totalSize:]
	}

	return messages
}

// parseFixedLengthWithMeta extracts fixed-size messages with metadata
func (f *Framer) parseFixedLengthWithMeta() []ParsedMessage {
	var messages []ParsedMessage

	for len(f.buffer) >= f.cfg.FixedSize {
		payload := make([]byte, f.cfg.FixedSize)
		copy(payload, f.buffer[:f.cfg.FixedSize])

		messages = append(messages, ParsedMessage{
			Payload:  payload,
			RawFrame: payload, // Same as payload for fixed-length
		})
		f.buffer = f.buffer[f.cfg.FixedSize:]
	}

	return messages
}

// SettingsDescription returns a human-readable description of the framing config
func (c Config) SettingsDescription() string {
	switch c.Mode {
	case "delimiter":
		delim := c.Delimiter
		switch delim {
		case "\\n":
			delim = "LF"
		case "\\r\\n":
			delim = "CRLF"
		case "\\0":
			delim = "NULL"
		}
		return fmt.Sprintf("Delimiter: %s", delim)
	case "length-prefix":
		if c.LengthEncoding == "binary" {
			endian := "Big Endian"
			if c.Endian == "little" {
				endian = "Little Endian"
			}
			return fmt.Sprintf("%d bytes, %s, %s", c.LengthBytes, endian, c.LengthEncoding)
		}
		return fmt.Sprintf("%d bytes, %s", c.LengthBytes, c.LengthEncoding)
	case "fixed-length":
		return fmt.Sprintf("Fixed %d bytes", c.FixedSize)
	default:
		return "No framing"
	}
}

// BytesToHex converts bytes to uppercase hex string with spaces
func BytesToHex(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	return strings.ToUpper(hex.EncodeToString(data))
}
