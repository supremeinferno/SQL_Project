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
| Database | Oracle Database 23c Free (Docker)  |
| Driver   | `sijms/go-ora/v2`                  |
| Frontend | HTML templates + Bootstrap 5       |

## Project Structure

```text
.
├── main.go                    # routes + server startup
├── handlers.go                # HTTP handlers (calls PL/SQL procedures)
├── db.go                      # Oracle connection setup
├── go.mod / go.sum
├── sql/
│   ├── plsql.sql              # Oracle schema + PL/SQL implementation
│   │                            (sequences, tables, views, triggers, package)
│   └── schema.dbml            # DBML diagram source
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

- [Go](https://go.dev/dl/) (any modern version)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (Mac / Windows / Linux)
- A free [Oracle Container Registry](https://container-registry.oracle.com) account (needed to pull the Oracle image)

> **Windows users:** All commands below work in **PowerShell** or **Git Bash**. If you are using Command Prompt (cmd.exe), use Git Bash instead — heredoc syntax (`<<`) is not supported in cmd.exe.

## Run Locally

### 1. Clone the repository

```bash
git clone https://github.com/supremeinferno/SQL_Project
cd SQL_Project
```

### 2. Log in to the Oracle Container Registry

Create a free account at https://container-registry.oracle.com, then run:

```bash
docker login container-registry.oracle.com
```

Enter your Oracle account email and password when prompted.

### 3. Start the Oracle Database container

```bash
docker run -d --name oracle-xe \
  -p 1521:1521 \
  -e ORACLE_PASSWORD=Oracle123 \
  container-registry.oracle.com/database/free:latest
```

> The image is ~2 GB — the first pull will take a few minutes depending on your connection.

Wait until the database is ready (look for `DATABASE IS READY TO USE!`):

```bash
docker logs -f oracle-xe
```

Press `Ctrl+C` to stop following logs once you see that message.

### 4. Load the schema and PL/SQL objects

**Mac / Linux / Git Bash (Windows):**
```bash
docker exec -i oracle-xe sqlplus system/Oracle123@FREEPDB1 @/dev/stdin < sql/plsql.sql
```

**PowerShell (Windows):**
```powershell
Get-Content sql/plsql.sql | docker exec -i oracle-xe sqlplus system/Oracle123@FREEPDB1 @/dev/stdin
```

### 5. Seed the admin account

**Mac / Linux / Git Bash (Windows):**
```bash
docker exec -i oracle-xe sqlplus system/Oracle123@FREEPDB1 <<'EOF'
INSERT INTO admins (username, password) VALUES ('admin', '123');
COMMIT;
EXIT;
EOF
```

**PowerShell (Windows):**
```powershell
"INSERT INTO admins (username, password) VALUES ('admin', '123');`nCOMMIT;`nEXIT;" |
  docker exec -i oracle-xe sqlplus system/Oracle123@FREEPDB1
```

### 6. Start the Go server

```bash
go run .
```

Open `http://localhost:8080` in your browser.

## Login

| Role    | Username | Password |
|---------|----------|----------|
| Admin   | `admin`  | `123`    |
| Student | set by admin when registering | set by admin |

## Custom Database Connection

By default the app connects as `system` to `FREEPDB1` on `localhost:1521`. Override with the `ORACLE_DSN` environment variable if your setup differs:

```bash
# Mac / Linux
export ORACLE_DSN="oracle://system:Oracle123@localhost:1521/FREEPDB1"
go run .

# PowerShell (Windows)
$env:ORACLE_DSN="oracle://system:Oracle123@localhost:1521/FREEPDB1"
go run .
```

## Troubleshooting

| Problem | Fix |
|---------|-----|
| `port is already allocated` | Another service is using 1521. Stop it, or change the port: `-p 1522:1521` and update `ORACLE_DSN` accordingly. |
| `docker: permission denied` | On Linux, prefix commands with `sudo` or add your user to the `docker` group. |
| Container exits immediately | Run `docker logs oracle-xe` to see the error. Usually not enough memory — Docker Desktop needs at least **4 GB RAM** allocated. |
| `go: command not found` | Install Go from https://go.dev/dl/ and make sure it is on your `PATH`. |
| `ORA-01017: invalid username/password` | The DB may still be initializing. Wait for `DATABASE IS READY TO USE!` in the logs, then retry step 4. |

## Notes

- Passwords are stored in plain text — this is a learning/demo project.
- Student operations (update/delete complaint) are scoped to their own `student_id` in SQL.
- The Go backend calls Oracle PL/SQL package procedures (e.g., `hostel_pkg.add_student`) instead of raw SQL where applicable.

## License

For learning/demo use only.
