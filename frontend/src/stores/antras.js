// frontend/src/stores/antras.js
import { defineStore } from 'pinia'
import { useLoggingStore } from './logging'
import { useSlogStore } from './slog'
import ExcelJS from 'exceljs'

/**
 * Antras Store - manages train movement data from Excel imports
 * 
 * This store handles the import and management of train schedule data
 * from external Excel files (182 sheets with 37 columns each).
 * Each sheet represents a network point/station with movement records.
 */
export const useAntrasStore = defineStore('antras', {
  state: () => ({
    records: [],                    // All imported records
    sheets: [],                     // List of sheet names (stations)
    isLoading: false,               // Loading state
    selectedSheet: null,            // Currently selected sheet/station
    selectedDate: null,             // Currently selected date filter
    lastImported: null,             // Timestamp of last import
    fieldMappings: null,            // Field name mappings from API
    fieldMappingsLoaded: false,     // Track if mappings are loaded
    fileName: null,                 // Name of imported file
  }),

  getters: {
    /**
     * Check if field mappings are loaded
     */
    hasFieldMappings(state) {
      return state.fieldMappings !== null && Object.keys(state.fieldMappings).length > 0
    },

    /**
     * Get available sheet names (stations)
     */
    availableSheets(state) {
      return state.sheets.sort()
    },

    /**
     * Get available dates from records
     */
    availableDates(state) {
      const dates = new Set()
      state.records.forEach(record => {
        if (record.validityIn) dates.add(record.validityIn)
        if (record.validityOut) dates.add(record.validityOut)
      })
      return Array.from(dates).sort()
    },

    /**
     * Get filtered records by selected sheet and date
     */
    filteredRecords(state) {
      let filtered = state.records

      if (state.selectedSheet) {
        filtered = filtered.filter(r => r.networkPointName === state.selectedSheet)
      }

      if (state.selectedDate) {
        filtered = filtered.filter(r => 
          r.validityIn === state.selectedDate || r.validityOut === state.selectedDate
        )
      }

      return filtered
    },

    /**
     * Get records count
     */
    recordsCount(state) {
      return state.records.length
    },
  },

  actions: {
    /**
     * Load field mappings from API
     */
    async loadFieldMappings() {
      const loggingStore = useLoggingStore()
      const slogStore = useSlogStore()

      try {
        const response = await fetch('/api/antras-field-mappings/map')
        
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}: ${response.statusText}`)
        }

        this.fieldMappings = await response.json()
        this.fieldMappingsLoaded = true

        loggingStore.info('Antras laukų atvaizdavimai įkelti', {
          component: 'antrasStore',
          action: 'load_field_mappings',
          count: Object.keys(this.fieldMappings).length
        })

        return true
      } catch (error) {
        loggingStore.error('Nepavyko įkelti Antras laukų atvaizdavimų', {
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
     * Import data from Excel file using ExcelJS
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
        this.records = []
        this.sheets = []

        let totalRecords = 0
        let sheetIndex = 0

        // Process each worksheet
        workbook.eachSheet((worksheet, id) => {
          const sheetName = worksheet.name
          
          // Skip first empty sheet
          if (sheetIndex === 0 && sheetName === 'Sheet1' && worksheet.rowCount <= 1) {
            sheetIndex++
            return
          }
          sheetIndex++

          // Check if sheet has data
          if (worksheet.rowCount < 2) {
            return
          }

          // Get headers from first row
          const headerRow = worksheet.getRow(1)
          const headers = []
          
          headerRow.eachCell({ includeEmpty: true }, (cell, colNumber) => {
            headers[colNumber - 1] = cell.value ? String(cell.value) : null
          })

          // Validate headers against field mappings
          const missingFields = this.validateHeaders(headers.filter(h => h !== null))
          if (missingFields.length > 0) {
            loggingStore.warn(`Лист ${sheetName}: trūksta laukų`, {
              component: 'antrasStore',
              sheet: sheetName,
              missingFields: missingFields
            })
          }

          // Add sheet to list
          this.sheets.push(sheetName)

          // Process data rows (starting from row 2)
          for (let rowNumber = 2; rowNumber <= worksheet.rowCount; rowNumber++) {
            const row = worksheet.getRow(rowNumber)
            
            // Skip empty rows
            if (!row.hasValues) continue

            // Extract cell values
            const rowValues = []
            row.eachCell({ includeEmpty: true }, (cell, colNumber) => {
              let value = cell.value
              
              // Handle different cell types
              if (value === null || value === undefined) {
                rowValues[colNumber - 1] = null
              } else if (typeof value === 'object') {
                // Handle dates, formulas, etc.
                if (value instanceof Date) {
                  // Format date as YYYY-MM-DD
                  rowValues[colNumber - 1] = value.toISOString().split('T')[0]
                } else if (value.result !== undefined) {
                  // Formula result
                  rowValues[colNumber - 1] = String(value.result)
                } else if (value.text !== undefined) {
                  // Rich text
                  rowValues[colNumber - 1] = value.text
                } else {
                  rowValues[colNumber - 1] = String(value)
                }
              } else {
                rowValues[colNumber - 1] = String(value)
              }
            })
            
            // Map row data to internal field names
            const record = this.mapRowToRecord(headers, rowValues, sheetName)
            
            if (record) {
              this.records.push(record)
              totalRecords++
            }
          }
        })

        this.fileName = file.name
        this.lastImported = new Date().toISOString()

        loggingStore.info('Antras duomenys sėkmingai importuoti', {
          component: 'antrasStore',
          action: 'import_success',
          fileName: file.name,
          sheetsCount: this.sheets.length,
          recordsCount: totalRecords
        })

        slogStore.addToast({
          message: `Sėkmingai importuota ${totalRecords} įrašų iš ${this.sheets.length} lapų`,
          type: 'alert-success'
        })

        return totalRecords

      } catch (error) {
        loggingStore.error('Klaida importuojant Antras duomenis', {
          component: 'antrasStore',
          action: 'import_error',
          fileName: file.name,
          error: error.message
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
     * @returns {Array} - Array of missing field names
     */
    validateHeaders(headers) {
      const missingFields = []
      
      // Check required fields
      for (const [externalName, internalName] of Object.entries(this.fieldMappings)) {
        if (!headers.includes(externalName)) {
          missingFields.push(externalName)
        }
      }
      
      return missingFields
    },

    /**
     * Map Excel row to internal record structure
     * @param {Array} headers - Array of header strings
     * @param {Array} row - Array of cell values
     * @param {string} sheetName - Sheet name (station code)
     * @returns {Object} - Mapped record object
     */
    mapRowToRecord(headers, row, sheetName) {
      const record = {
        id: this.generateRecordId(sheetName, row),
        sheetName: sheetName,
        createdAt: new Date().toISOString(),
      }

      // Map each column to internal field name
      headers.forEach((header, index) => {
        const internalName = this.fieldMappings[header]
        if (internalName) {
          record[internalName] = row[index] || null
        }
      })

      return record
    },

    /**
     * Generate unique ID for record
     * @param {string} sheetName - Sheet name
     * @param {Array} row - Row data
     * @returns {string} - Unique ID
     */
    generateRecordId(sheetName, row) {
      const timestamp = Date.now()
      const random = Math.floor(Math.random() * 10000)
      const networkPoint = row[0] || sheetName
      return `${sheetName}_${networkPoint}_${timestamp}_${random}`
    },

    /**
     * Set filters
     */
    setFilters({ sheet, date }) {
      if (sheet !== undefined) this.selectedSheet = sheet
      if (date !== undefined) this.selectedDate = date

      const loggingStore = useLoggingStore()
      loggingStore.info('Antras filtrai pakeisti', {
        component: 'antrasStore',
        sheet: this.selectedSheet,
        date: this.selectedDate
      })
    },

    /**
     * Clear all data
     */
    clearData() {
      this.records = []
      this.sheets = []
      this.selectedSheet = null
      this.selectedDate = null
      this.fileName = null
      this.lastImported = null

      const loggingStore = useLoggingStore()
      loggingStore.info('Antras duomenys išvalyti', {
        component: 'antrasStore',
        action: 'clear_data'
      })
    },

    /**
     * Export filtered records to JSON
     */
    exportToJson() {
      const data = {
        fileName: this.fileName,
        exportedAt: new Date().toISOString(),
        sheets: this.sheets,
        recordsCount: this.filteredRecords.length,
        records: this.filteredRecords
      }

      const json = JSON.stringify(data, null, 2)
      const blob = new Blob([json], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      
      const link = document.createElement('a')
      link.href = url
      link.download = `antras_export_${new Date().toISOString().split('T')[0]}.json`
      link.click()
      
      URL.revokeObjectURL(url)

      const loggingStore = useLoggingStore()
      loggingStore.info('Antras duomenys eksportuoti į JSON', {
        component: 'antrasStore',
        action: 'export_json',
        recordsCount: this.filteredRecords.length
      })
    },
  },
})
