<script setup lang="ts">
import { ref, computed } from 'vue'
import { useTabStore } from '@/stores/tab'
import { useSettingsStore } from '@/stores/settings'
import { useEchoStore } from '@/stores/echo'
import { COMMON_CHARSETS, DEFAULT_CONNECTION_SETTINGS } from '@/types'
import type { Charset } from '@/types'
import { SaveSettings as SaveSettingsApi } from '../../../wailsjs/go/main/App'

const tabStore = useTabStore()
const settingsStore = useSettingsStore()
const echoStore = useEchoStore()

const showSettingsModal = ref(false)
const charsetInput = ref('')
const showSuggestions = ref(false)

// Timeout settings
const connectTimeout = ref(DEFAULT_CONNECTION_SETTINGS.connectTimeout)
const readTimeout = ref(DEFAULT_CONNECTION_SETTINGS.readTimeout)

const commonCharsets = COMMON_CHARSETS

const filteredCharsets = computed(() => {
  const query = charsetInput.value.toLowerCase()
  if (!query) return commonCharsets
  return commonCharsets.filter(c => c.includes(query))
})

function selectCharset(charset: string) {
  charsetInput.value = charset
  showSuggestions.value = false
}

function handleInputBlur() {
  // Delay to allow click on suggestion
  setTimeout(() => {
    showSuggestions.value = false
  }, 150)
}

function openModal() {
  charsetInput.value = settingsStore.defaultCharset
  connectTimeout.value = settingsStore.connectionSettings.connectTimeout
  readTimeout.value = settingsStore.connectionSettings.readTimeout
  showSettingsModal.value = true
}

function cancelSettings() {
  showSettingsModal.value = false
}

async function saveSettings() {
  try {
    await SaveSettingsApi({
      defaultCharset: charsetInput.value.trim() || 'utf-8',
      connectionSettings: {
        connectTimeout: connectTimeout.value,
        readTimeout: readTimeout.value
      }
    })
    // Reload settings to update store
    await settingsStore.loadSettings()
  } catch (e) {
    console.error('Failed to save settings:', e)
  }
  showSettingsModal.value = false
}
</script>

<template>
  <header class="h-12 bg-bg-secondary border-b border-border-custom flex items-center justify-between px-4" style="--wails-draggable: drag">
    <!-- Logo -->
    <div class="flex items-center gap-3">
      <div class="flex items-center gap-2">
        <svg class="w-6 h-6 text-accent-primary" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 2L2 7l10 5 10-5-10-5z"/>
          <path d="M2 17l10 5 10-5"/>
          <path d="M2 12l10 5 10-5"/>
        </svg>
        <span class="text-lg font-semibold text-zinc-100">Wirey</span>
      </div>
      <span class="text-xs text-zinc-500 bg-bg-tertiary px-2 py-0.5 rounded">Socket Client</span>
    </div>

    <!-- Actions -->
    <div class="flex items-center gap-2" style="--wails-draggable: no-drag">
      <!-- Echo Server -->
      <button
        @click="tabStore.openTab('echo', 'echo')"
        class="relative p-2 rounded-lg hover:bg-bg-tertiary transition-colors text-zinc-400 hover:text-zinc-200"
        title="Echo Server"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
        </svg>
        <!-- Running indicator -->
        <span
          v-if="echoStore.isRunning"
          class="absolute top-1 right-1 w-2 h-2 bg-green-500 rounded-full"
        ></span>
      </button>

      <!-- Theme toggle (shows current state) -->
      <button
        @click="settingsStore.toggleTheme"
        class="p-2 rounded-lg hover:bg-bg-tertiary transition-colors text-zinc-400 hover:text-zinc-200"
        :title="settingsStore.settings.theme === 'dark' ? 'Dark mode' : 'Light mode'"
      >
        <!-- Moon for dark mode (current state) -->
        <svg v-if="settingsStore.settings.theme === 'dark'" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"/>
        </svg>
        <!-- Sun for light mode (current state) -->
        <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"/>
        </svg>
      </button>

      <!-- Settings -->
      <button
        @click="openModal"
        class="p-2 rounded-lg hover:bg-bg-tertiary transition-colors text-zinc-400 hover:text-zinc-200"
        title="Settings"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
        </svg>
      </button>
    </div>

    <!-- Settings Modal -->
    <Teleport to="body">
      <div
        v-if="showSettingsModal"
        class="fixed inset-0 z-50 flex items-center justify-center"
        @click.self="showSettingsModal = false"
      >
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm"></div>
        <div class="relative bg-bg-secondary border border-border-custom rounded-xl shadow-2xl w-full max-w-md p-6">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-lg font-medium text-zinc-100">Settings</h3>
            <button
              @click="showSettingsModal = false"
              class="p-1 rounded hover:bg-bg-tertiary text-zinc-400 hover:text-zinc-200"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
              </svg>
            </button>
          </div>

          <!-- Default Charset -->
          <div class="space-y-3">
            <div>
              <label class="text-sm font-medium text-zinc-300">Default Character Encoding</label>
              <p class="text-xs text-zinc-500 mt-1">Used when case/collection doesn't specify a charset</p>
            </div>
            <div class="relative">
              <input
                v-model="charsetInput"
                @focus="showSuggestions = true"
                @blur="handleInputBlur"
                type="text"
                placeholder="e.g. utf-8, euc-kr..."
                class="w-full px-4 py-2.5 bg-bg-tertiary border border-border-custom rounded-lg
                       text-zinc-200 placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-accent-primary/50"
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
                    cs === charsetInput ? 'text-accent-primary' : 'text-zinc-300'
                  ]"
                >
                  {{ cs }}
                </button>
              </div>
            </div>
          </div>

          <!-- Default Timeout Settings -->
          <div class="space-y-3 mt-6 pt-6 border-t border-border-custom">
            <div>
              <label class="text-sm font-medium text-zinc-300">Default Timeout Settings</label>
              <p class="text-xs text-zinc-500 mt-1">Used when case/collection doesn't specify timeout</p>
            </div>

            <!-- Connect Timeout -->
            <div class="space-y-1">
              <label class="text-xs text-zinc-400">Connect Timeout</label>
              <div class="flex items-center gap-2">
                <input
                  v-model.number="connectTimeout"
                  type="number"
                  min="0"
                  step="1000"
                  class="flex-1 px-3 py-2 bg-bg-tertiary border border-border-custom rounded-lg
                         text-zinc-200 text-sm focus:outline-none focus:ring-2 focus:ring-accent-primary/50"
                />
                <span class="text-xs text-zinc-500 w-8">ms</span>
              </div>
              <p class="text-xs text-zinc-600">Time to wait for connection (default: 10000ms)</p>
            </div>

            <!-- Read Timeout -->
            <div class="space-y-1">
              <label class="text-xs text-zinc-400">Read Timeout</label>
              <div class="flex items-center gap-2">
                <input
                  v-model.number="readTimeout"
                  type="number"
                  min="0"
                  step="1000"
                  class="flex-1 px-3 py-2 bg-bg-tertiary border border-border-custom rounded-lg
                         text-zinc-200 text-sm focus:outline-none focus:ring-2 focus:ring-accent-primary/50"
                />
                <span class="text-xs text-zinc-500 w-8">ms</span>
              </div>
              <p class="text-xs text-zinc-600">Time to wait for data (0 = unlimited, default: 60000ms)</p>
            </div>
          </div>

          <!-- Buttons -->
          <div class="flex justify-end gap-3 mt-6">
            <button
              @click="cancelSettings"
              class="px-4 py-2 text-sm rounded-lg bg-bg-tertiary hover:bg-border-custom text-zinc-400 transition-colors"
            >
              Cancel
            </button>
            <button
              @click="saveSettings"
              class="px-4 py-2 text-sm rounded-lg bg-accent-primary hover:bg-accent-primary/80 text-white transition-colors"
            >
              Save
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </header>
</template>
