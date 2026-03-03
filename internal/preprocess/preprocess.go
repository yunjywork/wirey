package preprocess

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Preprocessor handles message preprocessing with variables and escape sequences
type Preprocessor struct {
	counters map[string]int64
	mu       sync.Mutex
}

// NewPreprocessor creates a new preprocessor instance
func NewPreprocessor() *Preprocessor {
	return &Preprocessor{
		counters: make(map[string]int64),
	}
}

// Global default preprocessor
var defaultPreprocessor = NewPreprocessor()

// Process applies variable substitution and escape sequence processing
func Process(message string, useVariables bool) (string, error) {
	return defaultPreprocessor.Process(message, useVariables)
}

// Process applies variable substitution and escape sequence processing
func (p *Preprocessor) Process(message string, useVariables bool) (string, error) {
	result := message

	// 1. Apply variable substitution if enabled
	if useVariables {
		result = p.replaceVariables(result)
	}

	// 2. Always process escape sequences (inline hex)
	result = ProcessEscapeSequences(result)

	return result, nil
}

// ResetCounter resets the counter for a specific session
func (p *Preprocessor) ResetCounter(sessionID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.counters, sessionID)
}

// ResetAllCounters resets all counters
func (p *Preprocessor) ResetAllCounters() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.counters = make(map[string]int64)
}

// replaceVariables replaces {{variable}} patterns with actual values
func (p *Preprocessor) replaceVariables(message string) string {
	// Time-based variables
	now := time.Now()

	// {{timestamp}} - Unix timestamp (seconds)
	message = strings.ReplaceAll(message, "{{timestamp}}", strconv.FormatInt(now.Unix(), 10))

	// {{timestamp_ms}} - Unix timestamp (milliseconds)
	message = strings.ReplaceAll(message, "{{timestamp_ms}}", strconv.FormatInt(now.UnixMilli(), 10))

	// {{datetime}} - ISO 8601 format
	message = strings.ReplaceAll(message, "{{datetime}}", now.UTC().Format(time.RFC3339))

	// {{date}} - Date only
	message = strings.ReplaceAll(message, "{{date}}", now.UTC().Format("2006-01-02"))

	// {{time}} - Time only
	message = strings.ReplaceAll(message, "{{time}}", now.UTC().Format("15:04:05"))

	// {{uuid}} - UUID v4
	message = replaceUUID(message)

	// {{random:N}} - N bytes random hex
	message = replaceRandom(message)

	// {{counter}} and {{counter:N}} - Auto-incrementing counter
	message = p.replaceCounter(message)

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
var randomRegex = regexp.MustCompile(`\{\{random(?::(\d+))?\}\}`)

func replaceRandom(message string) string {
	return randomRegex.ReplaceAllStringFunc(message, func(match string) string {
		matches := randomRegex.FindStringSubmatch(match)

		// Default to 4 bytes if no size specified
		n := 4
		if len(matches) >= 2 && matches[1] != "" {
			parsed, err := strconv.Atoi(matches[1])
			if err != nil || parsed <= 0 || parsed > 256 {
				return match
			}
			n = parsed
		}

		bytes := make([]byte, n)
		if _, err := rand.Read(bytes); err != nil {
			return match
		}

		return strings.ToUpper(hex.EncodeToString(bytes))
	})
}

// replaceCounter replaces {{counter}} and {{counter:N}} with auto-incrementing values
var counterRegex = regexp.MustCompile(`\{\{counter(?::(\d+))?\}\}`)

func (p *Preprocessor) replaceCounter(message string) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	return counterRegex.ReplaceAllStringFunc(message, func(match string) string {
		matches := counterRegex.FindStringSubmatch(match)

		// Determine start value
		startValue := int64(1)
		if len(matches) >= 2 && matches[1] != "" {
			if v, err := strconv.ParseInt(matches[1], 10, 64); err == nil {
				startValue = v
			}
		}

		// Use match pattern as key to track different counters
		key := match
		if _, exists := p.counters[key]; !exists {
			p.counters[key] = startValue
		}

		value := p.counters[key]
		p.counters[key]++

		return strconv.FormatInt(value, 10)
	})
}

// ProcessEscapeSequences converts escape sequences to actual bytes
func ProcessEscapeSequences(message string) string {
	var result strings.Builder
	i := 0

	for i < len(message) {
		if message[i] == '\\' && i+1 < len(message) {
			switch message[i+1] {
			case 'n':
				result.WriteByte('\n')
				i += 2
			case 'r':
				result.WriteByte('\r')
				i += 2
			case 't':
				result.WriteByte('\t')
				i += 2
			case '0':
				result.WriteByte(0)
				i += 2
			case '\\':
				result.WriteByte('\\')
				i += 2
			case 'x':
				// \xNN - hex byte
				if i+3 < len(message) {
					hexStr := message[i+2 : i+4]
					if b, err := strconv.ParseUint(hexStr, 16, 8); err == nil {
						result.WriteByte(byte(b))
						i += 4
						continue
					}
				}
				// Invalid hex, keep as-is
				result.WriteByte(message[i])
				i++
			default:
				// Unknown escape, keep backslash
				result.WriteByte(message[i])
				i++
			}
		} else {
			result.WriteByte(message[i])
			i++
		}
	}

	return result.String()
}

// HasVariables checks if the message contains any variable patterns
func HasVariables(message string) bool {
	return strings.Contains(message, "{{") && strings.Contains(message, "}}")
}

// GetVariablesList returns a list of variable names found in the message
func GetVariablesList(message string) []string {
	varRegex := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := varRegex.FindAllStringSubmatch(message, -1)

	seen := make(map[string]bool)
	var result []string

	for _, match := range matches {
		if len(match) >= 2 && !seen[match[1]] {
			seen[match[1]] = true
			result = append(result, match[1])
		}
	}

	return result
}

// Preview returns the preprocessed message for preview (without incrementing counters)
func Preview(message string) string {
	// Create a temporary preprocessor for preview
	temp := NewPreprocessor()
	result, _ := temp.Process(message, true)
	return result
}

// SupportedVariables returns a list of supported built-in variables
func SupportedVariables() []string {
	return []string{
		"{{timestamp}}",
		"{{timestamp_ms}}",
		"{{datetime}}",
		"{{date}}",
		"{{time}}",
		"{{uuid}}",
		"{{random:N}}",
		"{{counter}}",
		"{{counter:N}}",
	}
}

// SupportedEscapeSequences returns a list of supported escape sequences
func SupportedEscapeSequences() map[string]string {
	return map[string]string{
		"\\n":   "LF (0x0A)",
		"\\r":   "CR (0x0D)",
		"\\t":   "TAB (0x09)",
		"\\0":   "NULL (0x00)",
		"\\\\":  "Backslash",
		"\\xNN": "Hex byte (e.g., \\x01 for SOH)",
	}
}

// FormatExample returns an example of FIX protocol usage
func FormatExample() string {
	return `8=FIX.4.4\x0135=D\x0149=SENDER\x0156=TARGET\x0134={{counter}}\x0152={{timestamp}}\x01`
}
