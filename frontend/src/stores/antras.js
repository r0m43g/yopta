// frontend/src/stores/antras.js
import { defineStore } from 'pinia'
import api from '../services/api'
import { useLoggingStore } from './logging'
import { useSlogStore } from './slog'
import { timeToDecimal } from '../utils/helpers'
import ExcelJS from 'exceljs'

/**
 * Antras Store - manages train movement data from Excel imports
 * 
 * This store handles the import and management of train schedule data
 * from external Excel files with complex parsing rules for:
 * - Vehicle identification from type/number combinations
 * - Staff information extraction
 * - Depot sheet merging (+D/-D → .D)
 * - Time parsing with day offsets (+1/-1)
 * - Trains collection by date with routes and staff
 */
export const useAntrasStore = defineStore('antras', {
  state: () => ({
    // Core data structures
    vehicles: [],           // { vehicle, vehicleNo:[], vehicleRegNo:[], vehicleWorkings:[] }
    vehicleWorkings: [],    // { vehicleWorking, startingTime, startingLocation, endingTime, endLocation }
    staff: [],              // { id, occ, name, phone, duties:[] }
    duties: [],             // { id, startingTime, endTime, trains:[] }
    stations: [],           // { code, networkPointName, arrivals:[], departures:[] }
    
    // Trains by date - NEW structure
    // { '2025-12-16': [ { no, startingLocation, endLocation, staff:[], stops:[] } ] }
    trains: {},
    
    // Raw records for debugging/display
    records: [],
    sheets: [],
    
    // UI state
    isLoading: false,
    selectedSheet: null,
    selectedDate: null,
    lastImported: null,
    fileName: null,
    
    // Field mappings from API
    fieldMappings: null,
    fieldMappingsLoaded: false,
  }),

  getters: {
    /**
     * Check if field mappings are loaded and valid
     */
    hasFieldMappings(state) {
      return state.fieldMappings !== null && Object.keys(state.fieldMappings).length > 0
    },

    /**
     * Get available station codes
     */
    availableSheets(state) {
      return state.stations.map(s => s.code).sort()
    },

    /**
     * Get available dates from all stations
     */
    availableDates(state) {
      const dates = new Set()
      state.stations.forEach(station => {
        station.arrivals.forEach(arr => {
          if (arr.arrivalDate) dates.add(arr.arrivalDate)
        })
        station.departures.forEach(dep => {
          if (dep.departureDate) dates.add(dep.departureDate)
        })
      })
      return Array.from(dates).sort()
    },

    /**
     * Get trains for a specific date
     */
    getTrainsByDate: (state) => (date) => {
      return state.trains[date] || []
    },

    /**
     * Get all unique train numbers
     */
    allTrainNumbers(state) {
      const trainNos = new Set()
      Object.values(state.trains).forEach(dayTrains => {
        dayTrains.forEach(train => trainNos.add(train.no))
      })
      return Array.from(trainNos).sort((a, b) => a - b)
    },

    /**
     * Get filtered records by selected sheet and date
     */
    filteredRecords(state) {
      let result = []
      
      state.stations.forEach(station => {
        if (state.selectedSheet && station.code !== state.selectedSheet) return
        
        const arrivals = station.arrivals.filter(arr => {
          if (!state.selectedDate) return true
          return arr.arrivalDate === state.selectedDate
        })
        
        const departures = station.departures.filter(dep => {
          if (!state.selectedDate) return true
          return dep.departureDate === state.selectedDate
        })
        
        result.push({
          code: station.code,
          networkPointName: station.networkPointName,
          arrivals,
          departures
        })
      })
      
      return result
    },

    /**
     * Total records count
     */
    recordsCount(state) {
      return state.records.length
    },

    /**
     * Get station by code
     */
    getStationByCode: (state) => (code) => {
      return state.stations.find(s => s.code === code)
    },

    /**
     * Get vehicle info by name
     */
    getVehicleByName: (state) => (name) => {
      return state.vehicles.find(v => v.vehicle === name)
    },

    /**
     * Get staff member by personnel number
     */
    getStaffById: (state) => (id) => {
      return state.staff.find(s => s.id === id)
    },

    /**
     * Get train by number and date
     */
    getTrainByNoAndDate: (state) => (trainNo, date) => {
      const dayTrains = state.trains[date]
      if (!dayTrains) return null
      return dayTrains.find(t => t.no === trainNo)
    },
  },

  actions: {
    /**
     * Load field mappings from API
     * Required before importing Excel data
     */
    async loadFieldMappings() {
      const loggingStore = useLoggingStore()
      const slogStore = useSlogStore()

      try {
        const response = await api.get('/antras-field-mappings/map')
        this.fieldMappings = response.data
        this.fieldMappingsLoaded = true

        loggingStore.info('Laukų atvaizdavimai įkelti', {
          component: 'antrasStore',
          action: 'load_field_mappings',
          fieldsCount: Object.keys(response.data).length
        })

        return true

      } catch (error) {
        loggingStore.error('Klaida įkeliant laukų atvaizdavimus', {
          component: 'antrasStore',
          action: 'load_field_mappings_error',
          error: error.message
        })

        // Use fallback mappings for development
        this.fieldMappings = this.getDefaultMappings()
        this.fieldMappingsLoaded = true

        slogStore.addToast({
          message: 'Naudojami numatytieji laukų atvaizdavimai',
          type: 'alert-warning'
        })

        return true
      }
    },

    /**
     * Default field mappings as fallback
     */
    getDefaultMappings() {
      return {
        'Network point name': 'networkPointName',
        'Technical vehicle type.in': 'technicalVehicleTypeIn',
        'Technical vehicle type.out': 'technicalVehicleTypeOut',
        'Vehicle no.in': 'vehicleNoIn',
        'Vehicle no.out': 'vehicleNoOut',
        'Train No.in': 'trainNoIn',
        'Train No.out': 'trainNoOut',
        'Arrival': 'arrival',
        'Departure': 'departure',
        'Duty.in': 'dutyIn',
        'Duty.out': 'dutyOut',
        'Driver.in': 'driverIn',
        'Phone.in': 'phoneIn',
        'Driver.out': 'driverOut',
        'Phone.out': 'phoneOut',
        'Driver.PersonnelNumber.in': 'driverPersonnelNumberIn',
        'Driver.PersonnelNumber.out': 'driverPersonnelNumberOut',
        'Duty.StartingTime.in': 'dutyStartingTimeIn',
        'Duty.StartingTime.out': 'dutyStartingTimeOut',
        'Duty.EndTime.in': 'dutyEndTimeIn',
        'Duty.EndTime.out': 'dutyEndTimeOut',
        'Validity.in': 'validityIn',
        'Validity.out': 'validityOut',
        'Starting location.in': 'startingLocationIn',
        'Starting location.out': 'startingLocationOut',
        'Starting time.in': 'startingTimeIn',
        'Starting time.out': 'startingTimeOut',
        'End location.in': 'endLocationIn',
        'End location.out': 'endLocationOut',
        'Ending time.in': 'endingTimeIn',
        'Ending time.out': 'endingTimeOut',
        'Vehicle working.in': 'vehicleWorkingIn',
        'Vehicle working.out': 'vehicleWorkingOut',
        'Vehicle reg. no.in': 'vehicleRegNoIn',
        'Vehicle reg. no.out': 'vehicleRegNoOut',
        'Train length.in': 'trainLengthIn',
        'Train length.out': 'trainLengthOut',
      }
    },

    /**
     * Parse vehicle name from technical type and number
     * Implements all rules from the specification
     * @param {string} technicalType - e.g., "620M", "Siemens", "*730ML m"
     * @param {string} vehicleNo - e.g., "620-010", "731-004,733-004"
     * @returns {string} - Formatted vehicle name
     */
        parseVehicleName(typeStr, vehicleStr) {
      if (!typeStr || !vehicleStr) return null

      const types = typeStr.split(',').map(t => t.trim().replace(/^\*/, ''))
      const vehicles = vehicleStr.split(',').map(v => v.trim())
      
      if (types.length === 0 || vehicles.length === 0) return null

      const firstType = types[0]
      const firstVehicle = vehicles[0]

      // Single vehicle types: 620M, Siemens
      if (firstType === '620M' || firstType === 'Siemens') {
        return firstVehicle
      }

      // Passenger cars: *Seat, *Coupe
      if (firstType === 'Seat' || firstType === 'Coupe') {
        const count = types.length
        const baseNum = firstVehicle.split('-')[0]
        return `${count} vag. ${baseNum}`
      }

      // 630 series: find 631-XXX, result = 630MiL-XXX
      if (firstType.startsWith('630')) {
        const vehicle631 = vehicles.find(v => v.startsWith('631-'))
        if (vehicle631) {
          const num = vehicle631.split('-')[1]
          return `630MiL-${num}`
        }
        return firstVehicle
      }

      // 730 series: find 731-XXX, result = 730ML-XXX
      if (firstType.startsWith('730')) {
        const vehicle731 = vehicles.find(v => v.startsWith('731-'))
        if (vehicle731) {
          const num = vehicle731.split('-')[1]
          return `730ML-${num}`
        }
        return firstVehicle
      }

      // EJ575: find 211-XXX, result = EJ575-XXX
      if (firstType.startsWith('EJ575')) {
        const vehicle211 = vehicles.find(v => v.startsWith('211-'))
        if (vehicle211) {
          const num = vehicle211.split('-')[1]
          return `EJ575-${num}`
        }
        return firstVehicle
      }

      // DR1A/DR1AM: find type with 'm', use type before space/m + vehicle number
      if (firstType.startsWith('DR1A')) {
        // Find index of type containing lowercase 'm'
        const mIndex = types.findIndex(t => /m[^c]?$/i.test(t) || t.includes(' m'))
        if (mIndex !== -1) {
          const mType = types[mIndex]
          const mVehicle = vehicles[mIndex]
          // Extract base type (before space or 'm')
          let baseType = mType.replace(/\s*m.*$/i, '').replace(/m$/i, '')
          // Handle DR1AMv → DR1AMv, DR1A 3 → DR1A
          if (mType.includes(' ')) {
            baseType = mType.split(' ')[0]
          } else {
            // DR1AMvm → DR1AMv, DR1AM m → DR1AM
            baseType = mType.replace(/m$/, '')
          }
          return `${baseType} ${mVehicle}`
        }
        return `${firstType} ${firstVehicle}`
      }

      // RA-2: find vehicle ending with -01
      if (firstType.startsWith('RA-2')) {
        const vehicle01 = vehicles.find(v => v.endsWith('-01'))
        if (vehicle01) {
          const baseType = firstType.split(' ')[0]
          return `${baseType}-${vehicle01.split('-')[0]}`
        }
        return firstVehicle
      }

      // Default: return first vehicle
      return firstVehicle
    },
    /**
     * Parse time string with optional day offset
     * Handles formats like: "08:48", "23:59 (+1)", "00:24 (-1)"
     * @param {string|Date} timeValue - Time string or Date object
     * @param {string|Date} baseDate - Base date from Validity field
     * @returns {Date|null} - Parsed Date object
     */
    parseTimeWithOffset(timeValue, baseDate) {
      if (!timeValue) return null

      // If already a Date object
      if (timeValue instanceof Date) return timeValue

      const timeStr = String(timeValue).trim()
      
      // Match time with optional day offset: "08:48" or "00:24 (+1)"
      const match = timeStr.match(/^(\d{1,2}):(\d{2})(?:\s*\(([+-]\d+)\))?$/)
      if (!match) return null

      const hours = parseInt(match[1], 10)
      const minutes = parseInt(match[2], 10)
      const dayOffset = match[3] ? parseInt(match[3], 10) : 0

      // Parse base date
      let date
      if (baseDate instanceof Date) {
        date = new Date(baseDate)
      } else if (baseDate) {
        date = new Date(baseDate)
      } else {
        // Use today as fallback
        date = new Date()
      }
      
      if (isNaN(date.getTime())) {
        date = new Date()
      }

      // Set time
      date.setHours(hours, minutes, 0, 0)
      
      // Apply day offset
      if (dayOffset !== 0) {
        date.setDate(date.getDate() + dayOffset)
      }

      return date
    },

    /**
     * Parse date from Validity field
     * Handles formats like: "2025-12-16", Date objects, Excel serial numbers
     * @param {*} value - Raw value from Excel
     * @returns {string|null} - ISO date string (YYYY-MM-DD)
     */
    parseValidityDate(value) {
      if (!value) return null

      // Already a string in expected format
      if (typeof value === 'string') {
        const match = value.match(/^(\d{4})-(\d{2})-(\d{2})/)
        if (match) return match[0]
      }

      // Date object
      if (value instanceof Date) {
        return value.toISOString().split('T')[0]
      }

      // Excel serial date number
      if (typeof value === 'number') {
        // Excel serial date: days since 1899-12-30
        const excelEpoch = new Date(1899, 11, 30)
        const date = new Date(excelEpoch.getTime() + value * 86400000)
        return date.toISOString().split('T')[0]
      }

      return null
    },

    /**
     * Parse staff information from comma-separated fields
     * @param {Object} row - Row data with driver, phone, personnelNumber, duty fields
     * @param {string} suffix - 'In' or 'Out'
     * @param {Date} baseDate - Base date for time parsing
     * @returns {Array} - Array of staff objects
     */
    parseStaff(row, suffix, baseDate) {
      const drivers = (row[`driver${suffix}`] || '').split(',').map(d => d.trim()).filter(Boolean)
      const phones = (row[`phone${suffix}`] || '').split(',').map(p => p.trim())
      const personnelNumbers = (row[`driverPersonnelNumber${suffix}`] || '').split(',').map(p => p.trim())
      const duties = (row[`duty${suffix}`] || '').split(',').map(d => d.trim())
      const startingTimes = (row[`dutyStartingTime${suffix}`] || '').split(',').map(t => t.trim())
      const endTimes = (row[`dutyEndTime${suffix}`] || '').split(',').map(t => t.trim())

      const staffList = []
      
      for (let i = 0; i < drivers.length; i++) {
        if (!drivers[i]) continue

        const duty = duties[i] || ''
        // First letter indicates occupation: M = mašinistas, K = konduktorius, R = reservas
        let occ = duty.charAt(0).toUpperCase()
        occ = occ === 'R' ? duty.charAt(1).toUpperCase() : occ

        staffList.push({
          id: personnelNumbers[i] || null,
          name: drivers[i],
          phone: phones[i] || null,
          occ: occ === 'M' ? 'M' : (occ === 'K' ? 'K' : null),
          duty: duty,
          dutyStartingTime: this.parseTimeWithOffset(startingTimes[i], baseDate),
          dutyEndTime: this.parseTimeWithOffset(endTimes[i], baseDate),
        })
      }

      return staffList
    },

    /**
     * Normalize depot sheet name
     * LTE+D or LTE-D → LTE.D
     * @param {string} sheetName - Original sheet name
     * @returns {string} - Normalized name
     */
    normalizeDepotCode(sheetName) {
      return sheetName.replace(/[+-]D$/, '.D')
    },

    /**
     * Check if sheet is a depot sheet (+D or -D suffix)
     * @param {string} sheetName - Sheet name
     * @returns {boolean}
     */
    isDepotSheet(sheetName) {
      return /[+-]D$/.test(sheetName)
    },

    /**
     * Generate unique key for deduplication
     * @param {Object} row - Row data
     * @returns {string} - Unique key
     */
    generateDedupeKey(row) {
      return [
        row.validityIn || '',
        row.validityOut || '',
        row.trainNoIn || '',
        row.trainNoOut || '',
        row.arrival || '',
        row.departure || '',
        row.vehicleNoIn || '',
        row.vehicleNoOut || ''
      ].join('|')
    },

    /**
     * Import data from Excel file
     * Main entry point implementing all rules
     * @param {File} file - Excel file object
     * @returns {number} - Number of imported records
     */
    async importFromExcel(file) {
      if (!file) return 0

      const loggingStore = useLoggingStore()
      const slogStore = useSlogStore()

      // Check if field mappings are loaded
      if (!this.fieldMappingsLoaded || !this.fieldMappings) {
        slogStore.addToast({
          message: 'Laukų atvaizdavimai dar neįkelti. Bandykite vėl.',
          type: 'alert-warning'
        })
        
        loggingStore.error('Bandoma importuoti be laukų atvaizdavimų', {
          component: 'antrasStore',
          action: 'import_without_mappings'
        })
        
        return 0
      }

      this.isLoading = true

      try {
        // Read file as array buffer
        const buffer = await file.arrayBuffer()
        
        // Create new workbook and load the file
        const workbook = new ExcelJS.Workbook()
        await workbook.xlsx.load(buffer)
        
        loggingStore.info('Excel failas nuskaitytas', {
          component: 'antrasStore',
          action: 'excel_read',
          fileName: file.name,
          sheetsCount: workbook.worksheets.length
        })

        // Clear existing data
        this.clearData()

        // Temporary storage for processing
        const stationsMap = new Map()      // code → station data
        const vehiclesMap = new Map()      // vehicle name → vehicle data
        const vehicleWorkingsMap = new Map() // vehicleWorking → working data
        const staffMap = new Map()         // personnelNumber → staff data
        const dutiesMap = new Map()        // duty code → duty data
        const trainsMap = new Map()        // date → Map(trainNo → train data)
        const processedKeys = new Set()    // For deduplication

        let totalRecords = 0
        let skippedSheets = 0
        let duplicatesSkipped = 0
        let globalRowId = 0              // Global counter for unique IDs

        // Process each worksheet
        workbook.eachSheet((worksheet, sheetId) => {
          const sheetName = worksheet.name
          
          // Skip first empty sheet
          if (sheetName === 'Sheet1' && worksheet.rowCount <= 1) {
            skippedSheets++
            return
          }

          // Check if sheet has data (at least headers + 1 row)
          if (worksheet.rowCount < 2) {
            skippedSheets++
            return
          }

          // Get headers from first row
          const headerRow = worksheet.getRow(1)
          const headers = []
          
          headerRow.eachCell({ includeEmpty: true }, (cell, colNumber) => {
            headers[colNumber - 1] = cell.value ? String(cell.value).trim() : null
          })

          // Validate headers - check required fields exist
          const requiredFields = ['Network point name', 'Arrival', 'Departure']
          const missingRequired = requiredFields.filter(f => !headers.includes(f))
          
          if (missingRequired.length > 0) {
            loggingStore.warn(`Лист ${sheetName}: netinkamas formatas`, {
              component: 'antrasStore',
              sheet: sheetName,
              missingFields: missingRequired
            })
            skippedSheets++
            return
          }

          // Determine station code
          const stationCode = this.normalizeDepotCode(sheetName)
          const isDepot = this.isDepotSheet(sheetName)

          // Process data rows (starting from row 2)
          for (let rowNumber = 2; rowNumber <= worksheet.rowCount; rowNumber++) {
            const row = worksheet.getRow(rowNumber)
            
            // Skip empty rows
            if (!row.hasValues) continue

            // Extract cell values and map to internal names
            const rawData = {}
            const mappedData = {}
            
            row.eachCell({ includeEmpty: true }, (cell, colNumber) => {
              const header = headers[colNumber - 1]
              if (!header) return
              
              let value = cell.value
              
              // Handle different cell types
              if (value === null || value === undefined) {
                value = null
              } else if (value.result !== undefined) {
                // Formula cell - use result
                value = value.result
              } else if (value.text !== undefined) {
                // Rich text cell
                value = value.text
              } else if (value instanceof Date) {
                // Keep as Date
              } else {
                value = String(value)
              }
              
              rawData[header] = value
              
              // Map to internal name
              const internalName = this.fieldMappings[header]
              if (internalName) {
                mappedData[internalName] = value
              }
            })

            // Deduplication for depot sheets
            if (isDepot) {
              const dedupeKey = this.generateDedupeKey(mappedData)
              if (processedKeys.has(dedupeKey)) {
                duplicatesSkipped++
                continue
              }
              processedKeys.add(dedupeKey)
            }

            // Increment global row ID
            globalRowId++

            // Parse validity dates - THIS IS THE KEY FIX
            const validityDateIn = this.parseValidityDate(mappedData.validityIn)
            const validityDateOut = this.parseValidityDate(mappedData.validityOut)

            // Use validity dates as base for time parsing
            const baseDateIn = validityDateIn ? new Date(validityDateIn) : new Date()
            const baseDateOut = validityDateOut ? new Date(validityDateOut) : new Date()

            // Parse times with proper date context
            const arrivalTime = this.parseTimeWithOffset(mappedData.arrival, baseDateIn)
            const arrivalDecimal = timeToDecimal(arrivalTime)
            const departureTime = this.parseTimeWithOffset(mappedData.departure, baseDateOut)
            const departureDecimal = timeToDecimal(departureTime)

            // Parse vehicle names
            const vehicleIn = this.parseVehicleName(
              mappedData.technicalVehicleTypeIn,
              mappedData.vehicleNoIn
            )
            const vehicleOut = this.parseVehicleName(
              mappedData.technicalVehicleTypeOut,
              mappedData.vehicleNoOut
            )

            // Parse staff information with correct date context
            const staffIn = this.parseStaff(mappedData, 'In', baseDateIn)
            const staffOut = this.parseStaff(mappedData, 'Out', baseDateOut)

            // Initialize station if not exists
            if (!stationsMap.has(stationCode)) {
              stationsMap.set(stationCode, {
                code: stationCode,
                networkPointName: mappedData.networkPointName || stationCode,
                arrivals: [],
                departures: []
              })
            }
            
            const station = stationsMap.get(stationCode)

            // Add arrival record with date binding
            if (mappedData.trainNoIn || vehicleIn) {
              // FIXED: Unique ID using global counter
              const arrivalId = `arr.${stationCode}---${globalRowId}`
              
              station.arrivals.push({
                id: arrivalId,
                rowID: globalRowId,
                trainNo: mappedData.trainNoIn,
                arrivalDate: validityDateIn,  // Date binding
                arrival: arrivalTime,
                arrivalDecimal,
                arrivalPlanned: arrivalTime ? arrivalTime.toLocaleTimeString('lt-LT', { hour: '2-digit', minute: '2-digit' }) : null,
                vehicle: vehicleIn,
                vehicleWorking: mappedData.vehicleWorkingIn,
                staff: staffIn,
                driverPersonnelNumber: staffIn.length > 0 ? staffIn[0].id : null,
                departureTrainNo: mappedData.trainNoOut,  // Link to departure
                targetTrack: null,
                startingLocation: mappedData.startingLocationIn,
                endLocation: mappedData.endLocationIn,
                // Compatibility fields
                arrivalTrainNumber: mappedData.trainNoIn,
                arrivalEmployee1: staffIn.filter(s => s.occ === 'M').map(s => s.name).join(', '),
              })
            }

            // Add departure record with date binding
            if (mappedData.trainNoOut || vehicleOut) {
              // FIXED: Unique ID using global counter
              const departureId = `dep.${stationCode}---${globalRowId}`
              
              station.departures.push({
                id: departureId,
                rowID: globalRowId,
                trainNo: mappedData.trainNoOut,
                departureDate: validityDateOut,  // Date binding
                departure: departureTime,
                departureDecimal,
                departurePlanned: departureTime ? departureTime.toLocaleTimeString('lt-LT', { hour: '2-digit', minute: '2-digit' }) : null,
                vehicle: vehicleOut,
                vehicleWorking: mappedData.vehicleWorkingOut,
                staff: staffOut,
                driverPersonnelNumber: staffOut.length > 0 ? staffOut[0].id : null,
                arrivalTrainNo: mappedData.trainNoIn,  // Link to arrival
                startingTrack: null,
                startingLocation: mappedData.startingLocationOut,
                endLocation: mappedData.endLocationOut,
                // Compatibility fields
                departureTrainNumber: mappedData.trainNoOut,
                departureEmployee1: staffOut.filter(s => s.occ === 'M').map(s => s.name).join(', '),
              })
            }

            // Collect trains by date - NEW FEATURE
            const collectTrain = (trainNo, date, stationCode, time, isArrival, staff, startLoc, endLoc) => {
              if (!trainNo || !date) return

              if (!trainsMap.has(date)) {
                trainsMap.set(date, new Map())
              }
              
              const dayTrains = trainsMap.get(date)
              
              if (!dayTrains.has(trainNo)) {
                dayTrains.set(trainNo, {
                  no: trainNo,
                  startingLocation: null,
                  endLocation: null,
                  staff: [],
                  stops: []
                })
              }
              
              const train = dayTrains.get(trainNo)
              
              // Update start/end locations
              if (startLoc && !train.startingLocation) train.startingLocation = startLoc
              if (endLoc) train.endLocation = endLoc
              
              // Add stop
              const stopExists = train.stops.some(s => s.station === stationCode && s.code === stationCode)
              if (!stopExists) {
                train.stops.push({
                  station: mappedData.networkPointName || stationCode,
                  code: stationCode,
                  arrival: isArrival ? time : null,
                  departure: !isArrival ? time : null
                })
              } else {
                // Update existing stop
                const stop = train.stops.find(s => s.code === stationCode)
                if (stop) {
                  if (isArrival && !stop.arrival) stop.arrival = time
                  if (!isArrival && !stop.departure) stop.departure = time
                }
              }
              
              // Add staff (avoiding duplicates)
              staff.forEach(s => {
                if (!s.id) return
                const exists = train.staff.some(ts => ts.id === s.id)
                if (!exists) {
                  train.staff.push({
                    id: s.id,
                    name: s.name,
                    phone: s.phone,
                    duty: s.duty,
                    dutyStartingTime: s.dutyStartingTime,
                    dutyEndTime: s.dutyEndTime
                  })
                }
              })
            }

            // Collect arrival train
            if (mappedData.trainNoIn && validityDateIn) {
              collectTrain(
                mappedData.trainNoIn,
                validityDateIn,
                stationCode,
                arrivalTime,
                true,
                staffIn,
                mappedData.startingLocationIn,
                mappedData.endLocationIn
              )
            }

            // Collect departure train
            if (mappedData.trainNoOut && validityDateOut) {
              collectTrain(
                mappedData.trainNoOut,
                validityDateOut,
                stationCode,
                departureTime,
                false,
                staffOut,
                mappedData.startingLocationOut,
                mappedData.endLocationOut
              )
            }

            // Collect vehicle information
            const collectVehicle = (vehicle, vehicleNo, vehicleRegNo, vehicleWorking, startLoc, startTime, endLoc, endTime, baseDate) => {
              if (!vehicle) return
              
              if (!vehiclesMap.has(vehicle)) {
                vehiclesMap.set(vehicle, {
                  vehicle,
                  vehicleNo: [],
                  vehicleRegNo: [],
                  vehicleWorkings: []
                })
              }
              
              const v = vehiclesMap.get(vehicle)
              
              // Add numbers if not present
              if (vehicleNo) {
                const nums = vehicleNo.split(',').map(n => n.trim())
                nums.forEach(n => {
                  if (!v.vehicleNo.includes(n)) v.vehicleNo.push(n)
                })
              }
              
              if (vehicleRegNo) {
                const regs = vehicleRegNo.split(',').map(r => r.trim())
                regs.forEach(r => {
                  if (!v.vehicleRegNo.includes(r)) v.vehicleRegNo.push(r)
                })
              }
              
              if (vehicleWorking && !v.vehicleWorkings.includes(vehicleWorking)) {
                v.vehicleWorkings.push(vehicleWorking)
                
                // Add to vehicle workings map
                if (!vehicleWorkingsMap.has(vehicleWorking)) {
                  vehicleWorkingsMap.set(vehicleWorking, {
                    vehicleWorking,
                    startingTime: this.parseTimeWithOffset(startTime, baseDate),
                    startingLocation: startLoc,
                    endingTime: this.parseTimeWithOffset(endTime, baseDate),
                    endLocation: endLoc
                  })
                }
              }
            }

            collectVehicle(
              vehicleIn,
              mappedData.vehicleNoIn,
              mappedData.vehicleRegNoIn,
              mappedData.vehicleWorkingIn,
              mappedData.startingLocationIn,
              mappedData.startingTimeIn,
              mappedData.endLocationIn,
              mappedData.endingTimeIn,
              baseDateIn
            )

            collectVehicle(
              vehicleOut,
              mappedData.vehicleNoOut,
              mappedData.vehicleRegNoOut,
              mappedData.vehicleWorkingOut,
              mappedData.startingLocationOut,
              mappedData.startingTimeOut,
              mappedData.endLocationOut,
              mappedData.endingTimeOut,
              baseDateOut
            )

            // Collect staff information
            const collectStaff = (staffList, date) => {
              staffList.forEach(person => {
                if (!person.id) return
                
                if (!staffMap.has(person.id)) {
                  staffMap.set(person.id, {
                    id: person.id,
                    occ: person.occ,
                    name: person.name,
                    phone: person.phone,
                    duties: [],
                    dates: []  // Track dates for this staff
                  })
                }
                
                const staff = staffMap.get(person.id)
                
                // Update info if more complete
                if (!staff.phone && person.phone) staff.phone = person.phone
                if (!staff.occ && person.occ) staff.occ = person.occ
                
                // Add duty with date binding
                if (person.duty) {
                  const dutyKey = `${date}:${person.duty}`
                  if (!staff.duties.includes(dutyKey)) {
                    staff.duties.push(dutyKey)
                  }
                  
                  // Track dates
                  if (date && !staff.dates.includes(date)) {
                    staff.dates.push(date)
                  }
                  
                  // Also add to dutiesMap with date
                  const dutyMapKey = `${date}:${person.duty}`
                  if (!dutiesMap.has(dutyMapKey)) {
                    dutiesMap.set(dutyMapKey, {
                      id: person.duty,
                      date: date,
                      startingTime: person.dutyStartingTime,
                      endTime: person.dutyEndTime,
                      trains: []
                    })
                  }
                  
                  // Add train to duty
                  const duty = dutiesMap.get(dutyMapKey)
                  const trainNo = mappedData.trainNoIn || mappedData.trainNoOut
                  if (trainNo && !duty.trains.includes(trainNo)) {
                    duty.trains.push(trainNo)
                  }
                }
              })
            }

            collectStaff(staffIn, validityDateIn)
            collectStaff(staffOut, validityDateOut)

            // Store raw record for display - FIXED ID
            this.records.push({
              id: `rec-${globalRowId}`,
              sheetName: stationCode,
              originalSheet: sheetName,
              rowNumber,
              ...mappedData,
              vehicleIn,
              vehicleOut,
              arrivalTime,
              arrivalDecimal,
              arrivalDate: validityDateIn,
              departureTime,
              departureDecimal,
              departureDate: validityDateOut,
            })

            totalRecords++
          }

          // Add sheet to list (use normalized code)
          if (!this.sheets.includes(stationCode)) {
            this.sheets.push(stationCode)
          }
        })

        // Convert maps to arrays
        this.stations = Array.from(stationsMap.values())
        this.vehicles = Array.from(vehiclesMap.values())
        this.vehicleWorkings = Array.from(vehicleWorkingsMap.values())
        this.staff = Array.from(staffMap.values())
        this.duties = Array.from(dutiesMap.values())
        
        // Convert trains map to object structure
        this.trains = {}
        trainsMap.forEach((dayTrains, date) => {
          this.trains[date] = Array.from(dayTrains.values())
            .sort((a, b) => a.no - b.no)
        })

        this.fileName = file.name
        this.lastImported = new Date().toISOString()

        loggingStore.info('Duomenys sėkmingai importuoti', {
          component: 'antrasStore',
          action: 'import_success',
          fileName: file.name,
          stationsCount: this.stations.length,
          vehiclesCount: this.vehicles.length,
          staffCount: this.staff.length,
          trainsCount: Object.values(this.trains).flat().length,
          recordsCount: totalRecords,
          skippedSheets,
          duplicatesSkipped
        })

        slogStore.addToast({
          message: `Sėkmingai importuota: ${totalRecords} įrašų, ${this.stations.length} stočių, ${this.vehicles.length} riedmenų`,
          type: 'alert-success'
        })

        return totalRecords

      } catch (error) {
        loggingStore.error('Klaida importuojant Antras duomenis', {
          component: 'antrasStore',
          action: 'import_error',
          fileName: file.name,
          error: error.message,
          stack: error.stack
        })

        slogStore.addToast({
          message: `Importavimo klaida: ${error.message}`,
          type: 'alert-error'
        })

        return 0

      } finally {
        this.isLoading = false
      }
    },

    /**
     * Validate Excel headers against field mappings
     * @param {Array} headers - Array of header strings
     * @returns {Object} - { valid: boolean, missing: [], extra: [] }
     */
    validateHeaders(headers) {
      const expectedHeaders = Object.keys(this.fieldMappings)
      const headerSet = new Set(headers.filter(h => h !== null))
      
      const missing = expectedHeaders.filter(h => !headerSet.has(h))
      const extra = headers.filter(h => h && !this.fieldMappings[h])
      
      return {
        valid: missing.length === 0,
        missing,
        extra
      }
    },

    /**
     * Set filters for data display
     * @param {Object} filters - { sheet, date }
     */
    setFilters({ sheet, date }) {
      this.selectedSheet = sheet || null
      this.selectedDate = date || null
    },

    /**
     * Clear all imported data
     */
    clearData() {
      this.records = []
      this.sheets = []
      this.stations = []
      this.vehicles = []
      this.vehicleWorkings = []
      this.staff = []
      this.duties = []
      this.trains = {}
      this.selectedSheet = null
      this.selectedDate = null
      this.lastImported = null
      this.fileName = null
    },

    /**
     * Export current data to JSON file
     */
    exportToJson() {
      const loggingStore = useLoggingStore()
      
      const exportData = {
        exportedAt: new Date().toISOString(),
        fileName: this.fileName,
        stations: this.stations,
        vehicles: this.vehicles,
        vehicleWorkings: this.vehicleWorkings,
        staff: this.staff,
        duties: this.duties,
        trains: this.trains
      }

      const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `antras-export-${new Date().toISOString().split('T')[0]}.json`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)

      loggingStore.info('Antras duomenys eksportuoti į JSON', {
        component: 'antrasStore',
        action: 'export_json',
        stationsCount: this.stations.length
      })
    },

    /**
     * Update a record (arrival or departure) with new data
     * Used for track assignment from TracksTimeLine
     * @param {string} recordId - Record ID
     * @param {Object} updates - Object with fields to update
     */
    updateRecord(recordId, updates) {
      if (!recordId || !updates) return false

      // Search in arrivals
      for (const station of this.stations) {
        const arrival = station.arrivals.find(arr => arr.id === recordId)
        if (arrival) {
          Object.assign(arrival, updates)
          return true
        }

        const departure = station.departures.find(dep => dep.id === recordId)
        if (departure) {
          Object.assign(departure, updates)
          return true
        }
      }

      return false
    },

    /**
     * Get summary statistics
     * @returns {Object} - Statistics object
     */
    getStatistics() {
      const totalArrivals = this.stations.reduce((sum, s) => sum + s.arrivals.length, 0)
      const totalDepartures = this.stations.reduce((sum, s) => sum + s.departures.length, 0)
      const totalTrains = Object.values(this.trains).flat().length
      
      return {
        stations: this.stations.length,
        vehicles: this.vehicles.length,
        vehicleWorkings: this.vehicleWorkings.length,
        staff: this.staff.length,
        duties: this.duties.length,
        arrivals: totalArrivals,
        departures: totalDepartures,
        records: this.records.length,
        trains: totalTrains,
        dates: this.availableDates.length
      }
    }
  }
})
