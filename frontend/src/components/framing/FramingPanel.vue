<script setup lang="ts">
import { computed } from 'vue'
import type { Case, FramingConfig } from '@/types'
import { useCaseStore } from '@/stores/case'
import FramingOptions from './FramingOptions.vue'

const props = defineProps<{
  caseData: Case
}>()

const caseStore = useCaseStore()

// Check if case is connected (disable editing)
const isDisabled = computed(() =>
  props.caseData.status === 'connected' || props.caseData.status === 'connecting'
)

// Get collection's shared framing for preview
const collectionFraming = computed<FramingConfig | undefined>(() => {
  const collection = caseStore.findCaseCollection(props.caseData.id)
  return collection?.sharedFraming
})

// Framing config with v-model support
const framingConfig = computed({
  get: () => props.caseData.framing,
  set: (value: FramingConfig) => {
    caseStore.updateCase(props.caseData.id, { framing: value })
  }
})
</script>

<template>
  <div class="p-6 space-y-6">
    <!-- Lock notice when connected -->
    <div v-if="isDisabled" class="p-3 bg-yellow-900/20 border border-yellow-700/50 rounded-lg text-xs text-yellow-400">
      Cannot change framing settings while connected. Disconnect first to modify.
    </div>

    <FramingOptions
      v-model="framingConfig"
      :disabled="isDisabled"
      :show-collection-option="true"
      :collection-framing="collectionFraming"
    />

    <!-- Info -->
    <div class="text-xs text-zinc-500 pt-4 border-t border-border-custom">
      <p>Framing settings apply to both send and receive operations.</p>
    </div>
  </div>
</template>
