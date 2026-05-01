# Hostel Management System (Go + Oracle SQL)

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
  - 5 Oracle triggers (validation, audit trail, cascade delete, admin protection)
  - 2 SQL views (`admin_complaint_view`, `student_complaint_summary`)
  - Oracle PL/SQL Package (`hostel_pkg`) with stored procedures, functions, cursors, and exception handling

## Tech Stack

| Layer    | Technology                         |
|----------|------------------------------------|
| Language | Go                                 |
| Database | Oracle SQL (via `sijms/go-ora/v2`) |
| Frontend | HTML templates + Bootstrap 5       |

## Project Structure

```text
.
├── main.go                    # routes + server startup
├── handlers.go                # HTTP handlers (calls PL/SQL procedures)
├── db.go                      # Oracle connection setup
├── go.mod / go.sum
├── sql/
│   ├── schema.sql             # base table definitions + seed data
│   ├── schema.dbml            # DBML diagram source
│   └── plsql.sql              # Oracle PL/SQL implementation
│                                (package, procedures, functions, cursors)
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

The database logic is heavily driven by Oracle PL/SQL:

| Construct | Details |
|-----------|---------|
| **Package** | `hostel_pkg` — specification + body |
| **Procedures** | `add_student`, `update_complaint_status`, `delete_student`, `list_pending_complaints` |
| **Function** | `get_complaint_count(student_id)` |
| **Cursor** | Explicit cursor in `list_pending_complaints` |
| **Exception handling** | Named exceptions, `NO_DATA_FOUND`, `OTHERS`, `ROLLBACK` |
| **Triggers** | 5 triggers (audit, cascade delete, validation, admin protection) |
| **Views** | `admin_complaint_view`, `student_complaint_summary` |

## Requirements

- Go (any modern version)
- An Oracle Database instance (e.g., Oracle Database Free, Express Edition, or Enterprise)

## Run Locally

```bash
# 1. Clone
git clone https://github.com/supremeinferno/SQL_Project
cd SQL_Project

# 2. Setup the Oracle Database
# Connect to your Oracle database using SQL*Plus or a similar tool
# and execute the PL/SQL script to create tables, views, triggers, and packages:
# sqlplus username/password@localhost:1521/XEPDB1 @sql/plsql.sql

# 3. Configure connection
# Set the ORACLE_DSN environment variable with your Oracle credentials:
export ORACLE_DSN="oracle://username:password@localhost:1521/XEPDB1"

# 4. Start the server
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
- The Go backend now calls Oracle PL/SQL package procedures (e.g., `hostel_pkg.add_student`) instead of running raw `INSERT`/`UPDATE` queries where applicable.

## License

For learning/demo use only.
