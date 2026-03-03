/**
 * CodeMirror configuration for Wirey script editor and message input
 */
import { javascript } from '@codemirror/lang-javascript'
import { oneDark } from '@codemirror/theme-one-dark'
import { EditorView, Decoration, DecorationSet, ViewPlugin, ViewUpdate, keymap } from '@codemirror/view'
import { autocompletion, CompletionContext, type CompletionResult } from '@codemirror/autocomplete'
import { RangeSetBuilder } from '@codemirror/state'

// Light theme styling
export const lightTheme = EditorView.theme({
  '&': {
    backgroundColor: '#ffffff',
    color: '#1e293b'
  },
  '.cm-content': {
    caretColor: '#0ea5e9'
  },
  '.cm-cursor': {
    borderLeftColor: '#0ea5e9'
  },
  '&.cm-focused .cm-selectionBackground, .cm-selectionBackground': {
    backgroundColor: '#bae6fd'
  },
  '.cm-gutters': {
    backgroundColor: '#f8fafc',
    color: '#94a3b8',
    borderRight: '1px solid #e2e8f0'
  },
  '.cm-activeLineGutter': {
    backgroundColor: '#f1f5f9'
  },
  '.cm-activeLine': {
    boxShadow: 'inset 0 0 0 1000px rgba(241, 245, 249, 0.5)'
  }
}, { dark: false })

// Wirey API completions
const wireyCompletions = [
  { label: 'wirey', type: 'variable', detail: 'Wirey API object' },
  { label: 'wirey.get', type: 'function', detail: '(key) → value', info: 'Get case variable' },
  { label: 'wirey.set', type: 'function', detail: '(key, value)', info: 'Set case variable' },
  { label: 'wirey.collection.get', type: 'function', detail: '(key) → value', info: 'Get collection variable (persisted)' },
  { label: 'wirey.collection.set', type: 'function', detail: '(key, value)', info: 'Set collection variable (persisted)' },
  { label: 'wirey.uuid', type: 'function', detail: '() → string', info: 'Generate UUID v4' },
  { label: 'wirey.log', type: 'function', detail: '(...args)', info: 'Log to message panel' },
  { label: 'wirey.randomHex', type: 'function', detail: '(n?) → string', info: 'Random hex string (n bytes)' },

  // Byte manipulation
  { label: 'wirey.toHex', type: 'function', detail: '(str) → string', info: 'String to hex "48 65 6C"' },
  { label: 'wirey.fromHex', type: 'function', detail: '(hex) → string', info: 'Hex "48 65 6C" to string' },
  { label: 'wirey.toBytes', type: 'function', detail: '(str) → byte[]', info: 'String to byte array' },
  { label: 'wirey.fromBytes', type: 'function', detail: '(bytes) → string', info: 'Byte array to string' },
  { label: 'wirey.subBytes', type: 'function', detail: '(str, start, end?) → string', info: 'Extract bytes as string' },
  { label: 'wirey.appendBytes', type: 'function', detail: '(str, bytes) → string', info: 'Append bytes to string' },
  { label: 'wirey.replaceBytes', type: 'function', detail: '(str, start, bytes) → string', info: 'Replace bytes at position' },
  { label: 'wirey.byteAt', type: 'function', detail: '(str, index) → number', info: 'Get byte value at position' },
  { label: 'wirey.setByteAt', type: 'function', detail: '(str, index, value) → string', info: 'Set byte value at position' },

  // HTTP
  { label: 'wirey.httpGet', type: 'function', detail: '(url) → {status, body, error}', info: 'HTTP GET request (10s timeout)' },
  { label: 'wirey.httpPost', type: 'function', detail: '(url, body, contentType?) → {status, body, error}', info: 'HTTP POST request (10s timeout)' },
]

// ES6 built-in completions (Goja runtime compatible)
const es6Completions = [
  // JSON
  { label: 'JSON.stringify', type: 'function', detail: '(value, replacer?, space?)' },
  { label: 'JSON.parse', type: 'function', detail: '(text, reviver?)' },

  // Object
  { label: 'Object.keys', type: 'function', detail: '(obj) → string[]' },
  { label: 'Object.values', type: 'function', detail: '(obj) → any[]' },
  { label: 'Object.entries', type: 'function', detail: '(obj) → [key, value][]' },
  { label: 'Object.assign', type: 'function', detail: '(target, ...sources)' },
  { label: 'Object.freeze', type: 'function', detail: '(obj)' },

  // Array
  { label: 'Array.isArray', type: 'function', detail: '(value) → boolean' },
  { label: 'Array.from', type: 'function', detail: '(iterable, mapFn?)' },
  { label: 'Array.of', type: 'function', detail: '(...items)' },

  // String
  { label: 'String.fromCharCode', type: 'function', detail: '(...codes) → string' },
  { label: 'String.fromCodePoint', type: 'function', detail: '(...codePoints) → string' },

  // Number
  { label: 'Number.parseInt', type: 'function', detail: '(string, radix?)' },
  { label: 'Number.parseFloat', type: 'function', detail: '(string)' },
  { label: 'Number.isNaN', type: 'function', detail: '(value) → boolean' },
  { label: 'Number.isFinite', type: 'function', detail: '(value) → boolean' },
  { label: 'Number.isInteger', type: 'function', detail: '(value) → boolean' },

  // Math
  { label: 'Math.floor', type: 'function', detail: '(x) → number' },
  { label: 'Math.ceil', type: 'function', detail: '(x) → number' },
  { label: 'Math.round', type: 'function', detail: '(x) → number' },
  { label: 'Math.abs', type: 'function', detail: '(x) → number' },
  { label: 'Math.min', type: 'function', detail: '(...values) → number' },
  { label: 'Math.max', type: 'function', detail: '(...values) → number' },
  { label: 'Math.random', type: 'function', detail: '() → number' },
  { label: 'Math.pow', type: 'function', detail: '(base, exp) → number' },
  { label: 'Math.sqrt', type: 'function', detail: '(x) → number' },

  // Global functions
  { label: 'parseInt', type: 'function', detail: '(string, radix?) → number' },
  { label: 'parseFloat', type: 'function', detail: '(string) → number' },
  { label: 'isNaN', type: 'function', detail: '(value) → boolean' },
  { label: 'isFinite', type: 'function', detail: '(value) → boolean' },
  { label: 'encodeURI', type: 'function', detail: '(uri) → string' },
  { label: 'decodeURI', type: 'function', detail: '(uri) → string' },
  { label: 'encodeURIComponent', type: 'function', detail: '(str) → string' },
  { label: 'decodeURIComponent', type: 'function', detail: '(str) → string' },
  { label: 'atob', type: 'function', detail: '(base64) → string' },
  { label: 'btoa', type: 'function', detail: '(string) → base64' },

  // Date
  { label: 'Date.now', type: 'function', detail: '() → number' },
  { label: 'Date.parse', type: 'function', detail: '(dateString) → number' },

  // Promise
  { label: 'Promise.resolve', type: 'function', detail: '(value)' },
  { label: 'Promise.reject', type: 'function', detail: '(reason)' },
  { label: 'Promise.all', type: 'function', detail: '(promises)' },
  { label: 'Promise.race', type: 'function', detail: '(promises)' },

  // Common string methods (instance methods shown as reminders)
  { label: '.split', type: 'method', detail: '(separator) → string[]' },
  { label: '.join', type: 'method', detail: '(separator) → string' },
  { label: '.slice', type: 'method', detail: '(start, end?) → string|array' },
  { label: '.substring', type: 'method', detail: '(start, end?) → string' },
  { label: '.substr', type: 'method', detail: '(start, length?) → string' },
  { label: '.replace', type: 'method', detail: '(search, replace) → string' },
  { label: '.trim', type: 'method', detail: '() → string' },
  { label: '.toLowerCase', type: 'method', detail: '() → string' },
  { label: '.toUpperCase', type: 'method', detail: '() → string' },
  { label: '.indexOf', type: 'method', detail: '(search) → number' },
  { label: '.lastIndexOf', type: 'method', detail: '(search) → number' },
  { label: '.includes', type: 'method', detail: '(search) → boolean' },
  { label: '.startsWith', type: 'method', detail: '(search) → boolean' },
  { label: '.endsWith', type: 'method', detail: '(search) → boolean' },
  { label: '.padStart', type: 'method', detail: '(length, pad?) → string' },
  { label: '.padEnd', type: 'method', detail: '(length, pad?) → string' },
  { label: '.charCodeAt', type: 'method', detail: '(index) → number' },
  { label: '.charAt', type: 'method', detail: '(index) → string' },

  // Common array methods
  { label: '.map', type: 'method', detail: '(fn) → array' },
  { label: '.filter', type: 'method', detail: '(fn) → array' },
  { label: '.reduce', type: 'method', detail: '(fn, init?) → any' },
  { label: '.forEach', type: 'method', detail: '(fn)' },
  { label: '.find', type: 'method', detail: '(fn) → element' },
  { label: '.findIndex', type: 'method', detail: '(fn) → number' },
  { label: '.some', type: 'method', detail: '(fn) → boolean' },
  { label: '.every', type: 'method', detail: '(fn) → boolean' },
  { label: '.push', type: 'method', detail: '(...items) → length' },
  { label: '.pop', type: 'method', detail: '() → element' },
  { label: '.shift', type: 'method', detail: '() → element' },
  { label: '.unshift', type: 'method', detail: '(...items) → length' },
  { label: '.concat', type: 'method', detail: '(...arrays) → array' },
  { label: '.reverse', type: 'method', detail: '() → array' },
  { label: '.sort', type: 'method', detail: '(compareFn?) → array' },
  { label: '.length', type: 'property', detail: 'number' },
]

// All completions combined
const allCompletions = [...wireyCompletions, ...es6Completions]

// Custom completion function
function scriptCompletion(context: CompletionContext): CompletionResult | null {
  const word = context.matchBefore(/[\w.]*/)
  if (!word || (word.from === word.to && !context.explicit)) return null

  const text = word.text.toLowerCase()

  // Filter completions based on input
  const filtered = allCompletions.filter(c =>
    c.label.toLowerCase().startsWith(text) ||
    c.label.toLowerCase().includes(text)
  )

  if (filtered.length === 0) return null

  return {
    from: word.from,
    options: filtered,
    validFor: /^[\w.]*$/
  }
}

/**
 * Get CodeMirror extensions for script editor
 * @param isDark - Whether to use dark theme
 */
export function getScriptExtensions(isDark: boolean) {
  const jsLang = javascript()
  const baseExtensions = [
    jsLang,
    autocompletion({ activateOnTyping: true }),
    jsLang.language.data.of({ autocomplete: scriptCompletion })
  ]

  if (isDark) {
    return [...baseExtensions, oneDark]
  }
  return [...baseExtensions, lightTheme]
}

// ========== Message Input Extensions ==========

// Decoration marks for variables and escape sequences
const variableMark = Decoration.mark({ class: 'cm-variable-highlight' })
const escapeMark = Decoration.mark({ class: 'cm-escape-highlight' })

// Plugin to highlight {{variables}} and \escape sequences
const messageHighlighter = ViewPlugin.fromClass(class {
  decorations: DecorationSet

  constructor(view: EditorView) {
    this.decorations = this.buildDecorations(view)
  }

  update(update: ViewUpdate) {
    if (update.docChanged || update.viewportChanged) {
      this.decorations = this.buildDecorations(update.view)
    }
  }

  buildDecorations(view: EditorView): DecorationSet {
    const builder = new RangeSetBuilder<Decoration>()
    const text = view.state.doc.toString()

    // Find {{variables}}
    const varRegex = /\{\{[^}]+\}\}/g
    let match
    while ((match = varRegex.exec(text)) !== null) {
      builder.add(match.index, match.index + match[0].length, variableMark)
    }

    // Find escape sequences: \n, \r, \t, \0, \\, \xNN
    const escRegex = /\\[nrt0\\]|\\x[0-9A-Fa-f]{2}/g
    while ((match = escRegex.exec(text)) !== null) {
      builder.add(match.index, match.index + match[0].length, escapeMark)
    }

    return builder.finish()
  }
}, {
  decorations: v => v.decorations
})

// Message input theme (dark) - matches original textarea design
const messageInputDarkTheme = EditorView.theme({
  '&': {
    backgroundColor: 'transparent',
    color: '#e4e4e7',
    fontSize: '14px',
    fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace',
    height: '100%'
  },
  '.cm-content': {
    caretColor: '#e4e4e7',
    padding: '12px 16px',
    minHeight: '100%'
  },
  '.cm-cursor': {
    borderLeftColor: '#e4e4e7',
    borderLeftWidth: '2px'
  },
  '&.cm-focused': {
    outline: 'none'
  },
  // Selection background - darker for dark theme
  '&.cm-focused .cm-selectionBackground, .cm-selectionBackground': {
    backgroundColor: 'rgba(99, 102, 241, 0.3) !important'
  },
  '.cm-selectionMatch': {
    backgroundColor: 'rgba(99, 102, 241, 0.2) !important'
  },
  // Active line - no highlight
  '.cm-activeLine': {
    backgroundColor: 'transparent'
  },
  '.cm-line': {
    padding: '0'
  },
  '.cm-placeholder': {
    color: '#71717a',
    fontFamily: 'inherit'
  },
  '.cm-scroller': {
    overflow: 'auto',
    fontFamily: 'inherit'
  },
  '.cm-gutters': {
    display: 'none'
  },
  // Variable highlighting
  '.cm-variable-highlight': {
    color: 'rgb(251, 191, 36)',
    backgroundColor: 'rgba(245, 158, 11, 0.25)',
    borderRadius: '2px',
    boxShadow: '0 0 0 2px rgba(245, 158, 11, 0.15)'
  },
  // Escape sequence highlighting
  '.cm-escape-highlight': {
    color: 'rgb(34, 211, 238)',
    backgroundColor: 'rgba(6, 182, 212, 0.25)',
    borderRadius: '2px',
    boxShadow: '0 0 0 2px rgba(6, 182, 212, 0.15)'
  }
}, { dark: true })

// Message input theme (light) - matches original textarea design
const messageInputLightTheme = EditorView.theme({
  '&': {
    backgroundColor: 'transparent',
    color: '#1e293b',
    fontSize: '14px',
    fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace',
    height: '100%'
  },
  '.cm-content': {
    caretColor: '#1e293b',
    padding: '12px 16px',
    minHeight: '100%'
  },
  '.cm-cursor': {
    borderLeftColor: '#1e293b',
    borderLeftWidth: '2px'
  },
  '&.cm-focused': {
    outline: 'none'
  },
  '&.cm-focused .cm-selectionBackground, .cm-selectionBackground': {
    backgroundColor: 'rgba(99, 102, 241, 0.2) !important'
  },
  '.cm-selectionMatch': {
    backgroundColor: 'rgba(99, 102, 241, 0.15) !important'
  },
  '.cm-activeLine': {
    backgroundColor: 'transparent'
  },
  '.cm-line': {
    padding: '0'
  },
  '.cm-placeholder': {
    color: '#a1a1aa',
    fontFamily: 'inherit'
  },
  '.cm-scroller': {
    overflow: 'auto',
    fontFamily: 'inherit'
  },
  '.cm-gutters': {
    display: 'none'
  },
  // Variable highlighting
  '.cm-variable-highlight': {
    color: 'rgb(217, 119, 6)',
    backgroundColor: 'rgba(245, 158, 11, 0.2)',
    borderRadius: '2px'
  },
  // Escape sequence highlighting
  '.cm-escape-highlight': {
    color: 'rgb(8, 145, 178)',
    backgroundColor: 'rgba(6, 182, 212, 0.2)',
    borderRadius: '2px'
  }
}, { dark: false })

// Variable completions for message input
const variableCompletions = [
  { label: '{{timestamp}}', type: 'variable', detail: 'Unix timestamp (seconds)' },
  { label: '{{timestamp_ms}}', type: 'variable', detail: 'Unix timestamp (milliseconds)' },
  { label: '{{datetime}}', type: 'variable', detail: 'ISO 8601 format' },
  { label: '{{date}}', type: 'variable', detail: 'Date only (YYYY-MM-DD)' },
  { label: '{{time}}', type: 'variable', detail: 'Time only (HH:MM:SS)' },
  { label: '{{uuid}}', type: 'variable', detail: 'UUID v4' },
  { label: '{{random}}', type: 'variable', detail: 'Random hex (4 bytes)' },
  { label: '{{random:8}}', type: 'variable', detail: 'Random hex (8 bytes)' },
  { label: '{{counter}}', type: 'variable', detail: 'Auto-increment from 1' },
  { label: '{{counter:100}}', type: 'variable', detail: 'Auto-increment from 100' },
  { label: '\\n', type: 'keyword', detail: 'Line Feed (0x0A)' },
  { label: '\\r', type: 'keyword', detail: 'Carriage Return (0x0D)' },
  { label: '\\t', type: 'keyword', detail: 'Tab (0x09)' },
  { label: '\\0', type: 'keyword', detail: 'NULL (0x00)' },
  { label: '\\\\', type: 'keyword', detail: 'Backslash (0x5C)' },
  { label: '\\x00', type: 'keyword', detail: 'Hex byte' },
]

function messageCompletion(context: CompletionContext): CompletionResult | null {
  // Match {{ for variables or \ for escapes
  const beforeVar = context.matchBefore(/\{\{[\w:]*/)
  const beforeEsc = context.matchBefore(/\\[xnrt0\\]?[0-9A-Fa-f]*/)

  if (beforeVar) {
    return {
      from: beforeVar.from,
      options: variableCompletions.filter(c => c.label.startsWith('{{')),
      validFor: /^\{\{[\w:]*$/
    }
  }

  if (beforeEsc) {
    return {
      from: beforeEsc.from,
      options: variableCompletions.filter(c => c.label.startsWith('\\')),
      validFor: /^\\[xnrt0\\]?[0-9A-Fa-f]*$/
    }
  }

  return null
}

/**
 * Get CodeMirror extensions for message input
 * @param isDark - Whether to use dark theme
 * @param enableHighlight - Whether to enable variable/escape highlighting
 */
export function getMessageInputExtensions(isDark: boolean, enableHighlight: boolean = true) {
  const baseExtensions = [
    EditorView.lineWrapping,
    autocompletion({ activateOnTyping: true, override: [messageCompletion] }),
  ]

  if (enableHighlight) {
    baseExtensions.push(messageHighlighter)
  }

  if (isDark) {
    return [...baseExtensions, messageInputDarkTheme]
  }
  return [...baseExtensions, messageInputLightTheme]
}
