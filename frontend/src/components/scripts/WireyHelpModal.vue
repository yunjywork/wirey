<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  visible: boolean
  insertDisabled?: boolean
}>()

const emit = defineEmits<{
  close: []
  insert: [code: string]
}>()

// Search filter
const searchQuery = ref('')

// Helper function definitions
const helpers = [
  {
    category: 'Variables',
    items: [
      { name: 'wirey.get', signature: '(key) → value', desc: 'Get case variable', example: `const val = wirey.get('myVar');` },
      { name: 'wirey.set', signature: '(key, value)', desc: 'Set case variable', example: `wirey.set('myVar', 123);` },
      { name: 'wirey.collection.get', signature: '(key) → value', desc: 'Get collection variable (persisted)', example: `const seq = wirey.collection.get('seq') ?? 0;` },
      { name: 'wirey.collection.set', signature: '(key, value)', desc: 'Set collection variable (persisted)', example: `wirey.collection.set('seq', seq + 1);` },
    ]
  },
  {
    category: 'Utilities',
    items: [
      { name: 'wirey.uuid', signature: '() → string', desc: 'Generate UUID v4', example: `const id = wirey.uuid();` },
      { name: 'wirey.log', signature: '(...args)', desc: 'Log to message panel', example: `wirey.log('Debug:', value);` },
      { name: 'wirey.randomHex', signature: '(n?) → string', desc: 'Random hex string (n bytes)', example: `const hex = wirey.randomHex(4); // "A1B2C3D4"` },
    ]
  },
  {
    category: 'Byte Conversion',
    items: [
      { name: 'wirey.toHex', signature: '(str) → string', desc: 'String to uppercase hex', example: `const hex = wirey.toHex('Hi'); // "4869"` },
      { name: 'wirey.fromHex', signature: '(hex) → string', desc: 'Hex to string (spaces allowed)', example: `const str = wirey.fromHex('48 69'); // "Hi"` },
      { name: 'wirey.toBytes', signature: '(str) → byte[]', desc: 'String to byte array', example: `const bytes = wirey.toBytes('Hi'); // [72, 105]` },
      { name: 'wirey.fromBytes', signature: '(bytes) → string', desc: 'Byte array to string', example: `const str = wirey.fromBytes([72, 105]); // "Hi"` },
    ]
  },
  {
    category: 'Byte Manipulation',
    items: [
      { name: 'wirey.byteAt', signature: '(str, index) → number', desc: 'Get byte value at position', example: `const b = wirey.byteAt(msg, 0); // first byte` },
      { name: 'wirey.setByteAt', signature: '(str, index, value) → string', desc: 'Set byte value at position', example: `msg = wirey.setByteAt(msg, 0, 0x01);` },
      { name: 'wirey.subBytes', signature: '(str, start, end?) → string', desc: 'Extract bytes as string', example: `const header = wirey.subBytes(msg, 0, 4);` },
      { name: 'wirey.appendBytes', signature: '(str, bytes) → string', desc: 'Append bytes to string', example: `msg = wirey.appendBytes(msg, [0x0D, 0x0A]);` },
      { name: 'wirey.replaceBytes', signature: '(str, start, bytes) → string', desc: 'Replace bytes at position', example: `msg = wirey.replaceBytes(msg, 0, [0x01, 0x02]);` },
    ]
  },
  {
    category: 'HTTP',
    items: [
      { name: 'wirey.httpGet', signature: '(url) → {status, body, error}', desc: 'HTTP GET request (10s timeout)', example: `const res = wirey.httpGet('https://api.myip.com');
if (!res.error) wirey.log('My IP:', JSON.parse(res.body).ip);` },
      { name: 'wirey.httpPost', signature: '(url, body, contentType?) → {status, body, error}', desc: 'HTTP POST request (10s timeout)', example: `const res = wirey.httpPost('https://jsonplaceholder.typicode.com/users', JSON.stringify({name: 'John', email: 'john@test.com'}));
if (res.status === 201) wirey.set('userId', JSON.parse(res.body).id);` },
    ]
  }
]

// Filtered helpers based on search
const filteredHelpers = computed(() => {
  if (!searchQuery.value.trim()) return helpers

  const query = searchQuery.value.toLowerCase()
  return helpers.map(cat => ({
    ...cat,
    items: cat.items.filter(item =>
      item.name.toLowerCase().includes(query) ||
      item.desc.toLowerCase().includes(query)
    )
  })).filter(cat => cat.items.length > 0)
})

function handleInsert(example: string) {
  if (props.insertDisabled) return
  emit('insert', example)
  emit('close')
}

function handleBackdropClick(e: MouseEvent) {
  if (e.target === e.currentTarget) {
    emit('close')
  }
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      @click="handleBackdropClick"
    >
      <div class="bg-bg-secondary border border-border-custom rounded-xl shadow-2xl w-[600px] max-h-[80vh] flex flex-col">
        <!-- Header -->
        <div class="flex items-center justify-between px-4 py-3 border-b border-border-custom">
          <h2 class="text-lg font-semibold text-zinc-100">? wirey Helper Functions</h2>
          <button
            @click="$emit('close')"
            class="p-1 rounded hover:bg-bg-tertiary text-zinc-400 hover:text-zinc-200 transition-colors"
          >
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <!-- Search -->
        <div class="px-4 py-2 border-b border-border-custom">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search functions..."
            class="w-full px-3 py-2 bg-bg-tertiary border border-border-custom rounded-lg text-zinc-200 text-sm placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-accent-primary/50"
          />
        </div>

        <!-- Content -->
        <div class="flex-1 overflow-y-auto p-4 space-y-4">
          <div v-for="category in filteredHelpers" :key="category.category">
            <h3 class="text-sm font-medium text-zinc-400 mb-2">{{ category.category }}</h3>
            <div class="space-y-2">
              <div
                v-for="item in category.items"
                :key="item.name"
                :class="[
                  'bg-bg-tertiary border border-border-custom rounded-lg p-3 transition-colors',
                  insertDisabled ? 'cursor-default' : 'hover:border-accent-primary/50 cursor-pointer group'
                ]"
                @click="handleInsert(item.example)"
              >
                <div class="flex items-start justify-between gap-2">
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center gap-2">
                      <code class="text-accent-primary font-mono text-sm">{{ item.name }}</code>
                      <span class="text-zinc-500 text-xs font-mono">{{ item.signature }}</span>
                    </div>
                    <p class="text-zinc-400 text-xs mt-1">{{ item.desc }}</p>
                    <pre class="text-zinc-300 text-xs mt-2 font-mono bg-bg-primary rounded px-2 py-1 overflow-x-auto">{{ item.example }}</pre>
                  </div>
                  <span v-if="!insertDisabled" class="text-xs text-zinc-500 group-hover:text-accent-primary shrink-0">Click to insert</span>
                </div>
              </div>
            </div>
          </div>

          <div v-if="filteredHelpers.length === 0" class="text-center text-zinc-500 py-8">
            No matching functions found
          </div>
        </div>

        <!-- Footer -->
        <div class="px-4 py-3 border-t border-border-custom text-xs text-zinc-500">
          <template v-if="insertDisabled">
            Enable the script to insert example code
          </template>
          <template v-else>
            Click any function to insert example code into the current script
          </template>
        </div>
      </div>
    </div>
  </Teleport>
</template>
