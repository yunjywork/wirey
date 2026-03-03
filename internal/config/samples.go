package config

// SampleCollection returns the "Getting Started" sample collection metadata
func SampleCollection() Collection {
	return Collection{
		Name:        "Getting Started",
		Description: "Sample cases demonstrating Wirey features",
		Order:       0,
		Notes: `# Getting Started with Wirey

Welcome! This collection demonstrates Wirey's key features.

## Sample Cases

| Case | Description |
|------|-------------|
| TCPBin Echo | Connect to external TCP echo server |
| Echo Server Basic | Test with built-in Echo Server |
| Delimiter Framing | Line-based message framing |
| Length Prefix | Binary length-prefixed protocol |
| Fixed Length | Fixed-size message framing |
| Variables | Dynamic variable substitution |
| Hex Mode | Send raw hexadecimal bytes |
| Scripts | Setup/Post-recv script example |

## Quick Start

1. **TCPBin Echo**: Click Connect to test with tcpbin.com
2. **Echo Server**: Start Echo Server (header button), then connect
3. Check Notes tab in each case for detailed explanations`,
	}
}

// SampleCases returns the sample cases for "Getting Started" collection
func SampleCases() []SavedCase {
	return []SavedCase{
		{
			ID:           "sample-tcpbin",
			Name:         "TCPBin Echo",
			Protocol:     "tcp",
			Host:         "tcpbin.com",
			Port:         4242,
			Order:        0,
			Framing:      &FramingConfig{Mode: "delimiter", Delimiter: "\\n"},
			DraftMessage: "Hello from Wirey!",
			DraftFormat:  "text",
			UseVariables: false,
			Notes: `# TCPBin Echo Server

[tcpbin.com](https://tcpbin.com) is a free TCP echo service for testing.

## How to Use

1. Click **Connect** to establish connection
2. Type a message and click **Send**
3. The server echoes back your message

## Connection Details

- **Host**: tcpbin.com
- **Port**: 4242
- **Protocol**: TCP

## Tips

- Add ` + "`\\n`" + ` at the end for line-based protocols
- Check the message log for sent/received data`,
		},
		{
			ID:           "sample-echo",
			Name:         "Echo Server Basic",
			Protocol:     "tcp",
			Host:         "127.0.0.1",
			Port:         39876,
			Order:        1,
			DraftMessage: "Test message for echo",
			DraftFormat:  "text",
			UseVariables: false,
			Notes: `# Built-in Echo Server

Wirey includes a built-in echo server for local testing.

## How to Use

1. Click **Echo Server** button in the sidebar (bottom)
2. Start the server on port 39876
3. Come back to this case and click **Connect**
4. Send any message - it will be echoed back

## Why Use This?

- Test without external dependencies
- Works offline
- Fast local testing`,
		},
		{
			ID:           "sample-delimiter",
			Name:         "Delimiter Framing",
			Protocol:     "tcp",
			Host:         "127.0.0.1",
			Port:         39876,
			Order:        2,
			Framing:      &FramingConfig{Mode: "delimiter", Delimiter: "\\n"},
			DraftMessage: "Line-based message",
			DraftFormat:  "text",
			UseVariables: false,
			Notes: `# Delimiter-Based Framing

Messages are separated by a delimiter character.

## Common Delimiters

| Delimiter | Use Case |
|-----------|----------|
| ` + "`\\n`" + ` | Line-based protocols (HTTP, SMTP) |
| ` + "`\\r\\n`" + ` | Windows-style line endings |
| ` + "`\\0`" + ` | Null-terminated strings |

## How It Works

1. **Sending**: Delimiter is appended to your message
2. **Receiving**: Data is split at delimiter boundaries

## Current Settings

- **Delimiter**: ` + "`\\n`" + ` (newline)`,
		},
		{
			ID:       "sample-length",
			Name:     "Length Prefix",
			Protocol: "tcp",
			Host:     "127.0.0.1",
			Port:     39876,
			Order:    3,
			Framing: &FramingConfig{
				Mode:           "length-prefix",
				LengthBytes:    4,
				Endian:         "big",
				LengthEncoding: "binary",
				IncludeHeader:  false,
				LengthMode:     "append",
			},
			DraftMessage: "Length prefixed payload",
			DraftFormat:  "text",
			UseVariables: false,
			Notes: `# Length-Prefix Framing

Each message is prefixed with its length.

## How It Works

` + "```" + `
[4 bytes: length][payload bytes...]
` + "```" + `

## Settings

| Option | Description |
|--------|-------------|
| **Length Bytes** | Size of length field (1, 2, or 4) |
| **Endian** | Byte order (big/little) |
| **Encoding** | binary, ascii, hex, or bcd |
| **Include Header** | Whether length includes header size |

## Current Settings

- **4 bytes**, Big Endian, Binary
- Length = payload size only`,
		},
		{
			ID:           "sample-fixed",
			Name:         "Fixed Length",
			Protocol:     "tcp",
			Host:         "127.0.0.1",
			Port:         39876,
			Order:        4,
			Framing:      &FramingConfig{Mode: "fixed-length", FixedSize: 64},
			DraftMessage: "Fixed size message",
			DraftFormat:  "text",
			UseVariables: false,
			Notes: `# Fixed-Length Framing

All messages have exactly the same size.

## How It Works

- **Sending**: Message is padded to fixed size
- **Receiving**: Data is read in fixed-size chunks

## Settings

| Option | Description |
|--------|-------------|
| **Fixed Size** | Exact message size in bytes |
| **Padding** | Byte used to fill remaining space (0x00) |
| **Padding Position** | Left or right padding |

## Current Settings

- **Fixed Size**: 64 bytes
- **Padding**: 0x00 (null byte)`,
		},
		{
			ID:           "sample-variables",
			Name:         "Variables",
			Protocol:     "tcp",
			Host:         "127.0.0.1",
			Port:         39876,
			Order:        5,
			DraftMessage: "SEQ={{counter}} TS={{timestamp}} ID={{uuid}}\\n",
			DraftFormat:  "text",
			UseVariables: true,
			Notes: `# Variable Substitution

Insert dynamic values into your messages.

## Available Variables

| Variable | Description |
|----------|-------------|
| ` + "`{{timestamp}}`" + ` | Unix timestamp (seconds) |
| ` + "`{{timestamp_ms}}`" + ` | Unix timestamp (milliseconds) |
| ` + "`{{date}}`" + ` | Current date (YYYY-MM-DD) |
| ` + "`{{time}}`" + ` | Current time (HH:MM:SS) |
| ` + "`{{datetime}}`" + ` | ISO 8601 datetime |
| ` + "`{{uuid}}`" + ` | Random UUID v4 |
| ` + "`{{random}}`" + ` | Random number (0-999999) |
| ` + "`{{counter}}`" + ` | Auto-incrementing counter |
| ` + "`{{counter:N}}`" + ` | Counter starting from N |

## Escape Sequences

| Sequence | Output |
|----------|--------|
| ` + "`\\n`" + ` | Newline |
| ` + "`\\r`" + ` | Carriage return |
| ` + "`\\t`" + ` | Tab |
| ` + "`\\0`" + ` | Null byte |
| ` + "`\\xNN`" + ` | Hex byte |

## Example

` + "```" + `
SEQ={{counter}} TS={{timestamp}} ID={{uuid}}\n
` + "```" + `

Produces: ` + "`SEQ=1 TS=1703318400 ID=a1b2c3d4-...\n`",
		},
		{
			ID:           "sample-hex",
			Name:         "Hex Mode",
			Protocol:     "tcp",
			Host:         "127.0.0.1",
			Port:         39876,
			Order:        6,
			DraftMessage: "48 65 6C 6C 6F",
			DraftFormat:  "hex",
			UseVariables: false,
			Notes: `# Hex Mode

Send raw bytes using hexadecimal notation.

## How to Use

1. Toggle input mode to **Hex**
2. Enter hex bytes separated by spaces

## Example

` + "```" + `
48 65 6C 6C 6F
` + "```" + `

This sends the ASCII string "Hello" (5 bytes).

## Tips

- Use spaces between bytes for readability
- Mix with escape sequences if needed
- Great for binary protocols`,
		},
		{
			ID:             "sample-scripts",
			Name:           "Scripts",
			Protocol:       "tcp",
			Host:           "127.0.0.1",
			Port:           39876,
			Order:          7,
			Framing:        &FramingConfig{Mode: "delimiter", Delimiter: "\\n"},
			DraftMessage:   "REQ|{{seq}}|{{reqId}}|PING",
			DraftFormat:    "text",
			UseVariables:   true,
			PostRecvSample: "REQ|1|a1b2c3d4|PING",
			ScriptConfig: &ScriptConfig{
				SetupScript: `// Generate unique request ID
const reqId = wirey.uuid().substring(0, 8);

// Increment sequence number
const seq = (wirey.collection.get('seq') ?? 0) + 1;
wirey.collection.set('seq', seq);

// Store for validation in post-recv
wirey.set('reqId', reqId);
wirey.set('seq', seq);

wirey.log('→ Sending: seq=' + seq + ', reqId=' + reqId);`,
				SetupEnabled:    true,
				PreSendScript:   "return msg;",
				PreSendEnabled:  false,
				PostRecvScript: `// Parse response: REQ|seq|reqId|PING
const parts = msg.split('|');
if (parts.length >= 3) {
  const respSeq = parseInt(parts[1], 10);
  const respId = parts[2];

  // Compare with sent values
  const sentSeq = wirey.get('seq');
  const sentId = wirey.get('reqId');

  if (respSeq === sentSeq && respId === sentId) {
    wirey.log('✓ Validated: seq=' + respSeq);
  } else {
    wirey.log('✗ Mismatch! sent=' + sentSeq + '/' + sentId + ' recv=' + respSeq + '/' + respId);
  }
}`,
				PostRecvEnabled: true,
			},
			Notes: `# Scripts Example

Demonstrates Setup and Post-recv scripts for request-response validation.

## Scenario

Each request includes a unique ID and sequence number. Post-recv validates that the response matches what was sent.

## Message Format

` + "```" + `
REQ|{{seq}}|{{reqId}}|PING
` + "```" + `

## Script Flow

### 1. Setup Script (before send)

` + "```" + `javascript
const reqId = wirey.uuid().substring(0, 8);
const seq = (wirey.collection.get('seq') ?? 0) + 1;
wirey.collection.set('seq', seq);
wirey.set('reqId', reqId);
wirey.set('seq', seq);
` + "```" + `

- Generates unique request ID
- Increments sequence (persisted in collection)
- Stores values for post-recv validation

### 2. Post-recv Script (after receive)

` + "```" + `javascript
const parts = msg.split('|');
const respSeq = parseInt(parts[1], 10);
const respId = parts[2];
if (respSeq === wirey.get('seq') && respId === wirey.get('reqId')) {
  wirey.log('✓ Validated');
}
` + "```" + `

- Parses response fields
- Compares with sent values
- Logs validation result

## wirey API

| Function | Description |
|----------|-------------|
| ` + "`wirey.get(key)`" + ` | Get case variable |
| ` + "`wirey.set(key, val)`" + ` | Set case variable |
| ` + "`wirey.collection.get(key)`" + ` | Get collection variable (persisted) |
| ` + "`wirey.collection.set(key, val)`" + ` | Set collection variable (persisted) |
| ` + "`wirey.uuid()`" + ` | Generate UUID |
| ` + "`wirey.log(...)`" + ` | Output to message log |

## Test

1. Start Echo Server
2. Connect and Send multiple times
3. Check message log for validation results
4. View Scripts tab to see/edit scripts`,
		},
	}
}
