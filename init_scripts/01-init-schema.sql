-- Create schema
CREATE SCHEMA IF NOT EXISTS godplan;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Set default search path
ALTER DATABASE defaultdb SET search_path TO godplan, public;