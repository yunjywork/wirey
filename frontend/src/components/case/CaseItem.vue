<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import type { Case } from '@/types'
import { useCaseStore } from '@/stores/case'

const props = defineProps<{
  caseData: Case
  isActive: boolean
}>()

const emit = defineEmits<{
  select: []
  remove: []
  rename: []
}>()

const caseStore = useCaseStore()

// Dropdown menu
const showMenu = ref(false)
const menuRef = ref<HTMLDivElement | null>(null)

const statusClass = {
  connected: 'status-connected',
  connecting: 'status-connecting',
  disconnected: 'status-disconnected',
  error: 'bg-accent-error'
}

function toggleMenu(e: Event) {
  e.stopPropagation()
  showMenu.value = !showMenu.value
}

function closeMenu() {
  showMenu.value = false
}

function handleClickOutside(e: MouseEvent) {
  if (menuRef.value && !menuRef.value.contains(e.target as Node)) {
    closeMenu()
  }
}

function handleRename() {
  closeMenu()
  emit('rename')
}

function handleDelete() {
  closeMenu()
  emit('remove')
}

async function handleSave(e: Event) {
  e.stopPropagation()
  try {
    await caseStore.saveCase(props.caseData.id)
  } catch (error) {
    console.error('Failed to save:', error)
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div
    @click="emit('select')"
    :class="[
      'group relative px-2 py-2 rounded-lg cursor-pointer transition-all duration-200',
      isActive
        ? 'bg-accent-primary/10 border border-subtle'
        : 'hover:bg-zinc-700/50 border border-transparent'
    ]"
  >
    <!-- Line 1: Drag Handle + Name + Save + Menu -->
    <div class="flex items-center gap-2">
      <!-- Drag Handle -->
      <svg
        class="case-drag-handle w-3 h-3 text-zinc-600 hover:text-zinc-400 cursor-grab shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
        fill="currentColor"
        viewBox="0 0 24 24"
      >
        <path d="M8 6a2 2 0 1 1-4 0 2 2 0 0 1 4 0zm0 6a2 2 0 1 1-4 0 2 2 0 0 1 4 0zm0 6a2 2 0 1 1-4 0 2 2 0 0 1 4 0zm8-12a2 2 0 1 1-4 0 2 2 0 0 1 4 0zm0 6a2 2 0 1 1-4 0 2 2 0 0 1 4 0zm0 6a2 2 0 1 1-4 0 2 2 0 0 1 4 0z"/>
      </svg>

      <span class="flex-1 text-sm font-medium text-zinc-200 truncate">
        {{ caseData.name }}
      </span>

      <!-- Save button / Saved indicator -->
      <button
        v-if="!caseData.isSaved"
        @click="handleSave"
        class="p-1 rounded hover:bg-accent-success/20 text-zinc-500 hover:text-accent-success transition-all opacity-0 group-hover:opacity-100"
        title="Save case"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4"/>
        </svg>
      </button>
      <svg
        v-else
        class="w-3.5 h-3.5 text-accent-success shrink-0"
        fill="currentColor"
        viewBox="0 0 24 24"
        title="Saved"
      >
        <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/>
      </svg>

      <!-- Menu button -->
      <div ref="menuRef" class="relative">
        <button
          @click="toggleMenu"
          class="p-1 rounded hover:bg-bg-tertiary text-zinc-500 hover:text-zinc-200 transition-all opacity-0 group-hover:opacity-100"
          :class="{ 'opacity-100': showMenu }"
          title="More options"
        >
          <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
            <circle cx="12" cy="5" r="2"/>
            <circle cx="12" cy="12" r="2"/>
            <circle cx="12" cy="19" r="2"/>
          </svg>
        </button>

        <!-- Dropdown menu -->
        <div
          v-if="showMenu"
          class="absolute right-0 top-full mt-1 z-50 bg-bg-secondary border border-border-custom rounded-lg shadow-xl py-1 min-w-[120px]"
        >
          <button
            @click="handleRename"
            class="w-full px-3 py-1.5 text-left text-sm text-zinc-300 hover:bg-bg-tertiary flex items-center gap-2"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
            </svg>
            Rename
          </button>
          <button
            @click="handleDelete"
            class="w-full px-3 py-1.5 text-left text-sm text-accent-error hover:bg-bg-tertiary flex items-center gap-2"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
            </svg>
            Delete
          </button>
        </div>
      </div>
    </div>

    <!-- Line 2: Status + Protocol + Address -->
    <div class="flex items-center gap-2 mt-1">
      <span class="status-dot shrink-0" :class="statusClass[caseData.status]"></span>
      <span
        :class="[
          'text-[10px] font-bold uppercase px-1 py-0.5 rounded shrink-0',
          caseData.protocol === 'tcp'
            ? 'bg-blue-500/20 text-blue-400'
            : 'bg-purple-500/20 text-purple-400'
        ]"
      >
        {{ caseData.protocol }}
      </span>
      <span class="font-mono text-xs text-zinc-500 truncate">
        {{ caseData.host }}:{{ caseData.port }}
      </span>
    </div>
  </div>
</template>
