-- +goose Up
-- Migration to create field_mappings table for dynamic field name management
-- This table stores the mapping between external field names (from import source)
-- and internal field names used in the application

CREATE TABLE field_mappings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    external_name VARCHAR(255) NOT NULL COMMENT 'Field name from external system (e.g., "Vehicle working number")',
    internal_name VARCHAR(100) NOT NULL COMMENT 'Internal field name used in application (e.g., "vehicleWorkingNumber")',
    display_name VARCHAR(255) NOT NULL COMMENT 'Human-readable display name (Lithuanian)',
    field_type ENUM('string', 'date', 'time', 'number', 'boolean') DEFAULT 'string' COMMENT 'Data type of the field',
    is_required BOOLEAN DEFAULT FALSE COMMENT 'Whether this field is required for import',
    sort_order INT DEFAULT 0 COMMENT 'Display order in UI',
    description TEXT COMMENT 'Field description for documentation',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY idx_external_name (external_name),
    UNIQUE KEY idx_internal_name (internal_name),
    KEY idx_sort_order (sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Stores field mapping configuration for data import';

-- Insert default mappings from current headers.js
INSERT INTO field_mappings (external_name, internal_name, display_name, field_type, is_required, sort_order) VALUES
('Vehicle working number', 'vehicleWorkingNumber', 'Riedmens darbo numeris', 'string', TRUE, 1),
('Vehicle working designation', 'vehicleWorkingDesignation', 'Riedmens darbo žymėjimas', 'string', TRUE, 2),
('Date', 'date', 'Data', 'date', TRUE, 3),
('Date (planned arrival)', 'arrivalDate', 'Atvykimo data (planuojama)', 'date', FALSE, 4),
('Planned arrival', 'arrivalPlanned', 'Planuojamas atvykimas', 'time', FALSE, 5),
('Actual arrival', 'arrivalActual', 'Faktinis atvykimas', 'time', FALSE, 6),
('Arrival time last revenue trip', 'arrivalTimeLastRevenueTrip', 'Paskutinio reiso atvykimo laikas', 'time', FALSE, 7),
('Depot arrival', 'arrivalDepot', 'Atvykimo depas', 'string', FALSE, 8),
('Trip number (arrival)', 'arrivalTripNumber', 'Reiso numeris (atvykimas)', 'string', FALSE, 9),
('Train reference (arrival)', 'arrivalTrainNumber', 'Traukinio numeris (atvykimas)', 'string', FALSE, 10),
('Network train number (arrival)', 'arrivalNetworkTrainNumber', 'Tinklo traukinio numeris (atvykimas)', 'string', FALSE, 11),
('Route (arrival)', 'arrivalRoute', 'Maršrutas (atvykimas)', 'string', FALSE, 12),
('Starting location', 'startingLocation', 'Pradinė vieta', 'string', FALSE, 13),
('Date (planned departure)', 'departureDate', 'Išvykimo data (planuojama)', 'date', FALSE, 14),
('Planned departure', 'departurePlanned', 'Planuojamas išvykimas', 'time', FALSE, 15),
('Actual departure', 'departureActual', 'Faktinis išvykimas', 'time', FALSE, 16),
('Departure time first revenue trip', 'departureTimeFirstRevenueTrip', 'Pirmo reiso išvykimo laikas', 'time', FALSE, 17),
('Depot departure', 'departureDepot', 'Išvykimo depas', 'string', FALSE, 18),
('Trip number (departure)', 'departureTripNumber', 'Reiso numeris (išvykimas)', 'string', FALSE, 19),
('Train reference (departure)', 'departureTrainNumber', 'Traukinio numeris (išvykimas)', 'string', FALSE, 20),
('Network train number (departure)', 'departureNetworkTrainNumber', 'Tinklo traukinio numeris (išvykimas)', 'string', FALSE, 21),
('End location', 'endLocation', 'Galutinė vieta', 'string', FALSE, 22),
('Home depot', 'homeDepot', 'Namų depas', 'string', FALSE, 23),
('Route', 'route', 'Maršrutas', 'string', FALSE, 24),
('Route (departure)', 'departureRoute', 'Maršrutas (išvykimas)', 'string', FALSE, 25),
('Route info', 'routeInfo', 'Maršruto informacija', 'string', FALSE, 26),
('Route (long name)', 'routeLongName', 'Maršrutas (pilnas pavadinimas)', 'string', FALSE, 27),
('Planned vehicle type', 'plannedVehicleType', 'Planuojamas riedmens tipas', 'string', FALSE, 28),
('Vehicle', 'vehicleName', 'Riedmuo', 'string', FALSE, 29),
('Name of the vehicle group', 'vehicle', 'Riedmens grupės pavadinimas', 'string', FALSE, 30),
('Vehicle working km', 'vehicleWorkingKm', 'Riedmens darbo km', 'number', FALSE, 31),
('Vehicle working series km', 'vehicleWorkingSeriesKm', 'Riedmens darbo serijos km', 'number', FALSE, 32),
('Start track', 'startingTrack', 'Pradinis kelias', 'string', FALSE, 33),
('Destination track', 'targetTrack', 'Tikslo kelias', 'string', FALSE, 34),
('Start track section', 'startTrackSection', 'Pradinio kelio sekcija', 'string', FALSE, 35),
('Destination track section', 'destinationTrackSection', 'Tikslo kelio sekcija', 'string', FALSE, 36),
('External vehicle number', 'externalVehicleNumber', 'Išorinis riedmens numeris', 'string', FALSE, 37),
('Registration number', 'registrationNumber', 'Registracijos numeris', 'string', FALSE, 38),
('Duty (departure)', 'departureDuty', 'Pareiga (išvykimas)', 'string', FALSE, 39),
('Duty (arrival)', 'arrivalDuty', 'Pareiga (atvykimas)', 'string', FALSE, 40),
('Description', 'description', 'Aprašymas', 'string', FALSE, 41),
('Duration', 'duration', 'Trukmė', 'string', FALSE, 42),
('Vehicle working vehicle type', 'vehicleWorkingVehicleType', 'Riedmens darbo tipas', 'string', FALSE, 43),
('Vehicle working vehicle type (description)', 'vehicleWorkingVehicleTypeDescription', 'Riedmens darbo tipo aprašymas', 'string', FALSE, 44),
('Actual vehicle type', 'actualVehicleType', 'Faktinis riedmens tipas', 'string', FALSE, 45),
('Employees (all)', 'employeesAll', 'Visi darbuotojai', 'string', FALSE, 46),
('Duty (all)', 'dutyAll', 'Visos pareigos', 'string', FALSE, 47),
('Employee 1 departure', 'departureEmployee1', 'Darbuotojas 1 (išvykimas)', 'string', FALSE, 48),
('Employee 2 departure', 'departureEmployee2', 'Darbuotojas 2 (išvykimas)', 'string', FALSE, 49),
('Employee 3 departure', 'departureEmployee3', 'Darbuotojas 3 (išvykimas)', 'string', FALSE, 50),
('Employee 1 arrival', 'arrivalEmployee1', 'Darbuotojas 1 (atvykimas)', 'string', FALSE, 51),
('Employee 2 arrival', 'arrivalEmployee2', 'Darbuotojas 2 (atvykimas)', 'string', FALSE, 52),
('Employee 3 arrival', 'arrivalEmployee3', 'Darbuotojas 3 (atvykimas)', 'string', FALSE, 53),
('Synchronisation status', 'synchronisationStatus', 'Sinchronizacijos būsena', 'string', FALSE, 54);

-- +goose Down
DROP TABLE IF EXISTS field_mappings;
