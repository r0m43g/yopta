<!-- frontend/src/components/common/DateSelector.vue -->
<template>
  <div class="date-selector">
    <div class="input-group">
      <input
        :value="formattedDate"
        @input="handleInput"
        @blur="handleBlur"
        type="text"
        class="input input-bordered"
        :class="{ 'input-error': !isValid && isDirty }"
        :placeholder="placeholder"
      />
      <button class="btn" @click="toggleCalendar">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
      </button>
    </div>

    <div v-if="showCalendar" class="calendar-popup">
      <div class="calendar-header">
        <button @click="prevMonth" class="btn btn-sm">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <div class="month-year">{{ monthYearLabel }}</div>
        <button @click="nextMonth" class="btn btn-sm">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </div>

      <div class="calendar-days">
        <div v-for="day in weekDays" :key="day" class="day-name">{{ day }}</div>
      </div>

      <div class="calendar-dates">
        <div
          v-for="(date, index) in calendarDates"
          :key="index"
          class="date-cell"
          :class="{
            'other-month': date.otherMonth,
            'selected': isSelectedDate(date.date),
            'today': isToday(date.date)
          }"
          @click="selectDate(date.date)"
        >
          {{ date.day }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, toRef } from 'vue';

/**
 * The Date Selecting Wizard - helping users navigate the calendar realm!
 *
 * This component offers a user-friendly date selection interface with:
 * - Text input for direct date entry (Lithuanian format)
 * - Calendar popup for visual date picking
 * - Validation to ensure dates make sense (no February 30th allowed!)
 *
 * It's like having a tiny calendar app right inside your form field.
 */

const props = defineProps({
  // The currently selected date
  modelValue: {
    type: [Date, String, null],
    default: null
  },
  // Placeholder text for empty input
  placeholder: {
    type: String,
    default: 'Pasirinkite datą'
  },
  // Format for displaying the date
  format: {
    type: String,
    default: 'lt' // Lithuanian format by default
  }
});

const emit = defineEmits(['update:modelValue']);

// Internal state
const showCalendar = ref(false);
const currentMonth = ref(new Date());
const inputValue = ref('');
const isValid = ref(true);
const isDirty = ref(false);

// Lithuanian week days (starting with Monday)
const weekDays = ['Pr', 'An', 'Tr', 'Kt', 'Pn', 'Št', 'Sk'];

/**
 * The Date Formatter - converts Date objects to human-readable strings!
 *
 * Formats the current date value according to the specified format.
 * Lithuanian format: YYYY-MM-DD
 *
 * @returns {String} - The formatted date string
 */
const formattedDate = computed(() => {
  if (!props.modelValue) return '';

  const date = props.modelValue instanceof Date
    ? props.modelValue
    : new Date(props.modelValue);

  if (isNaN(date.getTime())) return '';

  if (props.format === 'lt') {
    // Lithuanian format: YYYY-MM-DD
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    return `${year}-${month}-${day}`;
  }

  // Default to Lithuanian locale string
  return date.toLocaleDateString('lt-LT');
});

/**
 * The Calendar Labeler - creates a friendly month and year identifier!
 *
 * Generates a human-readable label for the current month and year.
 *
 * @returns {String} - Month and year label (e.g., "Kovas 2025")
 */
const monthYearLabel = computed(() => {
  const months = [
    'Sausis', 'Vasaris', 'Kovas', 'Balandis', 'Gegužė', 'Birželis',
    'Liepa', 'Rugpjūtis', 'Rugsėjis', 'Spalis', 'Lapkritis', 'Gruodis'
  ];

  const month = months[currentMonth.value.getMonth()];
  const year = currentMonth.value.getFullYear();

  return `${month} ${year}`;
});

/**
 * The Calendar Generator - builds the visual month grid!
 *
 * Creates an array of dates for the current month view,
 * including some days from the previous and next months
 * to fill out the calendar grid.
 *
 * @returns {Array} - Calendar dates for display
 */
const calendarDates = computed(() => {
  const year = currentMonth.value.getFullYear();
  const month = currentMonth.value.getMonth();

  // Get the first day of the month
  const firstDay = new Date(year, month, 1);
  // Get day of week (0-6, where 0 is Sunday)
  // Convert to Monday-based (0-6, where 0 is Monday)
  let firstDayOfWeek = firstDay.getDay() - 1;
  if (firstDayOfWeek < 0) firstDayOfWeek = 6; // Sunday becomes 6

  // Get the last day of the month
  const lastDay = new Date(year, month + 1, 0);
  const daysInMonth = lastDay.getDate();

  // Get the last day of the previous month
  const prevMonthLastDay = new Date(year, month, 0).getDate();

  const dates = [];

  // Add days from the previous month
  for (let i = 0; i < firstDayOfWeek; i++) {
    const day = prevMonthLastDay - firstDayOfWeek + i + 1;
    dates.push({
      day,
      date: new Date(year, month - 1, day),
      otherMonth: true
    });
  }

  // Add days of the current month
  for (let i = 1; i <= daysInMonth; i++) {
    dates.push({
      day: i,
      date: new Date(year, month, i),
      otherMonth: false
    });
  }

  // Add days from the next month to complete the grid
  const remainingCells = 42 - dates.length; // 6 rows * 7 days
  for (let i = 1; i <= remainingCells; i++) {
    dates.push({
      day: i,
      date: new Date(year, month + 1, i),
      otherMonth: true
    });
  }

  return dates;
});

/**
 * The Date Comparer - checks if two dates are the same day!
 *
 * Determines if the given date is the currently selected date.
 *
 * @param {Date} date - The date to check
 * @returns {Boolean} - True if it's the selected date
 */
function isSelectedDate(date) {
  if (!props.modelValue) return false;

  const selectedDate = props.modelValue instanceof Date
    ? props.modelValue
    : new Date(props.modelValue);

  return date.getFullYear() === selectedDate.getFullYear() &&
         date.getMonth() === selectedDate.getMonth() &&
         date.getDate() === selectedDate.getDate();
}

/**
 * The Today Finder - identifies the current day!
 *
 * Checks if a date is today's date.
 *
 * @param {Date} date - The date to check
 * @returns {Boolean} - True if it's today
 */
function isToday(date) {
  const today = new Date();
  return date.getFullYear() === today.getFullYear() &&
         date.getMonth() === today.getMonth() &&
         date.getDate() === today.getDate();
}

/**
 * The Time Traveler - jumps to the previous month!
 *
 * Updates the calendar to show the previous month.
 */
function prevMonth() {
  currentMonth.value = new Date(
    currentMonth.value.getFullYear(),
    currentMonth.value.getMonth() - 1,
    1
  );
}

/**
 * The Future Explorer - jumps to the next month!
 *
 * Updates the calendar to show the next month.
 */
function nextMonth() {
  currentMonth.value = new Date(
    currentMonth.value.getFullYear(),
    currentMonth.value.getMonth() + 1,
    1
  );
}

/**
 * The Calendar Toggler - shows or hides the date picker!
 *
 * Toggles the visibility of the calendar popup.
 */
function toggleCalendar() {
  showCalendar.value = !showCalendar.value;
}

/**
 * The Date Chooser - finalizes the user's selection!
 *
 * Handles the selection of a date from the calendar.
 *
 * @param {Date} date - The selected date
 */
function selectDate(date) {
  emit('update:modelValue', date);
  showCalendar.value = false;
  isDirty.value = true;
}

/**
 * The Text Parser - interprets keyboard input as dates!
 *
 * Handles direct text input in the date field,
 * attempting to parse it into a valid date.
 *
 * @param {Event} event - The input event
 */
function handleInput(event) {
  inputValue.value = event.target.value;
  isDirty.value = true;

  // If input is empty, clear the date
  if (!inputValue.value.trim()) {
    isValid.value = true;
    emit('update:modelValue', null);
    return;
  }

  // Try to parse the input as a date
  const parsedDate = parseDate(inputValue.value);
  isValid.value = parsedDate !== null;

  if (isValid.value) {
    emit('update:modelValue', parsedDate);
  }
}

/**
 * The Input Finalizer - validates the date when focus leaves!
 *
 * Handles the blur event for the input field,
 * finalizing validation and formatting.
 */
function handleBlur() {
  // If the input is empty, clear the date
  if (!inputValue.value.trim()) {
    isValid.value = true;
    emit('update:modelValue', null);
    return;
  }

  // Try to parse the input as a date
  const parsedDate = parseDate(inputValue.value);

  if (parsedDate) {
    // Valid date, update the model and reset the input
    isValid.value = true;
    emit('update:modelValue', parsedDate);
    inputValue.value = formattedDate.value;
  } else {
    // Invalid date, mark as invalid but don't change the model
    isValid.value = false;
  }
}

/**
 * The Date Interpreter - converts text into Date objects!
 *
 * Parses a date string in various formats,
 * with special handling for Lithuanian format.
 *
 * @param {String} dateStr - The date string to parse
 * @returns {Date|null} - Parsed date or null if invalid
 */
function parseDate(dateStr) {
  // Check for Lithuanian format: YYYY-MM-DD or YYYY.MM.DD
  const ltMatch = dateStr.match(/^(\d{4})[-.](\d{1,2})[-.](\d{1,2})$/);
  if (ltMatch) {
    const year = parseInt(ltMatch[1], 10);
    const month = parseInt(ltMatch[2], 10) - 1; // 0-11
    const day = parseInt(ltMatch[3], 10);

    // Validate date values
    if (month < 0 || month > 11 || day < 1 || day > 31) {
      return null;
    }

    const date = new Date(year, month, day);

    // Check if the date is valid (e.g., not February 30)
    if (date.getFullYear() === year && date.getMonth() === month && date.getDate() === day) {
      return date;
    }

    return null;
  }

  // Try the browser's native date parsing
  const date = new Date(dateStr);
  return isNaN(date.getTime()) ? null : date;
}

// Update input value when model value changes
watch(toRef(props, 'modelValue'), () => {
  inputValue.value = formattedDate.value;
});

// Close calendar when clicking outside
function handleClickOutside(event) {
  const element = event.target;
  const dateSelector = document.querySelector('.date-selector');

  if (dateSelector && !dateSelector.contains(element)) {
    showCalendar.value = false;
  }
}

// Initialize component state
onMounted(() => {
  // Set up initial month view based on model value or current date
  if (props.modelValue) {
    const initialDate = props.modelValue instanceof Date
      ? props.modelValue
      : new Date(props.modelValue);

    if (!isNaN(initialDate.getTime())) {
      currentMonth.value = new Date(
        initialDate.getFullYear(),
        initialDate.getMonth(),
        1
      );
    }
  }

  inputValue.value = formattedDate.value;

  // Add click outside listener
  document.addEventListener('click', handleClickOutside);
});

// Clean up event listeners
onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>

<style scoped>
.date-selector {
  position: relative;
  width: 100%;
}

.input-group {
  display: flex;
}

.input-group input {
  flex: 1;
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
}

.input-group button {
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
}

.calendar-popup {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background-color: white;
  border: 1px solid #ccc;
  border-radius: 0.5rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  z-index: 10;
  margin-top: 0.25rem;
  padding: 0.75rem;
  max-width: 300px;
}

.calendar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.month-year {
  font-weight: bold;
}

.calendar-days {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 0.25rem;
  margin-bottom: 0.25rem;
}

.day-name {
  text-align: center;
  font-size: 0.75rem;
  font-weight: bold;
  padding: 0.25rem 0;
  color: #666;
}

.calendar-dates {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  grid-auto-rows: minmax(2rem, auto);
  gap: 0.125rem;
}

.date-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.25rem;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: background-color 0.2s;
}

.date-cell:hover {
  background-color: rgba(0, 0, 0, 0.05);
}

.date-cell.other-month {
  color: #ccc;
}

.date-cell.selected {
  background-color: #3b82f6;
  color: white;
}

.date-cell.today:not(.selected) {
  border: 1px solid #3b82f6;
}

@media (max-width: 640px) {
  .calendar-popup {
    position: fixed;
    top: auto;
    bottom: 0;
    left: 0;
    right: 0;
    max-width: 100%;
    border-radius: 0.5rem 0.5rem 0 0;
  }

  .calendar-dates {
    grid-auto-rows: minmax(2.5rem, auto);
  }
}
</style>
