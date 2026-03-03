import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Case, Collection, Message, ConnectionStatus, MessageFormat, Protocol, FramingConfig, FramingMeta, SystemMessageType, DisconnectReason } from '@/types'
import { DEFAULT_FRAMING, DEFAULT_ECHO_PORT } from '@/types'
import {
  LoadCollections,
  CreateCollection as CreateCollectionApi,
  UpdateCollection as UpdateCollectionApi,
  DeleteCollection as DeleteCollectionApi,
  RenameCollection as RenameCollectionApi,
  SaveCase as SaveCaseApi,
  LoadCase as LoadCaseApi,
  DeleteCase as DeleteCaseApi,
  MoveCase as MoveCaseApi,
  ReorderCollections as ReorderCollectionsApi,
  ReorderCases as ReorderCasesApi
} from '../../wailsjs/go/main/App'
import { config } from '../../wailsjs/go/models'
import {
  toBackendSavedCase,
  toBackendCollection,
  fromBackendCollectionWithCases,
  fromBackendSavedCase
} from '@/utils/backend'
import { useTabStore } from './tab'

export const useCaseStore = defineStore('case', () => {
  // State
  const collections = ref<Collection[]>([])
  const showNewCollectionModal = ref(false)  // Shared modal state

  // Getters (using tabStore for active state)
  const activeCollection = computed(() => {
    const tabStore = useTabStore()
    return collections.value.find(c => c.name === tabStore.activeCollectionName)
  })

  const activeCase = computed(() => {
    const tabStore = useTabStore()
    if (!tabStore.activeCaseId) return undefined
    for (const collection of collections.value) {
      const found = collection.cases.find(c => c.id === tabStore.activeCaseId)
      if (found) return found
    }
    return undefined
  })

  const allCases = computed(() =>
    collections.value.flatMap(c => c.cases)
  )

  const connectedCases = computed(() =>
    allCases.value.filter(c => c.status === 'connected')
  )

  // Helper functions
  function generateId(): string {
    return `${Date.now()}-${Math.random().toString(36).substring(2, 9)}`
  }

  function getNextUntitledNumber(): number {
    const untitledPattern = /^Untitled (\d+)$/
    let maxNumber = 0

    for (const c of allCases.value) {
      const match = c.name.match(untitledPattern)
      if (match) {
        maxNumber = Math.max(maxNumber, parseInt(match[1], 10))
      }
    }

    return maxNumber + 1
  }

  function findCaseCollection(caseId: string): Collection | undefined {
    return collections.value.find(col =>
      col.cases.some(c => c.id === caseId)
    )
  }

  function getCaseById(caseId: string): Case | undefined {
    for (const collection of collections.value) {
      const found = collection.cases.find(c => c.id === caseId)
      if (found) return found
    }
    return undefined
  }

  function getCollectionByName(name: string): Collection | undefined {
    return collections.value.find(c => c.name === name)
  }

  // Collection Actions
  async function loadCollections() {
    try {
      const loaded = await LoadCollections()
      collections.value = (loaded || []).map(fromBackendCollectionWithCases)
      restoreExpandedState()
    } catch (error) {
      console.error('Failed to load collections:', error)
    }
  }

  async function createCollection(name: string) {
    const tabStore = useTabStore()
    try {
      await CreateCollectionApi(name)
      // Add to local state with next order number
      const now = new Date()
      const maxOrder = collections.value.reduce((max, c) => Math.max(max, c.order || 0), 0)
      collections.value.push({
        name,
        createdAt: now,
        updatedAt: now,
        order: maxOrder + 1,
        cases: [],
        isExpanded: true
      })
      // Open the new collection tab
      tabStore.openTab('collection', name)
    } catch (error) {
      console.error('Failed to create collection:', error)
      throw error
    }
  }

  async function updateCollection(name: string, updates: Partial<Collection>) {
    const collection = collections.value.find(c => c.name === name)
    if (!collection) return

    Object.assign(collection, updates, { updatedAt: new Date() })

    try {
      await UpdateCollectionApi(name, toBackendCollection(collection))
    } catch (error) {
      console.error('Failed to update collection:', error)
      throw error
    }
  }

  async function deleteCollection(name: string) {
    const tabStore = useTabStore()
    try {
      await DeleteCollectionApi(name)
      const index = collections.value.findIndex(c => c.name === name)
      if (index !== -1) {
        collections.value.splice(index, 1)
      }
      // Close the collection tab if open
      tabStore.closeTab({ type: 'collection', id: name })
    } catch (error) {
      console.error('Failed to delete collection:', error)
      throw error
    }
  }

  async function renameCollection(oldName: string, newName: string) {
    const tabStore = useTabStore()
    try {
      await RenameCollectionApi(oldName, newName)
      const collection = collections.value.find(c => c.name === oldName)
      if (collection) {
        collection.name = newName
        // Update all cases' collectionName
        collection.cases.forEach(c => {
          c.collectionName = newName
        })
      }
      // Update tab if this collection is open
      const tabIndex = tabStore.openTabs.findIndex(t => t.type === 'collection' && t.id === oldName)
      if (tabIndex !== -1) {
        tabStore.openTabs[tabIndex].id = newName
        if (tabStore.activeTabKey === `collection:${oldName}`) {
          tabStore.activeTabKey = `collection:${newName}`
        }
      }
    } catch (error) {
      console.error('Failed to rename collection:', error)
      throw error
    }
  }

  function toggleCollectionExpand(name: string) {
    const collection = collections.value.find(c => c.name === name)
    if (collection) {
      collection.isExpanded = !collection.isExpanded
      saveExpandedState()
    }
  }

  // Update only sharedVariables for a collection (used by collection:varsUpdated event)
  function updateCollectionVariables(name: string, variables: Record<string, unknown>) {
    const collection = collections.value.find(c => c.name === name)
    if (collection) {
      collection.sharedVariables = variables
    }
  }

  function saveExpandedState() {
    const expanded = collections.value
      .filter(c => c.isExpanded)
      .map(c => c.name)
    localStorage.setItem('collectionExpanded', JSON.stringify(expanded))
  }

  function restoreExpandedState() {
    try {
      const saved = localStorage.getItem('collectionExpanded')
      if (saved) {
        const expanded = JSON.parse(saved) as string[]
        collections.value.forEach(c => {
          c.isExpanded = expanded.includes(c.name)
        })
      }
    } catch {
      // ignore
    }
  }

  // Case Actions
  async function createCase(collectionName: string, caseConfig?: {
    name?: string
    protocol?: Protocol
    host?: string
    port?: number
  }): Promise<Case> {
    const tabStore = useTabStore()
    const collection = collections.value.find(c => c.name === collectionName)
    if (!collection) {
      throw new Error(`Collection ${collectionName} not found`)
    }

    const id = generateId()
    const name = caseConfig?.name || `Untitled ${getNextUntitledNumber()}`

    // Calculate next order number for this collection
    const maxOrder = collection.cases.reduce((max, c) => Math.max(max, c.order || 0), 0)

    const newCase: Case = {
      id,
      name,
      collectionName,
      protocol: caseConfig?.protocol || 'tcp',
      host: caseConfig?.host || '127.0.0.1',
      port: caseConfig?.port || DEFAULT_ECHO_PORT,
      status: 'disconnected',
      createdAt: new Date(),
      order: maxOrder + 1,
      messages: [],
      isSaved: true,  // Will be saved immediately
      framing: { ...DEFAULT_FRAMING },
      charset: 'collection',  // Use collection's charset by default
      useVariables: true      // Enable variables by default
    }

    collection.cases.push(newCase)

    // Save immediately to backend
    try {
      await SaveCaseApi(collection.name, toBackendSavedCase(newCase))
    } catch (error) {
      console.error('Failed to save new case:', error)
      // Remove from local state if save failed
      const index = collection.cases.findIndex(c => c.id === id)
      if (index !== -1) {
        collection.cases.splice(index, 1)
      }
      throw error
    }

    // Open the new case tab
    tabStore.openTab('case', newCase.id)

    return newCase
  }

  function updateCase(caseId: string, updates: Partial<Case>) {
    for (const collection of collections.value) {
      const caseItem = collection.cases.find(c => c.id === caseId)
      if (caseItem) {
        Object.assign(caseItem, updates, { updatedAt: new Date() })
        // Mark as unsaved if substantive change
        if (!('isSaved' in updates)) {
          caseItem.isSaved = false
        }
        break
      }
    }
  }

  function updateCaseStatus(caseId: string, status: ConnectionStatus, localAddr?: string) {
    for (const collection of collections.value) {
      const caseItem = collection.cases.find(c => c.id === caseId)
      if (caseItem) {
        caseItem.status = status
        if (status === 'connected' && localAddr) {
          caseItem.localAddr = localAddr
        } else if (status === 'disconnected') {
          caseItem.localAddr = undefined
        }
        break
      }
    }
  }

  function addMessage(
    caseId: string,
    direction: 'sent' | 'received',
    content: string,
    format: MessageFormat = 'text',
    extra?: {
      rawBytes?: string       // base64 encoded
      size?: number
      localAddr?: string
      remoteAddr?: string
      framingInfo?: FramingMeta
      timestamp?: number      // Unix milliseconds from Go
    }
  ) {
    for (const collection of collections.value) {
      const caseItem = collection.cases.find(c => c.id === caseId)
      if (caseItem) {
        const message: Message = {
          id: generateId(),
          direction,
          content,
          format,
          timestamp: extra?.timestamp ? new Date(extra.timestamp) : new Date(),
          rawBytes: extra?.rawBytes,
          size: extra?.size,
          localAddr: extra?.localAddr,
          remoteAddr: extra?.remoteAddr,
          framingInfo: extra?.framingInfo
        }
        caseItem.messages.push(message)
        break
      }
    }
  }

  function clearMessages(caseId: string) {
    for (const collection of collections.value) {
      const caseItem = collection.cases.find(c => c.id === caseId)
      if (caseItem) {
        caseItem.messages = []
        break
      }
    }
  }

  function addSystemMessage(
    caseId: string,
    systemType: SystemMessageType,
    content: string,
    extra?: {
      localAddr?: string
      remoteAddr?: string
      protocol?: string
      reason?: DisconnectReason
      duration?: number
      bytesSent?: number
      bytesRecv?: number
    }
  ) {
    for (const collection of collections.value) {
      const caseItem = collection.cases.find(c => c.id === caseId)
      if (caseItem) {
        const message: Message = {
          id: generateId(),
          direction: 'system',
          content,
          format: 'text',
          timestamp: new Date(),
          systemType,
          localAddr: extra?.localAddr,
          remoteAddr: extra?.remoteAddr,
          protocol: extra?.protocol,
          reason: extra?.reason,
          duration: extra?.duration,
          bytesSent: extra?.bytesSent,
          bytesRecv: extra?.bytesRecv
        }
        caseItem.messages.push(message)
        break
      }
    }
  }

  async function saveCase(caseId: string) {
    const collection = findCaseCollection(caseId)
    if (!collection) return

    const caseItem = collection.cases.find(c => c.id === caseId)
    if (!caseItem) return

    try {
      await SaveCaseApi(collection.name, toBackendSavedCase(caseItem))
      caseItem.isSaved = true
    } catch (error) {
      console.error('Failed to save case:', error)
      throw error
    }
  }

  async function reloadCase(caseId: string) {
    const collection = findCaseCollection(caseId)
    if (!collection) return

    const caseItem = collection.cases.find(c => c.id === caseId)
    if (!caseItem) return

    try {
      const savedCase = await LoadCaseApi(collection.name, caseId)
      const loadedData = fromBackendSavedCase(savedCase, collection.name)

      // Preserve runtime state (status, messages, localAddr)
      Object.assign(caseItem, loadedData, {
        status: caseItem.status,
        messages: caseItem.messages,
        localAddr: caseItem.localAddr,
        isSaved: true
      })
    } catch (error) {
      console.error('Failed to reload case:', error)
      throw error
    }
  }

  async function deleteCase(caseId: string) {
    const tabStore = useTabStore()
    const collection = findCaseCollection(caseId)
    if (!collection) return

    const caseItem = collection.cases.find(c => c.id === caseId)
    if (!caseItem) return

    // Remove from backend if saved
    if (caseItem.isSaved) {
      try {
        await DeleteCaseApi(collection.name, caseId)
      } catch (error) {
        console.error('Failed to delete case:', error)
        throw error
      }
    }

    // Close the tab if open
    tabStore.closeTab({ type: 'case', id: caseId })

    // Remove from local state
    const index = collection.cases.findIndex(c => c.id === caseId)
    if (index !== -1) {
      collection.cases.splice(index, 1)
    }
  }

  async function moveCase(caseId: string, toCollectionName: string) {
    const fromCollection = findCaseCollection(caseId)
    if (!fromCollection) return
    if (fromCollection.name === toCollectionName) return

    const toCollection = collections.value.find(c => c.name === toCollectionName)
    if (!toCollection) return

    const caseIndex = fromCollection.cases.findIndex(c => c.id === caseId)
    if (caseIndex === -1) return

    const caseItem = fromCollection.cases[caseIndex]

    // Move in backend if saved
    if (caseItem.isSaved) {
      try {
        await MoveCaseApi(fromCollection.name, toCollectionName, caseId)
      } catch (error) {
        console.error('Failed to move case:', error)
        throw error
      }
    }

    // Move in local state
    fromCollection.cases.splice(caseIndex, 1)
    caseItem.collectionName = toCollectionName
    toCollection.cases.push(caseItem)
  }

  // Get effective framing config (resolves 'collection' mode)
  function getEffectiveFraming(caseId: string): FramingConfig {
    const collection = findCaseCollection(caseId)
    const caseItem = collection?.cases.find(c => c.id === caseId)

    if (!caseItem) {
      return { mode: 'none' }
    }

    if (caseItem.framing.mode === 'collection') {
      // Use collection's shared framing
      return collection?.sharedFraming || { mode: 'none' }
    }

    return caseItem.framing
  }

  // Reorder collections (after drag-drop)
  async function reorderCollections(orderedNames: string[]) {
    // Update local order values
    orderedNames.forEach((name, index) => {
      const col = collections.value.find(c => c.name === name)
      if (col) col.order = index
    })

    // Sort locally
    collections.value.sort((a, b) => (a.order || 0) - (b.order || 0))

    // Persist to backend
    try {
      const orders = orderedNames.map((name, index) =>
        new config.CollectionOrder({ name, order: index })
      )
      await ReorderCollectionsApi(orders)
    } catch (error) {
      console.error('Failed to reorder collections:', error)
    }
  }

  // Reorder cases within a collection (after drag-drop)
  async function reorderCases(collectionName: string, orderedIds: string[]) {
    const collection = collections.value.find(c => c.name === collectionName)
    if (!collection) return

    // Update local order values
    orderedIds.forEach((id, index) => {
      const caseItem = collection.cases.find(c => c.id === id)
      if (caseItem) caseItem.order = index
    })

    // Sort locally
    collection.cases.sort((a, b) => (a.order || 0) - (b.order || 0))

    // Persist to backend
    try {
      const orders = orderedIds.map((id, index) =>
        new config.CaseOrder({ id, order: index })
      )
      await ReorderCasesApi(collectionName, orders)
    } catch (error) {
      console.error('Failed to reorder cases:', error)
    }
  }

  return {
    // State
    collections,
    showNewCollectionModal,
    // Getters
    activeCollection,
    activeCase,
    allCases,
    connectedCases,
    // Collection Actions
    loadCollections,
    createCollection,
    updateCollection,
    updateCollectionVariables,
    deleteCollection,
    renameCollection,
    toggleCollectionExpand,
    // Case Actions
    createCase,
    updateCase,
    updateCaseStatus,
    addMessage,
    addSystemMessage,
    clearMessages,
    saveCase,
    reloadCase,
    deleteCase,
    moveCase,
    reorderCollections,
    reorderCases,
    // Helpers
    findCaseCollection,
    getCaseById,
    getCollectionByName,
    getEffectiveFraming
  }
})
