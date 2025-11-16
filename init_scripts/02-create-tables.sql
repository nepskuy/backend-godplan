docker exec -it shared-godplan-db psql -U postgres -d defaultdb -c "
CREATE TABLE godplan.attendance_schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    start_time TIME,
    end_time TIME,
    working_days INTEGER DEFAULT 0,
    tolerance_late INTEGER DEFAULT 0,
    tolerance_early INTEGER DEFAULT 0,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);"