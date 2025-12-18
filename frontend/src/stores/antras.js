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
 */
export const useAntrasStore = defineStore('antras', {
  state: () => ({
    // Core data structures from todo.txt
    vehicles: [],           // { vehicle, vehicleNo:[], vehicleRegNo:[], vehicleWorkings:[] }
    vehicleWorkings: [],    // { vehicleWorking, startingTime, startingLocation, endingTime, endLocation }
    staff: [],              // { id, occ, name, phone, duties:[] }
    duties: [],             // { id, startingTime, endTime, trains:[] }
    stations: [],           // { code, networkPointName, arrivals:[], departures:[] }
    
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
          if (arr.arrival instanceof Date) {
            dates.add(arr.arrival.toISOString().split('T')[0])
          }
        })
        station.departures.forEach(dep => {
          if (dep.departure instanceof Date) {
            dates.add(dep.departure.toISOString().split('T')[0])
          }
        })
      })
      return Array.from(dates).sort()
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
          if (arr.arrival instanceof Date) {
            return arr.arrival.toISOString().split('T')[0] === state.selectedDate
          }
          return false
        })
        
        const departures = station.departures.filter(dep => {
          if (!state.selectedDate) return true
          if (dep.departure instanceof Date) {
            return dep.departure.toISOString().split('T')[0] === state.selectedDate
          }
          return false
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
          count: Object.keys(this.fieldMappings).length
        })

        return true
      } catch (error) {
        loggingStore.error('Nepavyko įkelti laukų atvaizdavimų', {
          component: 'antrasStore',
          action: 'load_field_mappings_failed',
          error: error.message
        })

        slogStore.addToast({
          message: 'Nepavyko įkelti laukų atvaizdavimų',
          type: 'alert-error'
        })

        return false
      }
    },

    /**
     * Parse vehicle name from Technical vehicle type and Vehicle no
     * Implements all rules from todo.txt
     * @param {string} typeStr - Technical vehicle type (comma-separated)
     * @param {string} vehicleStr - Vehicle numbers (comma-separated)
     * @returns {string} - Parsed vehicle name
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
     * Parse time string with day offset
     * Handles formats like "08:48 (+1)", "23:59 (-1)"
     * @param {string} timeStr - Time string
     * @param {Date|string} baseDate - Base date from Validity field
     * @returns {Date|null} - Parsed date or null
     */
    parseTimeWithOffset(timeStr, baseDate) {
      if (!timeStr || !baseDate) return null

      // Handle Date objects from Excel
      if (timeStr instanceof Date) {
        return timeStr
      }

      const str = String(timeStr).trim()
      
      // Extract time and optional offset
      const match = str.match(/^(\d{1,2}):(\d{2})(?:\s*\(([+-]\d+)\))?$/)
      if (!match) return null

      const hours = parseInt(match[1], 10)
      const minutes = parseInt(match[2], 10)
      const dayOffset = match[3] ? parseInt(match[3], 10) : 0

      // Parse base date
      let date
      if (baseDate instanceof Date) {
        date = new Date(baseDate)
      } else {
        date = new Date(baseDate)
      }
      
      if (isNaN(date.getTime())) return null

      // Set time
      date.setHours(hours, minutes, 0, 0)
      
      // Apply day offset
      if (dayOffset !== 0) {
        date.setDate(date.getDate() + dayOffset)
      }

      return date
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
        // First letter indicates occupation: M = mašinistas, K = konduktorius
        const occ = duty.charAt(0).toUpperCase()

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
      // Replace +D or -D with .D for depot sheets
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
     * Main entry point implementing all rules from todo.txt
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
        const processedKeys = new Set()    // For deduplication

        let totalRecords = 0
        let skippedSheets = 0
        let duplicatesSkipped = 0

        // Process each worksheet
        workbook.eachSheet((worksheet, id) => {
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
              } else if (value instanceof Date) {
                // Keep as Date object
              } else if (typeof value === 'object') {
                if (value.result !== undefined) {
                  value = value.result
                } else if (value.text !== undefined) {
                  value = value.text
                } else if (value.richText) {
                  value = value.richText.map(r => r.text).join('')
                } else {
                  value = String(value)
                }
              }
              
              rawData[header] = value
              
              // Map to internal name
              const internalName = this.fieldMappings[header]
              if (internalName) {
                mappedData[internalName] = value
              }
            })

            // Generate deduplication key
            const dedupeKey = this.generateDedupeKey(mappedData)
            if (processedKeys.has(dedupeKey)) {
              duplicatesSkipped++
              continue
            }
            processedKeys.add(dedupeKey)

            // Get base date from Validity field
            const baseDateIn = mappedData.validityIn || new Date()
            const baseDateOut = mappedData.validityOut || new Date()

            // Parse arrival time
            const arrivalTime = this.parseTimeWithOffset(mappedData.arrival, baseDateIn)
            const arrivalDecimal = timeToDecimal(arrivalTime)

            // Parse departure time
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

            // Parse staff information
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

            // Add arrival record
            if (mappedData.trainNoIn || vehicleIn) {
              const arrivalId = `${stationCode}-arr-${rowNumber}-${arrivalDecimal || Date.now()}`
              station.arrivals.push({
                id: arrivalId,
                trainNo: mappedData.trainNoIn,
                // Compatibility fields for TracksTimeLine
                arrivalTrainNumber: mappedData.trainNoIn,
                arrivalPlanned: arrivalTime ? arrivalTime.toLocaleTimeString('lt-LT', { hour: '2-digit', minute: '2-digit' }) : null,
                arrivalDate: arrivalTime ? arrivalTime.toISOString().split('T')[0] : null,
                arrivalEmployee1: staffIn.filter(s => s.occ === 'M').map(s => s.name).join(', '),
                arrival: arrivalTime,
                arrivalDecimal,
                vehicle: vehicleIn,
                driverPersonnelNumber: staffIn.length > 0 ? staffIn[0].id : null,
                departureTrainNo: mappedData.trainNoOut,
                staff: staffIn,
                // Track assignment (for TracksTimeLine)
                targetTrack: null,
                // Additional fields for display
                startingLocation: mappedData.startingLocationIn,
                endLocation: mappedData.endLocationIn,
                vehicleWorking: mappedData.vehicleWorkingIn,
              })
            }

            // Add departure record
            if (mappedData.trainNoOut || vehicleOut) {
              const departureId = `${stationCode}-dep-${rowNumber}-${departureDecimal || Date.now()}`
              station.departures.push({
                id: departureId,
                trainNo: mappedData.trainNoOut,
                // Compatibility fields for TracksTimeLine
                departureTrainNumber: mappedData.trainNoOut,
                departurePlanned: departureTime ? departureTime.toLocaleTimeString('lt-LT', { hour: '2-digit', minute: '2-digit' }) : null,
                departureDate: departureTime ? departureTime.toISOString().split('T')[0] : null,
                departureEmployee1: staffOut.filter(s => s.occ === 'M').map(s => s.name).join(', '),
                departure: departureTime,
                departureDecimal,
                vehicle: vehicleOut,
                driverPersonnelNumber: staffOut.length > 0 ? staffOut[0].id : null,
                staff: staffOut,
                // Track assignment (for TracksTimeLine)
                startingTrack: null,
                // Additional fields for display
                startingLocation: mappedData.startingLocationOut,
                endLocation: mappedData.endLocationOut,
                vehicleWorking: mappedData.vehicleWorkingOut,
              })
            }

            // Collect vehicle information
            const collectVehicle = (vehicleName, vehicleNo, vehicleRegNo, vehicleWorking, startLoc, startTime, endLoc, endTime, baseDate) => {
              if (!vehicleName) return
              
              if (!vehiclesMap.has(vehicleName)) {
                vehiclesMap.set(vehicleName, {
                  vehicle: vehicleName,
                  vehicleNo: [],
                  vehicleRegNo: [],
                  vehicleWorkings: []
                })
              }
              
              const veh = vehiclesMap.get(vehicleName)
              
              // Add vehicle numbers (split and dedupe)
              if (vehicleNo) {
                vehicleNo.split(',').forEach(no => {
                  const trimmed = no.trim()
                  if (trimmed && !veh.vehicleNo.includes(trimmed)) {
                    veh.vehicleNo.push(trimmed)
                  }
                })
              }
              
              // Add registration numbers
              if (vehicleRegNo) {
                vehicleRegNo.split(',').forEach(no => {
                  const trimmed = no.trim()
                  if (trimmed && !veh.vehicleRegNo.includes(trimmed)) {
                    veh.vehicleRegNo.push(trimmed)
                  }
                })
              }
              
              // Add vehicle working
              if (vehicleWorking && !veh.vehicleWorkings.includes(vehicleWorking)) {
                veh.vehicleWorkings.push(vehicleWorking)
                
                // Also add to vehicleWorkingsMap
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
              mappedData.baseDateIn
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
              mappedData.baseDateOut
            )

            // Collect staff information
            const collectStaff = (staffList) => {
              staffList.forEach(person => {
                if (!person.id) return
                
                if (!staffMap.has(person.id)) {
                  staffMap.set(person.id, {
                    id: person.id,
                    occ: person.occ,
                    name: person.name,
                    phone: person.phone,
                    duties: []
                  })
                }
                
                const staff = staffMap.get(person.id)
                
                // Update info if more complete
                if (!staff.phone && person.phone) staff.phone = person.phone
                if (!staff.occ && person.occ) staff.occ = person.occ
                
                // Add duty
                if (person.duty && !staff.duties.includes(person.duty)) {
                  staff.duties.push(person.duty)
                  
                  // Also add to dutiesMap
                  if (!dutiesMap.has(person.duty)) {
                    dutiesMap.set(person.duty, {
                      id: person.duty,
                      startingTime: person.dutyStartingTime,
                      endTime: person.dutyEndTime,
                      trains: []
                    })
                  }
                  
                  // Add train to duty
                  const duty = dutiesMap.get(person.duty)
                  const trainNo = mappedData.trainNoIn || mappedData.trainNoOut
                  if (trainNo && !duty.trains.includes(trainNo)) {
                    duty.trains.push(trainNo)
                  }
                }
              })
            }

            collectStaff(staffIn)
            collectStaff(staffOut)

            // Store raw record for display
            this.records.push({
              id: `${stationCode}-${rowNumber}`,
              sheetName: stationCode,
              ...mappedData,
              vehicleIn,
              vehicleOut,
              arrivalTime,
              arrivalDecimal,
              departureTime,
              departureDecimal,
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

        this.fileName = file.name
        this.lastImported = new Date().toISOString()

        loggingStore.info('Duomenys sėkmingai importuoti', {
          component: 'antrasStore',
          action: 'import_success',
          fileName: file.name,
          stationsCount: this.stations.length,
          vehiclesCount: this.vehicles.length,
          staffCount: this.staff.length,
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
        duties: this.duties
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
      
      return {
        stations: this.stations.length,
        vehicles: this.vehicles.length,
        vehicleWorkings: this.vehicleWorkings.length,
        staff: this.staff.length,
        duties: this.duties.length,
        arrivals: totalArrivals,
        departures: totalDepartures,
        records: this.records.length
      }
    }
  }
})
