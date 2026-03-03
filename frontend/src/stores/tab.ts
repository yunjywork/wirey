import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Tab, TabType } from '@/types'

export const useTabStore = defineStore('tab', () => {
  // State
  const openTabs = ref<Tab[]>([])
  const activeTabKey = ref<string | null>(null)  // 'case:id', 'collection:name', 'echo', 'connections'

  // Helpers
  function getTabKey(tab: Tab): string {
    return tab.type === 'echo' || tab.type === 'connections'
      ? tab.type
      : `${tab.type}:${tab.id}`
  }

  // Getters
  const activeTab = computed(() =>
    openTabs.value.find(t => getTabKey(t) === activeTabKey.value)
  )

  // Convenience getters for backward compatibility
  const activeCaseId = computed(() => {
    if (activeTabKey.value?.startsWith('case:')) {
      return activeTabKey.value.slice(5)
    }
    return null
  })

  const activeCollectionName = computed(() => {
    if (activeTabKey.value?.startsWith('collection:')) {
      return activeTabKey.value.slice(11)
    }
    return null
  })

  // Actions
  function openTab(type: TabType, id: string) {
    const tab: Tab = { type, id }
    const key = getTabKey(tab)

    const exists = openTabs.value.some(t => getTabKey(t) === key)
    if (!exists) {
      openTabs.value.push(tab)
    }
    activeTabKey.value = key
  }

  function closeTab(tab: Tab) {
    const key = getTabKey(tab)
    const index = openTabs.value.findIndex(t => getTabKey(t) === key)
    if (index === -1) return

    openTabs.value.splice(index, 1)

    // If closing the active tab, switch to another
    if (activeTabKey.value === key) {
      if (openTabs.value.length > 0) {
        const newIndex = Math.min(index, openTabs.value.length - 1)
        activeTabKey.value = getTabKey(openTabs.value[newIndex])
      } else {
        activeTabKey.value = null
      }
    }
  }

  function selectTab(tab: Tab) {
    activeTabKey.value = getTabKey(tab)
  }

  function isActiveTab(tab: Tab): boolean {
    return activeTabKey.value === getTabKey(tab)
  }

  function closeOthers(tab: Tab) {
    const key = getTabKey(tab)
    openTabs.value = openTabs.value.filter(t => getTabKey(t) === key)
    activeTabKey.value = key
  }

  function closeToLeft(tab: Tab) {
    const key = getTabKey(tab)
    const index = openTabs.value.findIndex(t => getTabKey(t) === key)
    if (index <= 0) return
    openTabs.value.splice(0, index)
    // If active tab was closed, switch to the target tab
    if (!openTabs.value.some(t => getTabKey(t) === activeTabKey.value)) {
      activeTabKey.value = key
    }
  }

  function closeToRight(tab: Tab) {
    const key = getTabKey(tab)
    const index = openTabs.value.findIndex(t => getTabKey(t) === key)
    if (index === -1 || index === openTabs.value.length - 1) return
    openTabs.value.splice(index + 1)
    // If active tab was closed, switch to the target tab
    if (!openTabs.value.some(t => getTabKey(t) === activeTabKey.value)) {
      activeTabKey.value = key
    }
  }

  return {
    // State
    openTabs,
    activeTabKey,
    // Getters
    activeTab,
    activeCaseId,
    activeCollectionName,
    // Actions
    openTab,
    closeTab,
    closeOthers,
    closeToLeft,
    closeToRight,
    selectTab,
    isActiveTab,
    getTabKey
  }
})
