-- backend/core/database/migrations/001_initial_schema.sql

-- Enable foreign key support
PRAGMA foreign_keys = ON;

-- Reports table
CREATE TABLE IF NOT EXISTS reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Folders table with self-referential relationship
CREATE TABLE IF NOT EXISTS folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    parent_id INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES folders(id) ON DELETE CASCADE
);

-- Files table
CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    size INTEGER NOT NULL,
    blurhash TEXT,  -- nullable, only for images
    folder_id INTEGER NOT NULL,
    mime_type TEXT NOT NULL,
    offset INTEGER NOT NULL,
    length INTEGER NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT 0, -- Deletion Marker
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (folder_id) REFERENCES folders(id)
);

-- Free spaces table for TVault management
CREATE TABLE IF NOT EXISTS free_spaces (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    offset INTEGER NOT NULL,  -- Start position in TVault
    length INTEGER NOT NULL,  -- Available space length
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Junction table for reports and files
CREATE TABLE IF NOT EXISTS report_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    report_id INTEGER NOT NULL,
    file_id INTEGER NOT NULL,
    FOREIGN KEY (report_id) REFERENCES reports(id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

-- Temporary files tracking table (for unlocked files)
CREATE TABLE IF NOT EXISTS temp_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL,
    temp_path TEXT NOT NULL,  -- Path to temporary decrypted file
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

-- Indexes for performance optimization
CREATE INDEX IF NOT EXISTS idx_folders_parent ON folders(parent_id);
CREATE INDEX IF NOT EXISTS idx_files_folder ON files(folder_id);
CREATE INDEX IF NOT EXISTS idx_files_uuid ON files(uuid);
CREATE INDEX IF NOT EXISTS idx_files_deleted ON files(is_deleted);
CREATE INDEX IF NOT EXISTS idx_report_files_report ON report_files(report_id);
CREATE INDEX IF NOT EXISTS idx_report_files_file ON report_files(file_id);
CREATE INDEX IF NOT EXISTS idx_free_spaces_length ON free_spaces(length);  -- For finding suitable free spaces

-- Triggers to update timestamps
CREATE TRIGGER IF NOT EXISTS reports_update_timestamp 
AFTER UPDATE ON reports
BEGIN
    UPDATE reports SET updated_at = CURRENT_TIMESTAMP 
    WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS folders_update_timestamp 
AFTER UPDATE ON folders
BEGIN
    UPDATE folders SET updated_at = CURRENT_TIMESTAMP 
    WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS files_update_timestamp 
AFTER UPDATE ON files
BEGIN
    UPDATE files SET updated_at = CURRENT_TIMESTAMP 
    WHERE id = NEW.id;
END;