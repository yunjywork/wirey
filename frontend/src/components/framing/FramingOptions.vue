<script setup lang="ts">
import { computed, ref } from 'vue'
import type { FramingConfig, FramingMode, LengthEncoding, Endian, LengthPrefixMode, PaddingPosition } from '@/types'
import { formatBytesHexDump } from '@/utils/hexdump'

const props = defineProps<{
  modelValue: FramingConfig
  disabled?: boolean
  showCollectionOption?: boolean  // Show "Collection Default" option
  collectionFraming?: FramingConfig  // For preview when mode is 'collection'
}>()

const emit = defineEmits<{
  'update:modelValue': [value: FramingConfig]
}>()

// Custom delimiter input mode (ascii or hex)
const customDelimiterMode = ref<'ascii' | 'hex'>('ascii')

function updateFraming(updates: Partial<FramingConfig>) {
  emit('update:modelValue', { ...props.modelValue, ...updates })
}

// Computed for two-way binding
const framingMode = computed({
  get: () => props.modelValue.mode || 'none',
  set: (value: FramingMode) => updateFraming({ mode: value })
})

const delimiter = computed({
  get: () => props.modelValue.delimiter || '\\n',
  set: (value: string) => updateFraming({ delimiter: value })
})

const lengthEncoding = computed({
  get: () => props.modelValue.lengthEncoding || 'binary',
  set: (value: LengthEncoding) => updateFraming({ lengthEncoding: value })
})

const lengthOffset = computed({
  get: () => props.modelValue.lengthOffset || 0,
  set: (value: number) => updateFraming({ lengthOffset: value })
})

const lengthBytes = computed({
  get: () => props.modelValue.lengthBytes || 4,
  set: (value: number) => updateFraming({ lengthBytes: value })
})

// Track if custom length bytes is selected
const isCustomLengthBytes = computed(() => {
  const bytes = props.modelValue.lengthBytes || 4
  return ![1, 2, 4, 8].includes(bytes)
})

// Length bytes select value (for UI)
const lengthBytesSelect = computed({
  get: () => {
    const bytes = props.modelValue.lengthBytes || 4
    if ([1, 2, 4, 8].includes(bytes)) {
      return bytes
    }
    return 'custom'
  },
  set: (value: number | 'custom') => {
    if (value === 'custom') {
      updateFraming({ lengthBytes: 3 })
    } else {
      updateFraming({ lengthBytes: value })
    }
  }
})

const endian = computed({
  get: () => props.modelValue.endian || 'big',
  set: (value: Endian) => updateFraming({ endian: value })
})

const includeHeader = computed({
  get: () => props.modelValue.includeHeader || false,
  set: (value: boolean) => updateFraming({ includeHeader: value })
})

const lengthMode = computed({
  get: () => props.modelValue.lengthMode || 'append',
  set: (value: LengthPrefixMode) => updateFraming({ lengthMode: value })
})

const fixedSize = computed({
  get: () => props.modelValue.fixedSize ?? 256,
  set: (value: number) => updateFraming({ fixedSize: value || 0 })
})

const paddingPosition = computed({
  get: () => props.modelValue.paddingPosition || 'right',
  set: (value: PaddingPosition) => updateFraming({ paddingPosition: value })
})

const paddingByte = computed({
  get: () => props.modelValue.paddingByte ?? 0,
  set: (value: number) => updateFraming({ paddingByte: value })
})

// Padding byte as hex string
const paddingByteHex = computed({
  get: () => (paddingByte.value).toString(16).toUpperCase().padStart(2, '0'),
  set: (value: string) => {
    const byte = parseInt(value, 16)
    if (!isNaN(byte) && byte >= 0 && byte <= 255) {
      updateFraming({ paddingByte: byte })
    }
  }
})

function filterHexOnly(event: Event) {
  const input = event.target as HTMLInputElement
  input.value = input.value.replace(/[^0-9A-Fa-f]/g, '').toUpperCase()
}

const delimiterOptions = [
  { value: '\\n', label: '\\n (LF)' },
  { value: '\\r\\n', label: '\\r\\n (CRLF)' },
  { value: '\\0', label: '\\0 (NULL)' },
  { value: 'custom', label: 'Custom' }
]

const presetDelimiters = ['\\n', '\\r\\n', '\\0']

const isCustomDelimiter = computed(() => {
  return !presetDelimiters.includes(delimiter.value)
})

// Delimiter select value (for UI - separate from actual value)
const delimiterSelect = computed({
  get: () => {
    if (presetDelimiters.includes(delimiter.value)) {
      return delimiter.value
    }
    return 'custom'
  },
  set: (value: string) => {
    if (value === 'custom') {
      updateFraming({ delimiter: 'END' })
    } else {
      updateFraming({ delimiter: value })
    }
  }
})

// Custom delimiter as ASCII string
const customDelimiterAscii = computed({
  get: () => {
    if (!isCustomDelimiter.value) return ''
    return delimiter.value
  },
  set: (value: string) => {
    updateFraming({ delimiter: value })
  }
})

// Convert hex string to actual bytes string
function hexToString(hex: string): string {
  const clean = hex.replace(/\s/g, '')
  if (clean.length % 2 !== 0) return ''
  let result = ''
  for (let i = 0; i < clean.length; i += 2) {
    const byte = parseInt(clean.substring(i, i + 2), 16)
    if (isNaN(byte)) return ''
    result += String.fromCharCode(byte)
  }
  return result
}

// Convert string to hex display
function stringToHex(str: string): string {
  return str.split('').map(c => c.charCodeAt(0).toString(16).toUpperCase().padStart(2, '0')).join(' ')
}

// Custom delimiter as hex string
const customDelimiterHex = computed({
  get: () => {
    if (!isCustomDelimiter.value) return ''
    return stringToHex(delimiter.value)
  },
  set: (value: string) => {
    const str = hexToString(value)
    if (str) {
      updateFraming({ delimiter: str })
    }
  }
})

// Filter hex input to only allow valid hex characters
function filterHexInput(event: Event) {
  const input = event.target as HTMLInputElement
  input.value = input.value.replace(/[^0-9A-Fa-f\s]/g, '').toUpperCase()
}

// Format framing mode for display
function formatFramingMode(mode: string | undefined): string {
  switch (mode) {
    case 'none': return 'None (Raw)'
    case 'delimiter': return 'Delimiter'
    case 'length-prefix': return 'Length Prefix'
    case 'fixed-length': return 'Fixed Length'
    default: return 'None'
  }
}

// ============ Preview Logic ============
const SAMPLE_TEXT = 'Hello World! This is sample.data'

function parseDelimiter(delim: string): number[] {
  switch (delim) {
    case '\\n': return [0x0A]
    case '\\r\\n': return [0x0D, 0x0A]
    case '\\0': return [0x00]
    default:
      return delim.split('').map(c => c.charCodeAt(0))
  }
}

function encodeLengthBinary(length: number, numBytes: number, endianVal: string): number[] {
  const bytes: number[] = []
  for (let i = 0; i < numBytes; i++) {
    bytes.push((length >> (8 * i)) & 0xFF)
  }
  return endianVal === 'big' ? bytes.reverse() : bytes
}

function encodeLengthAscii(length: number, numBytes: number): number[] {
  const str = length.toString().padStart(numBytes, '0')
  return str.split('').map(c => c.charCodeAt(0))
}

const samplePreview = computed(() => {
  const mode = framingMode.value
  const config = props.modelValue

  // Get effective config (for 'collection' mode, use collection's framing)
  let effectiveMode = mode
  let effectiveConfig = config
  if (mode === 'collection' && props.collectionFraming) {
    effectiveMode = props.collectionFraming.mode || 'none'
    effectiveConfig = props.collectionFraming
  }

  const sampleBytes = SAMPLE_TEXT.split('').map(c => c.charCodeAt(0))
  let headerBytes: number[] = []
  let footerBytes: number[] = []
  let payloadBytes = [...sampleBytes]

  switch (effectiveMode) {
    case 'delimiter': {
      const delim = effectiveConfig.delimiter || '\\n'
      footerBytes = parseDelimiter(delim)
      break
    }
    case 'length-prefix': {
      const lenBytes = effectiveConfig.lengthBytes || 4
      const encoding = effectiveConfig.lengthEncoding || 'binary'
      const endianVal = effectiveConfig.endian || 'big'
      const includeHdr = effectiveConfig.includeHeader || false
      const lpMode = effectiveConfig.lengthMode || 'append'
      const offset = effectiveConfig.lengthOffset || 0

      if (lpMode === 'rewrite') {
        const endPos = offset + lenBytes
        if (sampleBytes.length >= endPos) {
          let lengthValue = sampleBytes.length - endPos
          if (includeHdr) {
            lengthValue = sampleBytes.length
          }

          let lengthField: number[]
          if (encoding === 'binary') {
            lengthField = encodeLengthBinary(lengthValue, lenBytes, endianVal)
          } else if (encoding === 'ascii') {
            lengthField = encodeLengthAscii(lengthValue, lenBytes)
          } else if (encoding === 'hex') {
            const hexStr = lengthValue.toString(16).toUpperCase().padStart(lenBytes, '0')
            lengthField = hexStr.split('').map(c => c.charCodeAt(0))
          } else {
            lengthField = encodeLengthAscii(lengthValue, lenBytes)
          }

          payloadBytes = [
            ...sampleBytes.slice(0, offset),
            ...lengthField,
            ...sampleBytes.slice(endPos)
          ]
          headerBytes = new Array(lenBytes).fill(0)
          const rewriteOffset = offset

          const framedBytes = [...payloadBytes, ...footerBytes]
          const hexDumpHtml = formatBytesHexDump(framedBytes, lenBytes, 0, rewriteOffset)

          return {
            headerLen: lenBytes,
            footerLen: 0,
            totalLen: framedBytes.length,
            displayLen: framedBytes.length,
            isTruncated: false,
            hexDumpHtml
          }
        }
      } else {
        let lengthValue = sampleBytes.length
        if (includeHdr) {
          lengthValue += lenBytes
        }

        if (encoding === 'binary') {
          headerBytes = encodeLengthBinary(lengthValue, lenBytes, endianVal)
        } else if (encoding === 'ascii') {
          headerBytes = encodeLengthAscii(lengthValue, lenBytes)
        } else if (encoding === 'hex') {
          const hexStr = lengthValue.toString(16).toUpperCase().padStart(lenBytes, '0')
          headerBytes = hexStr.split('').map(c => c.charCodeAt(0))
        } else {
          headerBytes = encodeLengthAscii(lengthValue, lenBytes)
        }
      }
      break
    }
    case 'fixed-length': {
      const actualSize = effectiveConfig.fixedSize ?? 256
      const previewLimit = 64
      const displayLen = Math.min(actualSize, previewLimit)
      const isTruncated = actualSize > previewLimit
      const padPosition = effectiveConfig.paddingPosition || 'right'
      const padByte = effectiveConfig.paddingByte ?? 0

      if (displayLen > sampleBytes.length) {
        const paddingLen = displayLen - sampleBytes.length
        const padding = new Array(paddingLen).fill(padByte)
        if (padPosition === 'left') {
          payloadBytes = [...padding, ...sampleBytes]
        } else {
          payloadBytes = [...sampleBytes, ...padding]
        }
      } else {
        payloadBytes = sampleBytes.slice(0, displayLen)
      }

      const framedBytes = [...headerBytes, ...payloadBytes, ...footerBytes]
      const hexDumpHtml = formatBytesHexDump(framedBytes, headerBytes.length, footerBytes.length)

      return {
        headerLen: headerBytes.length,
        footerLen: footerBytes.length,
        totalLen: actualSize,
        displayLen: framedBytes.length,
        isTruncated,
        hexDumpHtml
      }
    }
    default:
      break
  }

  const framedBytes = [...headerBytes, ...payloadBytes, ...footerBytes]
  const hexDumpHtml = formatBytesHexDump(framedBytes, headerBytes.length, footerBytes.length)

  return {
    headerLen: headerBytes.length,
    footerLen: footerBytes.length,
    totalLen: framedBytes.length,
    displayLen: framedBytes.length,
    isTruncated: false,
    hexDumpHtml
  }
})

// Should show preview
const showPreview = computed(() => {
  if (framingMode.value === 'collection') {
    return props.collectionFraming && props.collectionFraming.mode !== 'collection'
  }
  return true
})
</script>

<template>
  <div class="space-y-6">
    <!-- Framing Mode Selection -->
    <div>
      <h3 class="text-sm font-medium text-zinc-300 mb-3">Framing Mode</h3>
      <div class="space-y-2">
        <!-- Collection Default (optional) -->
        <label
          v-if="showCollectionOption"
          :class="[
            'flex items-center gap-3 py-1.5 px-2 rounded-lg transition-colors',
            framingMode === 'collection' ? 'bg-accent-primary/10' : '',
            disabled ? 'cursor-not-allowed opacity-70' : 'cursor-pointer hover:bg-bg-tertiary'
          ]"
        >
          <input
            type="radio"
            v-model="framingMode"
            value="collection"
            :disabled="disabled"
            class="w-4 h-4 text-accent-primary bg-bg-tertiary border-border-custom focus:ring-accent-primary/50"
          />
          <span :class="['text-sm', framingMode === 'collection' ? 'text-accent-primary font-medium' : 'text-zinc-300']">Collection Default</span>
          <span class="text-xs text-zinc-500">- Use collection's shared framing</span>
        </label>

        <label :class="[
          'flex items-center gap-3 py-1.5 px-2 rounded-lg transition-colors',
          framingMode === 'none' ? 'bg-accent-primary/10' : '',
          disabled ? 'cursor-not-allowed opacity-70' : 'cursor-pointer hover:bg-bg-tertiary'
        ]">
          <input
            type="radio"
            v-model="framingMode"
            value="none"
            :disabled="disabled"
            class="w-4 h-4 text-accent-primary bg-bg-tertiary border-border-custom focus:ring-accent-primary/50"
          />
          <span :class="['text-sm', framingMode === 'none' ? 'text-accent-primary font-medium' : 'text-zinc-300']">None (Raw)</span>
          <span class="text-xs text-zinc-500">- No framing, send/receive as-is</span>
        </label>

        <label :class="[
          'flex items-center gap-3 py-1.5 px-2 rounded-lg transition-colors',
          framingMode === 'delimiter' ? 'bg-accent-primary/10' : '',
          disabled ? 'cursor-not-allowed opacity-70' : 'cursor-pointer hover:bg-bg-tertiary'
        ]">
          <input
            type="radio"
            v-model="framingMode"
            value="delimiter"
            :disabled="disabled"
            class="w-4 h-4 text-accent-primary bg-bg-tertiary border-border-custom focus:ring-accent-primary/50"
          />
          <span :class="['text-sm', framingMode === 'delimiter' ? 'text-accent-primary font-medium' : 'text-zinc-300']">Delimiter</span>
          <span class="text-xs text-zinc-500">- Split by delimiter character</span>
        </label>

        <label :class="[
          'flex items-center gap-3 py-1.5 px-2 rounded-lg transition-colors',
          framingMode === 'length-prefix' ? 'bg-accent-primary/10' : '',
          disabled ? 'cursor-not-allowed opacity-70' : 'cursor-pointer hover:bg-bg-tertiary'
        ]">
          <input
            type="radio"
            v-model="framingMode"
            value="length-prefix"
            :disabled="disabled"
            class="w-4 h-4 text-accent-primary bg-bg-tertiary border-border-custom focus:ring-accent-primary/50"
          />
          <span :class="['text-sm', framingMode === 'length-prefix' ? 'text-accent-primary font-medium' : 'text-zinc-300']">Length Prefix</span>
          <span class="text-xs text-zinc-500">- Length field before data</span>
        </label>

        <label :class="[
          'flex items-center gap-3 py-1.5 px-2 rounded-lg transition-colors',
          framingMode === 'fixed-length' ? 'bg-accent-primary/10' : '',
          disabled ? 'cursor-not-allowed opacity-70' : 'cursor-pointer hover:bg-bg-tertiary'
        ]">
          <input
            type="radio"
            v-model="framingMode"
            value="fixed-length"
            :disabled="disabled"
            class="w-4 h-4 text-accent-primary bg-bg-tertiary border-border-custom focus:ring-accent-primary/50"
          />
          <span :class="['text-sm', framingMode === 'fixed-length' ? 'text-accent-primary font-medium' : 'text-zinc-300']">Fixed Length</span>
          <span class="text-xs text-zinc-500">- Fixed byte size per message</span>
        </label>
      </div>
    </div>

    <!-- Delimiter Options -->
    <div
      v-if="framingMode === 'delimiter'"
      class="p-4 bg-bg-tertiary rounded-lg border border-border-custom space-y-4"
    >
      <h4 class="text-sm font-medium text-zinc-300">Delimiter Options</h4>

      <div class="flex items-center gap-4">
        <label class="text-xs text-zinc-400">Delimiter:</label>
        <select
          v-model="delimiterSelect"
          :disabled="disabled"
          class="input w-32"
        >
          <option v-for="opt in delimiterOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>

        <template v-if="isCustomDelimiter">
          <select v-model="customDelimiterMode" :disabled="disabled" class="input w-20">
            <option value="ascii">ASCII</option>
            <option value="hex">HEX</option>
          </select>
          <input
            v-if="customDelimiterMode === 'ascii'"
            v-model="customDelimiterAscii"
            :disabled="disabled"
            type="text"
            placeholder="e.g. END"
            class="input flex-1 font-mono"
          />
          <input
            v-else
            v-model="customDelimiterHex"
            @input="filterHexInput"
            :disabled="disabled"
            type="text"
            placeholder="e.g. 0D 0A"
            class="input flex-1 font-mono"
          />
        </template>
      </div>
    </div>

    <!-- Length Prefix Options -->
    <div
      v-if="framingMode === 'length-prefix'"
      class="p-4 bg-bg-tertiary rounded-lg border border-border-custom space-y-4"
    >
      <h4 class="text-sm font-medium text-zinc-300">Length Prefix Options</h4>

      <div class="grid grid-cols-2 gap-4">
        <!-- Encoding -->
        <div>
          <label class="block text-xs text-zinc-400 mb-1.5">Encoding</label>
          <select v-model="lengthEncoding" :disabled="disabled" class="input w-full">
            <option value="binary">Binary</option>
            <option value="ascii">ASCII</option>
            <option value="hex">HEX</option>
            <option value="bcd">BCD</option>
          </select>
        </div>

        <!-- Length Field Size -->
        <div>
          <label class="block text-xs text-zinc-400 mb-1.5">Length Field</label>
          <div class="flex gap-2">
            <select v-model="lengthBytesSelect" :disabled="disabled" class="input flex-1">
              <option :value="1">1 byte</option>
              <option :value="2">2 bytes</option>
              <option :value="4">4 bytes</option>
              <option v-if="lengthEncoding !== 'binary'" :value="8">8 bytes</option>
              <option value="custom">Custom</option>
            </select>
            <input
              v-if="isCustomLengthBytes || lengthBytesSelect === 'custom'"
              v-model.number="lengthBytes"
              :disabled="disabled"
              type="number"
              min="1"
              max="16"
              class="input w-16 font-mono"
            />
          </div>
        </div>

        <!-- Endian (only for binary) -->
        <div v-if="lengthEncoding === 'binary'">
          <label class="block text-xs text-zinc-400 mb-1.5">Endian</label>
          <select v-model="endian" :disabled="disabled" class="input w-full">
            <option value="big">Big Endian</option>
            <option value="little">Little Endian</option>
          </select>
        </div>
      </div>

      <div class="flex items-center gap-6 pt-2">
        <!-- Include Header -->
        <label :class="['flex items-center gap-2', disabled ? 'cursor-not-allowed' : 'cursor-pointer']">
          <input
            type="checkbox"
            v-model="includeHeader"
            :disabled="disabled"
            class="w-4 h-4 text-accent-primary bg-bg-tertiary border-border-custom rounded focus:ring-accent-primary/50 disabled:opacity-50"
          />
          <span class="text-sm text-zinc-300">Include header in length</span>
        </label>
      </div>

      <div class="pt-2">
        <label class="block text-xs text-zinc-400 mb-2">Mode</label>
        <div class="flex gap-4">
          <label :class="['flex items-center gap-2', disabled ? 'cursor-not-allowed' : 'cursor-pointer']">
            <input
              type="radio"
              v-model="lengthMode"
              value="append"
              :disabled="disabled"
              class="w-4 h-4 text-accent-primary bg-bg-tertiary border-border-custom focus:ring-accent-primary/50 disabled:opacity-50"
            />
            <span class="text-sm text-zinc-300">Append</span>
            <span class="text-xs text-zinc-500">- Add length prefix on send</span>
          </label>
          <label :class="['flex items-center gap-2', disabled ? 'cursor-not-allowed' : 'cursor-pointer']">
            <input
              type="radio"
              v-model="lengthMode"
              value="rewrite"
              :disabled="disabled"
              class="w-4 h-4 text-accent-primary bg-bg-tertiary border-border-custom focus:ring-accent-primary/50 disabled:opacity-50"
            />
            <span class="text-sm text-zinc-300">Rewrite</span>
            <span class="text-xs text-zinc-500">- Replace existing length field</span>
          </label>
        </div>
      </div>

      <!-- Offset (only for rewrite mode) -->
      <div v-if="lengthMode === 'rewrite'" class="pt-2">
        <label class="block text-xs text-zinc-400 mb-1.5">Length Field Offset (bytes)</label>
        <input
          v-model.number="lengthOffset"
          :disabled="disabled"
          type="number"
          min="0"
          class="input w-32 font-mono"
        />
        <p class="text-xs text-zinc-500 mt-1">Position where length field starts in the message</p>
      </div>
    </div>

    <!-- Fixed Length Options -->
    <div
      v-if="framingMode === 'fixed-length'"
      class="p-4 bg-bg-tertiary rounded-lg border border-border-custom space-y-4"
    >
      <h4 class="text-sm font-medium text-zinc-300">Fixed Length Options</h4>

      <div class="flex items-center gap-4">
        <label class="text-xs text-zinc-400">Size:</label>
        <input
          v-model.number="fixedSize"
          :disabled="disabled"
          type="number"
          min="1"
          class="input w-32 font-mono"
        />
        <span class="text-xs text-zinc-500">bytes per message</span>
      </div>

      <div class="flex items-center gap-6">
        <div class="flex items-center gap-2">
          <label class="text-xs text-zinc-400">Padding:</label>
          <select v-model="paddingPosition" :disabled="disabled" class="input w-24">
            <option value="right">Right</option>
            <option value="left">Left</option>
          </select>
        </div>
        <div class="flex items-center gap-2">
          <label class="text-xs text-zinc-400">Byte:</label>
          <div class="flex items-center">
            <span class="text-xs text-zinc-500 mr-1">0x</span>
            <input
              v-model="paddingByteHex"
              @input="filterHexOnly"
              :disabled="disabled"
              type="text"
              maxlength="2"
              class="input w-12 font-mono text-center"
              placeholder="00"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- Collection Framing Preview (when mode is 'collection') -->
    <div
      v-if="framingMode === 'collection' && showCollectionOption"
      class="p-4 bg-bg-tertiary rounded-lg border border-border-custom"
    >
      <h4 class="text-sm font-medium text-zinc-300 mb-3">Collection's Shared Framing</h4>

      <div v-if="collectionFraming && collectionFraming.mode !== 'collection'" class="space-y-2">
        <div class="flex items-center gap-2">
          <span class="text-xs text-zinc-400">Mode:</span>
          <span class="text-sm text-zinc-200">{{ formatFramingMode(collectionFraming.mode) }}</span>
        </div>

        <!-- Delimiter details -->
        <div v-if="collectionFraming.mode === 'delimiter'" class="flex items-center gap-2">
          <span class="text-xs text-zinc-400">Delimiter:</span>
          <span class="text-sm text-zinc-200 font-mono">{{ collectionFraming.delimiter || '\\n' }}</span>
        </div>

        <!-- Length Prefix details -->
        <template v-if="collectionFraming.mode === 'length-prefix'">
          <div class="flex items-center gap-2">
            <span class="text-xs text-zinc-400">Encoding:</span>
            <span class="text-sm text-zinc-200">{{ collectionFraming.lengthEncoding || 'binary' }}</span>
          </div>
          <div class="flex items-center gap-2">
            <span class="text-xs text-zinc-400">Length Field:</span>
            <span class="text-sm text-zinc-200">{{ collectionFraming.lengthBytes || 4 }} bytes</span>
          </div>
          <div v-if="collectionFraming.lengthEncoding === 'binary'" class="flex items-center gap-2">
            <span class="text-xs text-zinc-400">Endian:</span>
            <span class="text-sm text-zinc-200">{{ collectionFraming.endian === 'little' ? 'Little' : 'Big' }}</span>
          </div>
        </template>

        <!-- Fixed Length details -->
        <div v-if="collectionFraming.mode === 'fixed-length'" class="flex items-center gap-2">
          <span class="text-xs text-zinc-400">Size:</span>
          <span class="text-sm text-zinc-200">{{ collectionFraming.fixedSize ?? 256 }} bytes</span>
        </div>
      </div>

      <div v-else class="text-sm text-zinc-500 italic">
        No shared framing configured for this collection. Using "None (Raw)".
      </div>
    </div>

    <!-- Sample Preview -->
    <div
      v-if="showPreview"
      class="p-4 bg-bg-tertiary rounded-lg border border-border-custom"
    >
      <h4 class="text-sm font-medium text-zinc-300 mb-3">Sample Preview</h4>
      <pre
        class="sample-hex-dump text-xs text-zinc-300 font-mono overflow-x-auto whitespace-pre leading-relaxed"
        v-html="samplePreview.hexDumpHtml"
      ></pre>
      <div v-if="samplePreview.isTruncated" class="text-xs text-zinc-500 italic mt-2">
        ... ({{ samplePreview.totalLen - samplePreview.displayLen }} bytes omitted)
      </div>
      <div class="flex gap-4 mt-3 text-xs text-zinc-500">
        <span v-if="samplePreview.headerLen > 0">
          <span class="inline-block w-3 h-3 bg-blue-800/50 rounded mr-1 align-middle"></span>
          Header ({{ samplePreview.headerLen }}B)
        </span>
        <span v-if="samplePreview.footerLen > 0">
          <span class="inline-block w-3 h-3 bg-green-800/50 rounded mr-1 align-middle"></span>
          Footer ({{ samplePreview.footerLen }}B)
        </span>
        <span>Total: {{ samplePreview.totalLen }} bytes</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Frame header highlight - blue background */
.sample-hex-dump :deep(.hex-header) {
  background-color: rgba(30, 64, 175, 0.5);
}

/* Frame footer highlight - green background */
.sample-hex-dump :deep(.hex-footer) {
  background-color: rgba(22, 101, 52, 0.5);
}
</style>
