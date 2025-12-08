// src/stores/clip.js
import { defineStore } from 'pinia'
import { useLoggingStore } from './logging'
import { useSlogStore } from './slog'
import heads from '../models/headers';
import { timeToDecimal } from '../utils/helpers';
/**
 * Clip Store - the sacred vault of locomotive schedule data!
 *
 * This store manages train schedules imported from clipboard or stored in the database.
 * Like a meticulous train dispatcher keeping track of all locomotives, this store
 * maintains arrival and departure information organized by depot, date, and time range.
 */
export const useClipStore = defineStore('clip', {
  state: () => ({
    records: [],              // Full collection of records
    isLoading: false,         // Loading state tracker
    selectedDepot: null,      // Currently selected depot
    selectedDate: null,       // Currently selected date (ISO format)
    timeRange: 'all',         // Time range filter: 'day' (06:00-20:00), 'night' (18:00-08:00), 'all' (00:00-23:59)
    depotList: [],            // List of available depots
    dateList: [],             // List of available dates
    lastImported: null,       // Timestamp of last import
    stations: [],             // List of stations with tracks map
  }),

  getters: {
    availableDepots(state) {
      // If we already have a list, use it
      return state.depotList.sort();
    },

    availableDates(state) {
      return state.dateList.sort();
    },

    stationsAvailable(state) {
      return state.stations && state.stations.length > 0;
    },

    availableTracks(state) {
      if (!state.stations || state.stations.length === 0) return null;
      if (!state.selectedDepot) return null;

      const depot = state.stations.find(station => station.code === state.selectedDepot);
      if (!depot) return null;
      return depot.tracks.sort((a, b) => a.id - b.id);
    },

    getVehicles(state) {
      if (!state.selectedDepot || state.records.length === 0) return [];
      
      // Group records by vehicle for the selected station
      const vehicleGroups = {};
      
      state.records.forEach(record => {
        const vehicle = record.vehicle || record.vehicleName;
        if (!vehicle) return; // Skip records without vehicle information
        
        const start = record.startingLocation || record.departureDepot;
        const end = record.endLocation || record.arrivalDepot;
        
        // Skip records not related to the selected station
        if (start !== state.selectedDepot && end !== state.selectedDepot) return;
        
        // Initialize group for this vehicle if not exist
        if (!vehicleGroups[vehicle]) {
          vehicleGroups[vehicle] = {
            arrivals: [],
            departures: []
          };
        }
        
        // Add record to the appropriate array
        if (end === state.selectedDepot) {
          vehicleGroups[vehicle].arrivals.push(record);
        }
        if (start === state.selectedDepot) {
          vehicleGroups[vehicle].departures.push(record);
        }
      });
      
      // Create arrival-departure pairs
      const result = [];
      
      for (const vehicle in vehicleGroups) {
        const { arrivals, departures } = vehicleGroups[vehicle];
        
        // Sort by time
        arrivals.sort((a, b) => a.arrivalDecimal - b.arrivalDecimal);
        departures.sort((a, b) => a.departureDecimal - b.departureDecimal);
        
        // Create copies to avoid modifying the originals
        const arrivalsCopy = [...arrivals];
        const departuresCopy = [...departures];
        
        // Process arrivals - find matching departures
        for (const arrival of arrivalsCopy) {
          // Find the earliest departure after this arrival
          const departureIndex = departuresCopy.findIndex(dep => 
            dep.departureDecimal > arrival.arrivalDecimal
          );
          
          let departure = null;
          if (departureIndex !== -1) {
            departure = departuresCopy[departureIndex];
            // Remove used departure from the list
            departuresCopy.splice(departureIndex, 1);
          }
          
          // Add pair (arrival, departure)
          result.push({
            vehicle,
            arrival,
            departure
          });
        }
        
        // Process remaining departures (without arrivals)
        for (const departure of departuresCopy) {
          result.push({
            vehicle,
            arrival: null,
            departure
          });
        }
      }
      
      // Sort result by arrival/departure time
      result.sort((a, b) => {
        const timeA = a.arrival ? a.arrival.arrivalDecimal : a.departure.departureDecimal;
        const timeB = b.arrival ? b.arrival.arrivalDecimal : b.departure.departureDecimal;
        return timeA - timeB;
      });
      
      return result;
    },

    filteredRecords(state) {
      if (!state.selectedDepot || !state.selectedDate) return null;

      const intimeArrivals = [],
        intimeDepartures = [],
        arrivals = [],
        departures = []
      state.records.forEach(record => {
        // Check depot match
        const start = record.startingLocation || record.departureDepot;
        const end = record.endLocation || record.arrivalDepot;
        if (end === state.selectedDepot) {
          arrivals.push(record);
          if (isTimeInRange(record.arrivalDecimal, state.timeRange, state.selectedDate)) {
            intimeArrivals.push(record);
          }
        }
        if (start === state.selectedDepot) {
          departures.push(record);
          if (isTimeInRange(record.departureDecimal, state.timeRange, state.selectedDate)) {
            intimeDepartures.push(record);
          }
        }
      });
      return {
        intimeArrivals: intimeArrivals.sort((a, b) => a.arrivalDecimal - b.arrivalDecimal),
        intimeDepartures: intimeDepartures.sort((a, b) => a.departureDecimal - b.departureDecimal),
        arrivals: arrivals.sort((a, b) => a.arrivalDecimal - b.arrivalDecimal),
        departures: departures.sort((a, b) => a.departureDecimal - b.departureDecimal),
      }
    },

  },
  actions: {
    setFilters({ depot, date, timeRange }) {
      if (depot) this.selectedDepot = depot;
      if (date) this.selectedDate = date;
      if (timeRange) this.timeRange = timeRange;

      // Log filter change
      const loggingStore = useLoggingStore();
      loggingStore.info('Keisti filtrai', {
        component: 'Klasika',
        depot: this.selectedDepot,
        date: this.selectedDate,
        timeRange: this.timeRange
      });
    },

// Modified importFromText method for clip.js store
// This version preserves track assignments (targetTrack, startingTrack) 
// across clipboard imports by saving them to localStorage

importFromText(text) {
  if (!text || typeof text !== 'string') return 0;

  const loggingStore = useLoggingStore();
  const slogStore = useSlogStore();

  try {
    // STEP 1: Save track assignments to localStorage before clearing
    const trackAssignments = {};
    this.records.forEach(record => {
      if (record.targetTrack) {
        trackAssignments[record.id] = {
          targetTrack: record.targetTrack || null,
        };
      }
    });
    
    // Save to localStorage
    if (Object.keys(trackAssignments).length > 0) {
      try {
        localStorage.setItem('yopta_track_assignments', JSON.stringify(trackAssignments));
        loggingStore.info('Kelio priskyrimai išsaugoti', {
          component: 'clipStore',
          action: 'save_track_assignments',
          count: Object.keys(trackAssignments).length
        });
      } catch (e) {
        console.error('Klaida išsaugant kelio priskyrimus:', e);
        loggingStore.error('Nepavyko išsaugoti kelio priskyrimų', {
          component: 'clipStore',
          error: e.message
        });
      }
    }

    // STEP 2: Clear records completely
    this.records = [];
    
    loggingStore.info('Įrašai išvalyti prieš importą', {
      component: 'clipStore',
      action: 'clear_records'
    });

    // STEP 3: Parse and import new data
    const data = parseTabData(text);
    const rows = data.records;
    this.depotList = data.depotList;
    this.dateList = data.dateList;

    if (rows.length === 0) {
      slogStore.addToast({
        message: 'Nepavyko importuoti: netinkamas duomenų formatas',
        type: 'alert-warning'
      });
      return 0;
    }

    // Add new records
    this.addRecords(rows);

    // STEP 4: Restore track assignments from localStorage
    if (Object.keys(trackAssignments).length > 0) {
      let restoredArrivals = 0;
      let restoredDepartures = 0;
      
      this.records.forEach(record => {
        if (trackAssignments[record.id]) {
          // Restore targetTrack for arrivals
          if (trackAssignments[record.id].targetTrack) {
            record.targetTrack = trackAssignments[record.id].targetTrack;
            restoredArrivals++;
          }
        }
      });
      
      if (restoredArrivals > 0 || restoredDepartures > 0) {
        loggingStore.info('Kelio priskyrimai atkurti', {
          component: 'clipStore',
          action: 'restore_track_assignments',
          arrivals: restoredArrivals,
          departures: restoredDepartures,
          total: restoredArrivals + restoredDepartures
        });
        
        slogStore.addToast({
          message: `Atkurti ${restoredArrivals + restoredDepartures} kelio priskyrimai`,
          type: 'alert-info'
        });
      }
    }

    // Update last import timestamp
    this.lastImported = new Date().toISOString();

    // Log successful import
    loggingStore.info('Duomenys sėkmingai importuoti iš iškarpinės', {
      component: 'clipStore',
      action: 'clipboard_import',
      recordCount: rows.length
    });

    return rows.length;
  } catch (error) {
    console.error('Import error:', error);
    loggingStore.error('Klaida importuojant duomenis', {
      component: 'clipStore',
      action: 'clipboard_import_error',
      error: error.message
    });

    slogStore.addToast({
      message: `Importavimo klaida: ${error.message}`,
      type: 'alert-error'
    });

    return 0;
  }
},
    addRecords(newRecords) {
      if (!Array.isArray(newRecords) || newRecords.length === 0) return;

      newRecords.forEach(newRecord => {
        this.addRecord(newRecord);
      });

    },

    addRecord(record) {
      if (!record || typeof record !== 'object') return false;

      // Generate a unique ID if not present
      if (!record.id) {
        record.id = generateUniqueId(record);
      }

      // Check if the record already exists
      const existingIndex = this.records.findIndex(r => r.id === record.id);

      if (existingIndex >= 0) {
        // Update existing record
        this.records[existingIndex] = {
          ...this.records[existingIndex],
          ...record,
          updatedAt: new Date().toISOString()
        };
      } else {
        // Add new record
        this.records.push({
          ...record,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString()
        });
      }

      return true;
    },

    updateRecord(id, changes) {
      const index = this.records.findIndex(r => r.id === id);

      if (index === -1) return false;

      // Update specified fields
      this.records[index] = {
        ...this.records[index],
        ...changes,
        updatedAt: new Date().toISOString()
      };

      // Log the update
      const loggingStore = useLoggingStore();
      loggingStore.info('Įrašas atnaujintas', {
        action: 'update_record',
        recordId: id,
        changes: Object.keys(changes)
      });

      return true;
    },

    deleteRecord(id) {
      const index = this.records.findIndex(r => r.id === id);
      console.log('Deleting record:', id, index, this.records[index].arrivalTrainNumber);
      if (index === -1) return false;

      // Remove the record from the array
      this.records.splice(index, 1);

      // Log the deletion
      const loggingStore = useLoggingStore();
      loggingStore.info('Įrašas ištrintas', {
        action: 'delete_record',
        recordId: id
      });

      return true;
    },

    setStations(stations) {
      if (!Array.isArray(stations)) return false;

      this.stations = stations;
      // Log the station storage
      const loggingStore = useLoggingStore();
      loggingStore.info('Stotys su keliais įkrautos', {
        action: 'store_stations',
        stationCount: stations.length
      });

      return true;
    }
  }
});

function parseTabData(text, depotSource, dateSource) {
  const lines = text.split('\n');
  if (lines && lines.length < 2) return [];

  // Extract headers from first line
  const headers = lines[0].trim('\r').split('\t');

  // Check if all required fields are present
  const hasAllHeaders = Object.keys(heads).every(field => headers.includes(field));
  if (!hasAllHeaders) {
    console.error('Missing required headers.');
    return [];
  }

  // Process data rows
  const records = [];
  const depots = new Set(depotSource || []);
  const dates = new Set(dateSource || []);


  for (let i = 1; i < lines.length; i++) {
    if (!lines[i].trim()) continue;

    const values = lines[i].split('\t');
    if (values.length !== headers.length) continue;

    // Create record object
    const record = {};

    // Populate fields from headers and values
    for (let j = 0; j < headers.length; j++) {
      const header = headers[j].trim();
      const value = values[j].trim();
      if (heads[header]) record[heads[header]] = value;
    }

    try {
      // Process departure date/time
      const departureDate = record.departureDate || record.date;
      const departureTime = record.departurePlanned;
      if (departureDate && departureTime) {
        record.departureDateTime = new Date(`${departureDate}T${departureTime}Z`);
      }

      // Process arrival date/time
      const arrivalDate = record.arrivalDate || record.date;
      const arrivalTime = record.arrivalPlanned;
      if (arrivalDate && arrivalTime) {
        record.arrivalDateTime = new Date(`${arrivalDate}T${arrivalTime}Z`);
      }

      // Add the parsed record if dates are valid
      if (!isNaN(record.departureDateTime) || !isNaN(record.arrivalDateTime)) {
        if (record.arrivalDateTime) {
          dates.add(record.arrivalDate);
          record.arrivalDecimal = timeToDecimal(record.arrivalDateTime);
        }
        if (record.departureDateTime) {
          dates.add(record.departureDate);
          record.departureDecimal = timeToDecimal(record.departureDateTime);
        }
        if (record.startingLocation) depots.add(record.startingLocation);
        if (record.endLocation) depots.add(record.endLocation);
        records.push(record);
      }
    } catch (e) {
      console.error('Error parsing record:', e, record);
      // Skip this record if parsing fails
    }
  }

  console.log('Parsed records:', records);

  const depotList = Array.from(depots).filter(depot => depot);
  const dateList = Array.from(dates).sort();
  return {records, depotList, dateList};
}

function isTimeInRange(decimalTime, range, selectedDate) {
  if (!decimalTime || !range || !selectedDate) return false;

  let selectedDecimal = timeToDecimal(new Date(selectedDate));

  switch (range) {
    case 'day':
      return decimalTime >= selectedDecimal + (6 * 60)
        && decimalTime < selectedDecimal + (20 * 60); // 06:00 - 20:00
    case 'night':
      return decimalTime >= selectedDecimal + (18 * 60)
        && decimalTime < selectedDecimal + (32 * 60); // 18:00 - 08:00 next day
    case 'all':
      return decimalTime >= selectedDecimal
        && decimalTime < selectedDecimal + (24 * 60); // 00:00 - 23:59
    default:
      return false;
  }
}

function early(decimalTime, range, selectedDate) {
  if (!decimalTime || !range || !selectedDate) return false;

  let selectedDecimal = timeToDecimal(new Date(selectedDate));
  switch (range) {
    case 'day':
      return decimalTime < selectedDecimal + (6 * 60); // 06:00 - 20:00
    case 'night':
      return decimalTime < selectedDecimal + (18 * 60); // 18:00 - 08:00 next day
    case 'all':
      return decimalTime < selectedDecimal; // 00:00 - 23:59
    default:
      return false;
  }
}

function late(decimalTime, range, selectedDate) {
  if (!decimalTime || !range || !selectedDate) return false;

  let selectedDecimal = timeToDecimal(new Date(selectedDate));

  switch (range) {
    case 'day':
      return decimalTime > selectedDecimal + (20 * 60); // 06:00 - 20:00
    case 'night':
      return decimalTime > selectedDecimal + (32 * 60); // 18:00 - 08:00 next day
    case 'all':
      return decimalTime > selectedDecimal + (24 * 60); // 00:00 - 23:59
    default:
      return false;
  }
}

function generateUniqueId(record) {
  let str = [
    record.vehicleWorkingDesignation || '',
    record.departureDate || record.date || '',
    record.departureTripNumber || record.departureTrainNumber || record.departureNetworkTrainNumber || '',
  ];
  str = str.join('').split('-').join('')
  return str;
}
