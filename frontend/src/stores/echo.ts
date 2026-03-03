import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { EchoLogEntry, EchoStatus, EchoProtocol } from '@/types'
import { DEFAULT_ECHO_PORT } from '@/types'
import {
  StartEchoServer,
  StopEchoServer,
  GetEchoStatus,
  GetEchoLogs,
  ClearEchoLogs
} from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { useTabStore } from './tab'

export const useEchoStore = defineStore('echo', () => {
  // State
  const isRunning = ref(false)
  const port = ref(DEFAULT_ECHO_PORT)
  const protocol = ref<EchoProtocol>('tcp')
  const logs = ref<EchoLogEntry[]>([])
  const error = ref<string | null>(null)

  function setError(msg: string) {
    error.value = msg
  }

  function clearError() {
    error.value = null
  }

  // Getters
  const status = computed((): EchoStatus => ({
    running: isRunning.value,
    port: port.value,
    protocol: protocol.value,
    address: isRunning.value ? `:${port.value}` : ''
  }))

  const logCount = computed(() => logs.value.length)

  // Check if echo tab is currently active (derived from tabStore)
  const isTabOpen = computed(() => {
    const tabStore = useTabStore()
    return tabStore.activeTabKey === 'echo'
  })

  // Actions
  async function startServer() {
    try {
      clearError()
      await StartEchoServer(port.value, protocol.value)
      isRunning.value = true
    } catch (e) {
      setError(String(e))
      throw e
    }
  }

  async function stopServer() {
    try {
      clearError()
      await StopEchoServer()
      isRunning.value = false
    } catch (e) {
      setError(String(e))
      throw e
    }
  }

  async function refreshStatus() {
    try {
      const s = await GetEchoStatus()
      isRunning.value = s.running
      if (s.running) {
        port.value = s.port
        protocol.value = s.protocol as EchoProtocol
      }
    } catch (e) {
      console.error('Failed to get echo status:', e)
    }
  }

  async function refreshLogs() {
    try {
      const entries = await GetEchoLogs()
      logs.value = entries || []
    } catch (e) {
      console.error('Failed to get echo logs:', e)
    }
  }

  async function clearLogs() {
    try {
      await ClearEchoLogs()
      logs.value = []
    } catch (e) {
      console.error('Failed to clear echo logs:', e)
    }
  }

  function setPort(newPort: number) {
    port.value = newPort
  }

  function setProtocol(newProtocol: EchoProtocol) {
    protocol.value = newProtocol
  }

  // Event handlers
  function handleLogEvent(entry: EchoLogEntry) {
    logs.value.push(entry)
  }

  function handleStatusEvent(s: EchoStatus) {
    isRunning.value = s.running
    if (s.running) {
      port.value = s.port
      protocol.value = s.protocol as EchoProtocol
    }
  }

  // Initialize event listeners
  function initEvents() {
    EventsOn('echo:log', handleLogEvent)
    EventsOn('echo:status', handleStatusEvent)
  }

  function cleanupEvents() {
    EventsOff('echo:log')
    EventsOff('echo:status')
  }

  // Initialize
  initEvents()
  refreshStatus()

  return {
    // State
    isRunning,
    port,
    protocol,
    logs,
    error,
    clearError,
    // Getters
    status,
    logCount,
    isTabOpen,
    // Actions
    startServer,
    stopServer,
    refreshStatus,
    refreshLogs,
    clearLogs,
    setPort,
    setProtocol,
    // Events
    initEvents,
    cleanupEvents
  }
})
