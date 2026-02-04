# Hostel Management System (Go + SQLite)

A simple hostel management web app that supports **Admin** and **Student** logins.
Students can submit complaints, **view their complaints**, **update complaint status**, and **delete their own complaints**.
Admins can manage students and view/update/delete any complaint.

## Features

- **Authentication**
  - Admin login
  - Student login (stored in `students` table)
- **Student dashboard**
  - Submit a complaint
  - View “My Complaints”
  - Update status (Pending / In Progress / Resolved)
  - Delete own complaints
- **Admin dashboard**
  - Register students
  - View all complaints (with student info)
  - Update complaint status
  - Delete complaints
  - Search students by roll number
  - Export students to CSV

## Tech Stack

SQLite

## Project Structure

```text
.
├── main.go            # routes + server startup
├── handlers.go        # all HTTP handlers
├── db.go              # SQLite connection
├── hostel.db          # SQLite database file (local)
├── sql/
│   └── schema.sql     # DB schema + default admin insert
├── templates/         # HTML templates
└── static/            # CSS assets
```

## Requirements

- Go installed (any modern Go version should work)
- SQLite is embedded via the Go driver (no separate SQLite install required)

## Run Locally

### 1) Clone and enter the project

```bash
git clone <your-repo-url>
cd sql-projectupdated
```

### 2) (First time) Create the database/tables

This project uses `./hostel.db`.
If you already have `hostel.db`, you can skip this step.

Option A — using `sqlite3` CLI (recommended if you have it installed):

```bash
rm -f hostel.db
sqlite3 hostel.db < sql/schema.sql
```

Option B — keep the provided `hostel.db` (already in the repo), just run the app.

### 3) Start the server

```bash
go run .
```

Then open:

- `http://localhost:8080`

## Login

### Admin

Default admin from `sql/schema.sql`:

- **username**: `admin`
- **password**: `123`

### Student

Students are created by the admin on the Admin dashboard.
Use the **Username** and **Password** you set when registering the student.

## Notes / Security

- This is a learning project: passwords are stored in plain text and cookies are simple.
- Students can only **update/delete their own complaints** (enforced in SQL queries by `student_id`).

## License

For learning/demo use. Add a license if you plan to publish publicly.
