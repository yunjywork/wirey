<script setup lang="ts">
import { ref, computed, shallowRef, watch, onMounted } from 'vue'
import type { Case } from '@/types'
import { DEFAULT_SETUP_SCRIPT, DEFAULT_PRESEND_SCRIPT, DEFAULT_POSTRECV_SCRIPT } from '@/types'
import { useCaseStore } from '@/stores/case'
import { useSettingsStore } from '@/stores/settings'
import { Codemirror } from 'vue-codemirror'
import { EditorView } from '@codemirror/view'
import { getScriptExtensions } from '@/utils/codemirror'
import { DryRunScript, DryRunPostRecvScript } from '../../../wailsjs/go/main/App'
import { main } from '../../../wailsjs/go/models'
import { escapeForDisplay } from '@/utils/hexdump'
import WireyHelpModal from './WireyHelpModal.vue'
import { HelpCircle } from 'lucide-vue-next'

const props = defineProps<{
  caseData: Case
}>()

const caseStore = useCaseStore()
const settingsStore = useSettingsStore()

// Store EditorView instances for focus
const setupEditorView = shallowRef<EditorView>()
const preSendEditorView = shallowRef<EditorView>()
const postRecvEditorView = shallowRef<EditorView>()

function onSetupReady(payload: { view: EditorView }) {
  setupEditorView.value = payload.view
}

function onPreSendReady(payload: { view: EditorView }) {
  preSendEditorView.value = payload.view
}

function onPostRecvReady(payload: { view: EditorView }) {
  postRecvEditorView.value = payload.view
}

// Script tab state
type ScriptTab = 'setup' | 'presend' | 'postrecv'
const activeScriptTab = ref<ScriptTab>('setup')

// Variables panel tab state
type VarsTab = 'builtin' | 'collection'
const activeVarsTab = ref<VarsTab>('builtin')

// Local state for editing (synced with store)
const setupEnabled = ref(props.caseData.scriptConfig?.setupEnabled ?? false)
const preSendEnabled = ref(props.caseData.scriptConfig?.preSendEnabled ?? false)
const postRecvEnabled = ref(props.caseData.scriptConfig?.postRecvEnabled ?? false)
const setupScript = ref(props.caseData.scriptConfig?.setupScript ?? DEFAULT_SETUP_SCRIPT)
const preSendScript = ref(props.caseData.scriptConfig?.preSendScript ?? DEFAULT_PRESEND_SCRIPT)
const postRecvScript = ref(props.caseData.scriptConfig?.postRecvScript ?? DEFAULT_POSTRECV_SCRIPT)

// Hex mode check - scripts don't work in hex mode
const isHexMode = computed(() => props.caseData.draftFormat === 'hex')

// Collapse states (persisted to localStorage)
const helpCollapsed = ref(localStorage.getItem('wirey-scripts-help-collapsed') === 'true')
const varsCollapsed = ref(localStorage.getItem('wirey-scripts-vars-collapsed') === 'true')

// Wirey help modal
const showWireyHelp = ref(false)

// Check if current script tab is enabled
const isCurrentScriptEnabled = computed(() => {
  if (activeScriptTab.value === 'setup') return setupEnabled.value
  if (activeScriptTab.value === 'presend') return preSendEnabled.value
  return postRecvEnabled.value
})

function handleWireyHelpInsert(code: string) {
  // Don't insert if script is disabled
  if (!isCurrentScriptEnabled.value) return

  // Insert code into the current active script
  if (activeScriptTab.value === 'setup') {
    setupScript.value += '\n' + code
  } else if (activeScriptTab.value === 'presend') {
    // Insert before 'return msg;'
    const returnIdx = preSendScript.value.lastIndexOf('return msg;')
    if (returnIdx !== -1) {
      preSendScript.value =
        preSendScript.value.slice(0, returnIdx) +
        code + '\n' +
        preSendScript.value.slice(returnIdx)
    } else {
      preSendScript.value += '\n' + code
    }
  } else {
    postRecvScript.value += '\n' + code
  }
}

function toggleHelp() {
  helpCollapsed.value = !helpCollapsed.value
  localStorage.setItem('wirey-scripts-help-collapsed', String(helpCollapsed.value))
}

function toggleVars() {
  varsCollapsed.value = !varsCollapsed.value
  localStorage.setItem('wirey-scripts-vars-collapsed', String(varsCollapsed.value))
}

// Insert as log toggle (persisted)
const insertAsLog = ref(localStorage.getItem('wirey-scripts-insert-as-log') !== 'false')
watch(insertAsLog, (val) => localStorage.setItem('wirey-scripts-insert-as-log', String(val)))

// Context menu state
const contextMenu = ref<{
  show: boolean
  x: number
  y: number
  varName: string
  varType: 'builtin' | 'collection'
}>({ show: false, x: 0, y: 0, varName: '', varType: 'builtin' })

function showContextMenu(e: MouseEvent, varName: string, varType: 'builtin' | 'collection') {
  e.preventDefault()
  contextMenu.value = {
    show: true,
    x: e.clientX,
    y: e.clientY,
    varName,
    varType
  }
}

function hideContextMenu() {
  contextMenu.value.show = false
}

function insertFromContext(mode: 'get' | 'set' | 'log') {
  if (!isCurrentTabEnabled.value) return

  const { varName, varType } = contextMenu.value
  const getExpr = varType === 'collection'
    ? `wirey.collection.get('${varName}')`
    : `wirey.get('${varName}')`
  const setExpr = varType === 'collection'
    ? `wirey.collection.set('${varName}', value);`
    : `wirey.set('${varName}', value);`

  let snippet: string
  if (mode === 'get') {
    snippet = getExpr + ';'
  } else if (mode === 'set') {
    snippet = setExpr
  } else {
    snippet = `wirey.log('${varName}:', ${getExpr});`
  }

  if (activeScriptTab.value === 'setup') {
    setupScript.value += '\n' + snippet
  } else if (activeScriptTab.value === 'presend') {
    const returnIdx = preSendScript.value.lastIndexOf('return msg;')
    if (returnIdx !== -1) {
      preSendScript.value =
        preSendScript.value.slice(0, returnIdx) +
        snippet + '\n' +
        preSendScript.value.slice(returnIdx)
    } else {
      preSendScript.value += '\n' + snippet
    }
  } else {
    postRecvScript.value += '\n' + snippet
  }

  hideContextMenu()
}

// Watch for external changes to caseData
watch(() => props.caseData.scriptConfig, (newConfig) => {
  if (newConfig) {
    setupEnabled.value = newConfig.setupEnabled
    preSendEnabled.value = newConfig.preSendEnabled
    postRecvEnabled.value = newConfig.postRecvEnabled
    setupScript.value = newConfig.setupScript
    preSendScript.value = newConfig.preSendScript
    postRecvScript.value = newConfig.postRecvScript
  }
}, { deep: true })

// Watch for case change to sync postRecvSample
watch(() => props.caseData.postRecvSample, (newValue) => {
  postRecvSampleMessage.value = newValue || 'OK 00001234 SUCCESS'
})

// Sync local changes back to store
function updateScriptConfig() {
  caseStore.updateCase(props.caseData.id, {
    scriptConfig: {
      setupScript: setupScript.value,
      setupEnabled: setupEnabled.value,
      preSendScript: preSendScript.value,
      preSendEnabled: preSendEnabled.value,
      postRecvScript: postRecvScript.value,
      postRecvEnabled: postRecvEnabled.value
    }
  })
}

// Debounced update for script content
let updateTimeout: ReturnType<typeof setTimeout> | null = null
function debouncedUpdate() {
  if (updateTimeout) clearTimeout(updateTimeout)
  updateTimeout = setTimeout(updateScriptConfig, 300)
}

// Watch for local changes
watch([setupScript, preSendScript, postRecvScript], debouncedUpdate)
watch([setupEnabled, preSendEnabled, postRecvEnabled], updateScriptConfig)

// Built-in variable descriptions (these are auto-available via wirey.get())
const variableSnippets = {
  timestamp: { desc: 'Unix sec' },
  timestamp_ms: { desc: 'Unix ms' },
  datetime: { desc: 'ISO 8601' },
  date: { desc: 'YYYY-MM-DD' },
  time: { desc: 'HH:MM:SS' },
  uuid: { desc: 'UUID v4' },
  random: { desc: '0-999999' },
  counter: { desc: 'Auto incr' }
}

// Check if current tab's script is enabled
const isCurrentTabEnabled = computed(() => {
  if (activeScriptTab.value === 'setup') return setupEnabled.value
  if (activeScriptTab.value === 'presend') return preSendEnabled.value
  return postRecvEnabled.value
})

function formatVarName(name: string): string {
  return `{{${name}}}`
}

function insertSnippet(varName: keyof typeof variableSnippets) {
  if (!isCurrentTabEnabled.value) return

  const snippet = insertAsLog.value
    ? `wirey.log('${varName}:', wirey.get('${varName}'));`
    : `wirey.get('${varName}');`

  if (activeScriptTab.value === 'setup') {
    setupScript.value += '\n' + snippet
  } else if (activeScriptTab.value === 'presend') {
    // Insert before 'return msg;'
    const returnIdx = preSendScript.value.lastIndexOf('return msg;')
    if (returnIdx !== -1) {
      preSendScript.value =
        preSendScript.value.slice(0, returnIdx) +
        snippet + '\n' +
        preSendScript.value.slice(returnIdx)
    } else {
      preSendScript.value += '\n' + snippet
    }
  } else {
    postRecvScript.value += '\n' + snippet
  }
}

// Collection variables
const collectionVariables = computed(() => {
  const collection = caseStore.findCaseCollection(props.caseData.id)
  return collection?.sharedVariables || {}
})

function insertCollectionVar(varName: string) {
  if (!isCurrentTabEnabled.value) return

  const snippet = insertAsLog.value
    ? `wirey.log('${varName}:', wirey.collection.get('${varName}'));`
    : `wirey.collection.get('${varName}');`

  if (activeScriptTab.value === 'setup') {
    setupScript.value += '\n' + snippet
  } else if (activeScriptTab.value === 'presend') {
    const returnIdx = preSendScript.value.lastIndexOf('return msg;')
    if (returnIdx !== -1) {
      preSendScript.value =
        preSendScript.value.slice(0, returnIdx) +
        snippet + '\n' +
        preSendScript.value.slice(returnIdx)
    } else {
      preSendScript.value += '\n' + snippet
    }
  } else {
    postRecvScript.value += '\n' + snippet
  }
}

// CodeMirror extensions - reactive to theme
const extensions = computed(() => {
  return getScriptExtensions(settingsStore.settings.theme === 'dark')
})

// Toggle enable state with auto-focus
function toggleSetup() {
  setupEnabled.value = !setupEnabled.value
  if (setupEnabled.value) {
    setTimeout(() => {
      const view = setupEditorView.value
      if (view) {
        const endPos = view.state.doc.length
        view.dispatch({
          selection: { anchor: endPos, head: endPos }
        })
        view.focus()
      }
    }, 100)
  }
}

function togglePreSend() {
  preSendEnabled.value = !preSendEnabled.value
  if (preSendEnabled.value) {
    setTimeout(() => {
      const view = preSendEditorView.value
      if (view) {
        const content = view.state.doc.toString()
        const returnIdx = content.lastIndexOf('return msg;')
        const cursorPos = returnIdx !== -1 ? returnIdx : content.length
        view.dispatch({
          selection: { anchor: cursorPos, head: cursorPos }
        })
        view.focus()
      }
    }, 100)
  }
}

function togglePostRecv() {
  postRecvEnabled.value = !postRecvEnabled.value
  if (postRecvEnabled.value) {
    setTimeout(() => {
      const view = postRecvEditorView.value
      if (view) {
        const endPos = view.state.doc.length
        view.dispatch({
          selection: { anchor: endPos, head: endPos }
        })
        view.focus()
      }
    }, 100)
  }
}

// Dry Run state
const dryRunLoading = ref(false)
const dryRunResult = ref<{ result: string; logs: string[]; error?: string } | null>(null)

// Get draft message from case
const draftMessage = computed(() => props.caseData.draftMessage || '')

// Sample message for post-recv dry run (persisted per case)
const postRecvSampleMessage = ref(props.caseData.postRecvSample || 'OK 00001234 SUCCESS')

// Sync sample message changes back to store
watch(postRecvSampleMessage, (newValue) => {
  caseStore.updateCase(props.caseData.id, { postRecvSample: newValue })
})

async function runDryRun() {
  dryRunLoading.value = true

  try {
    const collection = caseStore.findCaseCollection(props.caseData.id)

    if (activeScriptTab.value === 'postrecv') {
      // Post-recv dry run
      if (!postRecvSampleMessage.value.trim()) return

      const req = new main.PostRecvDryRunRequest({
        message: postRecvSampleMessage.value,
        caseId: props.caseData.id,
        collectionName: props.caseData.collectionName,
        caseScriptConfig: {
          setupScript: '',
          setupEnabled: false,
          preSendScript: '',
          preSendEnabled: false,
          postRecvScript: postRecvScript.value,
          postRecvEnabled: postRecvEnabled.value
        },
        collectionVariables: collection?.sharedVariables,
        caseVariables: props.caseData.localVariables
      })

      const resp = await DryRunPostRecvScript(req)
      dryRunResult.value = {
        result: resp.result,
        logs: resp.logs || [],
        error: resp.error || undefined
      }
    } else {
      // Setup/Pre-send dry run
      if (!draftMessage.value.trim()) return

      const req = new main.ScriptDryRunRequest({
        message: draftMessage.value,
        caseId: props.caseData.id,
        collectionName: props.caseData.collectionName,
        caseScriptConfig: {
          setupScript: setupScript.value,
          setupEnabled: setupEnabled.value,
          preSendScript: preSendScript.value,
          preSendEnabled: preSendEnabled.value,
          postRecvScript: '',
          postRecvEnabled: false
        },
        collectionVariables: collection?.sharedVariables,
        caseVariables: props.caseData.localVariables
      })

      const resp = await DryRunScript(req)
      dryRunResult.value = {
        result: resp.result,
        logs: resp.logs || [],
        error: resp.error || undefined
      }
    }
  } catch (err) {
    dryRunResult.value = {
      result: '',
      logs: [],
      error: String(err)
    }
  } finally {
    dryRunLoading.value = false
  }
}
</script>

<template>
  <div class="px-4 py-3 space-y-3 h-full flex flex-col relative">
    <!-- Hex mode overlay -->
    <div
      v-if="isHexMode"
      class="absolute inset-0 bg-bg-primary/60 backdrop-blur-sm z-10 flex items-center justify-center rounded-lg"
    >
      <div class="text-center bg-zinc-900/80 backdrop-blur-md px-8 py-5 rounded-2xl shadow-[0_0_20px_rgba(0,0,0,0.3)]">
        <div class="text-white text-sm font-medium">Scripts disabled in HEX mode</div>
        <div class="text-white/60 text-xs mt-1.5">Switch to Text mode to use scripts</div>
      </div>
    </div>

    <!-- Header with Enable toggle -->
    <div class="flex items-center justify-between shrink-0">
      <!-- Segmented control for script type -->
      <div class="flex items-center gap-2">
        <div class="inline-flex rounded-lg bg-bg-tertiary p-[2px] border border-border-custom">
        <button
          @click="activeScriptTab = 'setup'"
          :class="[
            'px-3 py-1.5 text-sm font-medium rounded-md transition-all',
            activeScriptTab === 'setup'
              ? 'bg-accent-primary text-white shadow-sm'
              : 'text-zinc-400 hover:text-zinc-300'
          ]"
        >
          Setup
        </button>
        <button
          @click="activeScriptTab = 'presend'"
          :class="[
            'px-3 py-1.5 text-sm font-medium rounded-md transition-all',
            activeScriptTab === 'presend'
              ? 'bg-accent-primary text-white shadow-sm'
              : 'text-zinc-400 hover:text-zinc-300'
          ]"
        >
          Pre-send
        </button>
        <button
          @click="activeScriptTab = 'postrecv'"
          :class="[
            'px-3 py-1.5 text-sm font-medium rounded-md transition-all',
            activeScriptTab === 'postrecv'
              ? 'bg-accent-primary text-white shadow-sm'
              : 'text-zinc-400 hover:text-zinc-300'
          ]"
        >
          Post-recv
        </button>
        </div>

        <!-- Wirey help button -->
        <button
          @click="showWireyHelp = true"
          class="px-2 py-1.5 text-sm text-zinc-400 hover:text-zinc-200 hover:bg-bg-tertiary rounded-md transition-colors flex items-center gap-1"
          title="Wirey helper functions reference"
        >
          <HelpCircle class="w-4 h-4" />
          <span>wirey</span>
        </button>
      </div>

      <!-- iOS style toggle -->
      <label class="flex items-center gap-2 cursor-pointer">
        <span class="text-sm text-zinc-400">Enable</span>
        <button
          type="button"
          role="switch"
          :aria-checked="isCurrentTabEnabled"
          @click="activeScriptTab === 'setup' ? toggleSetup() : activeScriptTab === 'presend' ? togglePreSend() : togglePostRecv()"
          :class="[
            'relative inline-flex h-6 w-11 items-center rounded-full transition-colors duration-200 ease-in-out focus:outline-none focus:ring-2 focus:ring-accent-primary/50 focus:ring-offset-2 focus:ring-offset-bg-primary',
            isCurrentTabEnabled
              ? 'bg-accent-primary'
              : 'bg-zinc-600'
          ]"
        >
          <span
            :class="[
              'inline-block h-4 w-4 transform rounded-full bg-white transition-transform duration-200 ease-in-out shadow-sm',
              isCurrentTabEnabled
                ? 'translate-x-6'
                : 'translate-x-1'
            ]"
          />
        </button>
      </label>
    </div>

    <!-- Collapsible Help -->
    <div class="shrink-0">
      <button
        @click="toggleHelp"
        class="flex items-center gap-2 text-xs text-zinc-400 hover:text-zinc-300 transition-colors"
      >
        <svg
          :class="['w-3 h-3 transition-transform', helpCollapsed ? '' : 'rotate-90']"
          fill="currentColor"
          viewBox="0 0 20 20"
        >
          <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd"/>
        </svg>
        <span>Help</span>
      </button>
      <div v-if="!helpCollapsed" class="mt-2 bg-bg-tertiary/50 rounded-lg p-3 border border-border-custom">
        <div class="text-xs text-zinc-400 space-y-1">
          <div v-if="activeScriptTab === 'setup'">
            <p>Runs before variable substitution. Use <code class="px-1 py-0.5 bg-bg-tertiary rounded text-accent-primary">wirey.set('key', value)</code> to define custom variables.</p>
          </div>
          <div v-else-if="activeScriptTab === 'presend'">
            <p>Runs after variable substitution. Must <code class="px-1 py-0.5 bg-bg-tertiary rounded text-accent-primary">return</code> the final message. Return <code class="px-1 py-0.5 bg-bg-tertiary rounded text-zinc-400">null</code> to cancel.</p>
          </div>
          <div v-else>
            <p>Runs after receiving a message. Use to extract values: <code class="px-1 py-0.5 bg-bg-tertiary rounded text-accent-primary">msg</code> contains the received message.</p>
          </div>
          <p>Collection shared: <code class="px-1 py-0.5 bg-bg-tertiary rounded text-accent-primary">wirey.collection.get/set('key')</code></p>
          <p class="text-zinc-500">Send: Setup &rarr; Vars &rarr; Pre-send &rarr; Frame &rarr; Send</p>
          <p class="text-zinc-500">Recv: Receive &rarr; Unframe &rarr; Post-recv</p>
        </div>
      </div>
    </div>

    <!-- Editor + Variables Panel -->
    <div class="flex gap-2 flex-1 min-h-0">
      <!-- Script Editor -->
      <div class="flex-1 min-w-0">
        <!-- Setup script editor -->
        <div v-if="activeScriptTab === 'setup'" class="rounded-lg overflow-hidden border border-border-custom h-full">
          <Codemirror
            v-model="setupScript"
            :disabled="!setupEnabled"
            placeholder="// Write your setup script here..."
            :style="{ height: '100%' }"
            :autofocus="false"
            :indent-with-tab="true"
            :tab-size="2"
            :extensions="extensions"
            :class="{ 'opacity-50': !setupEnabled }"
            @ready="onSetupReady"
          />
        </div>

        <!-- Pre-send script editor -->
        <div v-else-if="activeScriptTab === 'presend'" class="rounded-lg overflow-hidden border border-border-custom h-full">
          <Codemirror
            v-model="preSendScript"
            :disabled="!preSendEnabled"
            placeholder="// Write your pre-send script here..."
            :style="{ height: '100%' }"
            :autofocus="false"
            :indent-with-tab="true"
            :tab-size="2"
            :extensions="extensions"
            :class="{ 'opacity-50': !preSendEnabled }"
            @ready="onPreSendReady"
          />
        </div>

        <!-- Post-recv script editor -->
        <div v-else class="rounded-lg overflow-hidden border border-border-custom h-full">
          <Codemirror
            v-model="postRecvScript"
            :disabled="!postRecvEnabled"
            placeholder="// Write your post-recv script here..."
            :style="{ height: '100%' }"
            :autofocus="false"
            :indent-with-tab="true"
            :tab-size="2"
            :extensions="extensions"
            :class="{ 'opacity-50': !postRecvEnabled }"
            @ready="onPostRecvReady"
          />
        </div>
      </div>

      <!-- Collapsible Variables Panel -->
      <div class="shrink-0 flex">
        <!-- Collapsed state: just a button -->
        <button
          v-if="varsCollapsed"
          @click="toggleVars"
          class="w-6 h-full bg-bg-secondary border border-border-custom rounded-lg flex items-center justify-center hover:bg-bg-tertiary transition-colors"
          title="Show Variables"
        >
          <svg class="w-3 h-3 text-zinc-400" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z" clip-rule="evenodd"/>
          </svg>
        </button>

        <!-- Expanded state: variables panel -->
        <div v-else class="w-56 bg-bg-secondary rounded-lg border border-border-custom p-2 flex flex-col">
          <div class="flex items-center justify-between mb-2">
            <!-- Tabs -->
            <div class="flex gap-1">
              <button
                @click="activeVarsTab = 'builtin'"
                :class="[
                  'px-2 py-0.5 text-xs rounded transition-colors',
                  activeVarsTab === 'builtin'
                    ? 'bg-bg-tertiary text-zinc-200'
                    : 'text-zinc-500 hover:text-zinc-300'
                ]"
              >Built-in</button>
              <button
                @click="activeVarsTab = 'collection'"
                :class="[
                  'px-2 py-0.5 text-xs rounded transition-colors',
                  activeVarsTab === 'collection'
                    ? 'bg-bg-tertiary text-zinc-200'
                    : 'text-zinc-500 hover:text-zinc-300'
                ]"
              >Collection</button>
            </div>
            <button
              @click="toggleVars"
              class="text-zinc-500 hover:text-zinc-300 transition-colors"
              title="Hide Variables"
            >
              <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd"/>
              </svg>
            </button>
          </div>
          <!-- Insert as log toggle -->
          <label class="flex items-center gap-1.5 mb-2 cursor-pointer">
            <div
              @click="insertAsLog = !insertAsLog"
              :class="[
                'relative w-7 h-4 rounded-full transition-colors',
                insertAsLog ? 'bg-accent-primary' : 'bg-zinc-600'
              ]"
            >
              <div
                :class="[
                  'absolute top-0.5 w-3 h-3 bg-white rounded-full transition-transform',
                  insertAsLog ? 'translate-x-3.5' : 'translate-x-0.5'
                ]"
              />
            </div>
            <span class="text-xs text-zinc-400">as log</span>
          </label>
          <div class="overflow-y-auto overflow-x-hidden flex-1" :class="{ 'opacity-50': !isCurrentTabEnabled }">
            <!-- Built-in variables -->
            <table v-if="activeVarsTab === 'builtin'" class="w-full text-xs table-fixed">
              <tbody>
                <tr
                  v-for="(snippet, varName) in variableSnippets"
                  :key="varName"
                  @click="insertSnippet(varName as keyof typeof variableSnippets)"
                  @contextmenu="showContextMenu($event, varName as string, 'builtin')"
                  :class="[
                    'cursor-pointer hover:bg-bg-tertiary transition-colors',
                    !isCurrentTabEnabled && 'cursor-not-allowed'
                  ]"
                >
                  <td class="py-1 px-2">
                    <code class="text-accent-primary whitespace-nowrap">{{ formatVarName(varName as string) }}</code>
                  </td>
                  <td class="py-1 px-2 text-zinc-500 text-right whitespace-nowrap">{{ snippet.desc }}</td>
                </tr>
              </tbody>
            </table>
            <!-- Collection variables -->
            <table v-else class="w-full text-xs table-fixed">
              <tbody>
                <tr
                  v-for="(value, varName) in collectionVariables"
                  :key="varName"
                  @click="insertCollectionVar(varName as string)"
                  @contextmenu="showContextMenu($event, varName as string, 'collection')"
                  :class="[
                    'cursor-pointer hover:bg-bg-tertiary transition-colors',
                    !isCurrentTabEnabled && 'cursor-not-allowed'
                  ]"
                >
                  <td class="py-1 px-2">
                    <code class="text-accent-primary whitespace-nowrap">{{ varName }}</code>
                  </td>
                  <td class="py-1 px-2 text-zinc-500 text-right truncate" :title="String(value)">{{ value }}</td>
                </tr>
                <tr v-if="Object.keys(collectionVariables).length === 0">
                  <td colspan="2" class="py-2 px-2 text-zinc-500 text-center">No variables</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- Dry Run Section -->
    <div class="shrink-0 border-t border-border-custom pt-3">
      <div class="flex items-center justify-between mb-2 gap-2">
        <!-- For Setup/Pre-send: Show draft message -->
        <div v-if="activeScriptTab !== 'postrecv'" class="flex items-center gap-2 text-xs text-zinc-400 min-w-0 flex-1">
          <span class="shrink-0">Message:</span>
          <span class="text-zinc-300 truncate">{{ draftMessage || '(empty)' }}</span>
        </div>
        <!-- For Post-recv: Editable sample message -->
        <div v-else class="flex items-center gap-2 text-xs text-zinc-400 min-w-0 flex-1">
          <span class="shrink-0">Sample:</span>
          <input
            v-model="postRecvSampleMessage"
            type="text"
            class="flex-1 min-w-0 px-2 py-1 text-xs text-zinc-300 bg-bg-tertiary border border-border-custom rounded focus:outline-none focus:ring-1 focus:ring-accent-primary"
            placeholder="Enter sample received message..."
          />
        </div>
        <button
          @click="runDryRun"
          :disabled="dryRunLoading || (activeScriptTab === 'postrecv' ? !postRecvSampleMessage.trim() : !draftMessage.trim())"
          class="shrink-0 px-3 py-1.5 text-sm font-medium bg-accent-primary text-white rounded-lg hover:bg-accent-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-1.5"
        >
          <svg v-if="dryRunLoading" class="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
          </svg>
          <span v-else>&#9654;</span>
          Dry Run
        </button>
      </div>

      <!-- Results -->
      <div v-if="dryRunResult" class="space-y-2 text-xs max-h-40 overflow-y-auto">
        <!-- Error -->
        <div v-if="dryRunResult.error" class="bg-red-500/10 border border-red-500/30 rounded-lg p-3">
          <div class="flex items-center gap-2 text-red-400 font-medium mb-1">
            <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
            </svg>
            Error
          </div>
          <pre class="text-red-300 whitespace-pre-wrap break-all font-mono">{{ dryRunResult.error }}</pre>
        </div>

        <!-- Result -->
        <div v-if="!dryRunResult.error" class="bg-green-500/10 border border-green-500/30 rounded-lg p-3">
          <div class="flex items-center gap-2 text-green-400 font-medium mb-1">
            <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
            </svg>
            Result
          </div>
          <pre class="text-zinc-200 whitespace-pre-wrap break-all font-mono">{{ escapeForDisplay(dryRunResult.result) }}</pre>
        </div>

        <!-- Logs -->
        <div v-if="dryRunResult.logs.length > 0" class="bg-amber-500/10 border border-amber-500/30 rounded-lg p-3">
          <div class="flex items-center gap-2 text-amber-400 font-medium mb-1">
            <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"/>
            </svg>
            Logs
          </div>
          <div v-for="(log, i) in dryRunResult.logs" :key="i" class="text-zinc-200 font-mono whitespace-pre-wrap break-all">
            <span class="text-amber-400">»</span> {{ escapeForDisplay(log) }}
          </div>
        </div>
      </div>
    </div>
  </div>

  <!-- Variable Context Menu -->
  <Teleport to="body">
    <div
      v-if="contextMenu.show"
      class="fixed inset-0 z-50"
      @click="hideContextMenu"
      @contextmenu.prevent="hideContextMenu"
    >
      <div
        class="absolute bg-bg-secondary border border-border-custom rounded-lg shadow-lg py-1 min-w-[120px]"
        :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
        @click.stop
      >
        <button
          @click="insertFromContext('get')"
          class="w-full px-3 py-1.5 text-left text-xs text-zinc-300 hover:bg-bg-tertiary transition-colors"
        >get</button>
        <button
          @click="insertFromContext('set')"
          class="w-full px-3 py-1.5 text-left text-xs text-zinc-300 hover:bg-bg-tertiary transition-colors"
        >set</button>
        <button
          @click="insertFromContext('log')"
          class="w-full px-3 py-1.5 text-left text-xs text-zinc-300 hover:bg-bg-tertiary transition-colors"
        >log get</button>
      </div>
    </div>
  </Teleport>

  <!-- Wirey Help Modal -->
  <WireyHelpModal
    :visible="showWireyHelp"
    :insertDisabled="!isCurrentScriptEnabled"
    @close="showWireyHelp = false"
    @insert="handleWireyHelpInsert"
  />
</template>

<style scoped>
/* CodeMirror disabled state */
:deep(.cm-editor.cm-focused) {
  outline: none;
}

:deep(.cm-editor) {
  font-size: 13px;
  height: 100%;
}

:deep(.cm-scroller) {
  font-family: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
}
</style>
