-- +goose Up
-- Migration to create antras_field_mappings table for Antras page field name management
-- This table stores the mapping for train movement data fields (37 columns from the Excel export)

CREATE TABLE antras_field_mappings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    external_name VARCHAR(255) NOT NULL COMMENT 'Field name from external system (e.g., "Network point name")',
    internal_name VARCHAR(100) NOT NULL COMMENT 'Internal field name used in application (e.g., "networkPointName")',
    display_name VARCHAR(255) NOT NULL COMMENT 'Human-readable display name (Lithuanian)',
    field_type ENUM('string', 'date', 'time', 'datetime', 'number', 'boolean') DEFAULT 'string' COMMENT 'Data type of the field',
    is_required BOOLEAN DEFAULT FALSE COMMENT 'Whether this field is required for import',
    sort_order INT DEFAULT 0 COMMENT 'Display order in UI (matches column order in Excel)',
    description TEXT COMMENT 'Field description for documentation',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY idx_external_name (external_name),
    UNIQUE KEY idx_internal_name (internal_name),
    KEY idx_sort_order (sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Stores field mapping configuration for Antras page train movement data import';

-- Insert default mappings based on Excel file structure (37 columns)
INSERT INTO antras_field_mappings (external_name, internal_name, display_name, field_type, is_required, sort_order, description) VALUES
-- Column A
('Network point name', 'networkPointName', 'Tinklo punkto pavadinimas', 'string', TRUE, 1, 'Station or network point name'),

-- Column B-C: Technical vehicle types
('Technical vehicle type.in', 'technicalVehicleTypeIn', 'Techninis riedmens tipas (atvykimas)', 'string', FALSE, 2, 'Technical vehicle type for incoming train'),
('Technical vehicle type.out', 'technicalVehicleTypeOut', 'Techninis riedmens tipas (išvykimas)', 'string', FALSE, 3, 'Technical vehicle type for outgoing train'),

-- Column D-E: Vehicle numbers
('Vehicle no.in', 'vehicleNoIn', 'Riedmens numeris (atvykimas)', 'string', FALSE, 4, 'Vehicle number for incoming train'),
('Vehicle no.out', 'vehicleNoOut', 'Riedmens numeris (išvykimas)', 'string', FALSE, 5, 'Vehicle number for outgoing train'),

-- Column F-G: Train numbers
('Train No.in', 'trainNoIn', 'Traukinio numeris (atvykimas)', 'string', FALSE, 6, 'Train number for incoming train'),
('Train No.out', 'trainNoOut', 'Traukinio numeris (išvykimas)', 'string', FALSE, 7, 'Train number for outgoing train'),

-- Column H-I: Arrival/Departure times
('Arrival', 'arrival', 'Atvykimas', 'time', TRUE, 8, 'Arrival time at the station'),
('Departure', 'departure', 'Išvykimas', 'time', TRUE, 9, 'Departure time from the station'),

-- Column J-K: Duty assignments
('Duty.in', 'dutyIn', 'Pareigos (atvykimas)', 'string', FALSE, 10, 'Duty assignment for incoming train crew'),
('Duty.out', 'dutyOut', 'Pareigos (išvykimas)', 'string', FALSE, 11, 'Duty assignment for outgoing train crew'),

-- Column L-O: Driver information
('Driver.in', 'driverIn', 'Mašinistas (atvykimas)', 'string', FALSE, 12, 'Driver name(s) for incoming train'),
('Phone.in', 'phoneIn', 'Telefonas (atvykimas)', 'string', FALSE, 13, 'Driver phone number(s) for incoming train'),
('Driver.out', 'driverOut', 'Mašinistas (išvykimas)', 'string', FALSE, 14, 'Driver name(s) for outgoing train'),
('Phone.out', 'phoneOut', 'Telefonas (išvykimas)', 'string', FALSE, 15, 'Driver phone number(s) for outgoing train'),

-- Column P-Q: Personnel numbers
('Driver.PersonnelNumber.in', 'driverPersonnelNumberIn', 'Mašinisto personalo numeris (atvykimas)', 'string', FALSE, 16, 'Driver personnel number for incoming train'),
('Driver.PersonnelNumber.out', 'driverPersonnelNumberOut', 'Mašinisto personalo numeris (išvykimas)', 'string', FALSE, 17, 'Driver personnel number for outgoing train'),

-- Column R-U: Duty start/end times
('Duty.StartingTime.in', 'dutyStartingTimeIn', 'Pareigų pradžia (atvykimas)', 'time', FALSE, 18, 'Duty starting time for incoming train crew'),
('Duty.StartingTime.out', 'dutyStartingTimeOut', 'Pareigų pradžia (išvykimas)', 'time', FALSE, 19, 'Duty starting time for outgoing train crew'),
('Duty.EndTime.in', 'dutyEndTimeIn', 'Pareigų pabaiga (atvykimas)', 'time', FALSE, 20, 'Duty ending time for incoming train crew'),
('Duty.EndTime.out', 'dutyEndTimeOut', 'Pareigų pabaiga (išvykimas)', 'time', FALSE, 21, 'Duty ending time for outgoing train crew'),

-- Column V-W: Validity dates
('Validity.in', 'validityIn', 'Galiojimas (atvykimas)', 'date', FALSE, 22, 'Validity date for incoming train schedule'),
('Validity.out', 'validityOut', 'Galiojimas (išvykimas)', 'date', FALSE, 23, 'Validity date for outgoing train schedule'),

-- Column X-AE: Starting/ending locations and times
('Starting location.in', 'startingLocationIn', 'Pradžios vieta (atvykimas)', 'string', FALSE, 24, 'Starting location for incoming train journey'),
('Starting location.out', 'startingLocationOut', 'Pradžios vieta (išvykimas)', 'string', FALSE, 25, 'Starting location for outgoing train journey'),
('Starting time.in', 'startingTimeIn', 'Pradžios laikas (atvykimas)', 'time', FALSE, 26, 'Starting time for incoming train journey'),
('Starting time.out', 'startingTimeOut', 'Pradžios laikas (išvykimas)', 'time', FALSE, 27, 'Starting time for outgoing train journey'),
('End location.in', 'endLocationIn', 'Pabaigos vieta (atvykimas)', 'string', FALSE, 28, 'Ending location for incoming train journey'),
('End location.out', 'endLocationOut', 'Pabaigos vieta (išvykimas)', 'string', FALSE, 29, 'Ending location for outgoing train journey'),
('Ending time.in', 'endingTimeIn', 'Pabaigos laikas (atvykimas)', 'time', FALSE, 30, 'Ending time for incoming train journey'),
('Ending time.out', 'endingTimeOut', 'Pabaigos laikas (išvykimas)', 'time', FALSE, 31, 'Ending time for outgoing train journey'),

-- Column AF-AG: Vehicle working numbers
('Vehicle working.in', 'vehicleWorkingIn', 'Riedmens darbo numeris (atvykimas)', 'string', FALSE, 32, 'Vehicle working number for incoming train'),
('Vehicle working.out', 'vehicleWorkingOut', 'Riedmens darbo numeris (išvykimas)', 'string', FALSE, 33, 'Vehicle working number for outgoing train'),

-- Column AH-AI: Registration numbers
('Vehicle reg. no.in', 'vehicleRegNoIn', 'Riedmens registracijos nr. (atvykimas)', 'string', FALSE, 34, 'Vehicle registration number for incoming train'),
('Vehicle reg. no.out', 'vehicleRegNoOut', 'Riedmens registracijos nr. (išvykimas)', 'string', FALSE, 35, 'Vehicle registration number for outgoing train'),

-- Column AJ-AK: Train lengths
('Train length.in', 'trainLengthIn', 'Traukinio ilgis (atvykimas)', 'number', FALSE, 36, 'Train length in meters for incoming train'),
('Train length.out', 'trainLengthOut', 'Traukinio ilgis (išvykimas)', 'number', FALSE, 37, 'Train length in meters for outgoing train');

-- +goose Down
DROP TABLE IF EXISTS antras_field_mappings;
