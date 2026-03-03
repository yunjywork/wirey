<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { COMMON_CHARSETS, DEFAULT_CONNECTION_SETTINGS } from '@/types'
import type { Collection, FramingConfig, Charset, ConnectionSettings } from '@/types'
import { useTabStore } from '@/stores/tab'
import { useCaseStore } from '@/stores/case'
import { useSettingsStore } from '@/stores/settings'
import FramingOptions from '@/components/framing/FramingOptions.vue'
import NotesEditor from '@/components/notes/NotesEditor.vue'
import { MdPreview } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'

const props = defineProps<{
  collection: Collection
}>()

const tabStore = useTabStore()
const caseStore = useCaseStore()
const settingsStore = useSettingsStore()

// Auto-save feedback
const showSaved = ref(false)
let savedTimeout: ReturnType<typeof setTimeout> | null = null

function showSavedIndicator() {
  showSaved.value = true
  if (savedTimeout) clearTimeout(savedTimeout)
  savedTimeout = setTimeout(() => {
    showSaved.value = false
  }, 2000)
}

type Tab = 'overview' | 'vars' | 'framing' | 'settings' | 'notes'
const activeTab = ref<Tab>('overview')

// Shared variables - convert from/to object format
const sharedVariables = ref<{ key: string; value: string }[]>([])

// Load variables from collection
function loadVariablesFromCollection() {
  const vars = props.collection.sharedVariables
  if (vars) {
    sharedVariables.value = Object.entries(vars).map(([key, value]) => ({
      key,
      value: String(value)
    }))
  } else {
    sharedVariables.value = []
  }
}

// Initialize from collection's sharedVariables
// Note: variables are updated in real-time via collection:varsUpdated event
watch(() => props.collection.sharedVariables, loadVariablesFromCollection, { immediate: true })

// Parse string value to appropriate type (number, boolean, or string)
function parseValue(value: string): string | number | boolean {
  // Boolean
  if (value === 'true') return true
  if (value === 'false') return false

  // Integer
  if (/^-?\d+$/.test(value)) {
    return parseInt(value, 10)
  }

  // Float
  if (/^-?\d+\.\d+$/.test(value)) {
    return parseFloat(value)
  }

  // String
  return value
}

// Debounced save for variables
let varSaveTimeout: ReturnType<typeof setTimeout> | null = null

async function saveVariables() {
  if (varSaveTimeout) clearTimeout(varSaveTimeout)
  varSaveTimeout = setTimeout(async () => {
    const varsObj: Record<string, unknown> = {}
    for (const v of sharedVariables.value) {
      if (v.key.trim()) {
        varsObj[v.key.trim()] = parseValue(v.value)
      }
    }
    await caseStore.updateCollection(props.collection.name, {
      sharedVariables: Object.keys(varsObj).length > 0 ? varsObj : undefined
    })
    showSavedIndicator()
  }, 500)
}

function addVariable() {
  sharedVariables.value.push({ key: '', value: '' })
}

function removeVariable(index: number) {
  sharedVariables.value.splice(index, 1)
  saveVariables()
}

// Notes handling
async function handleNotesChange(notes: string) {
  await caseStore.updateCollection(props.collection.name, { notes })
  showSavedIndicator()
}

const commonCharsets = COMMON_CHARSETS

// Charset mode: 'global' | 'custom'
type CharsetModeUI = 'global' | 'custom'

const charsetMode = computed<CharsetModeUI>(() => {
  return props.collection.sharedCharset ? 'custom' : 'global'
})

// Custom charset input
const customCharsetInput = ref('')
const showCharsetSuggestions = ref(false)

// Initialize custom input
watch(() => props.collection.sharedCharset, (newVal) => {
  if (newVal) {
    customCharsetInput.value = newVal
  }
}, { immediate: true })

const filteredCharsets = computed(() => {
  const query = customCharsetInput.value.toLowerCase()
  if (!query) return commonCharsets
  return commonCharsets.filter(c => c.includes(query))
})

async function setCharsetMode(mode: CharsetModeUI) {
  if (mode === 'global') {
    await caseStore.updateCollection(props.collection.name, { sharedCharset: undefined })
  } else {
    // custom - use current input or default to utf-8
    const charset = customCharsetInput.value.trim() || 'utf-8'
    customCharsetInput.value = charset
    await caseStore.updateCollection(props.collection.name, { sharedCharset: charset as Charset })
  }
  showSavedIndicator()
}

async function selectCharset(charset: string) {
  customCharsetInput.value = charset
  showCharsetSuggestions.value = false
  await caseStore.updateCollection(props.collection.name, { sharedCharset: charset as Charset })
  showSavedIndicator()
}

function handleCharsetInputBlur() {
  setTimeout(() => {
    showCharsetSuggestions.value = false
  }, 150)
}

async function handleCharsetInputChange() {
  if (customCharsetInput.value.trim()) {
    await caseStore.updateCollection(props.collection.name, { sharedCharset: customCharsetInput.value.trim() as Charset })
    showSavedIndicator()
  }
}

// Framing config with v-model support
const framingConfig = computed({
  get: () => props.collection.sharedFraming || { mode: 'none' as const },
  set: async (value: FramingConfig) => {
    await caseStore.updateCollection(props.collection.name, { sharedFraming: value })
    showSavedIndicator()
  }
})

// ========== Timeout Settings ==========

// Timeout mode: 'global' | 'custom'
type TimeoutModeUI = 'global' | 'custom'

const timeoutMode = computed<TimeoutModeUI>(() => {
  return props.collection.sharedConnectionSettings ? 'custom' : 'global'
})

// Local values for timeout inputs
const connectTimeout = ref(DEFAULT_CONNECTION_SETTINGS.connectTimeout)
const readTimeout = ref(DEFAULT_CONNECTION_SETTINGS.readTimeout)

// Initialize timeout values
watch(() => props.collection.sharedConnectionSettings, (newVal) => {
  if (newVal) {
    connectTimeout.value = newVal.connectTimeout
    readTimeout.value = newVal.readTimeout
  } else {
    connectTimeout.value = settingsStore.connectionSettings.connectTimeout
    readTimeout.value = settingsStore.connectionSettings.readTimeout
  }
}, { immediate: true })

async function setTimeoutMode(mode: TimeoutModeUI) {
  if (mode === 'global') {
    await caseStore.updateCollection(props.collection.name, { sharedConnectionSettings: undefined })
  } else {
    // custom - use current values
    await caseStore.updateCollection(props.collection.name, {
      sharedConnectionSettings: {
        connectTimeout: connectTimeout.value,
        readTimeout: readTimeout.value
      }
    })
  }
  showSavedIndicator()
}

async function updateTimeoutSettings() {
  if (timeoutMode.value === 'custom') {
    await caseStore.updateCollection(props.collection.name, {
      sharedConnectionSettings: {
        connectTimeout: connectTimeout.value,
        readTimeout: readTimeout.value
      }
    })
    showSavedIndicator()
  }
}

async function createNewCase() {
  try {
    await caseStore.createCase(props.collection.name)
  } catch (error) {
    console.error('Failed to create case:', error)
  }
}
</script>

<template>
  <div class="flex-1 flex flex-col h-full overflow-hidden bg-bg-primary">
    <!-- Collection Header -->
    <div class="p-4 border-b border-border-custom bg-bg-secondary/50">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-yellow-500/10 flex items-center justify-center">
          <svg class="w-5 h-5 text-yellow-500" fill="currentColor" viewBox="0 0 24 24">
            <path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
          </svg>
        </div>
        <div>
          <h1 class="text-lg font-semibold text-zinc-100">{{ collection.name }}</h1>
          <p class="text-xs text-zinc-500">{{ collection.cases.length }} cases</p>
        </div>
      </div>
    </div>

    <!-- Tab Bar -->
    <div class="flex border-b border-border-custom bg-bg-secondary/50">
      <button
        @click="activeTab = 'overview'"
        :class="[
          'px-4 py-2.5 text-sm font-medium transition-colors relative',
          activeTab === 'overview'
            ? 'text-accent-primary'
            : 'text-zinc-400 hover:text-zinc-200'
        ]"
      >
        Overview
        <span
          v-if="activeTab === 'overview'"
          class="absolute bottom-0 left-0 right-0 h-0.5 bg-accent-primary"
        ></span>
      </button>
      <button
        @click="activeTab = 'framing'"
        :class="[
          'px-4 py-2.5 text-sm font-medium transition-colors relative',
          activeTab === 'framing'
            ? 'text-accent-primary'
            : 'text-zinc-400 hover:text-zinc-200'
        ]"
      >
        Framing
        <span
          v-if="activeTab === 'framing'"
          class="absolute bottom-0 left-0 right-0 h-0.5 bg-accent-primary"
        ></span>
      </button>
      <button
        @click="activeTab = 'vars'"
        :class="[
          'px-4 py-2.5 text-sm font-medium transition-colors relative',
          activeTab === 'vars'
            ? 'text-accent-primary'
            : 'text-zinc-400 hover:text-zinc-200'
        ]"
      >
        Vars
        <span
          v-if="activeTab === 'vars'"
          class="absolute bottom-0 left-0 right-0 h-0.5 bg-accent-primary"
        ></span>
      </button>
      <button
        @click="activeTab = 'settings'"
        :class="[
          'px-4 py-2.5 text-sm font-medium transition-colors relative',
          activeTab === 'settings'
            ? 'text-accent-primary'
            : 'text-zinc-400 hover:text-zinc-200'
        ]"
      >
        Settings
        <span
          v-if="activeTab === 'settings'"
          class="absolute bottom-0 left-0 right-0 h-0.5 bg-accent-primary"
        ></span>
      </button>
      <button
        @click="activeTab = 'notes'"
        :class="[
          'px-4 py-2.5 text-sm font-medium transition-colors relative',
          activeTab === 'notes'
            ? 'text-accent-primary'
            : 'text-zinc-400 hover:text-zinc-200'
        ]"
      >
        Notes
        <span
          v-if="activeTab === 'notes'"
          class="absolute bottom-0 left-0 right-0 h-0.5 bg-accent-primary"
        ></span>
      </button>

      <!-- Auto-save indicator -->
      <transition name="fade">
        <span
          v-if="showSaved"
          class="ml-auto mr-3 self-center text-xs text-accent-success flex items-center gap-1"
        >
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
          </svg>
          Saved
        </span>
      </transition>
    </div>

    <!-- Tab Content -->
    <div class="flex-1 overflow-y-auto">
      <!-- Overview Tab -->
      <template v-if="activeTab === 'overview'">
        <div class="p-6 flex gap-6 h-full">
          <!-- Left: Cases Section (50%) -->
          <div class="w-1/2 bg-bg-secondary rounded-xl border border-border-custom p-5 flex flex-col">
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-sm font-medium text-zinc-300">Cases</h2>
              <button
                @click="createNewCase"
                class="px-3 py-1.5 text-xs font-medium rounded-lg bg-accent-primary/10 hover:bg-accent-primary/20 text-accent-primary transition-colors"
              >
                + New Case
              </button>
            </div>

            <div v-if="collection.cases.length > 0" class="flex-1 overflow-y-auto space-y-2">
              <div
                v-for="c in collection.cases"
                :key="c.id"
                @click="tabStore.openTab('case', c.id)"
                class="flex items-center gap-3 p-3 rounded-lg bg-bg-tertiary/50 hover:bg-bg-tertiary cursor-pointer transition-colors"
              >
                <div
                  class="w-2 h-2 rounded-full"
                  :class="{
                    'bg-accent-success': c.status === 'connected',
                    'bg-zinc-500': c.status === 'disconnected',
                    'bg-yellow-500': c.status === 'connecting',
                    'bg-accent-error': c.status === 'error'
                  }"
                ></div>
                <span class="text-sm text-zinc-300">{{ c.name }}</span>
                <span class="text-xs text-zinc-500 ml-auto">{{ c.protocol.toUpperCase() }}</span>
                <span class="text-xs text-zinc-600">{{ c.host }}:{{ c.port }}</span>
              </div>
            </div>

            <div v-else class="text-center py-8">
              <p class="text-sm text-zinc-500">No cases yet</p>
              <p class="text-xs text-zinc-600 mt-1">Create a new case to get started</p>
            </div>
          </div>

          <!-- Right: Notes Preview (50%) -->
          <div class="w-1/2 bg-bg-secondary rounded-xl border border-border-custom p-5 flex flex-col">
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-sm font-medium text-zinc-300">Notes</h2>
              <button
                @click="activeTab = 'notes'"
                class="text-xs text-accent-primary hover:text-accent-primary/80 transition-colors"
              >
                Edit
              </button>
            </div>
            <div class="flex-1 overflow-y-auto">
              <MdPreview
                v-if="collection.notes"
                :modelValue="collection.notes"
                :theme="settingsStore.settings.theme"
                class="notes-preview-compact"
              />
              <div v-else class="text-sm text-zinc-500 italic">
                No notes yet. Click "Edit" to add notes.
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- Framing Tab -->
      <template v-else-if="activeTab === 'framing'">
        <div class="p-6">
          <div class="text-xs text-zinc-500 mb-4 p-3 bg-bg-secondary rounded-lg border border-border-custom">
            This is the default Framing for the collection. When you select "Collection Default" in a case, this setting will be applied.
          </div>

          <FramingOptions v-model="framingConfig" />
        </div>
      </template>

      <!-- Vars Tab -->
      <template v-else-if="activeTab === 'vars'">
        <div class="p-6 space-y-4">
          <div>
            <h2 class="text-sm font-medium text-zinc-200">Shared Variables</h2>
            <p class="text-xs text-zinc-500 mt-1">Variables available to all cases in this collection</p>
          </div>

          <div class="bg-bg-secondary rounded-lg border border-border-custom overflow-hidden">
            <!-- Header -->
            <div class="grid grid-cols-[1fr_1fr_40px] gap-2 px-4 py-2 bg-bg-tertiary/50 border-b border-border-custom">
              <span class="text-xs font-medium text-zinc-400">Name</span>
              <span class="text-xs font-medium text-zinc-400">Value</span>
              <span></span>
            </div>

            <!-- Variable rows -->
            <div v-if="sharedVariables.length > 0" class="divide-y divide-border-custom">
              <div
                v-for="(variable, index) in sharedVariables"
                :key="index"
                class="grid grid-cols-[1fr_1fr_40px] gap-2 px-4 py-2 items-center"
              >
                <input
                  v-model="variable.key"
                  @input="saveVariables"
                  type="text"
                  placeholder="name"
                  class="px-2 py-1.5 bg-bg-tertiary border border-border-custom rounded text-sm text-zinc-200 focus:outline-none focus:ring-1 focus:ring-accent-primary/50"
                />
                <input
                  v-model="variable.value"
                  @input="saveVariables"
                  type="text"
                  placeholder="value"
                  class="px-2 py-1.5 bg-bg-tertiary border border-border-custom rounded text-sm text-zinc-200 focus:outline-none focus:ring-1 focus:ring-accent-primary/50"
                />
                <button
                  @click="removeVariable(index)"
                  class="p-1.5 text-zinc-500 hover:text-accent-error hover:bg-accent-error/10 rounded transition-colors"
                >
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                </button>
              </div>
            </div>

            <!-- Empty state -->
            <div v-else class="px-4 py-6 text-center text-sm text-zinc-500">
              No variables defined
            </div>

            <!-- Add button -->
            <div class="px-4 py-2 border-t border-border-custom">
              <button
                @click="addVariable"
                class="flex items-center gap-1 text-xs text-accent-primary hover:text-accent-primary/80 transition-colors"
              >
                <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                </svg>
                Add Variable
              </button>
            </div>
          </div>

          <div class="text-xs text-zinc-500 space-y-1">
            <p>Shared variables are accessible across all cases in this collection.</p>
            <p>In scripts: <code class="px-1 py-0.5 bg-bg-tertiary rounded text-accent-primary">wirey.collection.get('name')</code> / <code class="px-1 py-0.5 bg-bg-tertiary rounded text-accent-primary">wirey.collection.set('name', value)</code></p>
            <p>To use <code class="px-1 py-0.5 bg-bg-tertiary rounded text-accent-primary" v-pre>{{name}}</code> in messages, map to case: <code class="px-1 py-0.5 bg-bg-tertiary rounded text-accent-primary">wirey.set('name', wirey.collection.get('name'))</code></p>
          </div>
        </div>
      </template>

      <!-- Settings Tab -->
      <template v-else-if="activeTab === 'settings'">
        <div class="p-6 space-y-6">
          <div class="text-xs text-zinc-500 mb-4 p-3 bg-bg-secondary rounded-lg border border-border-custom">
            This is the default Charset for the collection. When you select "Collection" in a case, this setting will be applied.
          </div>

          <!-- Charset Setting -->
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <h3 class="text-sm font-medium text-zinc-200">Character Encoding</h3>
              <span class="text-xs text-zinc-500">
                Current: <span class="text-accent-primary">{{ collection.sharedCharset?.toUpperCase() || 'Global (' + settingsStore.defaultCharset + ')' }}</span>
              </span>
            </div>

            <p class="text-xs text-zinc-500">
              Set a shared charset for all cases in this collection. Cases can override this setting individually.
            </p>

            <!-- Mode selector -->
            <div class="space-y-2">
              <!-- Global Default -->
              <button
                @click="setCharsetMode('global')"
                :class="[
                  'w-full px-4 py-3 rounded-lg text-left transition-all border',
                  charsetMode === 'global'
                    ? 'bg-accent-primary/20 border-accent-primary text-zinc-100'
                    : 'bg-bg-tertiary border-border-custom text-zinc-400 hover:bg-bg-tertiary/80 hover:text-zinc-200'
                ]"
              >
                <div class="text-sm font-medium">Global Default</div>
                <div class="text-xs text-zinc-500">Use app default ({{ settingsStore.defaultCharset }})</div>
              </button>

              <!-- Custom -->
              <div
                :class="[
                  'w-full px-4 py-3 rounded-lg transition-all border',
                  charsetMode === 'custom'
                    ? 'bg-accent-primary/20 border-accent-primary'
                    : 'bg-bg-tertiary border-border-custom'
                ]"
              >
                <button
                  @click="setCharsetMode('custom')"
                  class="w-full text-left"
                >
                  <div :class="['text-sm font-medium', charsetMode === 'custom' ? 'text-zinc-100' : 'text-zinc-400']">
                    Custom
                  </div>
                  <div class="text-xs text-zinc-500">Override with specific charset for this collection</div>
                </button>

                <!-- Custom charset input with autocomplete -->
                <div v-if="charsetMode === 'custom'" class="mt-3 relative">
                  <input
                    v-model="customCharsetInput"
                    @focus="showCharsetSuggestions = true"
                    @blur="handleCharsetInputBlur"
                    @input="handleCharsetInputChange"
                    type="text"
                    placeholder="e.g. utf-8, euc-kr..."
                    class="w-full px-3 py-2 bg-bg-primary border border-border-custom rounded-lg
                           text-zinc-200 placeholder-zinc-500 text-sm
                           focus:outline-none focus:ring-2 focus:ring-accent-primary/50"
                  />
                  <!-- Suggestions dropdown -->
                  <div
                    v-if="showCharsetSuggestions && filteredCharsets.length > 0"
                    class="absolute top-full left-0 right-0 mt-1 bg-bg-tertiary border border-border-custom rounded-lg shadow-xl z-10 max-h-48 overflow-y-auto"
                  >
                    <button
                      v-for="cs in filteredCharsets"
                      :key="cs"
                      @mousedown.prevent="selectCharset(cs)"
                      :class="[
                        'w-full px-4 py-2 text-left text-sm hover:bg-accent-primary/20 transition-colors',
                        cs === customCharsetInput ? 'text-accent-primary' : 'text-zinc-300'
                      ]"
                    >
                      {{ cs }}
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Info box -->
          <div class="bg-bg-tertiary/50 rounded-lg p-3 border border-border-custom">
            <div class="flex items-start gap-2">
              <svg class="w-4 h-4 text-zinc-500 mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
              <div class="text-xs text-zinc-500">
                <p>Charset inheritance: Global → Collection → Case</p>
                <p class="mt-1">Cases using "Collection" charset mode will use this setting.</p>
              </div>
            </div>
          </div>

          <!-- Timeout Settings -->
          <div class="space-y-3 mt-6 pt-6 border-t border-border-custom">
            <div class="flex items-center justify-between">
              <h3 class="text-sm font-medium text-zinc-200">Timeout Settings</h3>
              <span class="text-xs text-zinc-500">
                Current: <span class="text-accent-primary">{{ collection.sharedConnectionSettings ? 'Custom' : 'Global' }}</span>
              </span>
            </div>

            <p class="text-xs text-zinc-500">
              Set shared timeout settings for all cases in this collection. Cases can override this setting individually.
            </p>

            <!-- Mode selector -->
            <div class="space-y-2">
              <!-- Global Default -->
              <button
                @click="setTimeoutMode('global')"
                :class="[
                  'w-full px-4 py-3 rounded-lg text-left transition-all border',
                  timeoutMode === 'global'
                    ? 'bg-accent-primary/20 border-accent-primary text-zinc-100'
                    : 'bg-bg-tertiary border-border-custom text-zinc-400 hover:bg-bg-tertiary/80 hover:text-zinc-200'
                ]"
              >
                <div class="text-sm font-medium">Global Default</div>
                <div class="text-xs text-zinc-500">
                  Use app defaults ({{ settingsStore.connectionSettings.connectTimeout }}ms / {{ settingsStore.connectionSettings.readTimeout }}ms)
                </div>
              </button>

              <!-- Custom -->
              <div
                :class="[
                  'w-full px-4 py-3 rounded-lg transition-all border',
                  timeoutMode === 'custom'
                    ? 'bg-accent-primary/20 border-accent-primary'
                    : 'bg-bg-tertiary border-border-custom'
                ]"
              >
                <button
                  @click="setTimeoutMode('custom')"
                  class="w-full text-left"
                >
                  <div :class="['text-sm font-medium', timeoutMode === 'custom' ? 'text-zinc-100' : 'text-zinc-400']">
                    Custom
                  </div>
                  <div class="text-xs text-zinc-500">Override with specific timeout for this collection</div>
                </button>

                <!-- Custom timeout inputs -->
                <div v-if="timeoutMode === 'custom'" class="mt-3 space-y-3">
                  <!-- Connect Timeout -->
                  <div class="space-y-1">
                    <label class="text-xs text-zinc-400">Connect Timeout</label>
                    <div class="flex items-center gap-2">
                      <input
                        v-model.number="connectTimeout"
                        @change="updateTimeoutSettings"
                        type="number"
                        min="0"
                        step="1000"
                        class="flex-1 px-3 py-2 bg-bg-primary border border-border-custom rounded-lg
                               text-zinc-200 text-sm focus:outline-none focus:ring-2 focus:ring-accent-primary/50"
                      />
                      <span class="text-xs text-zinc-500 w-8">ms</span>
                    </div>
                  </div>

                  <!-- Read Timeout -->
                  <div class="space-y-1">
                    <label class="text-xs text-zinc-400">Read Timeout</label>
                    <div class="flex items-center gap-2">
                      <input
                        v-model.number="readTimeout"
                        @change="updateTimeoutSettings"
                        type="number"
                        min="0"
                        step="1000"
                        class="flex-1 px-3 py-2 bg-bg-primary border border-border-custom rounded-lg
                               text-zinc-200 text-sm focus:outline-none focus:ring-2 focus:ring-accent-primary/50"
                      />
                      <span class="text-xs text-zinc-500 w-8">ms</span>
                    </div>
                    <p class="text-xs text-zinc-600">0 = unlimited</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- Notes Tab -->
      <template v-else-if="activeTab === 'notes'">
        <div class="h-full">
          <NotesEditor
            :modelValue="collection.notes || ''"
            @update:modelValue="handleNotesChange"
          />
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

<style>
/* Notes preview compact styling */
.notes-preview-compact {
  background: transparent !important;
  font-family: 'Malgun Gothic', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif !important;
}

.notes-preview-compact .md-editor-preview-wrapper {
  padding: 0 !important;
}

.notes-preview-compact .md-editor-preview {
  font-size: 0.875rem !important;
  color: rgb(var(--color-text-primary)) !important;
  font-family: 'Malgun Gothic', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif !important;
}
</style>
