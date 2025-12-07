<!-- frontend/components/PasswordStrength.vue -->
<template>
  <div class="flex flex-col mt-1">
    <div class="flex flex-row gap-1">
      <div v-for="n in 4" :key="n"
        class="h-1 rounded-sm w-1/4 transition-colors"
        :class="n <= strengthScore ? strengthColor : 'bg-red-500'">
      </div>
    </div>
    <span class="bg-green-500"></span>
    <span class="bg-yellow-500"></span>
    <span class="bg-orange-500"></span>
    <span class="bg-red-500"></span>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

/**
 * The Secret Decoder Ring - evaluates your password's resistance to evil-doers!
 *
 * This component visually represents password strength with a row of colored bars,
 * like a digital fortress wall that grows stronger with each security requirement met.
 * Red means "a child could guess this," while green says "even the NSA is sweating."
 */

// The sacred incantation: accept a password prop from the parent component
const props = defineProps(['password'])

// Current strength - starts at zero like a baby fortress
const strength = ref(0)

/**
 * The Digital Bouncer - judges if your password can enter the elite security club!
 *
 * Evaluates password strength based on several criteria:
 * 1. Length (is it long enough to be interesting?)
 * 2. Uppercase letters (SHOUTING adds security)
 * 3. Numbers (because 1337 speak is still cool in security circles)
 * 4. Special characters (because @#$% is how hackers feel when they see them)
 *
 * Returns a score from 0-4, where 0 means "change this immediately" and
 * 4 means "your password is wearing a digital kevlar vest"
 */
const strengthScore = computed(() => {
  let score = 0
  // Criterion 1: Length - Size matters in password security
  if (props.password.length >= 8) score++
  // Criterion 2: Uppercase - CAPS LOCK IS SECURITY
  if (/[A-Z]/.test(props.password)) score++
  // Criterion 3: Numbers - Because 123456 is still the most common password
  if (/[0-9]/.test(props.password)) score++
  // Criterion 4: Special characters - Punctuation that actually protects
  if (/[^A-Za-z0-9]/.test(props.password)) score++

  // Update the strength ref (for potential use elsewhere)
  strength.value = score

  return score
})

/**
 * The Color Coordinator - picks the perfect shade to represent your security status!
 *
 * Like a fashion consultant for your password security indicator, this function
 * selects colors that match your password's strength level.
 * Red = Danger Will Robinson!
 * Orange = Meh, could be worse.
 * Yellow = Not bad, but don't get cocky.
 * Green = Fort Knox would be impressed.
 */
const strengthColor = computed(() => {
  const colors = ['red', 'orange', 'yellow', 'green']
  return `bg-${colors[strength.value - 1]}-500`
})

/**
 * The Fashion Police - assigns CSS classes based on password strength!
 *
 * Provides appropriate CSS classes to style text that might accompany
 * the strength indicator. Red for embarrassing passwords, green for
 * passwords that deserve a security medal of honor.
 */
const strengthClass = computed(() => {
  return {
    'text-red-500': strength.value < 2,
    'text-orange-500': strength.value === 2,
    'text-green-500': strength.value > 2
  }
})
</script>
