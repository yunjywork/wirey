<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'
import { useCaseStore } from '@/stores/case'
import Header from '@/components/layout/Header.vue'
import Sidebar from '@/components/layout/Sidebar.vue'
import MainContent from '@/components/layout/MainContent.vue'

const caseStore = useCaseStore()

// Message event data type
interface SocketMessageEvent {
  caseId: string
  content: string
  rawBytes: string
  size: number
  localAddr: string
  remoteAddr: string
  timestamp: number  // Unix milliseconds from Go
  framingInfo: {
    mode: string
    frameHeader?: string
    frameFooter?: string
    payloadSize: number
    totalSize: number
    settings?: string
  }
}

onMounted(() => {
  // Clean up any existing listeners first
  EventsOff('socket:data', 'socket:sent', 'socket:status', 'socket:error', 'script:log', 'collection:varsUpdated')

  // Listen for received data from backend
  EventsOn('socket:data', (data: SocketMessageEvent) => {
    caseStore.addMessage(data.caseId, 'received', data.content, 'text', {
      rawBytes: data.rawBytes,
      size: data.size,
      localAddr: data.localAddr,
      remoteAddr: data.remoteAddr,
      framingInfo: data.framingInfo,
      timestamp: data.timestamp
    })
  })

  // Listen for sent data from backend
  EventsOn('socket:sent', (data: SocketMessageEvent) => {
    caseStore.addMessage(data.caseId, 'sent', data.content, 'text', {
      rawBytes: data.rawBytes,
      size: data.size,
      localAddr: data.localAddr,
      remoteAddr: data.remoteAddr,
      framingInfo: data.framingInfo,
      timestamp: data.timestamp
    })
  })

  // Listen for connection status changes
  EventsOn('socket:status', (data: {
    caseId: string
    status: string
    localAddr?: string
    remoteAddr?: string
    protocol?: string
    reason?: string
    duration?: number
    bytesSent?: number
    bytesRecv?: number
  }) => {
    const status = data.status as 'connected' | 'disconnected' | 'error'
    caseStore.updateCaseStatus(data.caseId, status, data.localAddr)

    // Add system message for connection status
    if (status === 'connected') {
      caseStore.addSystemMessage(data.caseId, 'connected', 'Connected', {
        localAddr: data.localAddr,
        remoteAddr: data.remoteAddr,
        protocol: data.protocol
      })
    } else if (status === 'disconnected') {
      caseStore.addSystemMessage(data.caseId, 'disconnected', 'Disconnected', {
        reason: data.reason as 'user' | 'server' | 'error',
        duration: data.duration,
        bytesSent: data.bytesSent,
        bytesRecv: data.bytesRecv
      })
    }
  })

  // Listen for errors
  EventsOn('socket:error', (data: { caseId: string; error: string }) => {
    caseStore.addSystemMessage(data.caseId, 'error', data.error)
    caseStore.updateCaseStatus(data.caseId, 'error')
  })

  // Listen for script logs (wirey.log())
  EventsOn('script:log', (data: { caseId: string; message: string; timestamp: number }) => {
    caseStore.addSystemMessage(data.caseId, 'script', data.message)
  })

  // Listen for collection variable updates (from wirey.collection.set())
  EventsOn('collection:varsUpdated', (data: { collectionName: string; variables: Record<string, unknown> }) => {
    caseStore.updateCollectionVariables(data.collectionName, data.variables)
  })
})

onUnmounted(() => {
  EventsOff('socket:data', 'socket:sent', 'socket:status', 'socket:error', 'script:log', 'collection:varsUpdated')
})
</script>

<template>
  <div class="h-screen flex flex-col bg-bg-primary">
    <Header />
    <div class="flex-1 flex overflow-hidden">
      <Sidebar />
      <MainContent />
    </div>
  </div>
</template>
