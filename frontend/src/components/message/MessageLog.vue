<script setup lang="ts">
import { ref, watch, nextTick, computed } from 'vue'
import type { Case } from '@/types'
import { useCaseStore } from '@/stores/case'
import { useSettingsStore } from '@/stores/settings'
import MessageItem from './MessageItem.vue'

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

const logContainer = ref<HTMLElement | null>(null)
const searchQuery = ref('')

const filteredMessages = computed(() => {
  if (!searchQuery.value) return props.caseData.messages
  const query = searchQuery.value.toLowerCase()
  return props.caseData.messages.filter(m =>
    m.content.toLowerCase().includes(query)
  )
})

// Auto-scroll to bottom
watch(
  () => props.caseData.messages.length,
  async () => {
    if (settingsStore.settings.autoScroll) {
      await nextTick()
      if (logContainer.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight
      }
    }
  }
)

function clearMessages() {
  caseStore.clearMessages(props.caseData.id)
}
</script>

<template>
  <div class="flex-1 flex flex-col overflow-hidden min-h-[120px]">
    <!-- Toolbar -->
    <div class="flex items-center justify-between px-4 py-2 border-b border-border-custom bg-bg-secondary/50">
      <div class="flex items-center gap-3">
        <h3 class="text-sm font-medium text-zinc-300">Message Log</h3>
        <span class="text-xs text-zinc-500 bg-bg-tertiary px-2 py-0.5 rounded">
          {{ props.caseData.messages.length }} messages
        </span>
      </div>

      <div class="flex items-center gap-2">
        <!-- Search -->
        <div class="relative">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search..."
            class="w-48 px-3 py-1.5 pl-8 text-xs bg-bg-tertiary border border-border-custom rounded-lg
                   text-zinc-300 placeholder-zinc-500 focus:outline-none focus:ring-1 focus:ring-accent-primary/50"
          />
          <svg class="w-3.5 h-3.5 absolute left-2.5 top-1/2 -translate-y-1/2 text-zinc-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
          </svg>
        </div>

        <!-- Clear button -->
        <button
          @click="clearMessages"
          class="px-3 py-1.5 text-xs rounded-lg bg-bg-tertiary hover:bg-border-custom text-zinc-400 hover:text-zinc-200 transition-colors flex items-center gap-1.5"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
          </svg>
          Clear
        </button>
      </div>
    </div>

    <!-- Message List -->
    <div
      ref="logContainer"
      class="flex-1 overflow-y-auto p-4 space-y-1 bg-bg-primary"
    >
      <!-- Messages -->
      <MessageItem
        v-for="message in filteredMessages"
        :key="message.id"
        :message="message"
        :show-timestamp="settingsStore.settings.showTimestamp"
        :charset="effectiveCharset"
      />

      <!-- Empty state -->
      <div
        v-if="props.caseData.messages.length === 0"
        class="h-full flex items-center justify-center"
      >
        <div class="text-center text-zinc-500">
          <svg class="w-12 h-12 mx-auto mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
          </svg>
          <p class="text-sm">No messages yet</p>
          <p class="text-xs mt-1">Connect and start sending messages</p>
        </div>
      </div>

      <!-- No results -->
      <div
        v-else-if="filteredMessages.length === 0 && searchQuery"
        class="h-full flex items-center justify-center"
      >
        <div class="text-center text-zinc-500">
          <p class="text-sm">No messages matching "{{ searchQuery }}"</p>
        </div>
      </div>
    </div>
  </div>
</template>
