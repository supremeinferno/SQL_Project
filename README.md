# Hostel Management System (Go + Oracle PL/SQL)

A hostel management web app with **Admin** and **Student** logins, featuring a full **PL/SQL implementation** (triggers, views, stored procedures, cursors, and exception handling) alongside the Go backend.

> Oracle-only — the project runs against Oracle Database Free 23ai. There is no SQLite fallback.

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
| Database | Oracle Database Free 23ai (Docker) |
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
│   └── plsql.sql              # Oracle schema + PL/SQL implementation
├── templates/
│   ├── base.html
│   ├── login.html
│   ├── admin.html
│   ├── student.html
│   └── audit.html
└── static/                    # CSS assets
```

## PL/SQL Implementation

| Construct | Details |
|-----------|---------|
| **Package** | `hostel_pkg` — specification + body |
| **Procedures** | `add_student`, `update_complaint_status`, `delete_student`, `list_pending_complaints` |
| **Function** | `get_complaint_count(student_id)` |
| **Cursor** | Explicit cursor in `list_pending_complaints` |
| **Exception handling** | Named exceptions, `NO_DATA_FOUND`, `OTHERS`, `ROLLBACK` |
| **Triggers** | 5 triggers (audit, cascade delete, validation, admin protection) |
| **Views** | `admin_complaint_view`, `student_complaint_summary` |

---

## Before You Start — Install These

You need three things installed before running the project:

### 1. Go
Download and install from https://go.dev/dl/

Verify it works:
```bash
go version
```

### 2. Docker Desktop
Download and install from https://www.docker.com/products/docker-desktop/

After installing, **open Docker Desktop and wait for it to start** (look for the whale icon in your taskbar/menu bar). Docker must be running before you use any `docker` commands.

Verify it works:
```bash
docker version
```

### 3. Oracle Container Registry account
The Oracle database image requires a free account to download.

- Sign up at https://container-registry.oracle.com
- After signing in, go to **Database → free** and click **Accept** on the license agreement

> **Windows users:** Use **PowerShell** or **Git Bash** for all commands below. Command Prompt (cmd.exe) does not support the heredoc syntax (`<<`) used in some steps.

---

## Run Locally

### Step 1 — Clone the repository

```bash
git clone https://github.com/supremeinferno/SQL_Project
cd SQL_Project
```

### Step 2 — Log in to the Oracle Container Registry

```bash
docker login container-registry.oracle.com
```

Enter the email and password of your Oracle account when prompted.

### Step 3 — Start the Oracle database

```bash
docker run -d --name oracle-free \
  -p 1521:1521 \
  -e ORACLE_PASSWORD=Oracle123 \
  container-registry.oracle.com/database/free:latest
```

> The image is ~2 GB — the first download will take a few minutes.

Wait until the database is ready before continuing. Run this command and watch the output:

```bash
docker logs -f oracle-free
```

When you see `DATABASE IS READY TO USE!`, press `Ctrl+C` to stop watching the logs.

### Step 4 — Load the schema and PL/SQL objects

**Mac / Linux / Git Bash:**
```bash
docker exec -i oracle-free sqlplus system/Oracle123@FREEPDB1 @/dev/stdin < sql/plsql.sql
```

**PowerShell (Windows):**
```powershell
Get-Content sql/plsql.sql | docker exec -i oracle-free sqlplus system/Oracle123@FREEPDB1 @/dev/stdin
```

### Step 5 — Create the admin account

**Mac / Linux / Git Bash:**
```bash
docker exec -i oracle-free sqlplus system/Oracle123@FREEPDB1 <<'EOF'
INSERT INTO admins (username, password) VALUES ('admin', '123');
COMMIT;
EXIT;
EOF
```

**PowerShell (Windows):**
```powershell
"INSERT INTO admins (username, password) VALUES ('admin', '123');`nCOMMIT;`nEXIT;" |
  docker exec -i oracle-free sqlplus system/Oracle123@FREEPDB1
```

### Step 6 — Start the app

```bash
go run .
```

Open `http://localhost:8080` in your browser.

---

## Login

| Role    | Username | Password |
|---------|----------|----------|
| Admin   | `admin`  | `123`    |
| Student | set by admin during registration | set by admin |

---

## Coming Back Later

Steps 4 and 5 only need to be done **once**. Next time you want to run the project:

```bash
# Start the database container (if it's not already running)
docker start oracle-free

# Start the app
go run .
```

To stop the database when you're done:
```bash
docker stop oracle-free
```

---

## Custom Database Connection

By default the app connects as `system` to `FREEPDB1` on `localhost:1521`. Set the `ORACLE_DSN` environment variable to override this:

```bash
# Mac / Linux
export ORACLE_DSN="oracle://system:Oracle123@localhost:1521/FREEPDB1"
go run .

# PowerShell (Windows)
$env:ORACLE_DSN="oracle://system:Oracle123@localhost:1521/FREEPDB1"
go run .
```

---

## Troubleshooting

| Problem | Fix |
|---------|-----|
| `docker: command not found` | Docker Desktop is not installed or not running. Install it from https://www.docker.com/products/docker-desktop/ and make sure it is open. |
| `port is already allocated` | Port 1521 is in use. Stop the conflicting service, or use a different port: `-p 1522:1521` and update `ORACLE_DSN` to match. |
| `docker: permission denied` | On Linux, run with `sudo` or add your user to the `docker` group: `sudo usermod -aG docker $USER` (then log out and back in). |
| Container exits immediately | Run `docker logs oracle-free` to see why. Most likely cause: not enough memory — Docker Desktop needs at least **4 GB RAM** allocated (check Docker Desktop → Settings → Resources). |
| `go: command not found` | Go is not installed or not on your PATH. Install from https://go.dev/dl/ and restart your terminal. |
| `ORA-01017: invalid username/password` | The database is still initializing. Wait for `DATABASE IS READY TO USE!` in `docker logs -f oracle-free`, then retry Step 4. |
| `unauthorized` when pulling image | You haven't accepted the license on the Oracle Container Registry website. Sign in at https://container-registry.oracle.com → Database → free → Accept. |

---

## Notes

- Passwords are stored in plain text — this is a learning/demo project.
- Student operations (update/delete complaint) are scoped to their own `student_id` in SQL.
- The Go backend calls Oracle PL/SQL package procedures (e.g., `hostel_pkg.add_student`) instead of raw SQL where applicable.

## License

For learning/demo use only.
