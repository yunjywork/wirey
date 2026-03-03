<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { COMMON_CHARSETS, DEFAULT_CONNECTION_SETTINGS } from '@/types'
import type { Case, Charset, CharsetMode, ConnectionSettings } from '@/types'
import { useCaseStore } from '@/stores/case'
import { useSettingsStore } from '@/stores/settings'
import { UpdateCharset } from '../../../wailsjs/go/main/App'

const props = defineProps<{
  caseData: Case
}>()

const caseStore = useCaseStore()
const settingsStore = useSettingsStore()

// Check if case is connected (disable editing)
const isDisabled = computed(() =>
  props.caseData.status === 'connected' || props.caseData.status === 'connecting'
)

const commonCharsets = COMMON_CHARSETS

// Charset mode: 'collection' | 'custom'
type CharsetModeUI = 'collection' | 'custom'

const charsetMode = computed<CharsetModeUI>(() => {
  const current = props.caseData.charset
  if (!current || current === 'global' || current === 'collection') return 'collection'
  return 'custom'
})

// Custom charset input
const customCharsetInput = ref('')
const showSuggestions = ref(false)

// Initialize custom input when switching to custom mode
watch(() => props.caseData.charset, (newVal) => {
  if (newVal && newVal !== 'global' && newVal !== 'collection') {
    customCharsetInput.value = newVal
  }
}, { immediate: true })

const filteredCharsets = computed(() => {
  const query = customCharsetInput.value.toLowerCase()
  if (!query) return commonCharsets
  return commonCharsets.filter(c => c.includes(query))
})

async function setMode(mode: CharsetModeUI) {
  if (mode === 'collection') {
    caseStore.updateCase(props.caseData.id, { charset: 'collection' })
  } else {
    // custom - use current input or default to utf-8
    const charset = customCharsetInput.value.trim() || 'utf-8'
    customCharsetInput.value = charset
    caseStore.updateCase(props.caseData.id, { charset: charset as CharsetMode })
  }

  // Update backend connection if connected
  if (props.caseData.status === 'connected') {
    try {
      await UpdateCharset(props.caseData.id, effectiveCharset.value)
    } catch (e) {
      console.error('Failed to update charset:', e)
    }
  }
}

async function selectCharset(charset: string) {
  customCharsetInput.value = charset
  showSuggestions.value = false
  caseStore.updateCase(props.caseData.id, { charset: charset as CharsetMode })

  // Update backend connection if connected
  if (props.caseData.status === 'connected') {
    try {
      await UpdateCharset(props.caseData.id, charset)
    } catch (e) {
      console.error('Failed to update charset:', e)
    }
  }
}

function handleInputBlur() {
  setTimeout(() => {
    showSuggestions.value = false
  }, 150)
}

async function handleInputChange() {
  if (customCharsetInput.value.trim()) {
    const charset = customCharsetInput.value.trim()
    caseStore.updateCase(props.caseData.id, { charset: charset as CharsetMode })

    // Update backend connection if connected
    if (props.caseData.status === 'connected') {
      try {
        await UpdateCharset(props.caseData.id, charset)
      } catch (e) {
        console.error('Failed to update charset:', e)
      }
    }
  }
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

// Get collection charset for display
const collectionCharset = computed(() => {
  const collection = caseStore.findCaseCollection(props.caseData.id)
  return collection?.sharedCharset || null
})

// ========== Connection Settings ==========

// Timeout mode: 'collection' | 'custom'
type TimeoutModeUI = 'collection' | 'custom'

const timeoutMode = computed<TimeoutModeUI>(() => {
  return props.caseData.connectionSettings ? 'custom' : 'collection'
})

// Get inherited connection settings (from collection or global)
const inheritedConnectionSettings = computed((): ConnectionSettings => {
  const collection = caseStore.findCaseCollection(props.caseData.id)
  if (collection?.sharedConnectionSettings) {
    return collection.sharedConnectionSettings
  }
  return settingsStore.connectionSettings
})

// Get effective connection settings (resolved through hierarchy)
const effectiveConnectionSettings = computed((): ConnectionSettings => {
  if (props.caseData.connectionSettings) {
    return props.caseData.connectionSettings
  }
  return inheritedConnectionSettings.value
})

// Local values for inputs
const connectTimeout = ref(effectiveConnectionSettings.value.connectTimeout)
const readTimeout = ref(effectiveConnectionSettings.value.readTimeout)

// Sync with props
watch(() => props.caseData.connectionSettings, () => {
  connectTimeout.value = effectiveConnectionSettings.value.connectTimeout
  readTimeout.value = effectiveConnectionSettings.value.readTimeout
}, { immediate: true })

function setTimeoutMode(mode: TimeoutModeUI) {
  if (mode === 'collection') {
    // Reset to inherit
    caseStore.updateCase(props.caseData.id, {
      connectionSettings: undefined
    })
    connectTimeout.value = inheritedConnectionSettings.value.connectTimeout
    readTimeout.value = inheritedConnectionSettings.value.readTimeout
  } else {
    // Custom - use current values
    caseStore.updateCase(props.caseData.id, {
      connectionSettings: {
        connectTimeout: connectTimeout.value,
        readTimeout: readTimeout.value
      }
    })
  }
}

function updateConnectionSettings() {
  caseStore.updateCase(props.caseData.id, {
    connectionSettings: {
      connectTimeout: connectTimeout.value,
      readTimeout: readTimeout.value
    }
  })
}

// Get collection timeout for display
const collectionTimeout = computed(() => {
  const collection = caseStore.findCaseCollection(props.caseData.id)
  return collection?.sharedConnectionSettings || null
})
</script>

<template>
  <div class="p-4 space-y-6">
    <!-- Lock notice when connected -->
    <div v-if="isDisabled" class="p-3 bg-yellow-900/20 border border-yellow-700/50 rounded-lg text-xs text-yellow-400">
      Cannot change charset while connected. Disconnect first to modify settings.
    </div>

    <!-- Case Charset Setting -->
    <div class="space-y-3">
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-medium text-zinc-200">Character Encoding</h3>
        <span class="text-xs text-zinc-500">
          Effective: <span class="text-accent-primary">{{ effectiveCharset.toUpperCase() }}</span>
        </span>
      </div>

      <p class="text-xs text-zinc-500">
        Select how text is encoded/decoded for this case. Inheritance: Global → Collection → Case
      </p>

      <!-- Mode selector -->
      <div class="space-y-2">
        <!-- Collection -->
        <button
          @click="setMode('collection')"
          :disabled="isDisabled"
          :class="[
            'w-full px-4 py-3 rounded-lg text-left transition-all border',
            charsetMode === 'collection'
              ? 'bg-accent-primary/20 border-accent-primary text-zinc-100'
              : 'bg-bg-tertiary border-border-custom text-zinc-400 hover:bg-bg-tertiary/80 hover:text-zinc-200',
            isDisabled ? 'opacity-60 cursor-not-allowed' : ''
          ]"
        >
          <div class="text-sm font-medium">Collection</div>
          <div class="text-xs text-zinc-500">
            Use collection setting ({{ collectionCharset || `not set → global (${settingsStore.defaultCharset})` }})
          </div>
        </button>

        <!-- Custom -->
        <div
          :class="[
            'w-full px-4 py-3 rounded-lg transition-all border',
            charsetMode === 'custom'
              ? 'bg-accent-primary/20 border-accent-primary'
              : 'bg-bg-tertiary border-border-custom',
            isDisabled ? 'opacity-60' : ''
          ]"
        >
          <button
            @click="setMode('custom')"
            :disabled="isDisabled"
            :class="['w-full text-left', isDisabled ? 'cursor-not-allowed' : '']"
          >
            <div :class="['text-sm font-medium', charsetMode === 'custom' ? 'text-zinc-100' : 'text-zinc-400']">
              Custom
            </div>
            <div class="text-xs text-zinc-500">Override with specific charset</div>
          </button>

          <!-- Custom charset input with autocomplete -->
          <div v-if="charsetMode === 'custom'" class="mt-3 relative">
            <input
              v-model="customCharsetInput"
              @focus="showSuggestions = true"
              @blur="handleInputBlur"
              @input="handleInputChange"
              :disabled="isDisabled"
              type="text"
              placeholder="e.g. utf-8, euc-kr..."
              class="w-full px-3 py-2 bg-bg-primary border border-border-custom rounded-lg
                     text-zinc-200 placeholder-zinc-500 text-sm
                     focus:outline-none focus:ring-2 focus:ring-accent-primary/50
                     disabled:opacity-50 disabled:cursor-not-allowed"
            />
            <!-- Suggestions dropdown -->
            <div
              v-if="showSuggestions && filteredCharsets.length > 0"
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
          <p>Charset affects how text messages are converted to bytes when sending, and how received bytes are decoded to text.</p>
          <p class="mt-1">For binary protocols, use HEX format in the message input instead.</p>
        </div>
      </div>
    </div>

    <!-- Timeout Settings -->
    <div class="space-y-3">
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-medium text-zinc-200">Timeout</h3>
        <span class="text-xs text-zinc-500">
          Effective: <span class="text-accent-primary">{{ effectiveConnectionSettings.connectTimeout }}ms / {{ effectiveConnectionSettings.readTimeout }}ms</span>
        </span>
      </div>

      <p class="text-xs text-zinc-500">
        Connection timeout settings. Inheritance: Global → Collection → Case
      </p>

      <!-- Mode selector -->
      <div class="space-y-2">
        <!-- Collection -->
        <button
          @click="setTimeoutMode('collection')"
          :disabled="isDisabled"
          :class="[
            'w-full px-4 py-3 rounded-lg text-left transition-all border',
            timeoutMode === 'collection'
              ? 'bg-accent-primary/20 border-accent-primary text-zinc-100'
              : 'bg-bg-tertiary border-border-custom text-zinc-400 hover:bg-bg-tertiary/80 hover:text-zinc-200',
            isDisabled ? 'opacity-60 cursor-not-allowed' : ''
          ]"
        >
          <div class="text-sm font-medium">Collection</div>
          <div class="text-xs text-zinc-500">
            Use collection setting ({{ collectionTimeout ? `${collectionTimeout.connectTimeout}ms / ${collectionTimeout.readTimeout}ms` : `not set → global (${settingsStore.connectionSettings.connectTimeout}ms / ${settingsStore.connectionSettings.readTimeout}ms)` }})
          </div>
        </button>

        <!-- Custom -->
        <div
          :class="[
            'w-full px-4 py-3 rounded-lg transition-all border',
            timeoutMode === 'custom'
              ? 'bg-accent-primary/20 border-accent-primary'
              : 'bg-bg-tertiary border-border-custom',
            isDisabled ? 'opacity-60' : ''
          ]"
        >
          <button
            @click="setTimeoutMode('custom')"
            :disabled="isDisabled"
            :class="['w-full text-left', isDisabled ? 'cursor-not-allowed' : '']"
          >
            <div :class="['text-sm font-medium', timeoutMode === 'custom' ? 'text-zinc-100' : 'text-zinc-400']">
              Custom
            </div>
            <div class="text-xs text-zinc-500">Override with specific timeout</div>
          </button>

          <!-- Custom timeout inputs -->
          <div v-if="timeoutMode === 'custom'" class="mt-3 space-y-3">
            <!-- Connect Timeout -->
            <div class="space-y-1">
              <label class="text-xs text-zinc-400">Connect Timeout</label>
              <div class="flex items-center gap-2">
                <input
                  v-model.number="connectTimeout"
                  @change="updateConnectionSettings"
                  :disabled="isDisabled"
                  type="number"
                  min="0"
                  step="1000"
                  class="flex-1 px-3 py-2 bg-bg-primary border border-border-custom rounded-lg
                         text-zinc-200 text-sm
                         focus:outline-none focus:ring-2 focus:ring-accent-primary/50
                         disabled:opacity-50 disabled:cursor-not-allowed"
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
                  @change="updateConnectionSettings"
                  :disabled="isDisabled"
                  type="number"
                  min="0"
                  step="1000"
                  class="flex-1 px-3 py-2 bg-bg-primary border border-border-custom rounded-lg
                         text-zinc-200 text-sm
                         focus:outline-none focus:ring-2 focus:ring-accent-primary/50
                         disabled:opacity-50 disabled:cursor-not-allowed"
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
