import { defineStore } from 'pinia'
import { ref, watch, computed } from 'vue'
import type { AppSettings, Charset, ConnectionSettings } from '@/types'
import { DEFAULT_APP_SETTINGS, DEFAULT_CONNECTION_SETTINGS } from '@/types'
import {
  LoadSettings as LoadSettingsApi,
  SaveSettings as SaveSettingsApi
} from '../../wailsjs/go/main/App'
import {
  WindowGetSize,
  WindowSetSize,
  WindowIsMaximised,
  WindowMaximise
} from '../../wailsjs/runtime/runtime'

interface WindowState {
  width: number
  height: number
  isMaximized: boolean
}

export const useSettingsStore = defineStore('settings', () => {
  // Load theme from localStorage synchronously to prevent flash
  const savedTheme = (() => {
    try {
      const saved = localStorage.getItem('wirey-settings')
      if (saved) {
        const parsed = JSON.parse(saved)
        return parsed.theme || 'dark'
      }
    } catch {}
    return 'dark'
  })()

  // Apply theme immediately
  if (savedTheme === 'dark') {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }

  // State
  const settings = ref<AppSettings>({ ...DEFAULT_APP_SETTINGS, theme: savedTheme })
  const isLoaded = ref(false)

  // Getters
  const defaultCharset = computed(() => settings.value.defaultCharset)
  const connectionSettings = computed(() => settings.value.connectionSettings)

  // Load settings from backend
  async function loadSettings() {
    try {
      // Load settings from Go backend
      const backendSettings = await LoadSettingsApi()
      if (backendSettings.defaultCharset) {
        settings.value.defaultCharset = backendSettings.defaultCharset as Charset
      }
      if (backendSettings.connectionSettings) {
        settings.value.connectionSettings = {
          connectTimeout: backendSettings.connectionSettings.connectTimeout || DEFAULT_CONNECTION_SETTINGS.connectTimeout,
          readTimeout: backendSettings.connectionSettings.readTimeout ?? DEFAULT_CONNECTION_SETTINGS.readTimeout
        }
      }

      // Load UI settings from localStorage
      const savedSettings = localStorage.getItem('wirey-settings')
      if (savedSettings) {
        const parsed = JSON.parse(savedSettings)
        settings.value = {
          ...settings.value,
          theme: parsed.theme || settings.value.theme,
          fontSize: parsed.fontSize || settings.value.fontSize,
          showTimestamp: parsed.showTimestamp ?? settings.value.showTimestamp,
          autoScroll: parsed.autoScroll ?? settings.value.autoScroll,
          maxMessages: parsed.maxMessages || settings.value.maxMessages
        }
      }

      // Apply theme after loading
      applyTheme()
      isLoaded.value = true
    } catch (e) {
      console.error('Failed to load settings:', e)
      isLoaded.value = true
    }
  }

  // Save UI settings to localStorage
  function saveLocalSettings() {
    try {
      localStorage.setItem('wirey-settings', JSON.stringify({
        theme: settings.value.theme,
        fontSize: settings.value.fontSize,
        showTimestamp: settings.value.showTimestamp,
        autoScroll: settings.value.autoScroll,
        maxMessages: settings.value.maxMessages
      }))
    } catch (e) {
      console.error('Failed to save settings:', e)
    }
  }

  // Save charset and connection settings to backend
  async function saveBackendSettings() {
    try {
      await SaveSettingsApi({
        defaultCharset: settings.value.defaultCharset,
        connectionSettings: settings.value.connectionSettings
      })
    } catch (e) {
      console.error('Failed to save backend settings:', e)
      throw e
    }
  }

  // Actions
  function updateSetting<K extends keyof AppSettings>(key: K, value: AppSettings[K]) {
    settings.value[key] = value
    saveLocalSettings()
  }

  async function setDefaultCharset(charset: Charset) {
    settings.value.defaultCharset = charset
    await saveBackendSettings()
  }

  function toggleTheme() {
    settings.value.theme = settings.value.theme === 'dark' ? 'light' : 'dark'
    saveLocalSettings()
    applyTheme()
  }

  function applyTheme() {
    if (settings.value.theme === 'dark') {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  // Initialize
  loadSettings()
  restoreWindowState()

  // Window state management (size only, position handled by OS)
  async function saveWindowState() {
    try {
      const isMaximized = await WindowIsMaximised()

      if (!isMaximized) {
        const size = await WindowGetSize()
        const state: WindowState = {
          width: size.w,
          height: size.h,
          isMaximized: false
        }
        localStorage.setItem('wirey-window', JSON.stringify(state))
      } else {
        // Save maximized state but keep previous size
        const saved = localStorage.getItem('wirey-window')
        if (saved) {
          const state = JSON.parse(saved) as WindowState
          state.isMaximized = true
          localStorage.setItem('wirey-window', JSON.stringify(state))
        } else {
          localStorage.setItem('wirey-window', JSON.stringify({ isMaximized: true }))
        }
      }
    } catch (e) {
      console.error('Failed to save window state:', e)
    }
  }

  async function restoreWindowState() {
    try {
      const saved = localStorage.getItem('wirey-window')
      if (!saved) return

      const state = JSON.parse(saved) as WindowState

      if (state.isMaximized) {
        WindowMaximise()
      } else if (state.width && state.height) {
        // Clamp size to screen bounds
        const screenWidth = window.screen.availWidth
        const screenHeight = window.screen.availHeight
        const width = Math.min(state.width, screenWidth)
        const height = Math.min(state.height, screenHeight)
        WindowSetSize(width, height)
        // Position is handled by OS (centered by default)
      }
    } catch (e) {
      console.error('Failed to restore window state:', e)
    }
  }

  // Auto-save window state on resize/move (debounced)
  let windowSaveTimeout: ReturnType<typeof setTimeout> | null = null
  function scheduleWindowSave() {
    if (windowSaveTimeout) clearTimeout(windowSaveTimeout)
    windowSaveTimeout = setTimeout(saveWindowState, 500)
  }

  // Listen to window resize
  if (typeof window !== 'undefined') {
    window.addEventListener('resize', scheduleWindowSave)
  }

  return {
    settings,
    defaultCharset,
    connectionSettings,
    isLoaded,
    updateSetting,
    setDefaultCharset,
    toggleTheme,
    loadSettings,
    saveWindowState
  }
})
