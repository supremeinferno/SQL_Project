-- =========================
-- Enable Foreign Keys
-- =========================
PRAGMA foreign_keys = ON;

-- =========================
-- ADMINS (ONLY ONE ADMIN)
-- =========================
CREATE TABLE IF NOT EXISTS admins (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL
);

-- Insert default admin (safe if run multiple times)
INSERT OR IGNORE INTO admins (id, username, password)
VALUES (1, 'admin', '123');

-- =========================
-- STUDENTS (ID PERMANENT)
-- =========================
CREATE TABLE IF NOT EXISTS students (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    roll_no TEXT,
    room_no TEXT,
    username TEXT UNIQUE,
    password TEXT
);

-- =========================
-- COMPLAINTS (LINKED TO STUDENT ID)
-- =========================
CREATE TABLE IF NOT EXISTS complaints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    status TEXT DEFAULT 'Pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (student_id)
        REFERENCES students(id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
);
