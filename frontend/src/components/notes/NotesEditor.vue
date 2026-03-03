<script setup lang="ts">
import { ref, watch, computed, onMounted, nextTick } from "vue";
import { MdEditor } from "md-editor-v3";
import type { ExposeParam } from "md-editor-v3";
import "md-editor-v3/lib/style.css";
import { useSettingsStore } from "@/stores/settings";

const props = defineProps<{
  modelValue: string;
}>();

const emit = defineEmits<{
  "update:modelValue": [value: string];
}>();

const settingsStore = useSettingsStore();

const editorTheme = computed(() => settingsStore.settings.theme);

const localValue = ref(props.modelValue || "");
const editorRef = ref<ExposeParam>();
const isPreviewOnly = ref(!!props.modelValue?.trim());
const isReady = ref(false);

// Set initial preview mode based on content
onMounted(async () => {
  await nextTick();
  if (isPreviewOnly.value) {
    editorRef.value?.togglePreviewOnly(true);
  }
  // Small delay to ensure mode is set before showing
  setTimeout(() => {
    isReady.value = true;
  }, 30);
});

// Sync with prop
watch(
  () => props.modelValue,
  (val) => {
    localValue.value = val || "";
  }
);

function toggleMode() {
  isPreviewOnly.value = !isPreviewOnly.value;
  editorRef.value?.togglePreviewOnly(isPreviewOnly.value);
}

// Debounced save
let saveTimeout: ReturnType<typeof setTimeout> | null = null;

function handleChange(value: string) {
  localValue.value = value;

  // Debounce auto-save
  if (saveTimeout) {
    clearTimeout(saveTimeout);
  }
  saveTimeout = setTimeout(() => {
    emit("update:modelValue", value);
  }, 800);
}
</script>

<template>
  <div
    class="notes-editor h-full"
    :class="{ 'opacity-0': !isReady, 'opacity-100': isReady, 'preview-mode': isPreviewOnly }"
    style="transition: opacity 0.1s"
  >
    <MdEditor
      ref="editorRef"
      v-model="localValue"
      @update:modelValue="handleChange"
      :theme="editorTheme"
      language="en-US"
      :toolbars="[
        0,
        '-',
        'bold',
        'underline',
        'italic',
        'strikeThrough',
        '-',
        'title',
        'sub',
        'sup',
        'quote',
        'unorderedList',
        'orderedList',
        'task',
        '-',
        'codeRow',
        'code',
        'link',
        'table',
        'mermaid',
        'katex',
        '-',
        'revoke',
        'next',
        '=',
        'preview',
      ]"
      :toolbarsExclude="[
        'github',
        'save',
        'fullscreen',
        'pageFullscreen',
        'catalog',
        'htmlPreview',
        'image',
        'previewOnly',
      ]"
      class="h-full !border-none"
    >
      <template #defToolbars>
        <div class="mode-switch" @click="toggleMode">
          <span :class="['mode-switch-btn', { active: !isPreviewOnly }]"
            >Edit</span
          >
          <span :class="['mode-switch-btn', { active: isPreviewOnly }]"
            >Preview</span
          >
        </div>
      </template>
    </MdEditor>
  </div>
</template>

<style>
/* Override md-editor dark theme to match app */
.md-editor-dark {
  --md-bk-color: rgb(var(--color-bg-tertiary)) !important;
  --md-color: #e4e4e7 !important;
  font-family: "Malgun Gothic", -apple-system, BlinkMacSystemFont, "Segoe UI",
    sans-serif !important;
}

.md-editor-dark .md-editor-input-wrapper textarea,
.md-editor-dark .cm-content {
  color: #e4e4e7 !important;
  font-family: "Malgun Gothic", -apple-system, BlinkMacSystemFont, "Segoe UI",
    Consolas, monospace !important;
}

.md-editor-dark .md-editor-preview {
  color: #e4e4e7 !important;
  font-family: "Malgun Gothic", -apple-system, BlinkMacSystemFont, "Segoe UI",
    sans-serif !important;
}

/* Override md-editor light theme to match app */
.md-editor-light {
  --md-bk-color: rgb(var(--color-bg-tertiary)) !important;
  --md-color: #18181b !important;
  font-family: "Malgun Gothic", -apple-system, BlinkMacSystemFont, "Segoe UI",
    sans-serif !important;
}

.md-editor-light .md-editor-input-wrapper textarea,
.md-editor-light .cm-content {
  color: #18181b !important;
  font-family: "Malgun Gothic", -apple-system, BlinkMacSystemFont, "Segoe UI",
    Consolas, monospace !important;
}

.md-editor-light .md-editor-preview {
  color: #18181b !important;
  font-family: "Malgun Gothic", -apple-system, BlinkMacSystemFont, "Segoe UI",
    sans-serif !important;
}

/* Mode switch styling - matches Text/Hex toggle */
.mode-switch {
  display: flex;
  border-radius: 8px;
  background: rgb(var(--color-bg-tertiary));
  padding: 2px;
  cursor: pointer;
}

.mode-switch-btn {
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 500;
  border-radius: 6px;
  color: #a1a1aa;
  transition: all 0.15s;
}

.mode-switch-btn.active {
  background: rgb(var(--color-accent-primary));
  color: white;
}

/* Disable toolbar buttons in preview mode (except Edit/Preview toggle) */
.preview-mode .md-editor-toolbar-wrapper .md-editor-toolbar-item:not(:has(.mode-switch)) {
  opacity: 0.3;
  pointer-events: none;
}
</style>
