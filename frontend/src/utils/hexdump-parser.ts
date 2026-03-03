/**
 * Hex dump parser utilities
 * Parses various hex dump formats and extracts raw bytes
 */

export interface ParseResult {
  bytes: Uint8Array
  format: string
  byteCount: number
}

/**
 * Auto-detect and parse hex dump input
 * Supports multiple formats:
 * 1. User format: | offset | hex (4-byte groups) | ascii |
 * 2. Wireshark format: offset  hex bytes (1-byte each)  ascii
 * 3. xxd format: 00000000: 4865 6c6c 6f20  Hello
 * 4. tcpdump format: 0x0000:  4500 0054 0000  E..T..
 * 5. Pure hex string (with or without spaces)
 */
export function parseHexDump(input: string): ParseResult | null {
  if (!input || !input.trim()) return null

  const trimmed = input.trim()

  // Try user format first (has | separators)
  if (trimmed.includes('|')) {
    const bytes = parseUserFormat(trimmed)
    if (bytes) {
      return { bytes, format: 'User Format (4-byte groups)', byteCount: bytes.length }
    }
  }

  // Try xxd format (offset with colon: "00000000: 4865 6c6c")
  if (/^[0-9A-Fa-f]+:\s/.test(trimmed)) {
    const bytes = parseXxdFormat(trimmed)
    if (bytes) {
      return { bytes, format: 'xxd Format', byteCount: bytes.length }
    }
  }

  // Try tcpdump format (0x prefix: "0x0000:  4500 0054")
  if (/^0x[0-9A-Fa-f]+:\s/.test(trimmed)) {
    const bytes = parseTcpdumpFormat(trimmed)
    if (bytes) {
      return { bytes, format: 'tcpdump Format', byteCount: bytes.length }
    }
  }

  // Try Wireshark format (starts with offset, has double space)
  if (/^[0-9A-Fa-f]{8}\s/.test(trimmed)) {
    const bytes = parseWiresharkFormat(trimmed)
    if (bytes) {
      return { bytes, format: 'Wireshark Format', byteCount: bytes.length }
    }
  }

  // Try pure hex string
  const bytes = parsePureHex(trimmed)
  if (bytes) {
    return { bytes, format: 'Pure Hex String', byteCount: bytes.length }
  }

  return null
}

/**
 * Parse user format: | 00000000 | 30303030  31323531 | ascii |
 */
function parseUserFormat(input: string): Uint8Array | null {
  const lines = input.split('\n')
  const allBytes: number[] = []

  for (const line of lines) {
    // Skip empty lines
    if (!line.trim()) continue

    // Split by | and get the hex part (second segment)
    const parts = line.split('|').map(p => p.trim()).filter(p => p)

    if (parts.length < 2) continue

    // First part is offset, second is hex data
    // Hex data is in 4-byte groups (8 hex chars) separated by spaces
    const hexPart = parts[1]

    // Remove spaces and parse hex
    const hexClean = hexPart.replace(/\s+/g, '')

    // Validate it's hex
    if (!/^[0-9A-Fa-f]+$/.test(hexClean)) continue

    // Convert to bytes
    for (let i = 0; i < hexClean.length; i += 2) {
      if (i + 1 < hexClean.length) {
        const byte = parseInt(hexClean.substring(i, i + 2), 16)
        if (!isNaN(byte)) {
          allBytes.push(byte)
        }
      }
    }
  }

  if (allBytes.length === 0) return null
  return new Uint8Array(allBytes)
}

/**
 * Parse Wireshark format: 00000000  48 65 6C 6C 6F 20 57 6F  72 6C 64 0A  Hello World.
 * Format: OFFSET(8) + SPACE(2) + HEX_BYTES(48 chars for 16 bytes) + SPACE(2) + ASCII(16)
 */
function parseWiresharkFormat(input: string): Uint8Array | null {
  const lines = input.split('\n')
  const allBytes: number[] = []

  for (const line of lines) {
    if (!line.trim()) continue

    // Match offset at start (8 hex digits)
    const match = line.match(/^([0-9A-Fa-f]{8})\s+(.+)$/)
    if (!match) continue

    // Get the rest after offset
    let rest = match[2]

    // Wireshark format: hex section is ~48 chars (16 bytes * 2 + spaces + gap)
    // Then 2+ spaces before ASCII column
    // Take only the hex section (first ~49 chars) to avoid ASCII column
    // The hex section pattern: "XX XX XX XX XX XX XX XX  XX XX XX XX XX XX XX XX"
    const hexSection = rest.substring(0, 49)

    // Extract all valid hex byte pairs from hex section only
    const hexMatches = hexSection.match(/\b[0-9A-Fa-f]{2}\b/g)
    if (!hexMatches) continue

    for (const hex of hexMatches) {
      const byte = parseInt(hex, 16)
      if (!isNaN(byte)) {
        allBytes.push(byte)
      }
    }
  }

  if (allBytes.length === 0) return null
  return new Uint8Array(allBytes)
}

/**
 * Parse xxd format: 00000000: 4865 6c6c 6f20 576f 726c 640a  Hello World.
 * Groups of 4 hex chars (2 bytes) separated by spaces
 * Format: OFFSET: + SPACE(2) + HEX_GROUPS(39 chars for 16 bytes) + SPACE(2) + ASCII
 */
function parseXxdFormat(input: string): Uint8Array | null {
  const lines = input.split('\n')
  const allBytes: number[] = []

  for (const line of lines) {
    if (!line.trim()) continue

    // Match offset with colon
    const match = line.match(/^([0-9A-Fa-f]+):\s+(.+)$/)
    if (!match) continue

    const rest = match[2]

    // xxd hex section: 8 groups of 4 chars + 7 spaces = 39 chars for 16 bytes
    // Take only the hex section to avoid ASCII column
    const hexSection = rest.substring(0, 40)

    // xxd uses 4-char groups (2 bytes), extract all hex groups
    // Split by whitespace and filter valid hex groups
    const parts = hexSection.split(/\s+/)

    for (const part of parts) {
      // Skip if it looks like ASCII (contains non-hex or is too short)
      if (!/^[0-9A-Fa-f]{2,}$/.test(part)) continue

      // Parse each pair of hex chars
      for (let i = 0; i < part.length; i += 2) {
        if (i + 1 < part.length) {
          const byte = parseInt(part.substring(i, i + 2), 16)
          if (!isNaN(byte)) {
            allBytes.push(byte)
          }
        }
      }
    }
  }

  if (allBytes.length === 0) return null
  return new Uint8Array(allBytes)
}

/**
 * Parse tcpdump format: 0x0000:  4500 0054 0000 4000 4001 b7cf  E..T..@.@...
 * Similar to xxd but with 0x prefix on offset
 * Format: 0xOFFSET: + SPACE(2) + HEX_GROUPS(~40 chars) + SPACE(2) + ASCII
 */
function parseTcpdumpFormat(input: string): Uint8Array | null {
  const lines = input.split('\n')
  const allBytes: number[] = []

  for (const line of lines) {
    if (!line.trim()) continue

    // Match 0x offset with colon
    const match = line.match(/^0x([0-9A-Fa-f]+):\s+(.+)$/)
    if (!match) continue

    const rest = match[2]

    // Take only the hex section to avoid ASCII column
    const hexSection = rest.substring(0, 40)

    // Split by whitespace and filter valid hex groups
    const parts = hexSection.split(/\s+/)

    for (const part of parts) {
      // Skip if it looks like ASCII
      if (!/^[0-9A-Fa-f]{2,}$/.test(part)) continue

      // Parse each pair of hex chars
      for (let i = 0; i < part.length; i += 2) {
        if (i + 1 < part.length) {
          const byte = parseInt(part.substring(i, i + 2), 16)
          if (!isNaN(byte)) {
            allBytes.push(byte)
          }
        }
      }
    }
  }

  if (allBytes.length === 0) return null
  return new Uint8Array(allBytes)
}

/**
 * Parse pure hex string (with or without spaces/newlines)
 * Examples:
 *   48656C6C6F
 *   48 65 6C 6C 6F
 *   48656C6C6F20576F726C64
 */
function parsePureHex(input: string): Uint8Array | null {
  // Remove all whitespace
  const hexClean = input.replace(/\s+/g, '')

  // Must be even length and all hex
  if (hexClean.length === 0 || hexClean.length % 2 !== 0) return null
  if (!/^[0-9A-Fa-f]+$/.test(hexClean)) return null

  const bytes: number[] = []
  for (let i = 0; i < hexClean.length; i += 2) {
    const byte = parseInt(hexClean.substring(i, i + 2), 16)
    if (!isNaN(byte)) {
      bytes.push(byte)
    }
  }

  if (bytes.length === 0) return null
  return new Uint8Array(bytes)
}

/**
 * Convert bytes to hex string
 * @param bytes - Uint8Array of bytes
 * @param separator - Separator between bytes (default: space)
 */
export function bytesToHexString(bytes: Uint8Array, separator: string = ' '): string {
  return Array.from(bytes)
    .map(b => b.toString(16).toUpperCase().padStart(2, '0'))
    .join(separator)
}

/**
 * Convert bytes to raw string (for message input)
 */
export function bytesToRawString(bytes: Uint8Array): string {
  return Array.from(bytes)
    .map(b => String.fromCharCode(b))
    .join('')
}
