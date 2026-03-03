<script setup lang="ts">
import { ref, onMounted, onUnmounted, onActivated, onDeactivated } from 'vue'
import type { Case } from '@/types'
import { useCaseStore } from '@/stores/case'
import ConnectionPanel from '@/components/connection/ConnectionPanel.vue'
import MessageLog from '@/components/message/MessageLog.vue'
import MessageInput from '@/components/message/MessageInput.vue'
import FramingPanel from '@/components/framing/FramingPanel.vue'
import SettingsPanel from '@/components/settings/SettingsPanel.vue'
import ScriptsPanel from '@/components/scripts/ScriptsPanel.vue'
import NotesEditor from '@/components/notes/NotesEditor.vue'

const props = defineProps<{
  caseData: Case
}>()

const caseStore = useCaseStore()

// Content tab state (preserved per instance via KeepAlive)
type ContentTab = 'messages' | 'scripts' | 'framing' | 'settings' | 'notes'
const activeContentTab = ref<ContentTab>('messages')

// Notes handling
function handleNotesChange(notes: string) {
  caseStore.updateCase(props.caseData.id, { notes })
}

// Ctrl+S to save
function handleKeyDown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 's') {
    e.preventDefault()
    caseStore.saveCase(props.caseData.id)
  }
}

onMounted(() => {
  document.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeyDown)
})

// KeepAlive: re-attach/detach on activate/deactivate
onActivated(() => {
  document.addEventListener('keydown', handleKeyDown)
})

onDeactivated(() => {
  document.removeEventListener('keydown', handleKeyDown)
})
</script>

<template>
  <div class="flex-1 flex flex-col overflow-hidden">
    <!-- Connection Panel -->
    <ConnectionPanel :case-data="caseData" />

    <!-- Content Tab Bar -->
    <div class="flex border-b border-border-custom bg-bg-secondary/50 shrink-0">
      <button
        @click="activeContentTab = 'messages'"
        :class="[
          'px-4 py-2.5 text-sm font-medium transition-colors relative',
          activeContentTab === 'messages'
            ? 'text-accent-primary'
            : 'text-zinc-400 hover:text-zinc-200'
        ]"
      >
        Messages
        <span
          v-if="activeContentTab === 'messages'"
          class="absolute bottom-0 left-0 right-0 h-0.5 bg-accent-primary"
        ></span>
      </button>
      <button
        @click="activeContentTab = 'framing'"
        :class="[
          'px-4 py-2.5 text-sm font-medium transition-colors relative',
          activeContentTab === 'framing'
            ? 'text-accent-primary'
            : 'text-zinc-400 hover:text-zinc-200'
        ]"
      >
        Framing
        <span
          v-if="activeContentTab === 'framing'"
          class="absolute bottom-0 left-0 right-0 h-0.5 bg-accent-primary"
        ></span>
      </button>
      <button
        @click="activeContentTab = 'scripts'"
        :class="[
          'px-4 py-2.5 text-sm font-medium transition-colors relative',
          activeContentTab === 'scripts'
            ? 'text-accent-primary'
            : 'text-zinc-400 hover:text-zinc-200'
        ]"
      >
        Scripts
        <span
          v-if="activeContentTab === 'scripts'"
          class="absolute bottom-0 left-0 right-0 h-0.5 bg-accent-primary"
        ></span>
      </button>
      <button
        @click="activeContentTab = 'settings'"
        :class="[
          'px-4 py-2.5 text-sm font-medium transition-colors relative',
          activeContentTab === 'settings'
            ? 'text-accent-primary'
            : 'text-zinc-400 hover:text-zinc-200'
        ]"
      >
        Settings
        <span
          v-if="activeContentTab === 'settings'"
          class="absolute bottom-0 left-0 right-0 h-0.5 bg-accent-primary"
        ></span>
      </button>
      <button
        @click="activeContentTab = 'notes'"
        :class="[
          'px-4 py-2.5 text-sm font-medium transition-colors relative',
          activeContentTab === 'notes'
            ? 'text-accent-primary'
            : 'text-zinc-400 hover:text-zinc-200'
        ]"
      >
        Notes
        <span
          v-if="activeContentTab === 'notes'"
          class="absolute bottom-0 left-0 right-0 h-0.5 bg-accent-primary"
        ></span>
      </button>
    </div>

    <!-- Tab Content -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- Messages Tab -->
      <template v-if="activeContentTab === 'messages'">
        <MessageLog :case-data="caseData" />
        <MessageInput :case-data="caseData" />
      </template>

      <!-- Framing Tab -->
      <template v-else-if="activeContentTab === 'framing'">
        <div class="flex-1 overflow-y-auto">
          <FramingPanel :case-data="caseData" />
        </div>
      </template>

      <!-- Scripts Tab -->
      <template v-else-if="activeContentTab === 'scripts'">
        <div class="flex-1 overflow-y-auto">
          <ScriptsPanel :case-data="caseData" />
        </div>
      </template>

      <!-- Settings Tab -->
      <template v-else-if="activeContentTab === 'settings'">
        <div class="flex-1 overflow-y-auto">
          <SettingsPanel :case-data="caseData" />
        </div>
      </template>

      <!-- Notes Tab -->
      <template v-else-if="activeContentTab === 'notes'">
        <div class="flex-1 overflow-hidden">
          <NotesEditor
            :modelValue="caseData.notes || ''"
            @update:modelValue="handleNotesChange"
          />
        </div>
      </template>
    </div>
  </div>
</template>
