<template lang="pug">
.tracks-timeline-container(
  ref="container" 
  :class="{ 'cursor-grabbing': isDragging }"
)
  canvas(
    ref="canvas" 
    @mousedown="handleMouseDown" 
    @mousemove="handleMouseMove" 
    @mouseup="handleMouseUp" 
    @mouseleave="handleMouseLeave"
    @wheel="handleWheel"
  )
  .timeline-empty(v-if="!hasTracksData") 
    p.text-xl.font-bold Nėra kelių duomenų pasirinktai stočiai
  .loading-indicator(v-if="isInitializing")
    span.loading.loading-spinner.loading-lg
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { useLoggingStore } from '@/stores/logging'
import { useSlogStore } from '@/stores/slog'

const props = defineProps({
  selectedDate: {
    type: String,
    required: true
  },
  selectedDepot: {
    type: String,
    required: true
  },
  tracks: {
    type: Array,
    default: () => []
  },
  arrivals: {
    type: Array,
    default: () => []
  },
  departures: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['track-assigned', 'movement-created'])

// Store references
const loggingStore = useLoggingStore()
const slogStore = useSlogStore()

// Canvas references
const container = ref(null)
const canvas = ref(null)
const ctx = ref(null)

// Component state
const isInitializing = ref(true)

// Timeline state
const timelineStartMinutes = ref(0)
const timelineEndMinutes = ref(0)
const minutesPerPixel = ref(0.5) // 30 seconds per pixel default
const pixelsPerTrack = ref(100)
const trackPositionWidth = ref(20)
const timelineOffsetY = ref(0)
const trackOffsetX = ref(0)
const height = ref(0)
const width = ref(0)
const movements = ref([])
const arrivalOffsets = ref(new Map())
const departureOffsets = ref(new Map())
// Interaction state
const isDragging = ref(false)
const dragStartX = ref(0)
const dragStartY = ref(0)
const lastMouseX = ref(0)
const lastMouseY = ref(0)
const selectedItem = ref(null)
const draggedItem = ref(null)
const hoverTrack = ref(null)
const hoverPosition = ref(null)
const isAltKeyPressed = ref(false)
const isShiftKeyPressed = ref(false)
const isCtrlKeyPressed = ref(false)

// Computed properties
const hasTracksData = computed(() => props.tracks && props.tracks.length > 0)
let animationFrame = null

const timelineHeightPixels = computed(() => {
  // 2 days in minutes (48 hours * 60 minutes)
  const totalMinutes = 2880
  return totalMinutes / minutesPerPixel.value
})

const trackAssignments = computed(() => {
  // Create a map of all track assignments
  const assignments = []
  movements.value = []
  // Process arrivals with assigned tracks
  props.arrivals.forEach(arrival => {
    if (arrival.movement) {
      arrival.notes = `${arrival.startingTrack} >>> ${arrival.targetTrack}`
      movements.value.push(arrival)
    }
    if (arrival.targetTrack) {
      const track = arrival.targetTrack.split('.')
      if (track.length < 2) return
      const trackNum = parseInt(track[0]) || 0
      const positionNum = parseInt(track[1]) || 1
      
      // Find matching departure
      const departure = props.departures.find(dep => 
        dep.vehicle === arrival.vehicle && 
        dep.departureDecimal > arrival.arrivalDecimal
      )
      if (departure) departure.startingTrack = `${trackNum}.${positionNum}`

      assignments.push({
        type: 'parking',
        vehicle: arrival.vehicle,
        arrivalInfo: arrival,
        departureInfo: departure,
        trackNum,
        positionNum,
        startTime: arrival.arrivalDecimal,
        endTime: departure ? departure.departureDecimal : timelineEndMinutes.value,
        isHighlighted: false,
        isConflict: false,
        conflicts: []
      })
    }
  })
  
  // Process departures with assigned tracks (no matching arrival)
  props.departures.forEach(departure => {
    if (departure.startingTrack) {
      // Check if already added through arrivals
      const existingAssignment = assignments.find(a => 
        a.departureInfo && a.departureInfo.departureDecimal === departure.departureDecimal &&
        a.vehicle === departure.vehicle
      )
      
      if (!existingAssignment) {
        const track = departure.startingTrack.split('.')
        if (track.length < 2) return
        const trackNum = parseInt(track[0]) || 0
        const positionNum = parseInt(track[1]) || 1
        
        // Find matching arrival
        let arrival = undefined
        for (let i = props.arrivals.length - 1; i >= 0; i--) {
          const arr = props.arrivals[i]
          if (arr.vehicle === departure.vehicle && arr.arrivalDecimal < departure.departureDecimal) {
            arr.targetTrack = `${trackNum}.${positionNum}`
            arrival = arr
            break
          }
        }

        assignments.push({
          type: 'parking',
          vehicle: departure.vehicle,
          arrivalInfo: arrival || null,
          departureInfo: departure,
          trackNum,
          positionNum,
          startTime: arrival ? arrival.arrivalDecimal : timelineStartMinutes.value,
          endTime: departure.departureDecimal,
          isHighlighted: false,
          isConflict: false,
          conflicts: []
        })
      }
    }
  })
  
  // Check for conflicts
  detectConflicts(assignments.sort((a, b) => a.startTime - b.startTime))
  return assignments
})

const formatName = (name) => {
  if (!name) return '...'
  const parts = name.split(',')
  return parts.length > 1 ? `${parts[1].trim()} ${parts[0].trim()}` : name.trim()
}

// Initialize timeline
const initTimeline = () => {
  if (!props.selectedDate) return
  
  isInitializing.value = true
  
  // Calculate start and end times for the timeline (2 days)
  const selectedDate = new Date(props.selectedDate)
  const startDate = new Date(selectedDate)
  startDate.setUTCHours(0, 0, 0, 0)
  
  const endDate = new Date(selectedDate)
  endDate.setUTCHours(0, 0, 0, 0)
  endDate.setUTCDate(endDate.getDate() + 2) // 2 days
  
  // Convert to minutes since 2025-01-01 00:00
  const baseDate = new Date('2025-01-01T00:00:00Z')
  timelineStartMinutes.value = Math.floor((startDate - baseDate) / (1000 * 60))
  timelineEndMinutes.value = Math.floor((endDate - baseDate) / (1000 * 60))
  
  // Reset view position
  timelineOffsetY.value = 0
  trackOffsetX.value = 0
  
  isInitializing.value = false
}

// Canvas setup and drawing
const setupCanvas = () => {
  if (!canvas.value || !container.value) return
  
  const dpr = window.devicePixelRatio || 1
  const rect = container.value.getBoundingClientRect()
  
  width.value = rect.width
  height.value = rect.height
  
  canvas.value.width = width.value * dpr
  canvas.value.height = height.value * dpr
  canvas.value.style.width = `${width.value}px`
  canvas.value.style.height = `${height.value}px`
  
  ctx.value = canvas.value.getContext('2d')
  ctx.value.scale(dpr, dpr)
}

const drawTimeline = () => {
  if (!ctx.value || !container.value || isInitializing.value) return
  
  // Clear canvas
  ctx.value.clearRect(0, 0, width.value, height.value)
  
  // Draw background
  ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--b1') || '#f3f4f6'
  ctx.value.fillRect(0, 0, width.value, height.value)
  
  // Draw time axis
  drawTimeAxis()
  
  // Draw tracks
  drawTracks()
  
  // Draw track assignments (occupied time slots)
  drawAssignments()
  drawMovements()
  
  // Draw arrival and departure markers
  drawMarkers()
  
  // Draw hover indicators
  if (hoverTrack.value !== null && hoverPosition.value !== null) {
    drawHoverIndicator()
  }
  
  // Draw track rules indicators
  drawTrackRules()
  
  // Draw legend
  // drawLegend()
  
  // Draw tooltip if hovering over an item
  drawTooltip()
}

// Drawing methods
const drawTimeAxis = () => {
  const hourHeight = 60 / minutesPerPixel.value
  const startHour = Math.floor(timelineStartMinutes.value / 60)
  const endHour = Math.ceil(timelineEndMinutes.value / 60)
  
  ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-content') || '#374151'
  ctx.value.font = '12px sans-serif'
  ctx.value.textAlign = 'right'
  
  for (let hour = startHour; hour <= endHour; hour++) {
    const hourMinutes = hour * 60
    const y = minutesToY(hourMinutes)
    
    if (y < 0 || y > height.value) continue
    
    // Draw hour line
    ctx.value.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-300') || '#d1d5db'
    ctx.value.beginPath()
    ctx.value.moveTo(0, y)
    ctx.value.lineTo(width.value, y)
    ctx.value.stroke()
    
    // Draw hour label
    const date = new Date('2025-01-01T00:00:00Z')
    date.setUTCMinutes(hourMinutes)
    const hourLabel = date.toISOString().substring(11, 16)

    // Draw date label if this is midnight
    if (hourLabel === '00:00') {
      const dateObj = new Date('2025-01-01T00:00:00Z')
      dateObj.setUTCMinutes(hourMinutes)
      const dateLabel = dateObj.toISOString().split('T')[0]
      ctx.value.font = 'bold 12px sans-serif'
      ctx.value.fillText(dateLabel, 80, y - 20)
      ctx.value.font = '12px sans-serif'
    }
    
    ctx.value.fillText(hourLabel, 80, y - 5)
  }
}

const drawTracks = () => {
  if (!props.tracks || props.tracks.length === 0) return
  
  const trackWidth = pixelsPerTrack.value
  const labelHeight = 25
  
  // Draw background for header and footer
  ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-200') || '#e5e7eb'
  ctx.value.fillRect(150, 0, width.value - 150, labelHeight)
  ctx.value.fillRect(150, height.value - labelHeight, width.value - 150, labelHeight)
  
  ctx.value.font = 'bold 14px sans-serif'
  ctx.value.textAlign = 'center'
  ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--primary') || '#3b82f6'
  
  // Draw track labels
  props.tracks.forEach((track, index) => {
    const x = 150 + (index * trackWidth) - trackOffsetX.value
    
    // Skip if outside visible area
    if (x + trackWidth < 150 || x > width.value) return
    
    // Draw track number at top
    ctx.value.fillText(track.track_number, x + trackWidth/2, labelHeight/2 + 5)
    
    // Draw track number at bottom
    ctx.value.fillText(track.track_number, x + trackWidth/2, height.value - labelHeight/2 + 10)
    
    // Draw track positions
    for (let pos = 1; pos <= track.positions; pos++) {
      const posX = x + (pos - 1) * trackPositionWidth.value
      
      // Draw position separators (lighter than track separators)
      ctx.value.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-200') || '#e5e7eb'
      ctx.value.beginPath()
      ctx.value.moveTo(posX, labelHeight)
      ctx.value.lineTo(posX, height.value - labelHeight)
      ctx.value.stroke()
      
      // Draw position number at bottom
      ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-content') || '#374151'
      ctx.value.font = '11px sans-serif'
      ctx.value.fillText(`${pos}`, posX + trackPositionWidth.value/2, height.value - 15)
      ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--primary') || '#3b82f6'
      ctx.value.font = 'bold 14px sans-serif'
    }
    
    // Draw track separators
    ctx.value.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-300') || '#d1d5db'
    ctx.value.lineWidth = 2
    ctx.value.beginPath()
    ctx.value.moveTo(x + trackWidth, labelHeight)
    ctx.value.lineTo(x + trackWidth, height.value - labelHeight)
    ctx.value.stroke()
    ctx.value.lineWidth = 1
  })
}

const drawTrackRules = () => {
  if (!props.tracks || props.tracks.length === 0) return
  
  const trackWidth = pixelsPerTrack.value
  
  props.tracks.forEach((track, index) => {
    const x = 150 + (index * trackWidth) - trackOffsetX.value
    
    // Skip if outside visible area
    if (x + trackWidth < 150 || x > width.value) return
    
    // Get track rule
    const isFiloRule = track.rule === 'filo'
    const color = isFiloRule 
      ? getComputedStyle(document.documentElement).getPropertyValue('--warning') || '#f59e0b'
      : getComputedStyle(document.documentElement).getPropertyValue('--success') || '#10b981'
    
    // Draw rule indicator
    ctx.value.fillStyle = color
    ctx.value.beginPath()
    ctx.value.arc(x + trackWidth - 10, 13, 5, 0, 2 * Math.PI)
    ctx.value.fill()
  })
}

const drawLegend = () => {
  const legendX = width.value - 210
  const legendY = 30
  const legendWidth = 200
  const legendHeight = 110
  
  // Draw legend background
  ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-100') || '#ffffff'
  ctx.value.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-300') || '#d1d5db'
  roundRect(ctx.value, legendX, legendY, legendWidth, legendHeight, 5, true, true)
  
  // Draw legend title
  ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-content') || '#374151'
  ctx.value.textAlign = 'left'
  ctx.value.font = 'bold 14px sans-serif'
  ctx.value.fillText('Grafikas', legendX + 10, legendY + 20)
  
  // Draw track rule indicators
  ctx.value.font = '12px sans-serif'
  
  // FIFO indicator
  ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--success') || '#10b981'
  ctx.value.beginPath()
  ctx.value.arc(legendX + 15, legendY + 40, 5, 0, 2 * Math.PI)
  ctx.value.fill()
  ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-content') || '#374151'
  ctx.value.fillText('FIFO - First In First Out', legendX + 25, legendY + 43)
  
  // FILO indicator
  ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--warning') || '#f59e0b'
  ctx.value.beginPath()
  ctx.value.arc(legendX + 15, legendY + 60, 5, 0, 2 * Math.PI)
  ctx.value.fill()
  ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-content') || '#374151'
  ctx.value.fillText('FILO - First In Last Out', legendX + 25, legendY + 63)
  
  // Conflict indicator
  ctx.value.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue('--error') || '#ef4444'
  ctx.value.lineWidth = 3
  ctx.value.beginPath()
  ctx.value.moveTo(legendX + 10, legendY + 80)
  ctx.value.lineTo(legendX + 20, legendY + 80)
  ctx.value.stroke()
  ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-content') || '#374151'
  ctx.value.fillText('Konfliktas', legendX + 25, legendY + 83)
  
  // Movement indicator
  ctx.value.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue('--secondary') || '#8b5cf6'
  ctx.value.beginPath()
  ctx.value.moveTo(legendX + 10, legendY + 100)
  ctx.value.lineTo(legendX + 20, legendY + 100)
  ctx.value.stroke()
  ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-content') || '#374151'
  ctx.value.fillText('Manevras', legendX + 25, legendY + 103)
  
  ctx.value.lineWidth = 1
}

const getVehicleColor = (vehicle) => {
  if (!vehicle) return getComputedStyle(document.documentElement).getPropertyValue('--primary') || '#3b82f6'
  
  // Generate color based on vehicle name hash
  const hash = vehicle.split('').reduce((acc, char) => char.charCodeAt(0) + acc, 0)
  
  // Use predefined theme colors from DaisyUI
  const colors = [
    getComputedStyle(document.documentElement).getPropertyValue('--primary') || '#3b82f6',
    getComputedStyle(document.documentElement).getPropertyValue('--secondary') || '#8b5cf6',
    getComputedStyle(document.documentElement).getPropertyValue('--accent') || '#d946ef',
    getComputedStyle(document.documentElement).getPropertyValue('--success') || '#10b981',
    getComputedStyle(document.documentElement).getPropertyValue('--info') || '#06b6d4',
    getComputedStyle(document.documentElement).getPropertyValue('--warning') || '#f59e0b',
  ]
  
  return colors[hash % colors.length]
}

const drawAssignments = () => {
  if (!trackAssignments.value.length) return
  
  trackAssignments.value.forEach(assignment => {
    const trackIndex = props.tracks.findIndex(t => parseInt(t.track_number) === assignment.trackNum)
    if (trackIndex === -1) return
    
    const startY = minutesToY(assignment.startTime)
    const endY = minutesToY(assignment.endTime)
    const trackX = 150 + (trackIndex * pixelsPerTrack.value) - trackOffsetX.value
    const posX = trackX + (assignment.positionNum - 1) * trackPositionWidth.value
    
    // Skip if completely outside visible area
    if (endY < 0 || startY > height.value) return
    
    // Draw parking line
    const lineWidth = 5
    ctx.value.lineWidth = lineWidth
    
    if (assignment.isConflict) {
      ctx.value.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue('--error') || '#ef4444'
    } else if (assignment.isHighlighted) {
      ctx.value.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue('--accent') || '#8b5cf6'
    // } else if (assignment.type === 'movement') {
    //   ctx.value.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue('--secondary') || '#8b5cf6'
    } else {
      ctx.value.strokeStyle = getVehicleColor(assignment.vehicle)
    }
    
    // Draw track occupancy line
    const lineX = posX + trackPositionWidth.value/2
    ctx.value.lineCap = 'round'
    ctx.value.beginPath()
    ctx.value.moveTo(lineX, Math.max(25, startY))
    ctx.value.lineTo(lineX, Math.min(height.value - 25, endY))
    ctx.value.stroke()
    ctx.value.lineCap = 'butt'
    
    // Draw vehicle label at the center of the line
    if (Math.abs(endY - startY) > 20) { // Only if line is tall enough
      const vehicleLabel = assignment.vehicle || 'Nežinomas'
      ctx.value.font = 'bold 12px sans-serif'
      ctx.value.fillStyle = ctx.value.strokeStyle
      ctx.value.textAlign = 'center'
      
      // Center text on the line
      const centerY = startY + (endY - startY) / 2
      ctx.value.fillText(vehicleLabel, lineX, centerY - 7)
    }
    
    // Draw connection lines to markers if available
    // if (assignment.arrivalInfo && assignment.arrivalInfo.arrivalDecimal >= timelineStartMinutes.value) {
    //   drawConnectorLine(lineX, startY, 100, startY, 'arrival')
    // 
    
    // if (assignment.departureInfo && assignment.departureInfo.departureDecimal <= timelineEndMinutes.value) {
    //   drawConnectorLine(lineX, endY, width.value - 100, endY, 'departure')
    // }
    
    // Draw conflict indicators if any
    if (assignment.conflicts && assignment.conflicts.length > 0) {
      assignment.conflicts.forEach(conflict => {
        const conflictY = minutesToY(conflict.time)
        
        // Skip if outside visible area
        if (conflictY < 0 || conflictY > height.value) return
        
        // Draw conflict indicator
        ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--error') || '#ef4444'
        ctx.value.beginPath()
        ctx.value.arc(lineX, conflictY, 6, 0, 2 * Math.PI)
        ctx.value.fill()
      })
    }
  })
  
  // Reset line width
  ctx.value.lineWidth = 1
}

const drawConnectorLine = (fromX, fromY, toX, toY, type) => {
  // Skip if outside visible area
  if (fromY < 0 || fromY > height.value) return
  ctx.value.lineWidth = 2
  ctx.value.setLineDash([2, 2])
  ctx.value.beginPath()
  ctx.value.moveTo(fromX, fromY)
  ctx.value.lineTo(toX, toY)
  ctx.value.stroke()
  ctx.value.setLineDash([])
}

const drawMovements = () => {
  if (!movements.value || movements.value.length === 0) return
  
  movements.value.forEach(movement => {
    const [startingTrack, startingPos] = movement.startingTrack.split('.')
    const [targetTrack, targetPos] = movement.targetTrack.split('.')
    const trackIndexA = props.tracks.findIndex(t => parseInt(t.track_number) === parseInt(startingTrack))
    const trackIndexB = props.tracks.findIndex(t => parseInt(t.track_number) === parseInt(targetTrack))
    if (trackIndexA === -1) return
    if (trackIndexB === -1) return
    
    const startY = minutesToY(movement.departureDecimal)
    const endY = minutesToY(movement.arrivalDecimal)
    const trackXA = 150 + (trackIndexA * pixelsPerTrack.value) - trackOffsetX.value
    const posXA = trackXA + (startingPos - 1) * trackPositionWidth.value
    const trackXB = 150 + (trackIndexB * pixelsPerTrack.value) - trackOffsetX.value
    const posXB = trackXB + (targetPos - 1) * trackPositionWidth.value
    
    // Skip if completely outside visible area
    if (endY < 0 || startY > height.value) return
    
    // Draw movement line
    ctx.value.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue('--secondary') || '#8b5cf6'
    ctx.value.lineWidth = 2
    ctx.value.setLineDash([2, 3])
    ctx.value.lineCap = 'round'
    
    const lineXA = posXA + trackPositionWidth.value/2
    const lineXB = posXB + trackPositionWidth.value/2
    ctx.value.beginPath()
    ctx.value.moveTo(lineXA, Math.max(25, startY))
    ctx.value.lineTo(lineXB, Math.min(height.value - 25, endY))
    ctx.value.stroke()
    

    ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-content') || '#374151'
    ctx.value.textAlign = 'left'
    ctx.value.font = '12px sans-serif'
    let min = Math.min(lineXA, lineXB)
    let max = Math.max(lineXA, lineXB)
    ctx.value.fillText(movement.notes || '.', lineXA + 5, startY - 2)
    ctx.value.setLineDash([])
  })
}

const drawMarkers = () => {
  // Сбрасываем смещения
  arrivalOffsets.value.clear()
  departureOffsets.value.clear()
  
  // Сначала группируем прибытия по времени
  const arrivalGroups = new Map()
  props.arrivals.forEach(arrival => {
    if (arrival.arrivalDecimal < timelineStartMinutes.value || 
        arrival.arrivalDecimal > timelineEndMinutes.value) return
    
    const timeKey = arrival.arrivalDecimal.toString()
    if (!arrivalGroups.has(timeKey)) {
      arrivalGroups.set(timeKey, [])
    }
    arrivalGroups.get(timeKey).push(arrival)
  })
  
  // Вычисляем смещения для прибытий
  arrivalGroups.forEach((arrivals, timeKey) => {
    if (arrivals.length <= 1) return
    
    // Распределяем маркеры равномерно по вертикали
    const offsetStep = 35 // Высота маркера + отступ
    for (let i = 0; i < arrivals.length; i++) {
      const arrival = arrivals[i]
      // Смещаем маркеры, центрируя группу (чтобы они были симметрично вокруг реального времени)
      const offset = (i - (arrivals.length - 1) / 2) * offsetStep
      arrivalOffsets.value.set(arrival.id, offset)
    }
  })
  
  // Аналогично для отправлений
  const departureGroups = new Map()
  props.departures.forEach(departure => {
    if (departure.departureDecimal < timelineStartMinutes.value || 
        departure.departureDecimal > timelineEndMinutes.value) return
    
    const timeKey = departure.departureDecimal.toString()
    if (!departureGroups.has(timeKey)) {
      departureGroups.set(timeKey, [])
    }
    departureGroups.get(timeKey).push(departure)
  })
  
  departureGroups.forEach((departures, timeKey) => {
    if (departures.length <= 1) return
    
    const offsetStep = 35
    for (let i = 0; i < departures.length; i++) {
      const departure = departures[i]
      const offset = (i - (departures.length - 1) / 2) * offsetStep
      departureOffsets.value.set(departure.id, offset)
    }
  })
  
  // Отрисовываем маркеры
  props.arrivals.forEach(arrival => {
    if (arrival.targetTrack || !arrival.vehicle) return
    if (arrival.arrivalDecimal < timelineStartMinutes.value || 
        arrival.arrivalDecimal > timelineEndMinutes.value) return
    
    const y = minutesToY(arrival.arrivalDecimal)
    // Применяем смещение, если есть
    const offset = arrivalOffsets.value.get(arrival.id) || 0
    
    // Skip if outside visible area
    if (y + offset < 0 || y + offset > height.value) return
    
    drawArrivalMarker(arrival, y + offset)
  })
  
  // Draw departure markers
  props.departures.forEach(departure => {
    if (departure.startingTrack || !departure.vehicle)return
    if (departure.departureDecimal < timelineStartMinutes.value || 
        departure.departureDecimal > timelineEndMinutes.value) return
    
    const y = minutesToY(departure.departureDecimal)
    // Применяем смещение, если есть
    const offset = departureOffsets.value.get(departure.id) || 0
    
    // Skip if outside visible area
    if (y + offset < 0 || y + offset > height.value) return
    
    drawDepartureMarker(departure, y + offset)
  })
}


const drawArrivalMarker = (arrival, y) => {
  const markerWidth = 130
  const markerHeight = 50
  const x = 10
  
  // Check if this marker is being dragged
  const isDraggingThisMarker = isDragging.value && 
    draggedItem.value && 
    draggedItem.value.type === 'arrival' && 
    draggedItem.value.id === arrival.id
  
  const color = isDraggingThisMarker 
    ? getComputedStyle(document.documentElement).getPropertyValue('--primary-focus') || 'rgba(59, 130, 246, 0.3)'
    : getComputedStyle(document.documentElement).getPropertyValue('--base-200') || '#e5e7eb'
  
  ctx.value.fillStyle = color
  ctx.value.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-300') || '#d1d5db'
  
  // Draw marker background
  roundRect(ctx.value, x, y - markerHeight/2, markerWidth, markerHeight, 5, true, true)
  
  // Draw marker content
  ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-content') || '#374151'
  ctx.value.textAlign = 'left'
  ctx.value.font = 'bold 14px sans-serif'

  ctx.value.fillText(arrival.arrivalTrainNumber || 'N/A', x + 6, y - 10)
  ctx.value.textAlign = 'right'
  ctx.value.fillText(arrival.arrivalPlanned || '---', x + 124, y - 10)
  
  ctx.value.textAlign = 'left'
  ctx.value.font = '12px sans-serif'
  ctx.value.fillText(arrival.vehicle || '---', x + 6, y + 5)
  ctx.value.fillText(formatName(arrival.arrivalEmployee1), x + 6, y + 20)

  ctx.value.textAlign = 'right'
  ctx.value.font = 'bold 14px sans-serif'
  ctx.value.fillStyle =getComputedStyle(document.documentElement).getPropertyValue('--primary-focus') || 'rgba(59, 130, 246, 0.9)'
  ctx.value.fillText(arrival.targetTrack || '---', x + 124, y + 5)
}

const drawDepartureMarker = (departure, y) => {
  const markerWidth = 130
  const markerHeight = 50
  const x = width.value - markerWidth - 10
  
  // Check if this marker is being dragged
  const isDraggingThisMarker = isDragging.value && 
    draggedItem.value && 
    draggedItem.value.type === 'departure' && 
    draggedItem.value.id === departure.id
  
  const color = isDraggingThisMarker 
    ? getComputedStyle(document.documentElement).getPropertyValue('--primary-focus') || 'rgba(59, 130, 246, 0.3)'
    : getComputedStyle(document.documentElement).getPropertyValue('--base-200') || '#e5e7eb'
  
  ctx.value.fillStyle = color
  ctx.value.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-300') || '#d1d5db'
  
  // Draw marker background
  roundRect(ctx.value, x, y - markerHeight/2, markerWidth, markerHeight, 5, true, true)
  
  // Draw marker content
  ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-content') || '#374151'
  ctx.value.textAlign = 'left'
  ctx.value.font = 'bold 14px sans-serif'
  ctx.value.fillText(departure.departureTrainNumber || 'N/A', x + 6, y - 10)
  ctx.value.textAlign = 'right'
  ctx.value.fillText(departure.departurePlanned || '---', x + 124, y - 10)
  
  ctx.value.textAlign = 'left'
  ctx.value.font = '12px sans-serif'
  ctx.value.fillText(departure.vehicle || '---', x + 6, y + 5)
  ctx.value.fillText(formatName(departure.departureEmployee1), x + 6, y + 20)
  
  ctx.value.textAlign = 'right'
  ctx.value.font = 'bold 14px sans-serif'
  ctx.value.fillStyle =getComputedStyle(document.documentElement).getPropertyValue('--primary-focus') || 'rgba(59, 130, 246, 0.9)'
  ctx.value.fillText(departure.startingTrack || '---', x + 124, y + 5)
}

const drawHoverIndicator = () => {
  const trackIndex = props.tracks.findIndex(t => parseInt(t.track_number) === hoverTrack.value)
  if (trackIndex === -1) return
  
  const trackX = 150 + (trackIndex * pixelsPerTrack.value) - trackOffsetX.value
  const posX = trackX + (hoverPosition.value - 1) * trackPositionWidth.value
  
  // Draw hover rectangle
  ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--primary-focus') || 'rgba(59, 130, 246, 0.3)'
  ctx.value.fillRect(posX, 25, trackPositionWidth.value, height.value - 50)
  
  // Draw ghost vertical line if dragging
  if (isDragging.value && draggedItem.value) {
    const ghostColor = isAltKeyPressed.value
      ? getComputedStyle(document.documentElement).getPropertyValue('--secondary') || '#8b5cf6'
      : getComputedStyle(document.documentElement).getPropertyValue('--primary') || '#3b82f6'
    
    ctx.value.strokeStyle = ghostColor
    ctx.value.lineWidth = 8
    ctx.value.globalAlpha = 0.5
    ctx.value.lineCap = 'round'
    
    let startY, endY
    
    if (draggedItem.value.type === 'arrival') {
      startY = minutesToY(draggedItem.value.arrivalDecimal)
      
      // Find matching departure
      const departure = props.departures.find(dep => 
        dep.vehicle === draggedItem.value.vehicle && 
        dep.departureDecimal > draggedItem.value.arrivalDecimal
      )
      
      endY = departure ? minutesToY(departure.departureDecimal) : height.value - 25
    } else if (draggedItem.value.type === 'departure') {
      endY = minutesToY(draggedItem.value.departureDecimal)
      
      // Find matching arrival
      const arrival = props.arrivals.find(arr => 
        arr.vehicle === draggedItem.value.vehicle && 
        arr.arrivalDecimal < draggedItem.value.departureDecimal
      )
      
      startY = arrival ? minutesToY(arrival.arrivalDecimal) : 25
    } else if (draggedItem.value.type === 'parking' && isAltKeyPressed.value) {
      // Alt+dragging for movement creation
      const cursorY = lastMouseY.value - container.value.getBoundingClientRect().top
      const minutes = yToMinutes(cursorY)
      
      // For movements, show a line from current position to new position
      if (minutes > draggedItem.value.startTime && minutes < draggedItem.value.endTime) {
        // Find original track and position
        const origTrackIndex = props.tracks.findIndex(t => parseInt(t.track_number) === draggedItem.value.trackNum)
        if (origTrackIndex !== -1) {
          const origTrackX = 150 + (origTrackIndex * pixelsPerTrack.value) - trackOffsetX.value
          const origPosX = origTrackX + (draggedItem.value.positionNum - 1) * trackPositionWidth.value
          const origLineX = origPosX + trackPositionWidth.value/2
          
          // Draw original line
          ctx.value.beginPath()
          ctx.value.moveTo(origLineX, Math.max(25, minutesToY(draggedItem.value.startTime)))
          ctx.value.lineTo(origLineX, minutesToY(minutes))
          ctx.value.stroke()
          
          // Draw dotted movement line
          const lineX = posX + trackPositionWidth.value/2
          ctx.value.setLineDash([5, 5])
          ctx.value.beginPath()
          ctx.value.moveTo(origLineX, minutesToY(minutes))
          ctx.value.lineTo(lineX, minutesToY(minutes + 10)) // 10 min movement
          ctx.value.stroke()
          ctx.value.setLineDash([])
          
          // Draw new line
          ctx.value.beginPath()
          ctx.value.moveTo(lineX, minutesToY(minutes + 10))
          ctx.value.lineTo(lineX, Math.min(height.value - 25, minutesToY(draggedItem.value.endTime)))
          ctx.value.stroke()
          
          // Draw movement label
          ctx.value.font = 'bold 12px sans-serif'
          ctx.value.fillStyle = ghostColor
          ctx.value.textAlign = 'center'
          ctx.value.fillText('Manevras', (origLineX + lineX) / 2, minutesToY(minutes) - 10)
          
          // No need to draw the normal ghost line
          ctx.value.globalAlpha = 1.0
          ctx.value.lineCap = 'butt'
          ctx.value.lineWidth = 1
          return
        }
      }
    }
    
    // Draw normal ghost line
    const lineX = posX + trackPositionWidth.value/2
    ctx.value.beginPath()
    ctx.value.moveTo(lineX, Math.max(25, startY))
    ctx.value.lineTo(lineX, Math.min(height.value - 25, endY))
    ctx.value.stroke()
    
    ctx.value.globalAlpha = 1.0
    ctx.value.lineCap = 'butt'
    ctx.value.lineWidth = 1
  }
}

const drawTooltip = () => {
  if (!container.value || !ctx.value) return
  
  const rect = container.value.getBoundingClientRect()
  const mouseX = lastMouseX.value - rect.left
  const mouseY = lastMouseY.value - rect.top
  
  const item = getItemAtCoords(mouseX, mouseY)
  
  if (item && item.type === 'parking') {
    const tooltipWidth = 220
    const tooltipHeight = 130
    const tooltipX = Math.min(mouseX + 10, width.value - tooltipWidth - 6)
    const tooltipY = Math.min(mouseY + 10, height.value - tooltipHeight - 5)
    
    // Draw tooltip background
    ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-100') || '#ffffff'
    ctx.value.strokeStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-300') || '#d1d5db'
    roundRect(ctx.value, tooltipX, tooltipY, tooltipWidth, tooltipHeight, 5, true, true)
    
    // Draw tooltip content
    ctx.value.fillStyle = getComputedStyle(document.documentElement).getPropertyValue('--base-content') || '#374151'
    ctx.value.textAlign = 'left'
    ctx.value.font = 'bold 14px sans-serif'
    ctx.value.fillText(`${item.vehicle || 'Nežinomas'}`, tooltipX + 10, tooltipY + 15)
    
    ctx.value.font = '12px sans-serif'
    ctx.value.fillText(`Kelias: `, tooltipX + 10, tooltipY + 32)
    ctx.value.font = 'bold 14px sans-serif'
    ctx.value.fillText(`${item.trackNum}.${item.positionNum}`, tooltipX + 50, tooltipY + 32)
    
    ctx.value.font = '12px sans-serif'
    if (item.arrivalInfo) {
      ctx.value.fillText(`Atvykimas: ${item.arrivalInfo.arrivalDate} ${item.arrivalInfo.arrivalPlanned || 'N/A'}`, tooltipX + 10, tooltipY + 48)
      ctx.value.fillText(`Traukinys: ${item.arrivalInfo.arrivalTrainNumber || 'N/A'}`, tooltipX + 10, tooltipY + 63)
      ctx.value.fillText(`Mašinistas: ${formatName(item.arrivalInfo.arrivalEmployee1) || 'N/A'}`, tooltipX + 10, tooltipY + 78)
    }
    
    if (item.departureInfo) {
      ctx.value.fillText(`Išvykimas: ${item.departureInfo.departureDate} ${item.departureInfo.departurePlanned || 'N/A'}`, tooltipX + 10, tooltipY + 95)
      ctx.value.fillText(`Traukinys: ${item.departureInfo.departureTrainNumber || 'N/A'}`, tooltipX + 10, tooltipY + 110)
      ctx.value.fillText(`Mašinistas: ${formatName(item.departureInfo.departureEmployee1) || 'N/A'}`, tooltipX + 10, tooltipY + 125)
    }
  }
}

// Helper function for drawing rounded rectangles
const roundRect = (ctx, x, y, width, height, radius, fill, stroke) => {
  ctx.beginPath()
  ctx.moveTo(x + radius, y)
  ctx.lineTo(x + width - radius, y)
  ctx.quadraticCurveTo(x + width, y, x + width, y + radius)
  ctx.lineTo(x + width, y + height - radius)
  ctx.quadraticCurveTo(x + width, y + height, x + width - radius, y + height)
  ctx.lineTo(x + radius, y + height)
  ctx.quadraticCurveTo(x, y + height, x, y + height - radius)
  ctx.lineTo(x, y + radius)
  ctx.quadraticCurveTo(x, y, x + radius, y)
  ctx.closePath()
  if (fill) {
    ctx.fill()
  }
  if (stroke) {
    ctx.stroke()
  }
}

// Coordinate conversion helpers
const minutesToY = (minutes) => {
  const timelineRange = timelineEndMinutes.value - timelineStartMinutes.value
  const pixelRange = timelineHeightPixels.value
  
  const relativeMinutes = minutes - timelineStartMinutes.value
  const y = (relativeMinutes / timelineRange) * pixelRange + 25 - timelineOffsetY.value
  
  return y
}

const yToMinutes = (y) => {
  const timelineRange = timelineEndMinutes.value - timelineStartMinutes.value
  const pixelRange = timelineHeightPixels.value
  
  const relativeY = y - 25 + timelineOffsetY.value
  const minutes = (relativeY / pixelRange) * timelineRange + timelineStartMinutes.value
  
  return Math.round(minutes)
}

const getTrackPositionFromCoords = (x, y) => {
  if (x < 150 || y < 25 || y > height.value - 25) return null
  
  // Calculate track index
  const tracksSection = x - 150 + trackOffsetX.value
  const trackIndex = Math.floor(tracksSection / pixelsPerTrack.value)
  
  // Ensure track exists
  if (trackIndex < 0 || trackIndex >= props.tracks.length) return null
  
  const track = props.tracks[trackIndex]
  const trackNum = parseInt(track.track_number)
  
  // Calculate position
  const positionOffset = tracksSection - (trackIndex * pixelsPerTrack.value)
  const positionIndex = Math.floor(positionOffset / trackPositionWidth.value) + 1
  
  // Check position bounds
  if (positionIndex < 1 || positionIndex > track.positions) return null
  
  return { trackNum, positionNum: positionIndex }
}

// Get item at coordinates
const getItemAtCoords = (x, y) => {
  // Check markers
  if (x < 140 && y > 25 && y < height.value - 25) {
    // Check arrival markers
    for (let i = props.arrivals.length - 1; i >= 0; i--) {
      const arrival = props.arrivals[i]
      if (arrival.arrivalDecimal < timelineStartMinutes.value || 
          arrival.arrivalDecimal > timelineEndMinutes.value) continue
      
      const markerY = minutesToY(arrival.arrivalDecimal)
      // Учитываем смещение маркера
      const offset = arrivalOffsets.value.get(arrival.id) || 0
      if (Math.abs(y - (markerY + offset)) < 25) { // Half height of marker
        return {
          type: 'arrival',
          ...arrival
        }
      }
    }
  } else if (x > width.value - 140 && y > 25 && y < height.value - 25) {
    // Check departure markers
    for (let i = props.departures.length - 1; i >= 0; i--) {
      const departure = props.departures[i]
      if (departure.departureDecimal < timelineStartMinutes.value || 
          departure.departureDecimal > timelineEndMinutes.value) continue
      
      const markerY = minutesToY(departure.departureDecimal)
      // Учитываем смещение маркера
      const offset = departureOffsets.value.get(departure.id) || 0
      if (Math.abs(y - (markerY + offset)) < 25) { // Half height of marker
        return {
          type: 'departure',
          ...departure
        }
      }
    }
  } else if (x >= 150 && x <= width.value - 150 && y >= 25 && y <= height.value - 25) {
    // Check track assignments (оставляем без изменений)
    for (const assignment of trackAssignments.value) {
      const trackIndex = props.tracks.findIndex(t => parseInt(t.track_number) === assignment.trackNum)
      if (trackIndex === -1) continue
      
      const trackX = 150 + (trackIndex * pixelsPerTrack.value) - trackOffsetX.value
      const posX = trackX + (assignment.positionNum - 1) * trackPositionWidth.value
      
      if (x >= posX && x <= posX + trackPositionWidth.value) {
        const startY = minutesToY(assignment.startTime)
        const endY = minutesToY(assignment.endTime)
        
        if (y >= startY && y <= endY) {
          return assignment
        }
      }
    }
  }
  
  return null
}

const clearTrackAssignment = (assignment) => {
  // Отправляем событие для удаления назначения пути
  if (assignment.arrivalInfo) {
    emit('track-assigned', {
      type: 'arrival',
      id: assignment.arrivalInfo.id,
      trackAssignment: '' // '' означает удаление назначения
    })
  }
  
  if (assignment.departureInfo) {
    emit('track-assigned', {
      type: 'departure',
      id: assignment.departureInfo.id,
      trackAssignment: '' // '' означает удаление назначения
    })
  }
}

// Conflict detection
const detectConflicts = (assignments) => {
  // Reset all conflicts
  assignments.forEach(assignment => {
    assignment.isConflict = false
    assignment.conflicts = []
  })
  
  // Check each pair of assignments for conflicts
  for (let i = 0; i < assignments.length; i++) {
    const a = assignments[i]
    for (let j = i + 1; j < assignments.length; j++) {
      const b = assignments[j]
      
      // Skip if not on the same track
      if (a.trackNum !== b.trackNum) continue
      
      // Find track info
      const trackObj = props.tracks.find(t => parseInt(t.track_number) === a.trackNum)
      if (!trackObj) continue
      
      const isFifo = trackObj.rule !== 'filo'
      
      // Check time overlap
      const overlaps = (a.startTime <= b.startTime && a.endTime > b.startTime)

      if (overlaps) {
        // Same position direct conflict
        if (a.positionNum === b.positionNum) {
          a.isConflict = true
          b.isConflict = true
          
          a.conflicts.push({
            time: Math.max(a.startTime, b.startTime),
            type: 'position',
            message: `Konfliktas: bandymas užimti tą pačią poziciją`
          })
          
          b.conflicts.push({
            time: Math.max(a.startTime, b.startTime),
            type: 'position',
            message: `Konfliktas: bandymas užimti tą pačią poziciją`
          })
          
          continue
        }
        
        // FIFO rule: Lower position can't be occupied if higher position is still occupied
        if (isFifo && a.positionNum > b.positionNum) {
          b.isConflict = true
          
          b.conflicts.push({
            time: b.startTime,
            type: 'fifo',
            message: 'Konfliktas: negalima užimti, nes užimta aukštesnė pozicija'
          })
        }
        
        // FIFO rule: Higher position can't leave if lower position is still occupied
        if (isFifo && a.positionNum < b.positionNum) {
          if (a.endTime > b.endTime) {
            b.isConflict = true
            
            b.conflicts.push({
              time: a.endTime,
              type: 'fifo',
              message: 'Konfliktas: negalima išvykti, nes užimta žemesnė pozicija'
            })
          }
        }
        
        // FILO rule: Higher position can't be occupied if lower position is still occupied
        if (!isFifo && a.positionNum < b.positionNum) {
          b.isConflict = true
          
          b.conflicts.push({
            time: b.startTime,
            type: 'filo',
            message: 'Konfliktas: negalima užimti, nes užimta žemesnė pozicija'
          })
        }
        
        // FILO rule: Lower position blocking higher position's departure
        if (!isFifo && a.positionNum > b.positionNum) {
          if (a.endTime < b.endTime) {
            b.isConflict = true
            
            b.conflicts.push({
              time: a.endTime,
              type: 'filo',
              message: 'Konfliktas: blokuojamas žemesnės pozicijos išvykimas'
            })
          }
        }
      }
    }
  }
}

// Event handlers
const handleMouseDown = (e) => {
  if (!container.value) return
  
  // Get mouse coordinates
  const rect = container.value.getBoundingClientRect()
  const mouseX = e.clientX - rect.left
  const mouseY = e.clientY - rect.top
  
  // Store mouse coordinates for reference
  dragStartX.value = mouseX
  dragStartY.value = mouseY
  lastMouseX.value = e.clientX
  lastMouseY.value = e.clientY
  
  // Check modifier keys
  isAltKeyPressed.value = e.altKey
  isShiftKeyPressed.value = e.shiftKey
  isCtrlKeyPressed.value = e.ctrlKey
  
  // Find item at cursor
  const item = getItemAtCoords(mouseX, mouseY)
  
  if (item) {
    selectedItem.value = item
    draggedItem.value = item
    isDragging.value = true
    
    // Highlight the selected item
    if (item.type === 'parking') {
      trackAssignments.value.forEach(assignment => {
        if (assignment === item) {
          assignment.isHighlighted = true
        } else {
          assignment.isHighlighted = false
        }
      })
    }
    
    // Log drag start
    loggingStore.info('Pradėtas elementų velkimas', {
      component: 'TracksTimeLine',
      type: item.type,
      vehicle: item.vehicle
    })
  } else {
    // No item clicked, start panning
    isDragging.value = true
    draggedItem.value = null
    
    // Deselect any selected item
    selectedItem.value = null
    trackAssignments.value.forEach(assignment => {
      assignment.isHighlighted = false
    })
  }
}

const handleMouseMove = (e) => {
  if (!container.value) return
  
  // Update mouse coordinates
  lastMouseX.value = e.clientX
  lastMouseY.value = e.clientY
  
  // Update modifier keys state
  isAltKeyPressed.value = e.altKey
  isShiftKeyPressed.value = e.shiftKey
  isCtrlKeyPressed.value = e.ctrlKey
  
  // Get mouse coordinates relative to canvas
  const rect = container.value.getBoundingClientRect()
  const mouseX = e.clientX - rect.left
  const mouseY = e.clientY - rect.top
  
  // Calculate track position under mouse
  const trackPosition = getTrackPositionFromCoords(mouseX, mouseY)
  
  if (trackPosition) {
    hoverTrack.value = trackPosition.trackNum
    hoverPosition.value = trackPosition.positionNum
  } else {
    hoverTrack.value = null
    hoverPosition.value = null
  }
  
  // Handle dragging
  if (isDragging.value) {
    if (draggedItem.value) {
      // Dragging an item - nothing to do here yet, just update UI
      drawTimeline()
    } else {
      // Panning the view
      const deltaX = mouseX - dragStartX.value
      const deltaY = mouseY - dragStartY.value
      
      // Update view offsets
      trackOffsetX.value = Math.max(0, trackOffsetX.value - deltaX)
      timelineOffsetY.value = Math.max(0, Math.min(timelineHeightPixels.value - height.value + 50, timelineOffsetY.value - deltaY))
      
      // Reset drag start for continuous movement
      dragStartX.value = mouseX
      dragStartY.value = mouseY
      
      drawTimeline()
    }
  } else {
    // Just hovering - redraw to update hover effects
    drawTimeline()
  }
}

const handleMouseUp = (e) => {
  if (!isDragging.value) return
  
  // Handle end of drag
  if (draggedItem.value && hoverTrack.value !== null && hoverPosition.value !== null) {
    // Item was dropped on a track position
    if (draggedItem.value.type === 'arrival') {
      // Assign track to arrival
      const trackAssignment = `${hoverTrack.value}.${hoverPosition.value}`
      
      // Emit event to update the model
      emit('track-assigned', {
        type: 'arrival',
        id: draggedItem.value.id,
        trackAssignment
      })
      
      // Update UI
      drawTimeline()
      
      loggingStore.info('Priskirtas kelias atvykimui', {
        component: 'TracksTimeLine',
        vehicle: draggedItem.value.vehicle,
        track: trackAssignment
      })
      
      slogStore.addToast({
        message: `Kelias ${trackAssignment} priskirtas ${draggedItem.value.vehicle}`,
        type: 'alert-success'
      })
    } else if (draggedItem.value.type === 'departure') {
      // Assign track to departure
      const trackAssignment = `${hoverTrack.value}.${hoverPosition.value}`
      
      // Emit event to update the model
      emit('track-assigned', {
        type: 'departure',
        id: draggedItem.value.id,
        trackAssignment
      })
      
      // Update UI
      drawTimeline()
      
      loggingStore.info('Priskirtas kelias išvykimui', {
        component: 'TracksTimeLine',
        vehicle: draggedItem.value.vehicle,
        track: trackAssignment
      })
      
      slogStore.addToast({
        message: `Kelias ${trackAssignment} priskirtas ${draggedItem.value.vehicle}`,
        type: 'alert-success'
      })
    } else if (draggedItem.value.type === 'parking') {
      // Get current track position
      const currentTrack = draggedItem.value.trackNum
      const currentPosition = draggedItem.value.positionNum
      
      // Check if position changed
      if (currentTrack !== hoverTrack.value || currentPosition !== hoverPosition.value) {
        if (isAltKeyPressed.value) {
          // Alt+drag for movement creation
          const rect = container.value.getBoundingClientRect()
          const mouseY = e.clientY - rect.top
          const minutes = yToMinutes(mouseY)
          
          // Only create movement if cursor is between arrival and departure
          if (minutes > draggedItem.value.startTime && minutes < draggedItem.value.endTime) {
            // Determine movement duration
            const movementDuration = currentTrack === hoverTrack.value ? 10 : 30 // 10 min same track, 30 min different track
            
            // Create movement event
            const movementEvent = {
              fromTrack: `${currentTrack}.${currentPosition}`,
              toTrack: `${hoverTrack.value}.${hoverPosition.value}`,
              vehicle: draggedItem.value.vehicle,
              startTime: minutes,
              endTime: minutes + movementDuration
            }
            
            // Emit movement created event
            emit('movement-created', movementEvent)
            
            loggingStore.info('Manevras sukurtas', {
              component: 'TracksTimeLine',
              movement: movementEvent
            })
            
            slogStore.addToast({
              message: `Manevras sukurtas: ${movementEvent.vehicle} iš ${movementEvent.fromTrack} į ${movementEvent.toTrack}`,
              type: 'alert-success'
            })
          } else {
            slogStore.addToast({
              message: 'Manevras turi būti tarp atvykimo ir išvykimo laiko',
              type: 'alert-warning'
            })
          }
        } else {
          // Regular drag - reassign track
          // Emit event to update the model
          if (draggedItem.value.arrivalInfo) {
            emit('track-assigned', {
              type: 'arrival',
              id: draggedItem.value.arrivalInfo.id,
              trackAssignment: `${hoverTrack.value}.${hoverPosition.value}`
            })
          }
          
          if (draggedItem.value.departureInfo) {
            emit('track-assigned', {
              type: 'departure',
              id: draggedItem.value.departureInfo.id,
              trackAssignment: `${hoverTrack.value}.${hoverPosition.value}`
            })
          }
          
          loggingStore.info('Traukinys perkeltas į kitą kelią', {
            component: 'TracksTimeLine',
            vehicle: draggedItem.value.vehicle,
            fromTrack: `${currentTrack}.${currentPosition}`,
            toTrack: `${hoverTrack.value}.${hoverPosition.value}`
          })
          
          slogStore.addToast({
            message: `${draggedItem.value.vehicle} perkeltas į kelią ${hoverTrack.value}.${hoverPosition.value}`,
            type: 'alert-success'
          })
        }
      }
    }
  }
  
  // Reset drag state
  isDragging.value = false
  draggedItem.value = null
  
  // Redraw timeline
  drawTimeline()
}

const handleMouseLeave = () => {
  // Cancel any dragging operation
  isDragging.value = false
  draggedItem.value = null
  hoverTrack.value = null
  hoverPosition.value = null
  
  // Redraw timeline
  drawTimeline()
}

const handleWheel = (e) => {
  // Prevent default scrolling
  e.preventDefault()
  if (e.ctrlKey) {
    // Get mouse coordinates relative to canvas
    const rect = container.value.getBoundingClientRect()
    const mouseY = e.clientY - rect.top
    
    // Mouse position relative to the entire timeline (with offset)
    const relativeY = mouseY - 25 + timelineOffsetY.value
    
    // Position as a percentage of the total timeline height
    const percentOfTimeline = relativeY / timelineHeightPixels.value
    
    // Apply zoom
    const zoomFactor = e.deltaY > 0 ? 1.1 : 0.9
    minutesPerPixel.value = Math.max(0.1, Math.min(2, minutesPerPixel.value * zoomFactor))
    
    // Calculate new position that corresponds to the same percentage
    const newRelativeY = percentOfTimeline * timelineHeightPixels.value
    
    // Adjust offset to keep cursor at the same relative position
    timelineOffsetY.value = newRelativeY - (mouseY - 25)
    
    // Ensure offset stays within valid range
    timelineOffsetY.value = Math.max(0, Math.min(
      timelineHeightPixels.value - height.value + 50,
      timelineOffsetY.value
    ))
    
  } else {
    // Scroll
    const scrollSpeed = 30
    if (e.shiftKey) {
      // Horizontal scroll
      trackOffsetX.value = Math.max(0, trackOffsetX.value + (e.deltaY > 0 ? scrollSpeed : -scrollSpeed))
    } else {
      // Vertical scroll
      timelineOffsetY.value = Math.max(0, Math.min(
        timelineHeightPixels.value - height.value + 50,
        timelineOffsetY.value + (e.deltaY > 0 ? scrollSpeed : -scrollSpeed)
      ))
    }
  }
  
  // Redraw timeline
  drawTimeline()
}

// Watch for props changes
watch(() => props.selectedDate, () => {
  initTimeline()
  drawTimeline()
})

watch(() => props.tracks, () => {
  drawTimeline()
})

watch(() => props.arrivals, () => {
  drawTimeline()
})

watch(() => props.departures, () => {
  drawTimeline()
})

// Window resize handler
const handleResize = () => {
  setupCanvas()
  drawTimeline()
}

// Keyboard event handlers
const handleKeyDown = (e) => {
  if (e.key === 'Alt') {
    isAltKeyPressed.value = true
    drawTimeline()
  } else if (e.key === 'Shift') {
    isShiftKeyPressed.value = true
    drawTimeline()
  } else if (e.key === 'Control') {
    isCtrlKeyPressed.value = true
    drawTimeline()
  } else if (e.key === 'Delete' || e.key === 'Backspace' || e.key === 'x' || e.key === 'х') {
    // Получаем элемент под курсором
    if (!container.value) return
    const rect = container.value.getBoundingClientRect()
    const mouseX = lastMouseX.value - rect.left
    const mouseY = lastMouseY.value - rect.top
    
    const hoveredItem = getItemAtCoords(mouseX, mouseY)
    
    // Если это назначение пути (вертикальная линия)
    if (hoveredItem && hoveredItem.type === 'parking') {
      clearTrackAssignment(hoveredItem)
    }
  }
}

const handleKeyUp = (e) => {
  if (e.key === 'Alt') {
    isAltKeyPressed.value = false
    drawTimeline()
  } else if (e.key === 'Shift') {
    isShiftKeyPressed.value = false
    drawTimeline()
  } else if (e.key === 'Control') {
    isCtrlKeyPressed.value = false
    drawTimeline()
  }
}

// Lifecycle hooks
onMounted(() => {
  // Initialize
  setupCanvas()
  initTimeline()
  drawTimeline()
  
  // Add event listeners
  window.addEventListener('resize', handleResize)
  window.addEventListener('keydown', handleKeyDown)
  window.addEventListener('keyup', handleKeyUp)
  
  // Set animation frame for smooth rendering
  let animate = () => {
    drawTimeline()
    animationFrame = requestAnimationFrame(animate)
  }
  
  animationFrame = requestAnimationFrame(animate)
  
  // Log component initialization
  loggingStore.info('TracksTimeLine komponentas inicializuotas', {
    component: 'TracksTimeLine',
    selectedDate: props.selectedDate,
    selectedDepot: props.selectedDepot
  })
})
onUnmounted(() => {
  // Clean up
  window.removeEventListener('resize', handleResize)
  window.removeEventListener('keydown', handleKeyDown)
  window.removeEventListener('keyup', handleKeyUp)
  
  if (animationFrame) {
    cancelAnimationFrame(animationFrame)
  }
  
  loggingStore.info('TracksTimeLine komponentas išmontuotas', {
    component: 'TracksTimeLine'
  })
})
</script>

<style scoped>
.tracks-timeline-container {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
  cursor: grab;
}

.tracks-timeline-container.cursor-grabbing {
  cursor: grabbing;
}

canvas {
  display: block;
}

.timeline-empty {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  background-color: var(--base-200);
  color: var(--base-content);
  text-align: center;
  padding: 2rem;
}

.loading-indicator {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: rgba(0, 0, 0, 0.2);
}
</style>
