<script setup lang="ts">
import { ref, computed } from 'vue'
import type { Message } from '@/types'
import { formatHexDumpWithHighlight, formatHexDump, formatTimestamp, formatSize, formatHexString } from '@/utils/hexdump'
import { Copy, Check } from 'lucide-vue-next'

const props = defineProps<{
  message: Message
  showTimestamp: boolean
  charset?: string
}>()

const isExpanded = ref(false)
const copiedHex = ref(false)
const copiedContent = ref(false)

function toggleExpand() {
  isExpanded.value = !isExpanded.value
}

// Copy hex dump to clipboard
async function copyHexDump() {
  if (!props.message.rawBytes) return
  const hexText = formatHexDump(props.message.rawBytes)
  try {
    await navigator.clipboard.writeText(hexText)
    copiedHex.value = true
    setTimeout(() => { copiedHex.value = false }, 2000)
  } catch (e) {
    console.error('Failed to copy:', e)
  }
}

// Copy content to clipboard
async function copyContent() {
  try {
    await navigator.clipboard.writeText(props.message.content)
    copiedContent.value = true
    setTimeout(() => { copiedContent.value = false }, 2000)
  } catch (e) {
    console.error('Failed to copy:', e)
  }
}

const timestamp = computed(() => formatTimestamp(new Date(props.message.timestamp)))

const isSystemMessage = computed(() => props.message.direction === 'system')

const directionIcon = computed(() => {
  if (props.message.direction === 'sent') return '->'
  if (props.message.direction === 'received') return '<-'
  // System messages
  if (props.message.systemType === 'connected') return '●'
  if (props.message.systemType === 'disconnected') return '○'
  if (props.message.systemType === 'script') return '»'
  return '!'  // error
})

const colorClass = computed(() => {
  if (props.message.direction === 'sent') return 'text-msg-sent'
  if (props.message.direction === 'received') return 'text-msg-received'
  // System messages
  if (props.message.systemType === 'connected') return 'text-green-500'
  if (props.message.systemType === 'disconnected') return 'text-zinc-500'
  if (props.message.systemType === 'script') return 'text-amber-500'
  return 'text-red-500'  // error
})

// Clean content for display (replace control chars), let CSS handle truncation
const preview = computed(() => props.message.content.replace(/[\x00-\x1F\x7F-\x9F]/g, '.'))

const size = computed(() => formatSize(props.message.size ?? props.message.content.length))

// Calculate header/footer byte counts from hex strings
const headerBytes = computed(() => {
  const hex = props.message.framingInfo?.frameHeader || ''
  return hex.replace(/\s/g, '').length / 2
})

const footerBytes = computed(() => {
  const hex = props.message.framingInfo?.frameFooter || ''
  return hex.replace(/\s/g, '').length / 2
})

// Hex dump with frame highlighting (returns HTML)
const hexDumpHtml = computed(() => {
  if (!props.message.rawBytes) return ''
  return formatHexDumpWithHighlight(props.message.rawBytes, headerBytes.value, footerBytes.value)
})

const hasFramingInfo = computed(() => {
  const info = props.message.framingInfo
  return info && info.mode !== 'none'
})

const frameHeaderFormatted = computed(() => {
  if (!props.message.framingInfo?.frameHeader) return ''
  return formatHexString(props.message.framingInfo.frameHeader)
})

const frameFooterFormatted = computed(() => {
  if (!props.message.framingInfo?.frameFooter) return ''
  return formatHexString(props.message.framingInfo.frameFooter)
})

// Format duration in human readable format
function formatDuration(ms?: number): string {
  if (!ms) return ''
  const seconds = Math.floor(ms / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m ${seconds % 60}s`
  } else if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`
  } else {
    return `${seconds}s`
  }
}

// Format bytes to human readable
function formatBytes(bytes?: number): string {
  if (!bytes) return '0 B'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

// Get disconnect reason text
const disconnectReasonText = computed(() => {
  const reason = props.message.reason
  if (reason === 'user') return 'User'
  if (reason === 'server') return 'Server'
  if (reason === 'error') return 'Error'
  return ''
})

// System message content
const systemMessageContent = computed(() => {
  const msg = props.message
  if (msg.systemType === 'connected') {
    // 127.0.0.1:4212 -> 127.0.0.1:8080 (TCP)
    const proto = msg.protocol?.toUpperCase() || ''
    return `${msg.localAddr || ''} -> ${msg.remoteAddr || ''} (${proto})`
  } else if (msg.systemType === 'disconnected') {
    // Duration: 5m 2s, Sent: 1.2KB, Recv: 3.4KB (User)
    const parts: string[] = []
    if (msg.duration) {
      parts.push(`Duration: ${formatDuration(msg.duration)}`)
    }
    if (msg.bytesSent !== undefined || msg.bytesRecv !== undefined) {
      parts.push(`Sent: ${formatBytes(msg.bytesSent)}`)
      parts.push(`Recv: ${formatBytes(msg.bytesRecv)}`)
    }
    if (disconnectReasonText.value) {
      parts.push(`(${disconnectReasonText.value})`)
    }
    return parts.join(', ')
  }
  return msg.content
})
</script>

<template>
  <div class="font-mono text-sm border-b border-subtle">
    <!-- System Message Row -->
    <div
      v-if="isSystemMessage"
      class="flex items-center gap-2 py-1.5 px-2"
    >
      <!-- Spacer for alignment -->
      <span class="w-3 shrink-0"></span>

      <!-- Timestamp -->
      <span v-if="showTimestamp" class="text-zinc-500 text-xs shrink-0">
        {{ timestamp }}
      </span>

      <!-- Status Icon -->
      <span :class="['shrink-0 w-5 text-center', colorClass]">
        {{ directionIcon }}
      </span>

      <!-- Status Label -->
      <span :class="['text-xs font-semibold shrink-0', colorClass]">
        {{ message.systemType === 'connected' ? 'CONNECTED' : message.systemType === 'disconnected' ? 'DISCONNECTED' : message.systemType === 'script' ? 'SCRIPT' : 'ERROR' }}
      </span>

      <!-- Details -->
      <span class="text-xs text-zinc-400 whitespace-pre-wrap break-all">
        {{ systemMessageContent }}
      </span>
    </div>

    <!-- Regular Message Row (sent/received) -->
    <div
      v-else
      @click="toggleExpand"
      class="flex items-center gap-2 py-1.5 px-2 cursor-pointer hover:bg-bg-secondary/50 transition-colors"
    >
      <!-- Expand/Collapse Arrow -->
      <span class="text-zinc-500 text-xs w-3 shrink-0">
        {{ isExpanded ? '\u25BC' : '\u25B6' }}
      </span>

      <!-- Timestamp -->
      <span v-if="showTimestamp" class="text-zinc-500 text-xs shrink-0">
        {{ timestamp }}
      </span>

      <!-- Local Address -->
      <span v-if="message.localAddr" class="text-zinc-400 text-xs shrink-0">
        {{ message.localAddr }}
      </span>

      <!-- Direction -->
      <span :class="['font-bold shrink-0 w-5', colorClass]">
        {{ directionIcon }}
      </span>

      <!-- Remote Address -->
      <span v-if="message.remoteAddr" class="text-zinc-400 text-xs shrink-0">
        {{ message.remoteAddr }}
      </span>

      <!-- Size -->
      <span class="text-zinc-500 text-xs shrink-0">
        {{ size }}
      </span>

      <!-- Content Preview -->
      <span class="text-zinc-300 flex-1 truncate">
        {{ preview }}
      </span>

      <!-- Format badge -->
      <span
        v-if="message.format !== 'text'"
        class="text-[10px] uppercase px-1.5 py-0.5 rounded bg-bg-tertiary text-zinc-500 shrink-0"
      >
        {{ message.format }}
      </span>
    </div>

    <!-- Expanded Content (only for sent/received messages) -->
    <div v-if="isExpanded && !isSystemMessage" class="px-6 pb-3 space-y-3">
      <!-- Framing Info Panel -->
      <div v-if="hasFramingInfo" class="bg-bg-tertiary/50 rounded-lg p-3 border border-muted">
        <div class="text-xs text-zinc-400 font-semibold mb-2">Framing Info</div>
        <div class="grid grid-cols-2 gap-x-4 gap-y-1 text-xs">
          <div class="flex justify-between">
            <span class="text-zinc-500">Mode:</span>
            <span class="text-zinc-300">{{ message.framingInfo?.mode }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-zinc-500">Settings:</span>
            <span class="text-zinc-300">{{ message.framingInfo?.settings || '-' }}</span>
          </div>
          <div v-if="frameHeaderFormatted" class="flex justify-between">
            <span class="text-zinc-500">Header:</span>
            <span class="text-zinc-300 font-mono">{{ frameHeaderFormatted }}</span>
          </div>
          <div v-if="frameFooterFormatted" class="flex justify-between">
            <span class="text-zinc-500">Footer:</span>
            <span class="text-zinc-300 font-mono">{{ frameFooterFormatted }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-zinc-500">Payload:</span>
            <span class="text-zinc-300">{{ message.framingInfo?.payloadSize }} bytes</span>
          </div>
          <div class="flex justify-between">
            <span class="text-zinc-500">Total:</span>
            <span class="text-zinc-300">{{ message.framingInfo?.totalSize }} bytes</span>
          </div>
        </div>
      </div>

      <!-- Hex Dump Panel -->
      <div class="bg-bg-tertiary/50 rounded-lg p-3 border border-muted">
        <div class="flex items-center justify-between mb-2">
          <span class="text-xs text-zinc-400 font-semibold">Hex Dump</span>
          <button
            v-if="hexDumpHtml"
            @click.stop="copyHexDump"
            class="flex items-center gap-1 px-2 py-0.5 text-[10px] rounded bg-bg-tertiary hover:bg-zinc-600 text-zinc-400 hover:text-zinc-200 transition-colors"
          >
            <Check v-if="copiedHex" class="w-3 h-3 text-green-400" />
            <Copy v-else class="w-3 h-3" />
            {{ copiedHex ? 'Copied!' : 'Copy' }}
          </button>
        </div>
        <pre
          v-if="hexDumpHtml"
          class="hex-dump text-xs text-zinc-300 font-mono overflow-x-auto whitespace-pre leading-relaxed"
          v-html="hexDumpHtml"
        ></pre>
        <div v-else class="text-xs text-zinc-500 italic">
          No raw bytes available
        </div>
      </div>

      <!-- Raw Content Panel -->
      <div class="bg-bg-tertiary/50 rounded-lg p-3 border border-muted">
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center gap-2">
            <span class="text-xs text-zinc-400 font-semibold">Content</span>
            <span v-if="charset" class="text-[10px] text-zinc-500 bg-bg-tertiary px-1.5 py-0.5 rounded">
              {{ charset }}
            </span>
          </div>
          <button
            @click.stop="copyContent"
            class="flex items-center gap-1 px-2 py-0.5 text-[10px] rounded bg-bg-tertiary hover:bg-zinc-600 text-zinc-400 hover:text-zinc-200 transition-colors"
          >
            <Check v-if="copiedContent" class="w-3 h-3 text-green-400" />
            <Copy v-else class="w-3 h-3" />
            {{ copiedContent ? 'Copied!' : 'Copy' }}
          </button>
        </div>
        <pre class="text-xs text-zinc-300 font-mono overflow-x-auto whitespace-pre-wrap break-all">{{ message.content }}</pre>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Frame header highlight - blue background */
.hex-dump :deep(.hex-header) {
  background-color: rgba(30, 64, 175, 0.5);
}

/* Frame footer highlight - green background */
.hex-dump :deep(.hex-footer) {
  background-color: rgba(22, 101, 52, 0.5);
}
</style>
