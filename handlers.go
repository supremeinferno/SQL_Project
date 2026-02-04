package main

import (
	"database/sql"
	"encoding/csv"
	"html/template"
	"net/http"
	"strconv"

	"hostel-management/database"
)

/* ---------- COMMON STRUCTS ---------- */

type PageData struct {
	Title string
	Data  interface{}
}

type ComplaintView struct {
	ID          int
	StudentName string
	RollNo      string
	RoomNo      string
	Title       string
	Description string
	Status      string
}

type StudentView struct {
	ID       int
	Name     string
	RollNo   string
	RoomNo   string
	Username string
	Password string
}

type AdminPageData struct {
	Complaints []ComplaintView
	Students   []StudentView
}

/* ---------- LOGIN ---------- */

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	role := r.FormValue("role")

	// -------- ADMIN LOGIN --------
	if role == "admin" {
		var id int
		err := database.DB.QueryRow(
			"SELECT id FROM admins WHERE username=? AND password=?",
			username, password,
		).Scan(&id)

		if err == nil {
			http.SetCookie(w, &http.Cookie{
				Name:  "admin",
				Value: "true",
				Path:  "/",
			})
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}
	}

	// -------- STUDENT LOGIN --------
	if role == "student" {
		var id int
		err := database.DB.QueryRow(
			"SELECT id FROM students WHERE username=? AND password=?",
			username, password,
		).Scan(&id)

		if err == nil {
			http.SetCookie(w, &http.Cookie{
				Name:  "student_id",
				Value: strconv.Itoa(id),
				Path:  "/",
			})
			http.Redirect(w, r, "/student", http.StatusSeeOther)
			return
		}
	}

	w.Write([]byte("Invalid login credentials"))
}

/* ---------- ADMIN DASHBOARD ---------- */

func adminDashboard(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("admin"); err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// -------- FETCH COMPLAINTS --------
	cRows, err := database.DB.Query(`
		SELECT c.id, s.name, s.roll_no, s.room_no,
		       c.title, c.description, c.status
		FROM complaints c
		LEFT JOIN students s ON c.student_id = s.id
		ORDER BY c.created_at DESC
	`)
	if err != nil {
		w.Write([]byte("Error loading complaints"))
		return
	}
	defer cRows.Close()

	var complaints []ComplaintView
	for cRows.Next() {
		var c ComplaintView
		var name, roll, room sql.NullString

		cRows.Scan(
			&c.ID,
			&name,
			&roll,
			&room,
			&c.Title,
			&c.Description,
			&c.Status,
		)

		c.StudentName = name.String
		c.RollNo = roll.String
		c.RoomNo = room.String
		complaints = append(complaints, c)
	}

	// -------- FETCH STUDENTS --------
	sRows, err := database.DB.Query(`
		SELECT id, name, roll_no, room_no, username, password
		FROM students
		ORDER BY id
	`)
	if err != nil {
		w.Write([]byte("Error loading students"))
		return
	}
	defer sRows.Close()

	var students []StudentView
	for sRows.Next() {
		var s StudentView
		sRows.Scan(
			&s.ID,
			&s.Name,
			&s.RollNo,
			&s.RoomNo,
			&s.Username,
			&s.Password,
		)
		students = append(students, s)
	}

	data := AdminPageData{
		Complaints: complaints,
		Students:   students,
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/admin.html",
	))

	tmpl.Execute(w, PageData{
		Title: "Admin Dashboard",
		Data:  data,
	})
}

/* ---------- STUDENT DASHBOARD ---------- */

func studentDashboard(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("student_id")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	rows, err := database.DB.Query(`
		SELECT id, title, description, status
		FROM complaints
		WHERE student_id = ?
		ORDER BY created_at DESC
	`, cookie.Value)
	if err != nil {
		w.Write([]byte("Error loading complaints"))
		return
	}
	defer rows.Close()

	type StudentComplaint struct {
		ID          int
		Title       string
		Description string
		Status      string
	}

	var complaints []StudentComplaint
	for rows.Next() {
		var c StudentComplaint
		rows.Scan(&c.ID, &c.Title, &c.Description, &c.Status)
		complaints = append(complaints, c)
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/student.html",
	))

	tmpl.Execute(w, PageData{
		Title: "Student Dashboard",
		Data:  complaints,
	})
}

/* ---------- STUDENT ACTIONS ---------- */

func addStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	_, err := database.DB.Exec(
		`INSERT INTO students (name, roll_no, room_no, username, password)
		 VALUES (?, ?, ?, ?, ?)`,
		r.FormValue("name"),
		r.FormValue("roll"),
		r.FormValue("room"),
		r.FormValue("username"),
		r.FormValue("password"),
	)

	if err != nil {
		w.Write([]byte("Error adding student"))
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func addComplaint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/student", http.StatusSeeOther)
		return
	}

	cookie, _ := r.Cookie("student_id")

	_, err := database.DB.Exec(
		`INSERT INTO complaints (student_id, title, description)
		 VALUES (?, ?, ?)`,
		cookie.Value,
		r.FormValue("title"),
		r.FormValue("description"),
	)

	if err != nil {
		w.Write([]byte("Error submitting complaint"))
		return
	}

	http.Redirect(w, r, "/student", http.StatusSeeOther)
}

/* ---------- ADMIN ACTIONS ---------- */

func updateComplaintStatus(w http.ResponseWriter, r *http.Request) {
	database.DB.Exec(
		"UPDATE complaints SET status=? WHERE id=?",
		r.FormValue("status"),
		r.FormValue("id"),
	)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func deleteComplaint(w http.ResponseWriter, r *http.Request) {
	database.DB.Exec("DELETE FROM complaints WHERE id=?", r.FormValue("id"))
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func deleteStudent(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	database.DB.Exec("DELETE FROM complaints WHERE student_id=?", id)
	database.DB.Exec("DELETE FROM students WHERE id=?", id)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

/* ---------- EXPORT ---------- */

func exportStudentsCSV(w http.ResponseWriter, r *http.Request) {
	rows, _ := database.DB.Query(`
		SELECT id, name, roll_no, room_no, username, password
		FROM students
		ORDER BY id
	`)
	defer rows.Close()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=students.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	writer.Write([]string{"ID", "Name", "Roll", "Room", "Username", "Password"})

	for rows.Next() {
		var s StudentView
		rows.Scan(&s.ID, &s.Name, &s.RollNo, &s.RoomNo, &s.Username, &s.Password)
		writer.Write([]string{
			strconv.Itoa(s.ID),
			s.Name,
			s.RollNo,
			s.RoomNo,
			s.Username,
			s.Password,
		})
	}
}

/* ---------- LOGOUT ---------- */

func logout(w http.ResponseWriter, r *http.Request) {

	// remove student cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "student_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// remove admin cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "admin",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
