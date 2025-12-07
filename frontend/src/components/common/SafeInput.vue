<!-- frontend/src/components/common/SafeInput.vue -->
<template>
 <div>
   <label v-if="label" class="input tooltip tooltip-right">
     <input
       :value="modelValue"
       @input="sanitizeInput"
       :type="type"
       :placeholder="label"
       class="grow"
       :class="{ 'input-error': hasError }"
     />
     <div class="tooltip-content" v-if="hasError">
       <span class="text-error">{{ errorMessage }}</span>
     </div>
   </label>
 </div>
</template>

<script setup>
import { computed } from 'vue';
import { sanitizeText } from '@/utils/xssSanitizer';

// Props for our super-safe input component
// These props define the component's API, making it flexible yet secure
const props = defineProps({
 // The v-model value - like a pet cat that comes and goes as it pleases
 modelValue: {
   type: String,
   default: ''
 },
 // Label text - the polite introduction to your input field
 label: {
   type: String,
   default: ''
 },
 // Input type - the secret identity of our input field
 type: {
   type: String,
   default: 'text'
 },
 // Placeholder - the ghostly hint that disappears when you start typing
 placeholder: {
   type: String,
   default: ''
 },
 // Validation pattern - the bouncer that decides which values get in
 pattern: {
   type: String,
   default: ''
 },
 // Error message - the passive-aggressive note when validation fails
 errorMessage: {
   type: String,
   default: 'Netinkama reikšmė'
 }
});

// Emit events for v-model to work its two-way binding magic
const emit = defineEmits(['update:modelValue']);

// Computed property to check if input matches the pattern
// It's like a spell-checker, but for your form's sanity
const hasError = computed(() => {
 if (!props.pattern || !props.modelValue) return false;
 const regex = new RegExp(props.pattern);
 return !regex.test(props.modelValue);
});

// Sanitize input to prevent XSS attacks
// Like washing your hands before eating, but for data
const sanitizeInput = (event) => {
 // Get raw, potentially dangerous value
 const value = event.target.value;

 // Clean it up like your room before company arrives
 const sanitizedValue = sanitizeText(value);

 // Check if someone tried to be sneaky with XSS
 if (value !== sanitizedValue) {
   console.warn('Aptikta ir sustabdyta galima XSS ataka', {
     original: value,
     sanitized: sanitizedValue
   });
 }

 // Emit the clean value to parent component
 emit('update:modelValue', sanitizedValue);
};
</script>
