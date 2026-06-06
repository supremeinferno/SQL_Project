# Hostel Management System

A clean, beginner-friendly **Go + Oracle PL/SQL** web application for managing hostel students and complaints. The project includes separate **Admin** and **Student** flows, an Oracle-backed schema, triggers, views, stored procedures, cursors, and an automatic complaint-status audit log.


---

## Features

### Admin
- Manage students (add/search/delete)
- View and update complaints
- Export student data (CSV)
- View audit logs

### Student
- Submit complaints
- Track complaint status
- Delete own complaints

### Database (Oracle PL/SQL)
- Triggers for validation & audit logging
- Views for simplified queries
- Package with procedures, functions, and cursors
---
## Preview

### Login Page
![Login](screenshots/login.jpeg)

### Admin Dashboard
![Admin](screenshots/admin.jpeg)

### Student Dashboard
![Student](screenshots/student.jpeg)

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go |
| Database | Oracle Database |
| Database image | `gvenzl/oracle-free` |
| Oracle driver | `github.com/sijms/go-ora/v2` |
| Frontend | HTML templates, Bootstrap 5, custom CSS |
| Dev setup | Docker Desktop |

---

## Application Flow

1. User logs in (Admin / Student)
2. Role-based dashboard is loaded
3. Actions trigger Go handlers
4. Handlers interact with Oracle DB
5. PL/SQL layer enforces rules (triggers, procedures)
6. Response rendered via HTML templates

## Project Structure

```text
.
├── main.go                 # Server startup and route registration
├── handlers.go             # HTTP handlers and template rendering
├── db.go                   # Oracle connection setup
├── go.mod / go.sum         # Go module files
├── sql/
│   └── plsql.sql           # Oracle schema, triggers, views, package, seed admin
├── templates/
│   ├── base.html           # Shared layout
│   ├── login.html          # Login page
│   ├── admin.html          # Admin dashboard
│   ├── student.html        # Student dashboard
│   └── audit.html          # Audit log page
└── static/
    ├── bootstrap.min.css
    └── style.css
```

---

## Prerequisites

Install these before running the project:

1. **Go 1.22 or newer**  
   Download: https://go.dev/dl/

   ```bash
   go version
   ```

2. **Docker Desktop**  
   Download: https://www.docker.com/products/docker-desktop/

   Open Docker Desktop before running the database commands.

   ```bash
   docker version
   ```

---

## Run Locally

### 1. Clone the repository

```bash
git clone https://github.com/Harbani007/hostel_management_system.git
cd hostel_management_system
```

### 2. Start Oracle Database

```bash
docker run -d --name oracle-free \
  -p 1521:1521 \
  -e ORACLE_PASSWORD=Oracle123 \
  gvenzl/oracle-free:latest
```

The image is large, so the first pull can take a few minutes.

Watch the logs until Oracle is ready:

```bash
docker logs -f oracle-free
```

Continue when you see:

```text
DATABASE IS READY TO USE!
```

Press `Ctrl + C` to stop watching logs.

### 3. Load the schema and PL/SQL objects

The SQL script creates all database objects and also seeds the default admin account.

**macOS / Linux / Git Bash**

```bash
docker exec -i oracle-free sqlplus system/Oracle123@FREEPDB1 @/dev/stdin < sql/plsql.sql
```

**Windows PowerShell**

```powershell
Get-Content sql/plsql.sql | docker exec -i oracle-free sqlplus system/Oracle123@FREEPDB1 @/dev/stdin
```

### 4. Start the Go app

```bash
go run .
```

Open this in your browser:

```text
http://localhost:8080
```

---

## Login Details

| Role | Username | Password |
|---|---|---|
| Admin | `admin` | `123` |
| Student | Created by admin | Created by admin |

---

## Useful Commands

Start the existing Oracle container:

```bash
docker start oracle-free
```

Stop the Oracle container:

```bash
docker stop oracle-free
```

Rebuild/run the Go app:

```bash
go run .
```

Build a local binary:

```bash
go build -o hostel-management .
```

The binary is ignored by Git and should not be committed.

---

## Custom Database Connection

By default, the app connects to:

```text
oracle://system:Oracle123@localhost:1521/FREEPDB1
```

Override it with `ORACLE_DSN`.

**macOS / Linux**

```bash
export ORACLE_DSN="oracle://system:Oracle123@localhost:1521/FREEPDB1"
go run .
```

**Windows PowerShell**

```powershell
$env:ORACLE_DSN="oracle://system:Oracle123@localhost:1521/FREEPDB1"
go run .
```

---

## PL/SQL Highlights

| PL/SQL concept | Where it is used |
|---|---|
| Sequences | Auto-generating IDs for admins, students, complaints, and audit rows |
| Views | `admin_complaint_view`, `student_complaint_summary` |
| Triggers | Status audit, status validation, cascade complaint cleanup, main admin protection |
| Package | `hostel_pkg` |
| Procedures | `add_student`, `update_complaint_status`, `delete_student`, `list_pending_complaints` |
| Function | `get_complaint_count(student_id)` |
| Cursor | Explicit cursor inside `list_pending_complaints` |
| Exceptions | Duplicate student, invalid status, missing complaint/student, rollback handling |

---

## Troubleshooting

| Problem | Fix |
|---|---|
| `docker: command not found` | Install/open Docker Desktop and restart the terminal. |
| `port is already allocated` | Another service is using port `1521`. Stop it, or map Oracle to another port and update `ORACLE_DSN`. |
| Container exits immediately | Check logs using `docker logs oracle-free`. Docker may need more memory. Allocate at least 4 GB RAM. |
| `ORA-01017: invalid username/password` | Oracle may still be starting. Wait for `DATABASE IS READY TO USE!`, then retry. |
| `table or view does not exist` | Run `sql/plsql.sql` again. The script is safe to rerun and recreates the demo schema. |
| `go: command not found` | Install Go and restart your terminal. |

---

## Notes

- This is a learning/demo project, so passwords are stored in plain text.
- Admin status updates are recorded automatically in `complaint_audit` by `trg_complaint_status_audit`.
- The SQL script is intentionally rerunnable to make local development easier.
- The project is structured to stay readable, practical, and easy to explain.

---
