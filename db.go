package main

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func connectDB() {
	var err error
	db, err = sql.Open("sqlite3", "./hostel.db")
	if err != nil {
		panic(err)
	}
	initPLSQL()
}

// initPLSQL creates the SQLite equivalents of PL/SQL constructs:
// an audit table, two views, and five triggers. Each statement is
// executed individually so multi-statement trigger bodies work correctly.
func initPLSQL() {
	stmts := []string{
		// Audit table
		`CREATE TABLE IF NOT EXISTS complaint_audit (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			complaint_id INTEGER NOT NULL,
			old_status   TEXT,
			new_status   TEXT,
			changed_at   DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// View: admin complaint list (replaces the raw JOIN in adminDashboard)
		`CREATE VIEW IF NOT EXISTS admin_complaint_view AS
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
		ORDER BY c.created_at DESC`,

		// View: per-student complaint statistics
		`CREATE VIEW IF NOT EXISTS student_complaint_summary AS
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
		GROUP BY s.id, s.name, s.roll_no, s.room_no`,

		// Trigger 1: reject invalid status on INSERT
		`CREATE TRIGGER IF NOT EXISTS trg_validate_complaint_status
		BEFORE INSERT ON complaints
		FOR EACH ROW
		WHEN NEW.status IS NOT NULL
		  AND NEW.status NOT IN ('Pending', 'In Progress', 'Resolved')
		BEGIN
			SELECT RAISE(ABORT,
				'Invalid complaint status. Allowed: Pending, In Progress, Resolved.');
		END`,

		// Trigger 2: reject invalid status on UPDATE
		`CREATE TRIGGER IF NOT EXISTS trg_validate_status_update
		BEFORE UPDATE OF status ON complaints
		FOR EACH ROW
		WHEN NEW.status NOT IN ('Pending', 'In Progress', 'Resolved')
		BEGIN
			SELECT RAISE(ABORT,
				'Invalid complaint status. Allowed: Pending, In Progress, Resolved.');
		END`,

		// Trigger 3: write to audit log on status change
		`CREATE TRIGGER IF NOT EXISTS trg_complaint_status_audit
		AFTER UPDATE OF status ON complaints
		FOR EACH ROW
		WHEN OLD.status != NEW.status
		BEGIN
			INSERT INTO complaint_audit (complaint_id, old_status, new_status, changed_at)
			VALUES (OLD.id, OLD.status, NEW.status, CURRENT_TIMESTAMP);
		END`,

		// Trigger 4: cascade-delete complaints before deleting a student
		`CREATE TRIGGER IF NOT EXISTS trg_cascade_delete_complaints
		BEFORE DELETE ON students
		FOR EACH ROW
		BEGIN
			DELETE FROM complaints WHERE student_id = OLD.id;
		END`,

		// Trigger 5: prevent deletion of the primary admin account
		`CREATE TRIGGER IF NOT EXISTS trg_protect_main_admin
		BEFORE DELETE ON admins
		FOR EACH ROW
		WHEN OLD.id = 1
		BEGIN
			SELECT RAISE(ABORT, 'Cannot delete the main admin account.');
		END`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			log.Printf("PL/SQL init: %v", err)
		}
	}
}
