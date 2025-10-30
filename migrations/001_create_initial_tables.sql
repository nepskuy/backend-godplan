-- Tabel users (jika belum ada)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'employee',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabel tasks (jika belum ada)
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    user_id INTEGER REFERENCES users(id),
    assigned_to INTEGER REFERENCES users(id),
    status VARCHAR(50) DEFAULT 'pending',
    completed BOOLEAN DEFAULT false,
    deadline TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabel office_locations (BARU - untuk validasi lokasi)
CREATE TABLE IF NOT EXISTS office_locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    radius INTEGER NOT NULL,
    address TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabel attendances (BARU - struktur sesuai UML)
CREATE TABLE IF NOT EXISTS attendances (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    type VARCHAR(10) NOT NULL CHECK (type IN ('in', 'out')),
    status VARCHAR(20) DEFAULT 'pending',
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    photo_selfie TEXT,
    in_range BOOLEAN DEFAULT false,
    force_attendance BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default office location (sesuaikan dengan lokasi kantor Anda)
INSERT INTO office_locations (name, latitude, longitude, radius, address, is_active) 
VALUES ('Kantor Pusat', -6.200000, 106.816666, 100, 'Jl. Sudirman No. 1', true)
ON CONFLICT DO NOTHING;

-- Insert sample user untuk testing
INSERT INTO users (username, name, email, password, role) 
VALUES 
    ('admin', 'Administrator', 'admin@godplan.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin'),
    ('karyawan1', 'Karyawan Satu', 'karyawan1@godplan.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'employee')
ON CONFLICT (email) DO NOTHING;