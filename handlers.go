package main

import (
	"database/sql"
	"encoding/csv"
	"html/template"
	"net/http"
	"strconv"
)

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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	role := r.FormValue("role")

	if role == "admin" {
		var id int
		err := db.QueryRow(
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

	if role == "student" {
		var id int
		err := db.QueryRow(
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

func adminDashboard(w http.ResponseWriter, r *http.Request) {
	// Admin access check (STEP D2)
	if _, err := r.Cookie("admin"); err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// ------------------ FETCH COMPLAINTS ------------------

	cRows, err := db.Query(`
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

	// ------------------ FETCH STUDENTS ------------------

	rollSearch := r.URL.Query().Get("roll_search")

	var sRows *sql.Rows

	if rollSearch != "" {
		sRows, err = db.Query(`
		SELECT id, name, roll_no, room_no, username, password
		FROM students
		WHERE roll_no LIKE ?
	`, "%"+rollSearch+"%")
	} else {
		sRows, err = db.Query(`
		SELECT id, name, roll_no, room_no, username, password
		FROM students
		ORDER BY id
	`)
	}

	if err != nil {
		w.Write([]byte("Error loading students"))
		return
	}

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

	// ------------------ SEND TO TEMPLATE ------------------

	data := AdminPageData{
		Complaints: complaints,
		Students:   students,
	}

	tmpl := template.Must(template.ParseFiles("templates/admin.html"))
	tmpl.Execute(w, data)
}

func studentDashboard(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("student_id")
	cookie, err := r.Cookie("student_id")
	if _, err := r.Cookie("admin"); err == nil {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	rows, err := db.Query(`
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

	tmpl := template.Must(template.ParseFiles("templates/student.html"))
	tmpl.Execute(w, complaints)
}

func addStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	roll := r.FormValue("roll")
	room := r.FormValue("room")
	username := r.FormValue("username")
	password := r.FormValue("password")

	_, err := db.Exec(
		`INSERT INTO students (name, roll_no, room_no, username, password)
		 VALUES (?, ?, ?, ?, ?)`,
		name, roll, room, username, password,
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

	cookie, err := r.Cookie("student_id")
	if err != nil {
		w.Write([]byte("Student not logged in"))
		return
	}

	studentID := cookie.Value
	title := r.FormValue("title")
	description := r.FormValue("description")

	_, err = db.Exec(
		`INSERT INTO complaints (student_id, title, description)
		 VALUES (?, ?, ?)`,
		studentID, title, description,
	)

	if err != nil {
		w.Write([]byte("Error submitting complaint"))
		return
	}

	http.Redirect(w, r, "/student", http.StatusSeeOther)
}
func logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "student_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:   "admin",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func updateComplaintStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	id := r.FormValue("id")
	status := r.FormValue("status")

	db.Exec(
		"UPDATE complaints SET status=? WHERE id=?",
		status, id,
	)

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func deleteComplaint(w http.ResponseWriter, r *http.Request) {
	// Allow only POST
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Admin check (important)
	if _, err := r.Cookie("admin"); err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id := r.FormValue("id")

	_, err := db.Exec("DELETE FROM complaints WHERE id = ?", id)
	if err != nil {
		w.Write([]byte("Error deleting complaint"))
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func deleteStudent(w http.ResponseWriter, r *http.Request) {
	// Only POST allowed
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Admin check
	if _, err := r.Cookie("admin"); err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	studentID := r.FormValue("id")

	// First delete complaints of the student
	db.Exec("DELETE FROM complaints WHERE student_id = ?", studentID)

	// Then delete the student
	db.Exec("DELETE FROM students WHERE id = ?", studentID)

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func exportStudentsCSV(w http.ResponseWriter, r *http.Request) {
	// Admin check
	if _, err := r.Cookie("admin"); err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	rows, err := db.Query(`
		SELECT id, name, roll_no, room_no, username, password
		FROM students
		ORDER BY id
	`)
	if err != nil {
		w.Write([]byte("Error exporting students"))
		return
	}
	defer rows.Close()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=students.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// CSV header
	writer.Write([]string{
		"ID", "Name", "Roll No", "Room", "Username", "Password",
	})

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
