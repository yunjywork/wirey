<script setup lang="ts">
import { useTabStore } from '@/stores/tab'
import { useCaseStore } from '@/stores/case'
import { Disconnect } from '../../../wailsjs/go/main/App'

const tabStore = useTabStore()
const caseStore = useCaseStore()

async function disconnectCase(caseId: string) {
  try {
    await Disconnect(caseId)
  } catch (e) {
    console.error('Failed to disconnect:', e)
  }
}

function openCase(caseId: string) {
  tabStore.openTab('case', caseId)
}
</script>

<template>
  <div class="flex-1 flex flex-col overflow-hidden bg-bg-primary">
    <!-- Header -->
    <div class="p-6 border-b border-border-custom">
      <h1 class="text-xl font-semibold text-zinc-100 flex items-center gap-3">
        <svg class="w-6 h-6 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0"/>
        </svg>
        Active Connections
        <span class="text-accent-success text-lg">({{ caseStore.connectedCases.length }})</span>
      </h1>
    </div>

    <!-- Content -->
    <div class="flex-1 overflow-auto p-6">
      <!-- Empty state -->
      <div
        v-if="caseStore.connectedCases.length === 0"
        class="flex flex-col items-center justify-center h-full text-zinc-500"
      >
        <svg class="w-16 h-16 mb-4 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M18.364 5.636a9 9 0 010 12.728m0 0l-2.829-2.829m2.829 2.829L21 21M15.536 8.464a5 5 0 010 7.072m0 0l-2.829-2.829m-4.243 2.829a5 5 0 01-7.072 0l2.829-2.829m4.243 2.829L12 12m-4.243 4.243L5.636 18.364"/>
        </svg>
        <p class="text-lg font-medium mb-1">No active connections</p>
        <p class="text-sm text-zinc-600">Connect to a server to see it here</p>
      </div>

      <!-- Connection list -->
      <div v-else class="space-y-3">
        <div
          v-for="c in caseStore.connectedCases"
          :key="c.id"
          class="p-4 bg-bg-secondary rounded-lg border border-border-custom hover:border-zinc-600 transition-colors"
        >
          <div class="flex items-start justify-between gap-4">
            <!-- Left: Connection info -->
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 mb-2">
                <span class="status-dot status-connected shrink-0"></span>
                <span class="font-medium text-zinc-100 truncate">{{ c.name }}</span>
                <span class="text-xs px-1.5 py-0.5 rounded bg-zinc-700 text-zinc-300 uppercase">{{ c.protocol }}</span>
              </div>

              <div class="space-y-1 text-sm text-zinc-400">
                <!-- Remote address -->
                <div class="flex items-center gap-2">
                  <svg class="w-4 h-4 text-zinc-500 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"/>
                  </svg>
                  <span class="font-mono">{{ c.host }}:{{ c.port }}</span>
                </div>

                <!-- Local address -->
                <div v-if="c.localAddr" class="flex items-center gap-2">
                  <svg class="w-4 h-4 text-zinc-500 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
                  </svg>
                  <span class="font-mono text-zinc-500">{{ c.localAddr }}</span>
                </div>
              </div>
            </div>

            <!-- Right: Actions -->
            <div class="flex items-center gap-2 shrink-0">
              <button
                @click="openCase(c.id)"
                class="px-3 py-1.5 text-sm rounded-lg bg-accent-primary hover:bg-accent-primary/80 text-white transition-colors"
              >
                Open
              </button>
              <button
                @click="disconnectCase(c.id)"
                class="px-3 py-1.5 text-sm rounded-lg bg-bg-tertiary hover:bg-zinc-600 text-zinc-300 transition-colors"
              >
                Disconnect
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
