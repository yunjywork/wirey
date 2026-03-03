/**
 * Hex dump utilities for Wireshark-style display
 */

/**
 * Decode base64 string to Uint8Array
 */
export function base64ToBytes(base64: string): Uint8Array {
  const binaryString = atob(base64)
  const bytes = new Uint8Array(binaryString.length)
  for (let i = 0; i < binaryString.length; i++) {
    bytes[i] = binaryString.charCodeAt(i)
  }
  return bytes
}

/**
 * Format a single byte as 2-digit uppercase hex
 */
function byteToHex(byte: number): string {
  return byte.toString(16).toUpperCase().padStart(2, '0')
}

/**
 * Convert byte to printable ASCII character or dot
 */
function byteToAscii(byte: number): string {
  // Printable ASCII range: 32-126
  if (byte >= 32 && byte <= 126) {
    return String.fromCharCode(byte)
  }
  return '.'
}

/**
 * Format bytes as Wireshark-style hex dump
 *
 * Example output:
 * 00000000  48 65 6C 6C 6F 20 57 6F  72 6C 64 0A              Hello World.
 * 00000010  00 00 00 0B 54 65 73 74  20 44 61 74 61           ....Test Data
 */
export function formatHexDump(base64Data: string): string {
  if (!base64Data) return ''

  const bytes = base64ToBytes(base64Data)
  const lines: string[] = []
  const bytesPerLine = 16

  for (let offset = 0; offset < bytes.length; offset += bytesPerLine) {
    const chunk = bytes.slice(offset, offset + bytesPerLine)

    // Offset (8 hex digits)
    const offsetStr = offset.toString(16).toUpperCase().padStart(8, '0')

    // Hex bytes (two groups of 8 bytes)
    const hexParts: string[] = []
    const asciiParts: string[] = []

    for (let i = 0; i < bytesPerLine; i++) {
      if (i < chunk.length) {
        hexParts.push(byteToHex(chunk[i]))
        asciiParts.push(byteToAscii(chunk[i]))
      } else {
        hexParts.push('  ')
        asciiParts.push(' ')
      }
    }

    // Format: first 8 bytes, space, next 8 bytes
    const hexStr = hexParts.slice(0, 8).join(' ') + '  ' + hexParts.slice(8).join(' ')
    const asciiStr = asciiParts.join('')

    lines.push(`${offsetStr}  ${hexStr}  ${asciiStr}`)
  }

  return lines.join('\n')
}

/**
 * Convert bytes to space-separated uppercase hex string
 * Example: "48 65 6C 6C 6F"
 */
export function bytesToHex(base64Data: string): string {
  if (!base64Data) return ''

  const bytes = base64ToBytes(base64Data)
  return Array.from(bytes).map(b => byteToHex(b)).join(' ')
}

/**
 * Format hex string for display (uppercase, with spaces)
 * Input can be with or without spaces
 */
export function formatHexString(hex: string): string {
  // Remove existing spaces and convert to uppercase
  const clean = hex.replace(/\s/g, '').toUpperCase()
  // Add spaces every 2 characters
  return clean.match(/.{1,2}/g)?.join(' ') || ''
}

/**
 * Get a short preview of message content (max 30 chars)
 */
export function getContentPreview(content: string, maxLength = 30): string {
  // Clean up for display (replace control chars with dots)
  const cleaned = content.replace(/[\x00-\x1F\x7F-\x9F]/g, '.')

  if (cleaned.length <= maxLength) {
    return cleaned
  }
  return cleaned.slice(0, maxLength - 3) + '...'
}

/**
 * Format timestamp as HH:MM:SS.mmm
 */
export function formatTimestamp(date: Date): string {
  const hours = date.getHours().toString().padStart(2, '0')
  const minutes = date.getMinutes().toString().padStart(2, '0')
  const seconds = date.getSeconds().toString().padStart(2, '0')
  const ms = date.getMilliseconds().toString().padStart(3, '0')
  return `${hours}:${minutes}:${seconds}.${ms}`
}

/**
 * Format byte size for display
 */
export function formatSize(bytes: number): string {
  if (bytes === 1) return '1 byte'
  return `${bytes} bytes`
}

/**
 * Highlight type for a byte
 */
type HighlightType = 'header' | 'footer' | 'none'

/**
 * Get highlight type for a byte at given index
 */
function getHighlightType(
  index: number,
  totalBytes: number,
  headerBytes: number,
  footerBytes: number,
  headerOffset: number = 0
): HighlightType {
  // Header: from headerOffset to headerOffset + headerBytes
  if (headerBytes > 0 && index >= headerOffset && index < headerOffset + headerBytes) {
    return 'header'
  }
  if (footerBytes > 0 && index >= totalBytes - footerBytes) {
    return 'footer'
  }
  return 'none'
}

/**
 * Wrap text with highlight span
 */
function wrapWithHighlight(text: string, type: HighlightType): string {
  if (type === 'header') {
    return `<span class="hex-header">${text}</span>`
  }
  if (type === 'footer') {
    return `<span class="hex-footer">${text}</span>`
  }
  return text
}

/**
 * Format bytes as Wireshark-style hex dump with frame highlighting
 * Returns HTML string with span elements for header/footer highlighting
 *
 * @param base64Data - Base64 encoded data
 * @param headerBytes - Number of bytes in frame header (0 if none)
 * @param footerBytes - Number of bytes in frame footer (0 if none)
 */
export function formatHexDumpWithHighlight(
  base64Data: string,
  headerBytes: number = 0,
  footerBytes: number = 0
): string {
  if (!base64Data) return ''

  const bytes = base64ToBytes(base64Data)
  return formatBytesHexDump(bytes, headerBytes, footerBytes)
}

/**
 * Format raw bytes as Wireshark-style hex dump with frame highlighting
 * Returns HTML string with span elements for header/footer highlighting
 *
 * @param bytes - Raw byte array (Uint8Array or number[])
 * @param headerBytes - Number of bytes in frame header (0 if none)
 * @param footerBytes - Number of bytes in frame footer (0 if none)
 * @param headerOffset - Offset where header starts (default 0)
 */
export function formatBytesHexDump(
  bytes: Uint8Array | number[],
  headerBytes: number = 0,
  footerBytes: number = 0,
  headerOffset: number = 0
): string {
  if (!bytes || bytes.length === 0) return ''

  const byteArray = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes)
  const totalBytes = byteArray.length
  const lines: string[] = []
  const bytesPerLine = 16

  for (let offset = 0; offset < byteArray.length; offset += bytesPerLine) {
    const chunk = byteArray.slice(offset, offset + bytesPerLine)

    // Offset (8 hex digits) - no highlight
    const offsetStr = offset.toString(16).toUpperCase().padStart(8, '0')

    // Build data for this line
    const lineData: { hex: string; ascii: string; type: HighlightType }[] = []

    for (let i = 0; i < bytesPerLine; i++) {
      const globalIndex = offset + i

      if (i < chunk.length) {
        const highlightType = getHighlightType(globalIndex, totalBytes, headerBytes, footerBytes, headerOffset)
        lineData.push({
          hex: byteToHex(chunk[i]),
          ascii: byteToAscii(chunk[i]),
          type: highlightType
        })
      } else {
        lineData.push({ hex: '  ', ascii: ' ', type: 'none' })
      }
    }

    // Build hex parts with individual highlights
    const hexParts = lineData.map(d => wrapWithHighlight(d.hex, d.type))

    // Build ASCII string by grouping consecutive same-type characters
    let asciiStr = ''
    let currentGroup = ''
    let currentType: HighlightType = 'none'

    for (const d of lineData) {
      if (d.type === currentType) {
        currentGroup += d.ascii
      } else {
        if (currentGroup) {
          asciiStr += wrapWithHighlight(currentGroup, currentType)
        }
        currentGroup = d.ascii
        currentType = d.type
      }
    }
    if (currentGroup) {
      asciiStr += wrapWithHighlight(currentGroup, currentType)
    }

    // Format: first 8 bytes, double space, next 8 bytes
    const hexStr = hexParts.slice(0, 8).join(' ') + '  ' + hexParts.slice(8).join(' ')

    lines.push(`${offsetStr}  ${hexStr}  ${asciiStr}`)
  }

  return lines.join('\n')
}

/**
 * Escape control characters for display
 * Control chars (0x00-0x1F, 0x7F) → \n, \t, \r, \0, \xNN
 * All other characters (including UTF-8) → as-is
 */
export function escapeForDisplay(str: string): string {
  let result = ''
  for (let i = 0; i < str.length; i++) {
    const code = str.charCodeAt(i)

    if (code < 0x20 || code === 0x7F) {
      // Control characters
      switch (code) {
        case 0x00: result += '\\0'; break
        case 0x09: result += '\\t'; break
        case 0x0A: result += '\\n'; break
        case 0x0D: result += '\\r'; break
        default:
          result += '\\x' + code.toString(16).padStart(2, '0')
      }
    } else {
      // Printable ASCII and all UTF-8 characters
      result += str[i]
    }
  }
  return result
}
