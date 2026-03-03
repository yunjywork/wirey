// Protocol types
export type Protocol = 'tcp' | 'udp'

// Tab types
export type TabType = 'case' | 'collection' | 'echo' | 'connections'

export interface Tab {
  type: TabType
  id: string
}

// Connection status
export type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'error'

// Message direction
export type MessageDirection = 'sent' | 'received' | 'system'

// Message format
export type MessageFormat = 'text' | 'hex'

// Charset types
export type Charset = 'utf-8' | 'ascii' | 'euc-kr' | 'iso-8859-1' | 'shift-jis' | 'gb2312'
export type CharsetMode = 'global' | 'collection' | Charset  // inheritance: global -> collection -> case

// Common charsets for autocomplete
export const COMMON_CHARSETS = [
  'utf-8', 'ascii', 'euc-kr', 'iso-8859-1', 'shift-jis', 'gb2312',
  'utf-16', 'utf-16le', 'utf-16be', 'windows-1252', 'big5', 'koi8-r'
] as const

// Connection settings
export interface ConnectionSettings {
  connectTimeout: number  // ms, default: 10000 (10s)
  readTimeout: number     // ms, default: 60000 (60s), 0 = unlimited
}

export const DEFAULT_CONNECTION_SETTINGS: ConnectionSettings = {
  connectTimeout: 10000,
  readTimeout: 60000
}

// Framing types
export type FramingMode = 'collection' | 'none' | 'delimiter' | 'length-prefix' | 'fixed-length'
export type LengthEncoding = 'binary' | 'ascii' | 'hex' | 'bcd'
export type Endian = 'big' | 'little'
export type LengthPrefixMode = 'append' | 'rewrite'
export type PaddingPosition = 'left' | 'right'

// Framing configuration
export interface FramingConfig {
  mode: FramingMode
  // Delimiter options
  delimiter?: string  // '\n', '\r\n', '\0', custom
  // Length Prefix options
  lengthEncoding?: LengthEncoding
  lengthOffset?: number
  lengthBytes?: number
  endian?: Endian
  includeHeader?: boolean
  lengthMode?: LengthPrefixMode
  // Fixed Length options
  fixedSize?: number
  paddingPosition?: PaddingPosition  // 'left' or 'right' (default: right)
  paddingByte?: number               // 0x00-0xFF (default: 0x00)
}

// Default framing config
export const DEFAULT_FRAMING: FramingConfig = {
  mode: 'collection'  // Use collection's shared framing by default
}

// Collection interface
export interface Collection {
  name: string  // folder name = ID
  description?: string
  createdAt: Date
  updatedAt: Date
  order: number  // Display order
  sharedFraming?: FramingConfig
  sharedCharset?: Charset                    // collection-level charset (overrides global)
  sharedConnectionSettings?: ConnectionSettings  // collection-level timeout settings
  notes?: string                             // Markdown notes
  // Script configuration
  sharedScriptConfig?: ScriptConfig
  sharedVariables?: Variables
  cases: Case[]  // runtime only (not saved)
  isExpanded?: boolean  // UI state
}

// Case interface
export interface Case {
  id: string
  name: string
  collectionName: string  // parent collection
  protocol: Protocol
  host: string
  port: number
  status: ConnectionStatus
  createdAt: Date
  updatedAt?: Date
  order: number  // Display order
  messages: Message[]
  isSaved: boolean
  framing: FramingConfig
  charset: CharsetMode                   // 'global' or specific charset
  connectionSettings?: ConnectionSettings // timeout settings
  notes?: string                         // Markdown notes
  // Runtime connection info
  localAddr?: string      // Local IP:Port when connected
  // Draft message (persisted per case)
  draftMessage?: string
  draftFormat?: MessageFormat
  useVariables?: boolean   // Enable variable substitution ({{timestamp}}, etc.)
  // Script configuration
  scriptConfig?: ScriptConfig
  localVariables?: Variables
  postRecvSample?: string  // Sample message for post-recv dry run
}

// Saved Case (persisted format)
export interface SavedCase {
  id: string
  name: string
  protocol: Protocol
  host: string
  port: number
  createdAt: string
  updatedAt: string
  order: number  // Display order
  messageTemplates?: MessageTemplate[]
  framing?: FramingConfig
  charset?: CharsetMode
  connectionSettings?: ConnectionSettings
  draftMessage?: string
  draftFormat?: MessageFormat
  useVariables?: boolean
  notes?: string
  // Script configuration
  scriptConfig?: ScriptConfig
  localVariables?: Variables
  postRecvSample?: string  // Sample message for post-recv dry run
}

// Framing metadata from backend
export interface FramingMeta {
  mode: string
  frameHeader?: string    // hex representation
  frameFooter?: string    // hex representation
  payloadSize: number
  totalSize: number
  settings?: string       // e.g., "4 bytes, Big Endian"
}

// System message type
export type SystemMessageType = 'connected' | 'disconnected' | 'error' | 'script'

// Disconnect reason
export type DisconnectReason = 'user' | 'server' | 'error'

// Connection status info from backend
export interface ConnectionStatusInfo {
  status: 'connected' | 'disconnected'
  localAddr?: string
  remoteAddr?: string
  protocol?: string
  reason?: DisconnectReason
  duration?: number  // milliseconds
  bytesSent?: number
  bytesRecv?: number
}

// Message interface
export interface Message {
  id: string
  direction: MessageDirection
  content: string
  format: MessageFormat
  timestamp: Date
  rawBytes?: string         // base64 encoded
  size?: number             // byte count
  localAddr?: string        // local "host:port"
  remoteAddr?: string       // remote "host:port"
  framingInfo?: FramingMeta
  systemType?: SystemMessageType  // for system messages
  // For system messages (connection/disconnection)
  protocol?: string              // 'tcp' or 'udp'
  reason?: DisconnectReason      // disconnect reason
  duration?: number              // connection duration in ms
  bytesSent?: number             // total bytes sent
  bytesRecv?: number             // total bytes received
}

// Message template (frequently used messages)
export interface MessageTemplate {
  id: string
  name: string
  content: string
  format: MessageFormat
}

// App settings
export interface AppSettings {
  theme: 'dark' | 'light'
  fontSize: number
  showTimestamp: boolean
  autoScroll: boolean
  maxMessages: number
  defaultCharset: Charset
  connectionSettings: ConnectionSettings
}

// Default app settings
export const DEFAULT_APP_SETTINGS: AppSettings = {
  theme: 'dark',
  fontSize: 14,
  showTimestamp: true,
  autoScroll: true,
  maxMessages: 1000,
  defaultCharset: 'utf-8',
  connectionSettings: { ...DEFAULT_CONNECTION_SETTINGS }
}

// Socket event types
export interface SocketEvent {
  caseId: string
  type: 'connected' | 'disconnected' | 'data' | 'error'
  data?: string
  error?: string
  timestamp: Date
}

// ========== Echo Server Types ==========

// Echo protocol type
export type EchoProtocol = 'tcp' | 'udp'

// Echo log entry
export interface EchoLogEntry {
  id: string
  timestamp: number  // Unix milliseconds
  direction: 'recv' | 'echo'
  remoteAddr: string
  data: string       // hex encoded
  size: number
}

// Echo server status
export interface EchoStatus {
  running: boolean
  port: number
  protocol: EchoProtocol
  address: string
}

// Default echo port
export const DEFAULT_ECHO_PORT = 39876

// ========== Script Types ==========

// Variables dictionary
export interface Variables {
  [key: string]: string | number | boolean
}

// Script configuration
export interface ScriptConfig {
  setupScript: string
  setupEnabled: boolean
  preSendScript: string
  preSendEnabled: boolean
  postRecvScript: string
  postRecvEnabled: boolean
}

// Default script templates
export const DEFAULT_SETUP_SCRIPT = `// Setup script runs before variable substitution.
// Use wirey.set('key', value) to set variables.

// Example: Use lastSeq from previous response, increment for next request
// const lastSeq = wirey.collection.get('lastSeq') ?? 0;
// wirey.set('seq', lastSeq + 1);
`

export const DEFAULT_PRESEND_SCRIPT = `// Pre-send script runs after variable substitution.
// Return the message to send. Return null to cancel.

return msg;
`

export const DEFAULT_POSTRECV_SCRIPT = `// Post-recv script runs after receiving a message.
// Use this to extract values from the response.
// 'msg' contains the received message content.

// Example: Extract a sequence number from response
// const seq = msg.substring(3, 11);  // e.g., "OK 00001234 SUCCESS" -> "00001234"
// wirey.collection.set('lastSeq', parseInt(seq, 10));
`

export const DEFAULT_SCRIPT_CONFIG: ScriptConfig = {
  setupScript: DEFAULT_SETUP_SCRIPT,
  setupEnabled: false,
  preSendScript: DEFAULT_PRESEND_SCRIPT,
  preSendEnabled: false,
  postRecvScript: DEFAULT_POSTRECV_SCRIPT,
  postRecvEnabled: false
}
