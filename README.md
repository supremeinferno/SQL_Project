# Hostel Management System (Go + SQLite)

A hostel management web app with **Admin** and **Student** logins, featuring a full **PL/SQL implementation** (triggers, views, stored procedures, cursors, and exception handling) alongside the Go backend.

## Features

- **Authentication** — separate admin and student logins
- **Student dashboard** — submit, view, update status, and delete own complaints
- **Admin dashboard**
  - Register / delete students
  - View and manage all complaints (via SQL view)
  - Search students by roll number
  - Export students to CSV
  - **Audit Log** — every complaint status change is recorded automatically by a trigger
- **PL/SQL layer**
  - 5 SQLite triggers (validation, audit trail, cascade delete, admin protection)
  - 2 SQL views (`admin_complaint_view`, `student_complaint_summary`)
  - Oracle PL/SQL reference (package, stored procedures, functions, cursors, exceptions)

## Tech Stack

| Layer    | Technology                         |
|----------|------------------------------------|
| Language | Go                                 |
| Database | SQLite (via `mattn/go-sqlite3`)    |
| Frontend | HTML templates + Bootstrap 5       |

## Project Structure

```text
.
├── main.go                    # routes + server startup
├── handlers.go                # HTTP handlers (uses views & triggers)
├── db.go                      # SQLite connection + PL/SQL init
├── hostel.db                  # SQLite database file (local only)
├── go.mod / go.sum
├── sql/
│   ├── schema.sql             # base table definitions + seed data
│   ├── schema.dbml            # DBML diagram source
│   ├── plsql.sql              # Oracle PL/SQL reference implementation
│   │                            (package, procedures, functions, cursors)
│   └── sqlite_plsql.sql       # SQLite-compatible equivalents
│                                (triggers + views — also loaded by db.go)
├── archive/                   # originals before PL/SQL was added
│   ├── db_original.go.bak
│   ├── handlers_original.go.bak
│   └── schema_original.sql
├── templates/
│   ├── base.html
│   ├── login.html
│   ├── admin.html
│   ├── student.html
│   └── audit.html             # audit log page (PL/SQL demo)
└── static/                    # CSS assets
```

## PL/SQL Implementation

### Oracle PL/SQL (`sql/plsql.sql`)

Full Oracle-style implementation for academic reference:

| Construct | Details |
|-----------|---------|
| **Package** | `hostel_pkg` — specification + body |
| **Procedures** | `add_student`, `update_complaint_status`, `delete_student`, `list_pending_complaints` |
| **Function** | `get_complaint_count(student_id)` |
| **Cursor** | Explicit cursor in `list_pending_complaints` |
| **Exception handling** | Named exceptions, `NO_DATA_FOUND`, `OTHERS`, `ROLLBACK` |
| **Triggers** | 5 triggers (audit, cascade delete, validation, admin protection) |
| **Views** | `admin_complaint_view`, `student_complaint_summary` |

### SQLite Equivalents (`sql/sqlite_plsql.sql`)

Loaded automatically at server startup by `initPLSQL()` in `db.go`:

| Name | Type | Purpose |
|------|------|---------|
| `complaint_audit` | Table | Stores audit trail entries |
| `admin_complaint_view` | View | Complaints joined with student info (used in admin dashboard) |
| `student_complaint_summary` | View | Per-student complaint statistics |
| `trg_validate_complaint_status` | Trigger | Rejects invalid status on INSERT |
| `trg_validate_status_update` | Trigger | Rejects invalid status on UPDATE |
| `trg_complaint_status_audit` | Trigger | Records every status change to `complaint_audit` |
| `trg_cascade_delete_complaints` | Trigger | Deletes a student's complaints before deleting the student |
| `trg_protect_main_admin` | Trigger | Prevents deletion of admin account id = 1 |

## Requirements

- Go (any modern version)
- SQLite is embedded via the Go driver — no separate install needed
- `gcc` / C compiler (required by `mattn/go-sqlite3` for CGO)

## Run Locally

```bash
# 1. Clone
git clone https://github.com/supremeinferno/SQL_Project
cd SQL_Project

# 2. (Optional) recreate the database from scratch
rm -f hostel.db
sqlite3 hostel.db < sql/schema.sql

# 3. Start the server
go run .
```

Open `http://localhost:8080`

## Login

| Role    | Username | Password |
|---------|----------|----------|
| Admin   | `admin`  | `123`    |
| Student | set by admin when registering | set by admin |

## Notes

- Passwords are stored in plain text — this is a learning/demo project.
- Student operations (update/delete complaint) are scoped to their own `student_id` in SQL.
- The `archive/` folder contains the original Go and SQL files before the PL/SQL additions.

## License

For learning/demo use only.
