<script setup lang="ts">
import { ref, computed, watch, nextTick, onActivated } from 'vue'
import { useEchoStore } from '@/stores/echo'
import { DEFAULT_ECHO_PORT } from '@/types'
import type { EchoProtocol, EchoLogEntry } from '@/types'

const echoStore = useEchoStore()

const portInput = ref(echoStore.port)
const protocolInput = ref<EchoProtocol>(echoStore.protocol)
const logsContainer = ref<HTMLElement | null>(null)
const autoScroll = ref(true)

// Sync inputs with store
watch(() => echoStore.port, (val) => { portInput.value = val })
watch(() => echoStore.protocol, (val) => { protocolInput.value = val })

// Auto-scroll when new logs arrive
watch(() => echoStore.logs.length, async () => {
  if (autoScroll.value && logsContainer.value) {
    await nextTick()
    logsContainer.value.scrollTop = logsContainer.value.scrollHeight
  }
})

// Scroll to bottom when tab becomes visible (for logs added while on other tabs)
onActivated(async () => {
  if (autoScroll.value && logsContainer.value) {
    await nextTick()
    logsContainer.value.scrollTop = logsContainer.value.scrollHeight
  }
})

async function handleStart() {
  echoStore.setPort(portInput.value)
  echoStore.setProtocol(protocolInput.value)
  await echoStore.startServer()
}

async function handleStop() {
  await echoStore.stopServer()
}

function formatTimestamp(ts: number): string {
  const date = new Date(ts)
  return date.toLocaleTimeString('en-US', { hour12: false })
}

function formatHexDump(hexData: string): string[] {
  const lines: string[] = []
  const bytes = hexData.match(/.{1,2}/g) || []

  for (let i = 0; i < bytes.length; i += 16) {
    const chunk = bytes.slice(i, i + 16)
    const offset = i.toString(16).padStart(8, '0')
    const hexPart = chunk.join(' ').padEnd(48, ' ')
    const asciiPart = chunk.map(b => {
      const code = parseInt(b, 16)
      return code >= 32 && code < 127 ? String.fromCharCode(code) : '.'
    }).join('')

    lines.push(`${offset}  ${hexPart}  ${asciiPart}`)
  }

  return lines
}

const formattedLogs = computed(() => {
  return echoStore.logs.map(log => ({
    ...log,
    hexDump: formatHexDump(log.data)
  }))
})
</script>

<template>
  <div class="flex-1 flex flex-col overflow-hidden bg-bg-primary">
    <!-- Header -->
    <div class="flex-shrink-0 p-4 border-b border-border-custom bg-bg-secondary">
      <h2 class="text-lg font-semibold text-zinc-100 mb-4">Echo Server</h2>

      <!-- Controls -->
      <div class="flex flex-wrap items-center gap-4">
        <!-- Port -->
        <div class="flex items-center gap-2">
          <label class="text-sm text-zinc-400">Port:</label>
          <input
            v-model.number="portInput"
            type="number"
            min="1"
            max="65535"
            :disabled="echoStore.isRunning"
            class="w-24 px-3 py-1.5 bg-bg-tertiary border border-border-custom rounded-lg
                   text-zinc-200 text-sm focus:outline-none focus:ring-2 focus:ring-accent-primary/50
                   disabled:opacity-50 disabled:cursor-not-allowed"
          />
        </div>

        <!-- Protocol -->
        <div class="flex items-center gap-3">
          <label class="text-sm text-zinc-400">Protocol:</label>
          <label class="flex items-center gap-1.5 cursor-pointer" :class="{ 'opacity-50': echoStore.isRunning }">
            <input
              type="radio"
              value="tcp"
              v-model="protocolInput"
              :disabled="echoStore.isRunning"
              class="w-4 h-4 text-accent-primary bg-bg-tertiary border-border-custom focus:ring-accent-primary"
            />
            <span class="text-sm text-zinc-300">TCP</span>
          </label>
          <label class="flex items-center gap-1.5 cursor-pointer" :class="{ 'opacity-50': echoStore.isRunning }">
            <input
              type="radio"
              value="udp"
              v-model="protocolInput"
              :disabled="echoStore.isRunning"
              class="w-4 h-4 text-accent-primary bg-bg-tertiary border-border-custom focus:ring-accent-primary"
            />
            <span class="text-sm text-zinc-300">UDP</span>
          </label>
        </div>

        <!-- Buttons -->
        <div class="flex items-center gap-2">
          <button
            v-if="!echoStore.isRunning"
            @click="handleStart"
            class="px-4 py-1.5 text-sm rounded-lg bg-green-600 hover:bg-green-500 text-white transition-colors"
          >
            Start
          </button>
          <button
            v-else
            @click="handleStop"
            class="px-4 py-1.5 text-sm rounded-lg bg-red-600 hover:bg-red-500 text-white transition-colors"
          >
            Stop
          </button>
        </div>

        <!-- Status -->
        <div class="flex items-center gap-2 ml-auto">
          <span
            :class="[
              'w-2 h-2 rounded-full',
              echoStore.isRunning ? 'bg-green-500' : echoStore.error ? 'bg-red-500' : 'bg-zinc-500'
            ]"
          ></span>
          <span class="text-sm text-zinc-400">
            {{ echoStore.isRunning ? `Running on :${echoStore.port} (${echoStore.protocol.toUpperCase()})` : 'Stopped' }}
          </span>
        </div>
      </div>

      <!-- Error message -->
      <div v-if="echoStore.error" class="mx-4 mt-3 p-3 bg-red-100 dark:bg-red-500/10 border border-red-300 dark:border-red-500/30 rounded-lg flex items-start gap-3">
        <svg class="w-5 h-5 text-red-500 dark:text-red-400 shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
        <div class="flex-1 min-w-0">
          <div class="text-sm font-medium text-red-600 dark:text-red-400">Error</div>
          <div class="text-sm text-red-800 dark:text-red-200 break-words">{{ echoStore.error }}</div>
        </div>
        <button @click="echoStore.clearError" class="shrink-0 text-red-500 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300 transition-colors">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- Logs Header -->
    <div class="flex-shrink-0 px-4 py-2 border-b border-border-custom bg-bg-secondary flex items-center justify-between">
      <div class="flex items-center gap-3">
        <span class="text-sm font-medium text-zinc-300">Logs</span>
        <span class="text-xs text-zinc-500">({{ echoStore.logCount }} entries)</span>
      </div>
      <div class="flex items-center gap-3">
        <label class="flex items-center gap-1.5 cursor-pointer">
          <input
            type="checkbox"
            v-model="autoScroll"
            class="w-4 h-4 text-accent-primary bg-bg-tertiary border-border-custom rounded focus:ring-accent-primary"
          />
          <span class="text-xs text-zinc-400">Auto-scroll</span>
        </label>
        <button
          @click="echoStore.clearLogs"
          class="px-3 py-1 text-xs rounded bg-bg-tertiary hover:bg-border-custom text-zinc-400 transition-colors"
        >
          Clear
        </button>
      </div>
    </div>

    <!-- Logs Content -->
    <div
      ref="logsContainer"
      class="flex-1 overflow-y-auto p-4 space-y-3 font-mono text-xs"
    >
      <template v-if="formattedLogs.length === 0">
        <div class="text-center text-zinc-500 py-8">
          No logs yet. Start the server and send some data.
        </div>
      </template>

      <template v-for="log in formattedLogs" :key="log.id">
        <div class="border border-border-custom rounded-lg overflow-hidden">
          <!-- Log Header -->
          <div
            :class="[
              'px-3 py-1.5 flex items-center gap-3',
              log.direction === 'recv' ? 'bg-blue-500/10' : 'bg-green-500/10'
            ]"
          >
            <span
              :class="[
                'text-sm font-medium',
                log.direction === 'recv' ? 'text-blue-400' : 'text-green-400'
              ]"
            >
              {{ log.direction === 'recv' ? '← RECV' : '→ ECHO' }}
            </span>
            <span class="text-zinc-400">{{ log.remoteAddr }}</span>
            <span class="text-zinc-500">({{ log.size }} bytes)</span>
            <span class="text-zinc-500 ml-auto">{{ formatTimestamp(log.timestamp) }}</span>
          </div>

          <!-- Hex Dump -->
          <div class="bg-bg-tertiary px-3 py-2">
            <pre class="text-zinc-300 leading-relaxed"><template v-for="(line, i) in log.hexDump" :key="i">{{ line }}
</template></pre>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
