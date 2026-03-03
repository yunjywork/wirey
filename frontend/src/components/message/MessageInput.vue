<script setup lang="ts">
import { ref, computed, watch, shallowRef, nextTick } from 'vue'
import type { Case, MessageFormat } from '@/types'
import { useCaseStore } from '@/stores/case'
import { useSettingsStore } from '@/stores/settings'
import { Send, EncodeToHex, DecodeFromHex, ProcessMessageWithScripts } from '../../../wailsjs/go/main/App'
import { main } from '../../../wailsjs/go/models'
import HexDumpParserModal from './HexDumpParserModal.vue'
import { FileInput } from 'lucide-vue-next'
import { Codemirror } from 'vue-codemirror'
import { EditorView, keymap, placeholder } from '@codemirror/view'
import { getMessageInputExtensions } from '@/utils/codemirror'

// Variables help modal
const showHelpModal = ref(false)

// Hex dump parser modal
const showHexDumpParser = ref(false)

// CodeMirror editor view reference
const editorView = shallowRef<EditorView>()

function onEditorReady(payload: { view: EditorView }) {
  editorView.value = payload.view
}

// Send handler for keymap
let handleSendRef: () => void = () => {}

// Variable definitions for tooltips and help
const variableDefinitions: Record<string, { description: string; example: string }> = {
  'timestamp': { description: 'Unix timestamp (seconds)', example: '1703318400' },
  'timestamp_ms': { description: 'Unix timestamp (milliseconds)', example: '1703318400123' },
  'datetime': { description: 'ISO 8601 format', example: '2024-12-23T14:30:00Z' },
  'date': { description: 'Date only', example: '2024-12-23' },
  'time': { description: 'Time only', example: '14:30:00' },
  'uuid': { description: 'UUID v4', example: '550e8400-e29b-41d4-a716-446655440000' },
  'random': { description: 'Random hex (4 bytes default)', example: 'A37BF201' },
  'random:N': { description: 'N bytes random hex', example: '{{random:8}} → A37BF201C9D2E4F6' },
  'counter': { description: 'Auto-increment from 1', example: '1, 2, 3...' },
  'counter:N': { description: 'Auto-increment from N', example: '{{counter:100}} → 100, 101...' }
}

const escapeDefinitions: Record<string, { description: string; hex: string }> = {
  '\\n': { description: 'Line Feed', hex: '0x0A' },
  '\\r': { description: 'Carriage Return', hex: '0x0D' },
  '\\t': { description: 'Tab', hex: '0x09' },
  '\\0': { description: 'NULL', hex: '0x00' },
  '\\\\': { description: 'Backslash', hex: '0x5C' },
  '\\xNN': { description: 'Hex byte (e.g., \\x01 for SOH)', hex: '0x00-0xFF' }
}

const props = defineProps<{
  caseData: Case
}>()

const caseStore = useCaseStore()
const settingsStore = useSettingsStore()

// Cursor position saved before modal opens
const lastCursorPos = ref(0)

// Save cursor position before modal opens
function saveCursorPosition() {
  if (editorView.value) {
    lastCursorPos.value = editorView.value.state.selection.main.head
  }
}

// Insert text at cursor position (using CodeMirror API)
function insertAtCursor(text: string) {
  const view = editorView.value
  if (!view) {
    // Fallback: just append
    message.value += text
    return
  }

  const pos = lastCursorPos.value
  view.dispatch({
    changes: { from: pos, insert: text },
    selection: { anchor: pos + text.length }
  })
  showHelpModal.value = false

  // Focus editor after insert
  setTimeout(() => {
    view.focus()
  }, 50)
}

// Handle hex dump parser insert
function handleHexDumpInsert(data: string) {
  const view = editorView.value
  if (!view) {
    // Fallback
    message.value += data
    return
  }

  // Save current position
  const pos = view.state.selection.main.head

  view.dispatch({
    changes: { from: pos, insert: data },
    selection: { anchor: pos + data.length }
  })

  // Focus editor after insert
  setTimeout(() => {
    view.focus()
  }, 50)
}

// Insert variable
function insertVariable(key: string) {
  insertAtCursor('{{' + key + '}}')
}

// Insert escape sequence
function insertEscape(seq: string) {
  insertAtCursor(seq)
}

// Get effective charset (resolved through hierarchy)
const effectiveCharset = computed(() => {
  const caseCharset = props.caseData.charset

  if (caseCharset && caseCharset !== 'global' && caseCharset !== 'collection') {
    return caseCharset
  }

  if (caseCharset === 'collection') {
    const collection = caseStore.findCaseCollection(props.caseData.id)
    if (collection?.sharedCharset) {
      return collection.sharedCharset
    }
  }

  return settingsStore.defaultCharset
})

const isSending = ref(false)
const errorMessage = ref('')
const clearAfterSend = ref(false)

// Use case's draft message (persisted)
const message = computed({
  get: () => props.caseData.draftMessage || '',
  set: (value: string) => {
    caseStore.updateCase(props.caseData.id, { draftMessage: value })
  }
})

const format = computed({
  get: () => props.caseData.draftFormat || 'text',
  set: (value: MessageFormat) => {
    caseStore.updateCase(props.caseData.id, { draftFormat: value })
  }
})

const useVariables = computed({
  get: () => props.caseData.useVariables ?? true,
  set: (value: boolean) => {
    caseStore.updateCase(props.caseData.id, { useVariables: value })
  }
})

// Check if message contains variables
const hasVariables = computed(() => {
  const msg = message.value
  return msg.includes('{{') && msg.includes('}}')
})

// Get tooltip for a variable
function getVariableTooltip(varName: string): string {
  // Handle parameterized variables
  if (varName.startsWith('random:')) {
    const def = variableDefinitions['random:N']
    return `${def.description}\nExample: ${def.example}`
  }
  if (varName.startsWith('counter:')) {
    const def = variableDefinitions['counter:N']
    return `${def.description}\nExample: ${def.example}`
  }
  const def = variableDefinitions[varName]
  if (def) {
    return `${def.description}\nExample: ${def.example}`
  }
  return ''
}

// Get tooltip for escape sequence
function getEscapeTooltip(seq: string): string {
  // Handle \xNN pattern
  if (seq.startsWith('\\x')) {
    return `Hex byte: ${seq.slice(2).toUpperCase()}\n→ 0x${seq.slice(2).toUpperCase()}`
  }
  const def = escapeDefinitions[seq]
  if (def) {
    return `${def.description}\n→ ${def.hex}`
  }
  return seq
}


// Panel resize - persist per case in localStorage
const PANEL_HEIGHT_KEY = 'wirey-input-panel-height'
const DEFAULT_PANEL_HEIGHT = 160
const MIN_PANEL_HEIGHT = 100
const MAX_PANEL_HEIGHT_RATIO = 0.4 // 40% of viewport height

function loadPanelHeight(): number {
  try {
    const stored = localStorage.getItem(PANEL_HEIGHT_KEY)
    if (stored) {
      const heights = JSON.parse(stored)
      const savedHeight = heights[props.caseData.id] ?? DEFAULT_PANEL_HEIGHT
      // Clamp to current viewport constraints
      const maxHeight = Math.floor(window.innerHeight * MAX_PANEL_HEIGHT_RATIO)
      return Math.max(MIN_PANEL_HEIGHT, Math.min(maxHeight, savedHeight))
    }
  } catch (e) {
    console.error('Failed to load panel height:', e)
  }
  return DEFAULT_PANEL_HEIGHT
}

function savePanelHeight(height: number) {
  try {
    const stored = localStorage.getItem(PANEL_HEIGHT_KEY)
    const heights = stored ? JSON.parse(stored) : {}
    heights[props.caseData.id] = height
    localStorage.setItem(PANEL_HEIGHT_KEY, JSON.stringify(heights))
  } catch (e) {
    console.error('Failed to save panel height:', e)
  }
}

const panelHeight = ref(loadPanelHeight())
const isResizing = ref(false)

// Reload panel height when case changes
watch(() => props.caseData.id, () => {
  panelHeight.value = loadPanelHeight()
})

function startResize(e: MouseEvent) {
  e.preventDefault()
  isResizing.value = true
  const startY = e.clientY
  const startHeight = panelHeight.value
  const maxHeight = Math.floor(window.innerHeight * MAX_PANEL_HEIGHT_RATIO)

  function onMouseMove(e: MouseEvent) {
    const deltaY = startY - e.clientY
    const newHeight = Math.max(MIN_PANEL_HEIGHT, Math.min(maxHeight, startHeight + deltaY))
    panelHeight.value = newHeight
  }

  function onMouseUp() {
    isResizing.value = false
    savePanelHeight(panelHeight.value)
    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseup', onMouseUp)
  }

  document.addEventListener('mousemove', onMouseMove)
  document.addEventListener('mouseup', onMouseUp)
}

const isConverting = ref(false)

async function setFormat(newFormat: MessageFormat) {
  if (newFormat === format.value) return

  // Disable HEX mode if variables are enabled and message contains variables
  if (newFormat === 'hex' && useVariables.value && hasVariables.value) {
    return
  }

  const currentMessage = message.value.trim()
  if (currentMessage) {
    isConverting.value = true
    try {
      if (newFormat === 'hex') {
        // Text to Hex: encode using charset
        const hexStr = await EncodeToHex(currentMessage, effectiveCharset.value)
        message.value = hexStr
      } else {
        // Hex to Text: decode using charset
        const textStr = await DecodeFromHex(currentMessage, effectiveCharset.value)
        message.value = textStr
      }
    } catch (e) {
      console.error('Format conversion failed:', e)
      // Keep original message on error
    } finally {
      isConverting.value = false
    }
  }
  format.value = newFormat
}

// HEX button disabled when variables are used
const hexDisabled = computed(() => {
  return useVariables.value && hasVariables.value
})

async function handleSend() {
  if (!message.value.trim() || isSending.value) return
  if (props.caseData.status !== 'connected') return

  let content = message.value.trim()
  isSending.value = true
  errorMessage.value = ''

  try {
    // Check if scripts are enabled
    const caseScript = props.caseData.scriptConfig
    const collection = caseStore.findCaseCollection(props.caseData.id)
    const collectionScript = collection?.sharedScriptConfig

    const hasEnabledScripts =
      (caseScript?.setupEnabled || caseScript?.preSendEnabled) ||
      (collectionScript?.setupEnabled || collectionScript?.preSendEnabled)

    // Run script pipeline if any scripts are enabled
    if (hasEnabledScripts && format.value === 'text') {
      const req = new main.ScriptProcessRequest({
        message: content,
        caseId: props.caseData.id,
        collectionName: props.caseData.collectionName,
        collectionScriptConfig: collectionScript ? {
          setupScript: collectionScript.setupScript,
          setupEnabled: collectionScript.setupEnabled,
          preSendScript: collectionScript.preSendScript,
          preSendEnabled: collectionScript.preSendEnabled
        } : undefined,
        caseScriptConfig: caseScript ? {
          setupScript: caseScript.setupScript,
          setupEnabled: caseScript.setupEnabled,
          preSendScript: caseScript.preSendScript,
          preSendEnabled: caseScript.preSendEnabled
        } : undefined,
        collectionVariables: collection?.sharedVariables,
        caseVariables: props.caseData.localVariables
      })

      const processed = await ProcessMessageWithScripts(req)
      if (processed === null) {
        // Script returned null - cancel send
        isSending.value = false
        return
      }
      content = processed
      // Send processed message without additional variable substitution
      await Send(props.caseData.id, content, format.value, false)
    } else {
      // No scripts enabled, use normal send with variable substitution
      await Send(props.caseData.id, content, format.value, useVariables.value)
    }

    if (clearAfterSend.value) {
      message.value = ''
    }
  } catch (error) {
    console.error('Send failed:', error)
    errorMessage.value = `Send failed: ${error}`
  } finally {
    isSending.value = false
  }
}

const isConnected = computed(() => props.caseData.status === 'connected')

// Assign handleSend to ref for keymap
handleSendRef = handleSend

// CodeMirror extensions (computed based on theme and settings)
const editorExtensions = computed(() => {
  const isDark = settingsStore.settings.theme === 'dark'
  const enableHighlight = useVariables.value && format.value !== 'hex'

  return [
    ...getMessageInputExtensions(isDark, enableHighlight),
    placeholder(isConnected.value
      ? 'Type your message... (Ctrl+Enter to send)'
      : 'Type your message... (Connect to send)'
    ),
    keymap.of([
      {
        key: 'Ctrl-Enter',
        mac: 'Cmd-Enter',
        run: () => {
          handleSendRef()
          return true
        }
      }
    ])
  ]
})

const formats: { value: MessageFormat; label: string }[] = [
  { value: 'text', label: 'Text' },
  { value: 'hex', label: 'Hex' }
]
</script>

<template>
  <div
    class="border-t border-border-custom bg-bg-secondary flex flex-col shrink-0"
    :style="{ height: `${panelHeight}px`, maxHeight: '40vh' }"
  >
    <!-- Resize handle -->
    <div
      @mousedown="startResize"
      class="h-1 cursor-ns-resize hover:bg-accent-primary/30 transition-colors flex items-center justify-center group"
      :class="{ 'bg-accent-primary/50': isResizing }"
    >
      <div class="w-8 h-0.5 rounded-full bg-zinc-600 group-hover:bg-zinc-400 transition-colors"></div>
    </div>

    <!-- Content -->
    <div class="flex-1 flex flex-col p-3 pt-2 overflow-hidden">
      <!-- Format selector and options -->
      <div class="flex items-center justify-between mb-2">
        <div class="flex items-center gap-4">
          <!-- Format selector -->
          <div class="flex items-center gap-2">
            <span class="text-xs text-zinc-400">Format:</span>
            <div class="flex rounded-lg bg-bg-tertiary p-0.5">
              <button
                v-for="f in formats"
                :key="f.value"
                @click="setFormat(f.value)"
                :disabled="f.value === 'hex' && hexDisabled"
                :class="[
                  'px-3 py-1 text-xs rounded-md transition-all',
                  format === f.value
                    ? 'bg-accent-primary text-white'
                    : f.value === 'hex' && hexDisabled
                      ? 'text-zinc-600 cursor-not-allowed'
                      : 'text-zinc-400 hover:text-zinc-200'
                ]"
                :title="f.value === 'hex' && hexDisabled ? 'Disable variables to use HEX mode' : ''"
              >
                {{ f.label }}
              </button>
            </div>
          </div>

          <!-- Variables toggle -->
          <div class="flex items-center gap-1">
            <button
              @click="useVariables = !useVariables"
              :disabled="format === 'hex'"
              :class="[
                'flex items-center gap-1.5 px-2 py-1 rounded-md text-xs transition-all',
                format === 'hex'
                  ? 'text-zinc-600 cursor-not-allowed'
                  : useVariables
                    ? 'bg-amber-500/20 text-amber-400'
                    : 'text-zinc-500 hover:text-zinc-300 hover:bg-bg-tertiary'
              ]"
              :title="format === 'hex' ? 'Variables not available in HEX mode' : 'Enable {{timestamp}}, {{uuid}}, \\x00 etc.'"
            >
              <span
                :class="[
                  'w-6 h-3.5 rounded-full relative transition-colors',
                  format === 'hex' ? 'bg-zinc-700' : useVariables ? 'bg-amber-500' : 'bg-zinc-600'
                ]"
              >
                <span
                  :class="[
                    'absolute top-0.5 w-2.5 h-2.5 rounded-full bg-white transition-all',
                    useVariables && format !== 'hex' ? 'left-3' : 'left-0.5'
                  ]"
                ></span>
              </span>
              Variables
            </button>
            <!-- Help button -->
            <button
              @click="saveCursorPosition(); showHelpModal = true"
              class="w-5 h-5 rounded-full bg-zinc-700 hover:bg-zinc-600 text-zinc-400 hover:text-zinc-200
                     flex items-center justify-center text-xs font-medium transition-all"
              title="Show variables reference"
            >
              ?
            </button>

            <!-- Separator -->
            <div class="w-px h-4 bg-zinc-600 mx-1"></div>

            <!-- Hex dump parser button -->
            <button
              @click="saveCursorPosition(); showHexDumpParser = true"
              class="px-2 py-1 rounded bg-zinc-700 hover:bg-zinc-600 text-zinc-400 hover:text-zinc-200
                     flex items-center gap-1.5 text-xs transition-all"
              title="Parse hex dump"
            >
              <FileInput class="w-3 h-3" />
              Parse Hex
            </button>
          </div>
        </div>

        <!-- Clear after send toggle -->
        <button
          @click="clearAfterSend = !clearAfterSend"
          :class="[
            'flex items-center gap-1.5 px-2 py-1 rounded-md text-xs transition-all',
            clearAfterSend
              ? 'bg-accent-primary/20 text-accent-primary'
              : 'text-zinc-500 hover:text-zinc-300 hover:bg-bg-tertiary'
          ]"
        >
          <span
            :class="[
              'w-6 h-3.5 rounded-full relative transition-colors',
              clearAfterSend ? 'bg-accent-primary' : 'bg-zinc-600'
            ]"
          >
            <span
              :class="[
                'absolute top-0.5 w-2.5 h-2.5 rounded-full bg-white transition-all',
                clearAfterSend ? 'left-3' : 'left-0.5'
              ]"
            ></span>
          </span>
          Clear after send
        </button>
      </div>

      <!-- Input area -->
      <div class="flex-1 flex gap-3 min-h-0 overflow-hidden">
        <div class="flex-1 relative overflow-hidden">
          <!-- CodeMirror Editor -->
          <Codemirror
            v-model="message"
            :extensions="editorExtensions"
            @ready="onEditorReady"
            class="message-editor h-full w-full rounded-lg border border-border-custom
                   focus-within:ring-2 focus-within:ring-accent-primary/50 focus-within:border-accent-primary
                   transition-all duration-200 overflow-hidden"
          />
          <span class="absolute bottom-2 right-3 text-[10px] text-zinc-600 z-20 pointer-events-none">
            Ctrl+Enter
          </span>
        </div>

        <button
          @click="handleSend"
          :disabled="!message.trim() || !isConnected || isSending"
          class="btn-primary px-5 h-fit self-end flex items-center gap-2"
        >
          <svg v-if="isSending" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
          </svg>
          <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"/>
          </svg>
          Send
        </button>
      </div>
    </div>
  </div>

  <!-- Variables Help Modal -->
  <Teleport to="body">
    <div
      v-if="showHelpModal"
      class="fixed inset-0 bg-black/60 flex items-center justify-center z-50"
      @click.self="showHelpModal = false"
    >
      <div class="bg-bg-secondary border border-border-custom rounded-xl shadow-2xl w-[600px] max-h-[80vh] overflow-hidden">
        <!-- Header -->
        <div class="flex items-center justify-between px-5 py-4 border-b border-border-custom">
          <h2 class="text-lg font-semibold text-zinc-100">Variables & Escape Sequences</h2>
          <button
            @click="showHelpModal = false"
            class="w-8 h-8 rounded-lg hover:bg-bg-tertiary flex items-center justify-center text-zinc-400 hover:text-zinc-200 transition-colors"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <!-- Content -->
        <div class="p-5 overflow-y-auto max-h-[calc(80vh-120px)]">
          <!-- Variables Section -->
          <div class="mb-6">
            <h3 class="text-sm font-semibold text-amber-400 mb-3 flex items-center gap-2">
              <span class="w-2 h-2 rounded-full bg-amber-400"></span>
              Variables
            </h3>
            <div class="bg-bg-tertiary rounded-lg overflow-hidden">
              <table class="w-full text-sm">
                <thead>
                  <tr class="border-b border-border-custom text-zinc-400">
                    <th class="text-left px-4 py-2 font-medium">Variable</th>
                    <th class="text-left px-4 py-2 font-medium">Description</th>
                    <th class="text-left px-4 py-2 font-medium">Example</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="(def, key) in variableDefinitions"
                    :key="key"
                    @click="insertVariable(String(key))"
                    class="border-b border-border-custom/50 last:border-0 hover:bg-amber-500/10 cursor-pointer transition-colors"
                  >
                    <td class="px-4 py-2">
                      <code class="text-amber-400 bg-amber-500/20 px-1.5 py-0.5 rounded text-xs" v-text="'{{' + key + '}}'"></code>
                    </td>
                    <td class="px-4 py-2 text-zinc-300">{{ def.description }}</td>
                    <td class="px-4 py-2 text-zinc-500 font-mono text-xs">{{ def.example }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Escape Sequences Section -->
          <div class="mb-6">
            <h3 class="text-sm font-semibold text-cyan-400 mb-3 flex items-center gap-2">
              <span class="w-2 h-2 rounded-full bg-cyan-400"></span>
              Escape Sequences
            </h3>
            <div class="bg-bg-tertiary rounded-lg overflow-hidden">
              <table class="w-full text-sm">
                <thead>
                  <tr class="border-b border-border-custom text-zinc-400">
                    <th class="text-left px-4 py-2 font-medium">Sequence</th>
                    <th class="text-left px-4 py-2 font-medium">Description</th>
                    <th class="text-left px-4 py-2 font-medium">Hex</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="(def, key) in escapeDefinitions"
                    :key="key"
                    @click="insertEscape(String(key))"
                    class="border-b border-border-custom/50 last:border-0 hover:bg-cyan-500/10 cursor-pointer transition-colors"
                  >
                    <td class="px-4 py-2">
                      <code class="text-cyan-400 bg-cyan-500/20 px-1.5 py-0.5 rounded text-xs">{{ key }}</code>
                    </td>
                    <td class="px-4 py-2 text-zinc-300">{{ def.description }}</td>
                    <td class="px-4 py-2 text-zinc-500 font-mono text-xs">{{ def.hex }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Example Section -->
          <div>
            <h3 class="text-sm font-semibold text-zinc-300 mb-3">Example</h3>
            <div class="bg-bg-tertiary rounded-lg p-4 font-mono text-xs space-y-2">
              <div class="text-zinc-300">
                <span class="text-zinc-500">Hello </span><span class="text-amber-400" v-text="'{{uuid}}'"></span><span class="text-cyan-400">\r\n</span>
              </div>
              <div class="text-zinc-500 text-[11px]">
                → Hello 550e8400-e29b-41d4-a716-446655440000\r\n
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-5 py-3 border-t border-border-custom bg-bg-tertiary/50 flex justify-end">
          <button
            @click="showHelpModal = false"
            class="px-4 py-2 bg-accent-primary hover:bg-accent-primary/80 text-white text-sm rounded-lg transition-colors"
          >
            Got it
          </button>
        </div>
      </div>
    </div>
  </Teleport>

  <!-- Hex Dump Parser Modal -->
  <HexDumpParserModal
    :visible="showHexDumpParser"
    :charset="effectiveCharset"
    @close="showHexDumpParser = false"
    @insert="handleHexDumpInsert"
  />
</template>

<style scoped>
/* CodeMirror editor container */
.message-editor {
  border: 1px solid rgb(var(--color-border-custom));
  border-radius: 0.5rem;
  overflow: hidden;
}

.message-editor :deep(.cm-editor) {
  height: 100%;
  background: rgb(var(--color-bg-tertiary));
  border-radius: 0.5rem;
}

.message-editor :deep(.cm-scroller) {
  overflow: auto;
}
</style>
