-- ============================================================
-- HOSTEL MANAGEMENT SYSTEM - SQLite PL/SQL Equivalents
-- ============================================================
-- SQLite does not support stored procedures or PL/SQL packages.
-- This file provides the closest SQLite-native equivalents:
--   - Triggers  (replaces PL/SQL triggers & stored procedures)
--   - Views     (replaces PL/SQL views / cursor queries)
--   - Audit table (supports the status-change audit trigger)
--
-- Run manually:  sqlite3 hostel.db < sql/sqlite_plsql.sql
-- Loaded automatically by db.go via initPLSQL()
-- ============================================================

-- ------------------------------------------------------------
-- AUDIT TABLE
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS complaint_audit (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    complaint_id INTEGER NOT NULL,
    old_status   TEXT,
    new_status   TEXT,
    changed_at   DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ------------------------------------------------------------
-- VIEWS
-- ------------------------------------------------------------

-- View used by adminDashboard instead of a raw JOIN query
CREATE VIEW IF NOT EXISTS admin_complaint_view AS
SELECT
    c.id          AS complaint_id,
    s.name        AS student_name,
    s.roll_no,
    s.room_no,
    c.title,
    c.description,
    c.status,
    c.created_at
FROM complaints c
LEFT JOIN students s ON c.student_id = s.id
ORDER BY c.created_at DESC;

-- Summary view: per-student complaint statistics
CREATE VIEW IF NOT EXISTS student_complaint_summary AS
SELECT
    s.id   AS student_id,
    s.name,
    s.roll_no,
    s.room_no,
    COUNT(c.id)                                               AS total_complaints,
    SUM(CASE WHEN c.status = 'Pending'     THEN 1 ELSE 0 END) AS pending_count,
    SUM(CASE WHEN c.status = 'In Progress' THEN 1 ELSE 0 END) AS in_progress_count,
    SUM(CASE WHEN c.status = 'Resolved'    THEN 1 ELSE 0 END) AS resolved_count
FROM students s
LEFT JOIN complaints c ON s.id = c.student_id
GROUP BY s.id, s.name, s.roll_no, s.room_no;

-- ------------------------------------------------------------
-- TRIGGER 1 — Validate status on INSERT
-- Equivalent to: hostel_pkg.add_student validation +
--                trg_validate_insert_status (Oracle)
-- ------------------------------------------------------------
CREATE TRIGGER IF NOT EXISTS trg_validate_complaint_status
BEFORE INSERT ON complaints
FOR EACH ROW
WHEN NEW.status IS NOT NULL
  AND NEW.status NOT IN ('Pending', 'In Progress', 'Resolved')
BEGIN
    SELECT RAISE(ABORT,
        'Invalid complaint status. Allowed: Pending, In Progress, Resolved.');
END;

-- ------------------------------------------------------------
-- TRIGGER 2 — Validate status on UPDATE
-- Equivalent to: hostel_pkg.update_complaint_status validation +
--                trg_validate_update_status (Oracle)
-- ------------------------------------------------------------
CREATE TRIGGER IF NOT EXISTS trg_validate_status_update
BEFORE UPDATE OF status ON complaints
FOR EACH ROW
WHEN NEW.status NOT IN ('Pending', 'In Progress', 'Resolved')
BEGIN
    SELECT RAISE(ABORT,
        'Invalid complaint status. Allowed: Pending, In Progress, Resolved.');
END;

-- ------------------------------------------------------------
-- TRIGGER 3 — Audit log on status change
-- Equivalent to: trg_complaint_status_audit (Oracle)
-- ------------------------------------------------------------
CREATE TRIGGER IF NOT EXISTS trg_complaint_status_audit
AFTER UPDATE OF status ON complaints
FOR EACH ROW
WHEN OLD.status != NEW.status
BEGIN
    INSERT INTO complaint_audit (complaint_id, old_status, new_status, changed_at)
    VALUES (OLD.id, OLD.status, NEW.status, CURRENT_TIMESTAMP);
END;

-- ------------------------------------------------------------
-- TRIGGER 4 — Cascade-delete complaints before student delete
-- Equivalent to: hostel_pkg.delete_student +
--                trg_cascade_delete_complaints (Oracle)
-- ------------------------------------------------------------
CREATE TRIGGER IF NOT EXISTS trg_cascade_delete_complaints
BEFORE DELETE ON students
FOR EACH ROW
BEGIN
    DELETE FROM complaints WHERE student_id = OLD.id;
END;

-- ------------------------------------------------------------
-- TRIGGER 5 — Protect the primary admin account (id = 1)
-- Equivalent to: trg_protect_main_admin (Oracle)
-- ------------------------------------------------------------
CREATE TRIGGER IF NOT EXISTS trg_protect_main_admin
BEFORE DELETE ON admins
FOR EACH ROW
WHEN OLD.id = 1
BEGIN
    SELECT RAISE(ABORT, 'Cannot delete the main admin account.');
END;
