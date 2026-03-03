<script setup lang="ts">
import { ref, watch, computed, nextTick } from 'vue'
import type { Case, Protocol } from '@/types'
import { useCaseStore } from '@/stores/case'
import { useSettingsStore } from '@/stores/settings'
import { Connect, Disconnect } from '../../../wailsjs/go/main/App'
import { toBackendFramingConfig, toBackendConnectionSettings } from '@/utils/backend'
import { DEFAULT_CONNECTION_SETTINGS } from '@/types'
import type { ConnectionSettings } from '@/types'

const props = defineProps<{
  caseData: Case
}>()

const caseStore = useCaseStore()
const settingsStore = useSettingsStore()

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

// Get effective framing (resolved through hierarchy)
const effectiveFraming = computed(() => {
  const caseFraming = props.caseData.framing

  // If mode is 'collection', use collection's shared framing
  if (caseFraming.mode === 'collection') {
    const collection = caseStore.findCaseCollection(props.caseData.id)
    if (collection?.sharedFraming) {
      return collection.sharedFraming
    }
    // Fallback to 'none' if collection has no shared framing
    return { mode: 'none' as const }
  }

  return caseFraming
})

// Get effective connection settings (resolved through hierarchy)
const effectiveConnectionSettings = computed((): ConnectionSettings => {
  // Case-level settings take priority
  if (props.caseData.connectionSettings) {
    return props.caseData.connectionSettings
  }

  // Collection-level settings
  const collection = caseStore.findCaseCollection(props.caseData.id)
  if (collection?.sharedConnectionSettings) {
    return collection.sharedConnectionSettings
  }

  // Global defaults
  return settingsStore.connectionSettings
})

// Local form state
const protocol = ref<Protocol>(props.caseData.protocol)
const host = ref(props.caseData.host)
const port = ref(props.caseData.port)
const isLoading = ref(false)

// Name editing
const isEditingName = ref(false)
const editName = ref('')
const nameInput = ref<HTMLInputElement | null>(null)

// Save modal
const showSaveModal = ref(false)
const modalName = ref('')
const modalInput = ref<HTMLInputElement | null>(null)
const isSaving = ref(false)

// Check if name is untitled
const isUntitledName = computed(() => /^Untitled \d+$/.test(props.caseData.name))

// Watch for case changes
watch(() => props.caseData, (newCase) => {
  protocol.value = newCase.protocol
  host.value = newCase.host
  port.value = newCase.port
}, { immediate: true })

// Sync changes back to store
watch([protocol, host, port], () => {
  if (props.caseData.status === 'disconnected') {
    caseStore.updateCase(props.caseData.id, {
      protocol: protocol.value,
      host: host.value,
      port: port.value
    })
  }
})

async function handleConnect() {
  if (isLoading.value) return

  isLoading.value = true
  caseStore.updateCaseStatus(props.caseData.id, 'connecting')

  try {
    const framingConfig = toBackendFramingConfig(effectiveFraming.value)
    const connSettings = toBackendConnectionSettings(effectiveConnectionSettings.value)
    await Connect(props.caseData.id, protocol.value, host.value, port.value, framingConfig, effectiveCharset.value, connSettings)
    caseStore.updateCaseStatus(props.caseData.id, 'connected')
  } catch (error) {
    console.error('Connection failed:', error)
    caseStore.updateCaseStatus(props.caseData.id, 'error')
    caseStore.addMessage(props.caseData.id, 'received', `Error: ${error}`, 'text')
  } finally {
    isLoading.value = false
  }
}

async function handleDisconnect() {
  if (isLoading.value) return

  isLoading.value = true

  try {
    await Disconnect(props.caseData.id)
    caseStore.updateCaseStatus(props.caseData.id, 'disconnected')
  } catch (error) {
    console.error('Disconnect failed:', error)
  } finally {
    isLoading.value = false
  }
}

const isConnected = () => props.caseData.status === 'connected'
const isConnecting = () => props.caseData.status === 'connecting'

// Name editing functions
async function startEditName() {
  editName.value = props.caseData.name
  isEditingName.value = true
  await nextTick()
  nameInput.value?.focus()
  nameInput.value?.select()
}

function saveEditName() {
  if (editName.value.trim()) {
    caseStore.updateCase(props.caseData.id, { name: editName.value.trim() })
  }
  isEditingName.value = false
}

function cancelEditName() {
  isEditingName.value = false
}

function handleNameKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    saveEditName()
  } else if (e.key === 'Escape') {
    cancelEditName()
  }
}

// Save functions
async function handleSave() {
  if (isUntitledName.value) {
    modalName.value = ''
    showSaveModal.value = true
    await nextTick()
    modalInput.value?.focus()
  } else {
    await doSave()
  }
}

async function doSave(name?: string) {
  if (isSaving.value) return

  isSaving.value = true
  try {
    if (name) {
      caseStore.updateCase(props.caseData.id, { name })
    }
    await caseStore.saveCase(props.caseData.id)
    showSaveModal.value = false
  } catch (error) {
    console.error('Save failed:', error)
  } finally {
    isSaving.value = false
  }
}

function handleModalKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && modalName.value.trim()) {
    doSave(modalName.value.trim())
  } else if (e.key === 'Escape') {
    showSaveModal.value = false
  }
}
</script>

<template>
  <div class="p-4 border-b border-border-custom bg-bg-secondary">
    <!-- Case name with edit and save -->
    <div class="mb-3 flex items-center gap-3 min-h-[40px]">
      <!-- Editable name -->
      <div v-if="isEditingName" class="flex items-center gap-2">
        <input
          ref="nameInput"
          v-model="editName"
          @keydown="handleNameKeydown"
          @blur="saveEditName"
          class="px-2 py-1 text-lg font-medium bg-bg-tertiary border border-accent-primary rounded text-zinc-100 focus:outline-none"
        />
      </div>
      <h2 v-else class="text-lg font-medium text-zinc-100">{{ props.caseData.name }}</h2>

      <!-- Edit name button -->
      <button
        v-if="!isEditingName"
        @click="startEditName"
        class="p-1.5 rounded-lg hover:bg-bg-tertiary text-zinc-500 hover:text-zinc-200 transition-colors"
        title="Edit name"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
        </svg>
      </button>

      <!-- Save button / Saved indicator (same position) -->
      <button
        v-if="!props.caseData.isSaved"
        @click="handleSave"
        :disabled="isSaving"
        class="ml-auto px-3 py-1.5 text-sm rounded-lg bg-accent-primary/10 hover:bg-accent-primary/20 text-accent-primary transition-colors flex items-center gap-1.5"
      >
        <svg v-if="isSaving" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
        </svg>
        <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4"/>
        </svg>
        Save
      </button>
      <span v-else class="ml-auto flex items-center gap-1 text-xs text-accent-success px-3 py-1.5">
        <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 24 24">
          <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
        </svg>
        Saved
      </span>
    </div>

    <div class="flex items-end gap-4">
      <!-- Protocol Select -->
      <div class="w-28">
        <label class="block text-xs font-medium text-zinc-400 mb-1.5">Protocol</label>
        <select
          v-model="protocol"
          :disabled="isConnected() || isConnecting()"
          class="input"
        >
          <option value="tcp">TCP</option>
          <option value="udp">UDP</option>
        </select>
      </div>

      <!-- Host Input -->
      <div class="flex-1">
        <label class="block text-xs font-medium text-zinc-400 mb-1.5">Host</label>
        <input
          v-model="host"
          type="text"
          placeholder="127.0.0.1"
          :disabled="isConnected() || isConnecting()"
          class="input font-mono"
        />
      </div>

      <!-- Port Input -->
      <div class="w-28">
        <label class="block text-xs font-medium text-zinc-400 mb-1.5">Port</label>
        <input
          v-model.number="port"
          type="number"
          placeholder="8080"
          min="1"
          max="65535"
          :disabled="isConnected() || isConnecting()"
          class="input font-mono"
        />
      </div>

      <!-- Connect/Disconnect Buttons -->
      <div class="flex gap-2">
        <button
          v-if="!isConnected()"
          @click="handleConnect"
          :disabled="isConnecting() || !host || !port"
          class="btn-success flex items-center gap-2"
        >
          <svg v-if="isConnecting()" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
          </svg>
          <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
          </svg>
          {{ isConnecting() ? 'Connecting...' : 'Connect' }}
        </button>

        <button
          v-else
          @click="handleDisconnect"
          :disabled="isLoading"
          class="btn-danger flex items-center gap-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
          Disconnect
        </button>
      </div>
    </div>

    <!-- Status indicator -->
    <div class="mt-3 flex items-center gap-2">
      <span class="status-dot" :class="{
        'status-connected': props.caseData.status === 'connected',
        'status-connecting': props.caseData.status === 'connecting',
        'status-disconnected': props.caseData.status === 'disconnected',
        'bg-accent-error': props.caseData.status === 'error'
      }"></span>
      <span class="text-xs text-zinc-400 capitalize">{{ props.caseData.status }}</span>
      <!-- Protocol badge -->
      <span v-if="isConnected()" class="text-[10px] font-medium px-1.5 py-0.5 rounded bg-accent-primary/20 text-accent-primary">
        {{ protocol.toUpperCase() }}
      </span>
      <!-- Address info -->
      <span v-if="isConnected() && props.caseData.localAddr" class="text-xs text-zinc-400 font-mono">
        {{ props.caseData.localAddr }} <span class="text-accent-primary">→</span> {{ host }}:{{ port }}
      </span>
    </div>

    <!-- Save Modal -->
    <Teleport to="body">
      <div
        v-if="showSaveModal"
        class="fixed inset-0 z-50 flex items-center justify-center"
        @click.self="showSaveModal = false"
      >
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm"></div>

        <!-- Modal -->
        <div class="relative bg-bg-secondary border border-border-custom rounded-xl shadow-2xl w-full max-w-md p-6">
          <h3 class="text-lg font-medium text-zinc-100 mb-2">Save Case</h3>
          <p class="text-sm text-zinc-400 mb-4">Enter a name for this case</p>

          <input
            ref="modalInput"
            v-model="modalName"
            @keydown="handleModalKeydown"
            type="text"
            placeholder="Case name..."
            class="w-full px-4 py-3 bg-bg-tertiary border border-border-custom rounded-lg
                   text-zinc-200 placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-accent-primary/50 mb-4"
          />

          <div class="flex justify-end gap-3">
            <button
              @click="showSaveModal = false"
              class="px-4 py-2 text-sm rounded-lg bg-bg-tertiary hover:bg-border-custom text-zinc-400 transition-colors"
            >
              Cancel
            </button>
            <button
              @click="doSave(modalName.trim())"
              :disabled="!modalName.trim() || isSaving"
              class="px-4 py-2 text-sm rounded-lg bg-accent-primary hover:bg-accent-primary/80 text-white disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center gap-2"
            >
              <svg v-if="isSaving" class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
              </svg>
              Save
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
