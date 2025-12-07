<!-- frontend/src/components/common/EditableCell.vue -->
<template>
  <div class="editable-cell" :class="{ 'is-editing': isEditing }">
    <div v-if="!isEditing" class="cell-content" @click="startEditing">
      <slot name="display">{{ modelValue }}</slot>
      <span class="edit-icon" v-if="!disabled">✎</span>
    </div>
    <div v-else class="cell-editor">
      <input
        ref="inputRef"
        v-model="editValue"
        :type="type"
        class="cell-input"
        @blur="finishEditing"
        @keydown.enter="finishEditing"
        @keydown.esc="cancelEditing"
      />
      <div class="editor-actions">
        <button class="action-button save" @click="finishEditing" title="Išsaugoti">✓</button>
        <button class="action-button cancel" @click="cancelEditing" title="Atšaukti">✗</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, toRef } from 'vue'
import { useLoggingStore } from '@/stores/logging'

/**
 * The Cell Shapeshifter - turns boring static data into interactive adventures!
 *
 * This component makes any table cell editable with a click, transforming
 * mundane data viewing into exciting data modification. It's like giving your
 * users a digital eraser and pencil for each piece of information.
 *
 * Just don't give this power to users who've had too much coffee.
 */

const props = defineProps({
  // The current value (for v-model binding)
  modelValue: {
    type: [String, Number],
    default: ''
  },
  // Input field type
  type: {
    type: String,
    default: 'text'
  },
  // Whether editing is allowed
  disabled: {
    type: Boolean,
    default: false
  },
  // The record ID and field name (for logging)
  recordId: {
    type: String,
    default: ''
  },
  fieldName: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue', 'edit-complete'])

// Internal state
const isEditing = ref(false)
const editValue = ref('')
const inputRef = ref(null)
const initialValue = ref('')
const loggingStore = useLoggingStore()

// Watch for external value changes
watch(toRef(props, 'modelValue'), (newVal) => {
  if (!isEditing.value) {
    editValue.value = newVal
  }
})

/**
 * The Edit Initiator - begins the magical transformation!
 * Switches the cell to edit mode and focuses the input field.
 * Like opening a secret door to a hidden editing dimension.
 */
function startEditing() {
  if (props.disabled) return

  isEditing.value = true
  initialValue.value = props.modelValue
  editValue.value = props.modelValue

  // Hocus pocus, focus!
  setTimeout(() => {
    if (inputRef.value) {
      inputRef.value.focus()
      inputRef.value.select()
    }
  }, 10)

  // Log the edit action
  loggingStore.info('Pradėtas redagavimas', {
    component: 'EditableCell',
    recordId: props.recordId,
    fieldName: props.fieldName,
    action: 'start_editing'
  })
}

/**
 * The Change Confirmer - solidifies the user's masterpiece!
 * Completes the editing process and updates the value.
 * Like a wizard turning liquid potion ingredients into a solid crystal.
 */
function finishEditing() {
  if (!isEditing.value) return

  isEditing.value = false

  // Only emit update if value changed
  if (editValue.value !== props.modelValue) {
    emit('update:modelValue', editValue.value)
    emit('edit-complete', {
      recordId: props.recordId,
      field: props.fieldName,
      value: editValue.value,
      previousValue: initialValue.value
    })

    // Log the completed edit
    loggingStore.info('Redagavimas baigtas', {
      component: 'EditableCell',
      recordId: props.recordId,
      fieldName: props.fieldName,
      oldValue: initialValue.value,
      newValue: editValue.value,
      action: 'finish_editing'
    })
  }
}

/**
 * The Discarder of Changes - the "oops, nevermind" button!
 * Cancels editing and restores the original value.
 * Like hitting ctrl+z on that embarrassing typo before anyone sees it.
 */
function cancelEditing() {
  isEditing.value = false
  editValue.value = props.modelValue

  // Log the cancelled edit
  loggingStore.info('Redagavimas atšauktas', {
    component: 'EditableCell',
    recordId: props.recordId,
    fieldName: props.fieldName,
    action: 'cancel_editing'
  })
}

// Initialize the component
onMounted(() => {
  editValue.value = props.modelValue
})
</script>

<style scoped>
.editable-cell {
  position: relative;
  padding: 2px;
  border-radius: 4px;
  min-height: 22px;
  transition: all 0.2s ease;
}

.editable-cell:not(.is-editing):hover {
  background-color: rgba(0, 0, 0, 0.05);
  cursor: pointer;
}

.cell-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  min-height: 22px;
}

.edit-icon {
  opacity: 0;
  margin-left: 8px;
  font-size: 12px;
  color: #666;
  transition: opacity 0.2s ease;
}

.editable-cell:hover .edit-icon {
  opacity: 1;
}

.cell-editor {
  display: flex;
  align-items: center;
  width: 100%;
}

.cell-input {
  flex: 1;
  min-width: 50px;
  padding: 4px 8px;
  border: 1px solid #ccc;
  border-radius: 4px;
  font-size: inherit;
  font-family: inherit;
}

.editor-actions {
  display: flex;
  margin-left: 4px;
}

.action-button {
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 4px;
  background: none;
  cursor: pointer;
  padding: 0;
  font-size: 14px;
  margin-left: 2px;
}

.action-button.save {
  color: #48bb78;
}

.action-button.save:hover {
  background-color: rgba(72, 187, 120, 0.1);
}

.action-button.cancel {
  color: #f56565;
}

.action-button.cancel:hover {
  background-color: rgba(245, 101, 101, 0.1);
}

/* Adding responsive behavior */
@media (max-width: 768px) {
  .cell-input {
    font-size: 14px;
    padding: 6px;
  }

  .action-button {
    width: 30px;
    height: 30px;
    font-size: 16px;
  }
}
</style>
