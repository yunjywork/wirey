<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { parseHexDump, bytesToHexString, type ParseResult } from '@/utils/hexdump-parser'
import { Copy, FileInput, X } from 'lucide-vue-next'
import { DecodeFromHex } from '../../../wailsjs/go/main/App'

const props = defineProps<{
  visible: boolean
  charset?: string
}>()

const emit = defineEmits<{
  close: []
  insert: [data: string]
}>()

// Input state
const inputText = ref('')
const parseResult = ref<ParseResult | null>(null)
const copied = ref(false)
const decodedText = ref('')

// Parse input whenever it changes
watch(inputText, (value) => {
  parseResult.value = parseHexDump(value)
})

// Decode with charset when parse result changes
watch([parseResult, () => props.charset], async ([result, charset]) => {
  if (!result) {
    decodedText.value = ''
    return
  }
  try {
    const hexStr = bytesToHexString(result.bytes, '')
    decodedText.value = await DecodeFromHex(hexStr, charset || 'utf-8')
  } catch (e) {
    decodedText.value = ''
  }
}, { immediate: true })

// Reset when modal opens
watch(() => props.visible, (visible) => {
  if (visible) {
    inputText.value = ''
    parseResult.value = null
    copied.value = false
    copiedText.value = false
    decodedText.value = ''
  }
})

// Preview hex string (limited)
const previewHex = computed(() => {
  if (!parseResult.value) return ''
  const hex = bytesToHexString(parseResult.value.bytes)
  if (hex.length > 300) {
    return hex.substring(0, 300) + '...'
  }
  return hex
})

// Handle insert
function handleInsert() {
  if (!parseResult.value || !decodedText.value) return
  // Remove NULL bytes that could cause issues
  const cleanText = decodedText.value.replace(/\x00/g, '')
  emit('insert', cleanText)
  emit('close')
}

// Handle copy hex to clipboard
const copiedText = ref(false)

async function handleCopyHex() {
  if (!parseResult.value) return
  const hex = bytesToHexString(parseResult.value.bytes)
  try {
    await navigator.clipboard.writeText(hex)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch (e) {
    console.error('Failed to copy:', e)
  }
}

async function handleCopyText() {
  if (!decodedText.value) return
  try {
    // Remove NULL bytes that would truncate clipboard content
    const cleanText = decodedText.value.replace(/\x00/g, '')
    await navigator.clipboard.writeText(cleanText)
    copiedText.value = true
    setTimeout(() => { copiedText.value = false }, 2000)
  } catch (e) {
    console.error('Failed to copy:', e)
  }
}

function handleBackdropClick(e: MouseEvent) {
  if (e.target === e.currentTarget) {
    emit('close')
  }
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      @click="handleBackdropClick"
    >
      <div class="bg-bg-secondary border border-border-custom rounded-xl shadow-2xl w-[980px] max-h-[85vh] flex flex-col">
        <!-- Header -->
        <div class="flex items-center justify-between px-4 py-3 border-b border-border-custom">
          <div class="flex items-center gap-2">
            <FileInput class="w-5 h-5 text-accent-primary" />
            <h2 class="text-lg font-semibold text-zinc-100">Hex Dump Parser</h2>
          </div>
          <button
            @click="$emit('close')"
            class="p-1 rounded hover:bg-bg-tertiary text-zinc-400 hover:text-zinc-200 transition-colors"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <!-- Content -->
        <div class="flex-1 overflow-y-auto p-4 space-y-4">
          <!-- Input textarea -->
          <div>
            <label class="block text-sm text-zinc-400 mb-2">Paste hex dump here:</label>
            <textarea
              v-model="inputText"
              class="w-full h-56 px-3 py-2 bg-bg-tertiary border border-border-custom rounded-lg text-zinc-200 text-sm font-mono placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-accent-primary/50 resize-none"
              placeholder="Supported formats:
| 00000000 | 30303030  31323531 | ascii |
00000000  48 65 6C 6C 6F  Hello (Wireshark)
00000000: 4865 6c6c 6f20  Hello (xxd)
0x0000:  4500 0054 0000  E..T.. (tcpdump)
48 65 6C 6C 6F (pure hex)"
            ></textarea>
          </div>

          <!-- Parse result -->
          <div v-if="parseResult" class="bg-green-500/10 border border-green-500/30 rounded-lg p-3">
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-2 text-green-400 font-medium">
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
                </svg>
                Parsed Successfully
              </div>
              <div class="text-xs text-zinc-400">
                <span class="text-zinc-200">{{ parseResult.format }}</span>
                <span class="mx-2">•</span>
                <span class="text-zinc-200">{{ parseResult.byteCount }}</span> bytes
              </div>
            </div>
            <div class="mb-2">
              <div class="text-xs text-zinc-500 mb-1">Hex:</div>
              <pre class="text-xs text-zinc-400 font-mono bg-bg-primary rounded px-2 py-1 overflow-x-auto whitespace-nowrap max-h-12 overflow-y-hidden">{{ previewHex }}</pre>
            </div>
            <div>
              <div class="text-xs text-zinc-500 mb-1">Text ({{ props.charset || 'utf-8' }}):</div>
              <pre class="text-sm text-zinc-200 font-mono bg-bg-primary rounded px-3 py-2 overflow-x-auto whitespace-pre-wrap break-all h-36 overflow-y-auto">{{ decodedText }}</pre>
            </div>
          </div>

          <!-- No result / error -->
          <div v-else-if="inputText.trim()" class="bg-amber-500/10 border border-amber-500/30 rounded-lg p-3">
            <div class="flex items-center gap-2 text-amber-400 font-medium">
              <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
              </svg>
              Could not parse hex dump
            </div>
            <p class="text-xs text-zinc-400 mt-1">Check the format and try again</p>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-4 py-3 border-t border-border-custom flex items-center justify-between">
          <span class="text-xs text-zinc-500">
            Supports: User format, Wireshark, pure hex
          </span>
          <div class="flex items-center gap-2">
            <button
              @click="handleCopyHex"
              :disabled="!parseResult"
              class="px-3 py-1.5 text-sm rounded-lg transition-colors flex items-center gap-1.5"
              :class="parseResult
                ? 'bg-bg-tertiary text-zinc-300 hover:bg-zinc-600'
                : 'bg-bg-tertiary/50 text-zinc-500 cursor-not-allowed'"
            >
              <Copy class="w-3.5 h-3.5" />
              {{ copied ? 'Copied!' : 'Copy Hex' }}
            </button>
            <button
              @click="handleCopyText"
              :disabled="!decodedText"
              class="px-3 py-1.5 text-sm rounded-lg transition-colors flex items-center gap-1.5"
              :class="decodedText
                ? 'bg-bg-tertiary text-zinc-300 hover:bg-zinc-600'
                : 'bg-bg-tertiary/50 text-zinc-500 cursor-not-allowed'"
            >
              <Copy class="w-3.5 h-3.5" />
              {{ copiedText ? 'Copied!' : 'Copy Text' }}
            </button>
            <button
              @click="handleInsert"
              :disabled="!parseResult"
              class="px-3 py-1.5 text-sm rounded-lg transition-colors flex items-center gap-1.5"
              :class="parseResult
                ? 'bg-accent-primary text-white hover:bg-accent-primary/80'
                : 'bg-accent-primary/30 text-zinc-400 cursor-not-allowed'"
            >
              <FileInput class="w-3.5 h-3.5" />
              Insert
            </button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
