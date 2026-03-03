<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useTabStore } from '@/stores/tab'
import { useCaseStore } from '@/stores/case'
import { useEchoStore } from '@/stores/echo'
import CaseTabContent from '@/components/case/CaseTabContent.vue'
import CollectionPanel from '@/components/collection/CollectionPanel.vue'
import EchoPanel from '@/components/echo/EchoPanel.vue'
import ConnectionsPanel from '@/components/connections/ConnectionsPanel.vue'
import type { Tab, Case } from '@/types'

const tabStore = useTabStore()
const caseStore = useCaseStore()
const echoStore = useEchoStore()

// Unsaved changes dialog state
const unsavedDialog = ref<{
  show: boolean
  pendingTabs: Tab[]
  currentCase: Case | null
  afterAction: (() => void) | null
}>({
  show: false,
  pendingTabs: [],
  currentCase: null,
  afterAction: null
})

// Context menu state
const contextMenu = ref<{ show: boolean; x: number; y: number; tab: Tab | null }>({
  show: false,
  x: 0,
  y: 0,
  tab: null
})

function handleTabContextMenu(e: MouseEvent, tab: Tab) {
  e.preventDefault()
  contextMenu.value = {
    show: true,
    x: e.clientX,
    y: e.clientY,
    tab
  }
}

function closeContextMenu() {
  contextMenu.value.show = false
}

function handleContextClose() {
  if (contextMenu.value.tab) {
    const tab = contextMenu.value.tab
    const unsavedCase = hasUnsavedChanges(tab)
    if (unsavedCase) {
      unsavedDialog.value = {
        show: true,
        pendingTabs: [tab],
        currentCase: unsavedCase,
        afterAction: () => tabStore.closeTab(tab)
      }
    } else {
      tabStore.closeTab(tab)
    }
  }
  closeContextMenu()
}

function handleContextCloseOthers() {
  if (contextMenu.value.tab) {
    const keepTab = contextMenu.value.tab
    const tabsToClose = tabStore.openTabs.filter(
      t => tabStore.getTabKey(t) !== tabStore.getTabKey(keepTab)
    )
    showUnsavedDialogForTabs(tabsToClose, () => tabStore.closeOthers(keepTab))
  }
  closeContextMenu()
}

function handleContextCloseToLeft() {
  if (contextMenu.value.tab) {
    const tab = contextMenu.value.tab
    const index = tabStore.openTabs.findIndex(
      t => tabStore.getTabKey(t) === tabStore.getTabKey(tab)
    )
    const tabsToClose = tabStore.openTabs.slice(0, index)
    showUnsavedDialogForTabs(tabsToClose, () => tabStore.closeToLeft(tab))
  }
  closeContextMenu()
}

function handleContextCloseToRight() {
  if (contextMenu.value.tab) {
    const tab = contextMenu.value.tab
    const index = tabStore.openTabs.findIndex(
      t => tabStore.getTabKey(t) === tabStore.getTabKey(tab)
    )
    const tabsToClose = tabStore.openTabs.slice(index + 1)
    showUnsavedDialogForTabs(tabsToClose, () => tabStore.closeToRight(tab))
  }
  closeContextMenu()
}

// Close context menu when clicking elsewhere
function handleClickOutside() {
  if (contextMenu.value.show) {
    closeContextMenu()
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

const statusClass: Record<string, string> = {
  connected: 'status-connected',
  connecting: 'status-connecting',
  disconnected: 'status-disconnected',
  error: 'bg-accent-error'
}

// Get tab display info
function getTabInfo(tab: Tab) {
  if (tab.type === 'case') {
    const c = caseStore.getCaseById(tab.id)
    return {
      name: c?.name || 'Unknown',
      status: c?.status || 'disconnected',
      isSaved: c?.isSaved ?? true,
      isCase: true
    }
  } else if (tab.type === 'echo') {
    return {
      name: 'Echo Server',
      isEcho: true,
      isSaved: true,
      isRunning: echoStore.isRunning
    }
  } else if (tab.type === 'connections') {
    return {
      name: 'Connections',
      isConnections: true,
      isSaved: true,
      count: caseStore.connectedCases.length
    }
  } else {
    // collection
    const col = caseStore.getCollectionByName(tab.id)
    return {
      name: col?.name || 'Unknown',
      caseCount: col?.cases.length || 0,
      isSaved: true,
      isCollection: true
    }
  }
}

// Check if a tab has unsaved changes
function hasUnsavedChanges(tab: Tab): Case | null {
  if (tab.type !== 'case') return null
  const caseItem = caseStore.getCaseById(tab.id)
  if (caseItem && !caseItem.isSaved) {
    return caseItem
  }
  return null
}

// Get all unsaved cases from a list of tabs
function getUnsavedTabs(tabs: Tab[]): Tab[] {
  return tabs.filter(tab => hasUnsavedChanges(tab) !== null)
}

// Show unsaved dialog for tabs
function showUnsavedDialogForTabs(tabs: Tab[], afterAction: () => void) {
  const unsavedTabs = getUnsavedTabs(tabs)
  if (unsavedTabs.length === 0) {
    afterAction()
    return
  }

  const firstUnsaved = hasUnsavedChanges(unsavedTabs[0])
  unsavedDialog.value = {
    show: true,
    pendingTabs: unsavedTabs,
    currentCase: firstUnsaved,
    afterAction
  }
}

function handleCloseTab(e: Event, tab: Tab) {
  e.stopPropagation()
  const unsavedCase = hasUnsavedChanges(tab)
  if (unsavedCase) {
    unsavedDialog.value = {
      show: true,
      pendingTabs: [tab],
      currentCase: unsavedCase,
      afterAction: () => tabStore.closeTab(tab)
    }
  } else {
    tabStore.closeTab(tab)
  }
}

function handleSelectTab(tab: Tab) {
  tabStore.selectTab(tab)
}

// Check if showing specific tab types
const showingCase = computed(() => tabStore.activeCaseId !== null)
const showingCollection = computed(() =>
  tabStore.activeCollectionName !== null && tabStore.activeCaseId === null
)
const showingEcho = computed(() => tabStore.activeTabKey === 'echo')
const showingConnections = computed(() => tabStore.activeTabKey === 'connections')

// Get current active collection name for new case button
const currentCollectionName = computed(() => {
  if (tabStore.activeCaseId) {
    const c = caseStore.getCaseById(tabStore.activeCaseId)
    return c?.collectionName || null
  }
  return tabStore.activeCollectionName
})

// Unsaved dialog handlers
function handleDialogCancel() {
  unsavedDialog.value = {
    show: false,
    pendingTabs: [],
    currentCase: null,
    afterAction: null
  }
}

async function handleDialogDontSave() {
  const { pendingTabs, afterAction } = unsavedDialog.value

  // Reload all unsaved cases from backend (discard changes)
  for (const tab of pendingTabs) {
    if (tab.type === 'case') {
      try {
        await caseStore.reloadCase(tab.id)
      } catch {
        // Case might not exist in backend yet - just mark as saved
        const caseItem = caseStore.getCaseById(tab.id)
        if (caseItem) {
          caseItem.isSaved = true
        }
      }
    }
  }

  // Execute the pending action
  if (afterAction) {
    afterAction()
  }

  handleDialogCancel()
}

async function handleDialogSave() {
  const { pendingTabs, afterAction } = unsavedDialog.value

  // Save all unsaved cases
  for (const tab of pendingTabs) {
    if (tab.type === 'case') {
      try {
        await caseStore.saveCase(tab.id)
      } catch (error) {
        console.error('Failed to save case:', error)
      }
    }
  }

  // Execute the pending action
  if (afterAction) {
    afterAction()
  }

  handleDialogCancel()
}

// Create new case
async function createNewCase() {
  if (!currentCollectionName.value) return
  try {
    await caseStore.createCase(currentCollectionName.value)
  } catch (error) {
    console.error('Failed to create case:', error)
  }
}

function openCreateCollectionModal() {
  caseStore.showNewCollectionModal = true
}
</script>

<template>
  <main class="flex-1 flex flex-col h-full overflow-hidden bg-bg-primary">
    <!-- Tabs Bar -->
    <div
      v-if="tabStore.openTabs.length > 0"
      class="flex items-center border-b border-border-custom bg-bg-secondary shrink-0"
    >
      <!-- Tabs -->
      <div class="flex-1 flex items-center overflow-x-auto">
        <button
          v-for="tab in tabStore.openTabs"
          :key="`${tab.type}-${tab.id}`"
          @click="handleSelectTab(tab)"
          @contextmenu="handleTabContextMenu($event, tab)"
          :class="[
            'group flex items-center gap-2 px-3 py-2 text-sm border-r border-border-custom transition-colors relative shrink-0',
            tabStore.isActiveTab(tab)
              ? 'bg-bg-primary text-zinc-100'
              : 'bg-bg-secondary hover:bg-bg-tertiary text-zinc-400'
          ]"
        >
          <!-- Case: Status dot -->
          <span
            v-if="tab.type === 'case'"
            class="status-dot shrink-0"
            :class="statusClass[getTabInfo(tab).status as string]"
          ></span>

          <!-- Echo: Terminal icon with running indicator -->
          <span v-else-if="tab.type === 'echo'" class="relative shrink-0">
            <svg class="w-3.5 h-3.5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
            </svg>
            <span
              v-if="getTabInfo(tab).isRunning"
              class="absolute -top-0.5 -right-0.5 w-1.5 h-1.5 bg-green-500 rounded-full"
            ></span>
          </span>

          <!-- Connections: Network icon -->
          <span v-else-if="tab.type === 'connections'" class="relative shrink-0">
            <svg class="w-3.5 h-3.5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0"/>
            </svg>
            <span
              v-if="getTabInfo(tab).count && getTabInfo(tab).count > 0"
              class="absolute -top-1 -right-1.5 min-w-[14px] h-[14px] bg-green-500 rounded-full text-[9px] text-white flex items-center justify-center font-medium"
            >{{ getTabInfo(tab).count }}</span>
          </span>

          <!-- Collection: Folder icon -->
          <svg
            v-else
            class="w-3.5 h-3.5 text-yellow-500 shrink-0"
            fill="currentColor"
            viewBox="0 0 24 24"
          >
            <path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
          </svg>

          <!-- Tab name -->
          <span
            class="max-w-[120px] truncate"
            :class="{ 'text-accent-warning': !getTabInfo(tab).isSaved }"
          >{{ getTabInfo(tab).name }}</span>

          <!-- Unsaved indicator / Close button -->
          <span
            @click="handleCloseTab($event, tab)"
            class="p-0.5 rounded hover:bg-zinc-600 transition-colors relative"
            :class="[
              getTabInfo(tab).isSaved
                ? 'text-zinc-500 hover:text-zinc-200'
                : 'text-accent-warning hover:text-zinc-200'
            ]"
          >
            <!-- Unsaved dot (hidden on hover) -->
            <span
              v-if="!getTabInfo(tab).isSaved"
              class="w-3 h-3 flex items-center justify-center group-hover:hidden"
            >
              <span class="w-2 h-2 bg-accent-warning rounded-full"></span>
            </span>
            <!-- Close X (always visible when saved, show on hover when unsaved) -->
            <svg
              class="w-3 h-3"
              :class="{ 'hidden group-hover:block': !getTabInfo(tab).isSaved }"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </span>

          <!-- Active indicator -->
          <span
            v-if="tabStore.isActiveTab(tab)"
            class="absolute bottom-0 left-0 right-0 h-0.5 bg-accent-primary"
          ></span>
        </button>
      </div>

      <!-- New tab button -->
      <button
        v-if="currentCollectionName"
        @click="createNewCase"
        class="p-2 text-zinc-500 hover:text-zinc-200 hover:bg-bg-tertiary transition-colors shrink-0"
        title="New case"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
        </svg>
      </button>
    </div>

    <!-- Case tabs with KeepAlive (preserves state per tab) -->
    <!-- v-show keeps DOM alive so KeepAlive cache persists when switching to collection/echo tabs -->
    <div v-show="showingCase" class="flex-1 flex flex-col overflow-hidden">
      <KeepAlive :max="10">
        <CaseTabContent
          v-if="caseStore.activeCase"
          :key="tabStore.activeCaseId"
          :case-data="caseStore.activeCase"
        />
      </KeepAlive>
    </div>

    <!-- Collection tabs with KeepAlive (preserves state per collection) -->
    <div v-show="showingCollection" class="flex-1 flex flex-col overflow-hidden">
      <KeepAlive :max="10">
        <CollectionPanel
          v-if="caseStore.activeCollection"
          :key="tabStore.activeCollectionName"
          :collection="caseStore.activeCollection"
        />
      </KeepAlive>
    </div>

    <!-- Echo Server page -->
    <KeepAlive>
      <EchoPanel v-if="showingEcho" />
    </KeepAlive>

    <!-- Connections page -->
    <ConnectionsPanel v-if="showingConnections" />

    <!-- Landing page (no tabs open) -->
    <div v-if="tabStore.openTabs.length === 0" class="flex-1 flex items-center justify-center">
      <div class="text-center">
        <!-- Logo -->
        <div class="w-20 h-20 mx-auto mb-6 rounded-2xl bg-bg-secondary border border-border-custom flex items-center justify-center">
          <svg class="w-10 h-10 text-accent-primary" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2L2 7l10 5 10-5-10-5z"/>
            <path d="M2 17l10 5 10-5"/>
            <path d="M2 12l10 5 10-5"/>
          </svg>
        </div>
        <!-- Title -->
        <h1 class="text-4xl font-bold text-accent-primary mb-2">Wirey</h1>
        <!-- Tagline -->
        <p class="text-lg text-zinc-400 mb-8">The Modern TCP/Socket Test Client</p>
        <!-- CTA Button -->
        <button
          @click="openCreateCollectionModal"
          class="inline-flex items-center gap-2 px-6 py-3 bg-accent-primary hover:opacity-80 text-white font-medium rounded-lg transition-opacity"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
          </svg>
          Create Collection
        </button>
      </div>
    </div>

    <!-- Tab Context Menu -->
    <Teleport to="body">
      <div
        v-if="contextMenu.show"
        class="fixed z-50 bg-bg-secondary border border-border-custom rounded-lg shadow-xl py-1 min-w-[160px]"
        :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
        @click.stop
      >
        <button
          @click="handleContextClose"
          class="w-full px-3 py-1.5 text-left text-sm text-zinc-300 hover:bg-bg-tertiary transition-colors"
        >
          Close
        </button>
        <button
          @click="handleContextCloseOthers"
          :disabled="tabStore.openTabs.length <= 1"
          class="w-full px-3 py-1.5 text-left text-sm text-zinc-300 hover:bg-bg-tertiary transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Close Others
        </button>
        <div class="h-px bg-border-custom my-1"></div>
        <button
          @click="handleContextCloseToLeft"
          :disabled="!contextMenu.tab || tabStore.openTabs.findIndex(t => tabStore.getTabKey(t) === tabStore.getTabKey(contextMenu.tab!)) === 0"
          class="w-full px-3 py-1.5 text-left text-sm text-zinc-300 hover:bg-bg-tertiary transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Close to the Left
        </button>
        <button
          @click="handleContextCloseToRight"
          :disabled="!contextMenu.tab || tabStore.openTabs.findIndex(t => tabStore.getTabKey(t) === tabStore.getTabKey(contextMenu.tab!)) === tabStore.openTabs.length - 1"
          class="w-full px-3 py-1.5 text-left text-sm text-zinc-300 hover:bg-bg-tertiary transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Close to the Right
        </button>
      </div>
    </Teleport>

    <!-- Unsaved Changes Dialog -->
    <Teleport to="body">
      <div
        v-if="unsavedDialog.show"
        class="fixed inset-0 z-50 flex items-center justify-center"
      >
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/50" @click="handleDialogCancel"></div>

        <!-- Dialog -->
        <div class="relative bg-bg-secondary border border-border-custom rounded-lg shadow-xl w-full max-w-md p-6">
          <!-- Icon -->
          <div class="w-12 h-12 mx-auto mb-4 rounded-full bg-yellow-500/10 flex items-center justify-center">
            <svg class="w-6 h-6 text-yellow-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
            </svg>
          </div>

          <!-- Title -->
          <h3 class="text-lg font-semibold text-zinc-100 text-center mb-2">
            Unsaved Changes
          </h3>

          <!-- Message -->
          <p class="text-sm text-zinc-400 text-center mb-6">
            <template v-if="unsavedDialog.pendingTabs.length === 1">
              Do you want to save changes to "{{ unsavedDialog.currentCase?.name }}"?
            </template>
            <template v-else>
              {{ unsavedDialog.pendingTabs.length }} cases have unsaved changes. Save all?
            </template>
          </p>

          <!-- Buttons -->
          <div class="flex justify-end gap-3">
            <button
              @click="handleDialogCancel"
              class="px-4 py-2 text-sm text-zinc-400 hover:text-zinc-200 transition-colors"
            >
              Cancel
            </button>
            <button
              @click="handleDialogDontSave"
              class="px-4 py-2 text-sm bg-zinc-700 hover:bg-zinc-600 text-zinc-200 rounded-lg transition-colors"
            >
              Don't Save
            </button>
            <button
              @click="handleDialogSave"
              class="px-4 py-2 text-sm bg-accent-primary hover:opacity-80 text-white rounded-lg transition-opacity"
            >
              Save
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </main>
</template>
