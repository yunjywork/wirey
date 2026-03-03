<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { VueDraggable } from 'vue-draggable-plus'
import { useTabStore } from '@/stores/tab'
import { useCaseStore } from '@/stores/case'
import CaseItem from '@/components/case/CaseItem.vue'
import type { Collection, Case } from '@/types'

const tabStore = useTabStore()
const caseStore = useCaseStore()

// Sidebar width with localStorage persistence
const SIDEBAR_WIDTH_KEY = 'wirey-sidebar-width'
const DEFAULT_WIDTH = 288 // 18rem = 288px
const MIN_WIDTH = 200
const MAX_WIDTH = 500

const sidebarWidth = ref(DEFAULT_WIDTH)
const isResizing = ref(false)

function loadSidebarWidth() {
  const saved = localStorage.getItem(SIDEBAR_WIDTH_KEY)
  if (saved) {
    const width = parseInt(saved, 10)
    if (!isNaN(width) && width >= MIN_WIDTH && width <= MAX_WIDTH) {
      sidebarWidth.value = width
    }
  }
}

function saveSidebarWidth() {
  localStorage.setItem(SIDEBAR_WIDTH_KEY, String(sidebarWidth.value))
}

function startResize(e: MouseEvent) {
  e.preventDefault()
  isResizing.value = true
  document.addEventListener('mousemove', handleResize)
  document.addEventListener('mouseup', stopResize)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}

function handleResize(e: MouseEvent) {
  if (!isResizing.value) return
  const newWidth = Math.min(MAX_WIDTH, Math.max(MIN_WIDTH, e.clientX))
  sidebarWidth.value = newWidth
}

function stopResize() {
  isResizing.value = false
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
  document.body.style.cursor = ''
  document.body.style.userSelect = ''
  saveSidebarWidth()
}

// Collection menu state
const showCollectionMenu = ref<string | null>(null)
const collectionMenuRef = ref<HTMLDivElement | null>(null)

// Context menu state
const contextMenu = ref<{
  show: boolean
  x: number
  y: number
  type: 'collection' | 'case' | null
  target: string | null
}>({
  show: false,
  x: 0,
  y: 0,
  type: null,
  target: null
})

function handleCollectionContextMenu(e: MouseEvent, collectionName: string) {
  e.preventDefault()
  e.stopPropagation()
  contextMenu.value = {
    show: true,
    x: e.clientX,
    y: e.clientY,
    type: 'collection',
    target: collectionName
  }
}

function handleCaseContextMenu(e: MouseEvent, caseId: string) {
  e.preventDefault()
  e.stopPropagation()
  contextMenu.value = {
    show: true,
    x: e.clientX,
    y: e.clientY,
    type: 'case',
    target: caseId
  }
}

function closeContextMenu() {
  contextMenu.value.show = false
}

function toggleCollectionMenu(e: Event, collectionName: string) {
  e.stopPropagation()
  showCollectionMenu.value = showCollectionMenu.value === collectionName ? null : collectionName
}

function closeCollectionMenu() {
  showCollectionMenu.value = null
}

function handleClickOutside(e: MouseEvent) {
  // Close collection dropdown menu
  if (collectionMenuRef.value && !collectionMenuRef.value.contains(e.target as Node)) {
    closeCollectionMenu()
  }
  // Close context menu
  if (contextMenu.value.show) {
    closeContextMenu()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  loadSidebarWidth()
  caseStore.loadCollections()
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

// Handle collection reorder after drag
function onCollectionUpdate(newList: Collection[]) {
  // Update local state first
  caseStore.collections.splice(0, caseStore.collections.length, ...newList)
  // Persist to backend
  const orderedNames = newList.map(c => c.name)
  caseStore.reorderCollections(orderedNames)
}

// Handle case drag end within a collection
function onCaseDragEnd(collection: Collection) {
  const orderedIds = collection.cases.map(c => c.id)
  caseStore.reorderCases(collection.name, orderedIds)
}

// Collection creation modal (state shared via store)
const newCollectionName = ref('')
const newCollectionInput = ref<HTMLInputElement | null>(null)

// Delete confirmation
const showDeleteModal = ref(false)
const deleteType = ref<'case' | 'collection'>('case')
const deleteTargetId = ref<string | null>(null)
const deleteTargetName = ref('')

// Rename modal
const showRenameModal = ref(false)
const renameType = ref<'case' | 'collection'>('case')
const renameTargetId = ref<string | null>(null)
const renameName = ref('')
const renameInput = ref<HTMLInputElement | null>(null)


// Collection actions
async function openNewCollectionModal() {
  newCollectionName.value = ''
  caseStore.showNewCollectionModal = true
  await nextTick()
  newCollectionInput.value?.focus()
}

async function createCollection() {
  const name = newCollectionName.value.trim()
  if (!name) return

  try {
    await caseStore.createCollection(name)
  } catch (error) {
    console.error('Failed to create collection:', error)
  }
  caseStore.showNewCollectionModal = false
  newCollectionName.value = ''
}

function cancelCreateCollection() {
  caseStore.showNewCollectionModal = false
  newCollectionName.value = ''
}

function handleNewCollectionKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && newCollectionName.value.trim()) {
    createCollection()
  } else if (e.key === 'Escape') {
    cancelCreateCollection()
  }
}

function toggleCollection(name: string) {
  caseStore.toggleCollectionExpand(name)
}

function selectCollection(name: string) {
  tabStore.openTab('collection', name)
}

function selectCase(caseId: string) {
  tabStore.openTab('case', caseId)
}

function openConnectionsTab() {
  tabStore.openTab('connections', 'connections')
}

function confirmDeleteCollection(collection: Collection) {
  closeCollectionMenu()
  deleteType.value = 'collection'
  deleteTargetId.value = collection.name
  deleteTargetName.value = collection.name
  showDeleteModal.value = true
}

// Case actions
async function createNewCase(collectionName: string) {
  try {
    await caseStore.createCase(collectionName)
  } catch (error) {
    console.error('Failed to create case:', error)
  }
}

function confirmDeleteCase(caseId: string) {
  const caseItem = caseStore.allCases.find(c => c.id === caseId)
  if (caseItem) {
    deleteType.value = 'case'
    deleteTargetId.value = caseId
    deleteTargetName.value = caseItem.name
    showDeleteModal.value = true
  }
}

async function executeDelete() {
  if (!deleteTargetId.value) return

  try {
    if (deleteType.value === 'collection') {
      await caseStore.deleteCollection(deleteTargetId.value)
    } else {
      await caseStore.deleteCase(deleteTargetId.value)
    }
  } catch (error) {
    console.error('Delete failed:', error)
  }

  showDeleteModal.value = false
  deleteTargetId.value = null
}

function cancelDelete() {
  showDeleteModal.value = false
  deleteTargetId.value = null
}

// Rename actions
async function startRenameCase(caseId: string) {
  const caseItem = caseStore.allCases.find(c => c.id === caseId)
  if (caseItem) {
    renameType.value = 'case'
    renameTargetId.value = caseId
    renameName.value = caseItem.name
    showRenameModal.value = true
    await nextTick()
    renameInput.value?.focus()
    renameInput.value?.select()
  }
}

async function startRenameCollection(collectionName: string) {
  closeCollectionMenu()
  renameType.value = 'collection'
  renameTargetId.value = collectionName
  renameName.value = collectionName
  showRenameModal.value = true
  await nextTick()
  renameInput.value?.focus()
  renameInput.value?.select()
}

async function executeRename() {
  if (!renameTargetId.value || !renameName.value.trim()) return

  if (renameType.value === 'collection') {
    await caseStore.renameCollection(renameTargetId.value, renameName.value.trim())
  } else {
    caseStore.updateCase(renameTargetId.value, { name: renameName.value.trim() })
  }

  showRenameModal.value = false
  renameTargetId.value = null
}

function cancelRename() {
  showRenameModal.value = false
  renameTargetId.value = null
}

function handleRenameKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    executeRename()
  } else if (e.key === 'Escape') {
    cancelRename()
  }
}

// Context menu actions
function handleContextRename() {
  if (!contextMenu.value.target) return
  if (contextMenu.value.type === 'collection') {
    startRenameCollection(contextMenu.value.target)
  } else {
    startRenameCase(contextMenu.value.target)
  }
  closeContextMenu()
}

function handleContextDelete() {
  if (!contextMenu.value.target) return
  if (contextMenu.value.type === 'collection') {
    const collection = caseStore.collections.find(c => c.name === contextMenu.value.target)
    if (collection) confirmDeleteCollection(collection)
  } else {
    confirmDeleteCase(contextMenu.value.target)
  }
  closeContextMenu()
}

function handleContextNewCase() {
  if (!contextMenu.value.target || contextMenu.value.type !== 'collection') return
  createNewCase(contextMenu.value.target)
  closeContextMenu()
}

function handleContextOpen() {
  if (!contextMenu.value.target) return
  if (contextMenu.value.type === 'collection') {
    selectCollection(contextMenu.value.target)
  } else {
    selectCase(contextMenu.value.target)
  }
  closeContextMenu()
}
</script>

<template>
  <aside
    class="bg-bg-secondary border-r border-border-custom flex flex-col h-full relative"
    :style="{ width: sidebarWidth + 'px' }"
  >
    <!-- Resize Handle -->
    <div
      class="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-accent-primary/50 transition-colors z-10"
      :class="{ 'bg-accent-primary/50': isResizing }"
      @mousedown="startResize"
    ></div>

    <!-- Collections Header -->
    <div class="p-3 border-b border-border-custom">
      <div class="flex items-center justify-between">
        <h2 class="text-xs font-semibold text-zinc-400 uppercase tracking-wider">Collections</h2>
        <button
          @click="openNewCollectionModal"
          class="p-1.5 rounded-lg bg-accent-primary/10 hover:bg-accent-primary/20 text-accent-primary transition-colors"
          title="New collection"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
          </svg>
        </button>
      </div>
    </div>

    <!-- Collection Tree -->
    <div class="flex-1 overflow-y-auto p-2 space-y-1">
      <!-- Draggable Collection Items -->
      <VueDraggable
        :model-value="caseStore.collections"
        :animation="150"
        handle=".collection-drag-handle"
        ghost-class="opacity-50"
        @update:model-value="onCollectionUpdate"
      >
        <div
          v-for="collection in caseStore.collections"
          :key="collection.name"
          class="group"
        >
          <!-- Collection Header -->
          <div
            class="flex items-center gap-2 px-2 py-1.5 rounded-lg cursor-pointer hover:bg-zinc-700/50 transition-colors"
            :class="{ 'bg-bg-tertiary/30': tabStore.activeCollectionName === collection.name && !tabStore.activeCaseId }"
            @click="selectCollection(collection.name)"
            @contextmenu="handleCollectionContextMenu($event, collection.name)"
          >
            <!-- Drag Handle -->
            <svg
              class="collection-drag-handle w-3 h-3 text-zinc-600 hover:text-zinc-400 cursor-grab shrink-0 opacity-0 group-hover:opacity-100 transition-opacity"
              fill="currentColor"
              viewBox="0 0 24 24"
            >
              <path d="M8 6a2 2 0 1 1-4 0 2 2 0 0 1 4 0zm0 6a2 2 0 1 1-4 0 2 2 0 0 1 4 0zm0 6a2 2 0 1 1-4 0 2 2 0 0 1 4 0zm8-12a2 2 0 1 1-4 0 2 2 0 0 1 4 0zm0 6a2 2 0 1 1-4 0 2 2 0 0 1 4 0zm0 6a2 2 0 1 1-4 0 2 2 0 0 1 4 0z"/>
            </svg>

            <!-- Expand/Collapse Arrow -->
            <svg
              class="w-3 h-3 text-zinc-500 transition-transform shrink-0 hover:text-zinc-300"
              :class="{ 'rotate-90': collection.isExpanded }"
              fill="currentColor"
              viewBox="0 0 24 24"
              @click.stop="toggleCollection(collection.name)"
            >
              <path d="M8 5l8 7-8 7V5z"/>
            </svg>

            <!-- Folder Icon -->
            <svg class="w-4 h-4 text-yellow-500 shrink-0" fill="currentColor" viewBox="0 0 24 24">
              <path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
            </svg>

            <!-- Collection Name -->
            <span class="flex-1 text-sm text-zinc-300 truncate">{{ collection.name }}</span>

            <!-- Case Count -->
            <span class="text-xs text-zinc-500">{{ collection.cases.length }}</span>

            <!-- Actions -->
            <div class="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
              <button
                @click.stop="createNewCase(collection.name)"
                class="p-1 rounded hover:bg-bg-tertiary text-zinc-500 hover:text-zinc-300"
                title="Add case"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                </svg>
              </button>

              <!-- Menu button -->
              <div ref="collectionMenuRef" class="relative">
                <button
                  @click.stop="toggleCollectionMenu($event, collection.name)"
                  class="p-1 rounded hover:bg-bg-tertiary text-zinc-500 hover:text-zinc-200 transition-all"
                  :class="{ 'opacity-100': showCollectionMenu === collection.name }"
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
                  v-if="showCollectionMenu === collection.name"
                  class="absolute right-0 top-full mt-1 z-50 bg-bg-secondary border border-border-custom rounded-lg shadow-xl py-1 min-w-[120px]"
                >
                  <button
                    @click.stop="startRenameCollection(collection.name)"
                    class="w-full px-3 py-1.5 text-left text-sm text-zinc-300 hover:bg-bg-tertiary flex items-center gap-2"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
                    </svg>
                    Rename
                  </button>
                  <button
                    @click.stop="confirmDeleteCollection(collection)"
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
          </div>

          <!-- Cases in Collection (Draggable) -->
          <div v-if="collection.isExpanded" class="ml-4 mt-1">
            <VueDraggable
              :model-value="collection.cases"
              :animation="150"
              handle=".case-drag-handle"
              ghost-class="opacity-50"
              @end="() => onCaseDragEnd(collection)"
              @update:model-value="(val: Case[]) => { collection.cases = val }"
            >
              <CaseItem
                v-for="c in collection.cases"
                :key="c.id"
                :case-data="c"
                :is-active="c.id === tabStore.activeCaseId"
                @select="selectCase(c.id)"
                @remove="confirmDeleteCase(c.id)"
                @rename="startRenameCase(c.id)"
                @contextmenu.native="handleCaseContextMenu($event, c.id)"
              />
            </VueDraggable>

            <!-- Empty collection state -->
            <div
              v-if="collection.cases.length === 0"
              class="px-2 py-2 text-xs text-zinc-600 italic"
            >
              No cases
            </div>
          </div>
        </div>
      </VueDraggable>

      <!-- Empty state (no collections) -->
      <div
        v-if="caseStore.collections.length === 0"
        class="flex flex-col items-center justify-center py-8 text-zinc-500"
      >
        <svg class="w-12 h-12 mb-3 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
        </svg>
        <p class="text-sm">No collections</p>
        <p class="text-xs text-zinc-600">Click + to create one</p>
      </div>
    </div>

    <!-- Stats (clickable to open connections tab) -->
    <div
      class="p-3 border-t border-border-custom bg-bg-primary/50 cursor-pointer hover:bg-bg-tertiary transition-colors"
      @click="openConnectionsTab"
    >
      <div class="flex items-center justify-between text-xs">
        <span class="text-zinc-500 group-hover:text-zinc-300">Active connections</span>
        <span class="text-accent-success font-medium">{{ caseStore.connectedCases.length }}</span>
      </div>
    </div>

    <!-- Context Menu -->
    <Teleport to="body">
      <!-- Backdrop to catch clicks outside -->
      <div
        v-if="contextMenu.show"
        class="fixed inset-0 z-40"
        @click="closeContextMenu"
        @contextmenu.prevent="closeContextMenu"
      ></div>
      <div
        v-if="contextMenu.show"
        class="fixed z-50 bg-bg-secondary border border-border-custom rounded-lg shadow-xl py-1 min-w-[140px]"
        :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
        @click.stop
      >
        <!-- Open -->
        <button
          @click="handleContextOpen"
          class="w-full px-3 py-1.5 text-left text-sm text-zinc-300 hover:bg-bg-tertiary transition-colors flex items-center gap-2"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
          </svg>
          Open
        </button>

        <!-- New Case (only for collections) -->
        <button
          v-if="contextMenu.type === 'collection'"
          @click="handleContextNewCase"
          class="w-full px-3 py-1.5 text-left text-sm text-zinc-300 hover:bg-bg-tertiary transition-colors flex items-center gap-2"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
          </svg>
          New Case
        </button>

        <div class="h-px bg-border-custom my-1"></div>

        <!-- Rename -->
        <button
          @click="handleContextRename"
          class="w-full px-3 py-1.5 text-left text-sm text-zinc-300 hover:bg-bg-tertiary transition-colors flex items-center gap-2"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/>
          </svg>
          Rename
        </button>

        <!-- Delete -->
        <button
          @click="handleContextDelete"
          class="w-full px-3 py-1.5 text-left text-sm text-accent-error hover:bg-bg-tertiary transition-colors flex items-center gap-2"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
          </svg>
          Delete
        </button>
      </div>
    </Teleport>

    <!-- Delete Confirmation Modal -->
    <Teleport to="body">
      <div
        v-if="showDeleteModal"
        class="fixed inset-0 z-50 flex items-center justify-center"
        @click.self="cancelDelete"
      >
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm"></div>

        <!-- Modal -->
        <div class="relative bg-bg-secondary border border-border-custom rounded-xl shadow-2xl w-full max-w-sm p-6">
          <h3 class="text-lg font-medium text-zinc-100 mb-2">
            Delete {{ deleteType === 'collection' ? 'Collection' : 'Case' }}
          </h3>
          <p class="text-sm text-zinc-400 mb-4">
            Are you sure you want to delete "<span class="text-zinc-200">{{ deleteTargetName }}</span>"?
            <span v-if="deleteType === 'collection'" class="block mt-1 text-accent-error">
              This will delete all cases inside.
            </span>
          </p>

          <div class="flex justify-end gap-3">
            <button
              @click="cancelDelete"
              class="px-4 py-2 text-sm rounded-lg bg-bg-tertiary hover:bg-border-custom text-zinc-400 transition-colors"
            >
              Cancel
            </button>
            <button
              @click="executeDelete"
              class="px-4 py-2 text-sm rounded-lg bg-accent-error hover:bg-accent-error/80 text-white transition-colors"
            >
              Delete
            </button>
          </div>
        </div>
      </div>

      <!-- Rename Modal -->
      <div
        v-if="showRenameModal"
        class="fixed inset-0 z-50 flex items-center justify-center"
        @click.self="cancelRename"
      >
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm"></div>

        <!-- Modal -->
        <div class="relative bg-bg-secondary border border-border-custom rounded-xl shadow-2xl w-full max-w-sm p-6">
          <h3 class="text-lg font-medium text-zinc-100 mb-4">
            Rename {{ renameType === 'collection' ? 'Collection' : 'Case' }}
          </h3>

          <input
            ref="renameInput"
            v-model="renameName"
            @keydown="handleRenameKeydown"
            type="text"
            :placeholder="renameType === 'collection' ? 'Collection name...' : 'Case name...'"
            class="w-full px-4 py-3 bg-bg-tertiary border border-border-custom rounded-lg
                   text-zinc-200 placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-accent-primary/50 mb-4"
          />

          <div class="flex justify-end gap-3">
            <button
              @click="cancelRename"
              class="px-4 py-2 text-sm rounded-lg bg-bg-tertiary hover:bg-border-custom text-zinc-400 transition-colors"
            >
              Cancel
            </button>
            <button
              @click="executeRename"
              :disabled="!renameName.trim()"
              class="px-4 py-2 text-sm rounded-lg bg-accent-primary hover:bg-accent-primary/80 text-white disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Rename
            </button>
          </div>
        </div>
      </div>

      <!-- New Collection Modal -->
      <div
        v-if="caseStore.showNewCollectionModal"
        class="fixed inset-0 z-50 flex items-center justify-center"
        @click.self="cancelCreateCollection"
      >
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/60 backdrop-blur-sm"></div>

        <!-- Modal -->
        <div class="relative bg-bg-secondary border border-border-custom rounded-xl shadow-2xl w-full max-w-sm p-6">
          <h3 class="text-lg font-medium text-zinc-100 mb-4">New Collection</h3>

          <input
            ref="newCollectionInput"
            v-model="newCollectionName"
            @keydown="handleNewCollectionKeydown"
            type="text"
            placeholder="Collection name..."
            class="w-full px-4 py-3 bg-bg-tertiary border border-border-custom rounded-lg
                   text-zinc-200 placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-accent-primary/50 mb-4"
          />

          <div class="flex justify-end gap-3">
            <button
              @click="cancelCreateCollection"
              class="px-4 py-2 text-sm rounded-lg bg-bg-tertiary hover:bg-border-custom text-zinc-400 transition-colors"
            >
              Cancel
            </button>
            <button
              @click="createCollection"
              :disabled="!newCollectionName.trim()"
              class="px-4 py-2 text-sm rounded-lg bg-accent-primary hover:bg-accent-primary/80 text-white disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Create
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </aside>
</template>
