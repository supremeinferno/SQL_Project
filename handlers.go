package main

import (
	"database/sql"
	"encoding/csv"
	"html/template"
	"net/http"
	"fmt"
	"strconv"
)

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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	role := r.FormValue("role")

	fmt.Println("Username:", username)
	fmt.Println("Password:", password)
	fmt.Println("Role:", role)

	if role == "admin" {
	var id int

		err := db.QueryRow(
			"SELECT id FROM admins WHERE username=:1 AND password=:2",
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

		fmt.Println("Admin login error:", err) // leaving this in while we sort out login flakiness
	}

	if role == "student" {
		var id int
		err := db.QueryRow(
			"SELECT id FROM students WHERE username=:1 AND password=:2",
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
	// Bounce anyone who isn't logged in as admin.
	if _, err := r.Cookie("admin"); err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Pull complaints. We use the admin_complaint_view so the join with
	// students is already done for us on the DB side.

	cRows, err := db.Query(`
		SELECT complaint_id, student_name, roll_no, room_no,
		       title, description, status
		FROM admin_complaint_view
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

	// Now grab the student list. If the admin typed something into the search
	// box we filter by roll number, otherwise just show everyone.

	rollSearch := r.URL.Query().Get("roll_search")

	var sRows *sql.Rows

	if rollSearch != "" {
		sRows, err = db.Query(`
		SELECT id, name, roll_no, room_no, username, password
		FROM students
		WHERE roll_no LIKE :1
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

	// Hand it all off to the template.
	data := AdminPageData{
		Complaints: complaints,
		Students:   students,
	}

	// Old single-template version, kept around in case the layout split
	// causes problems and we need to fall back quickly:
	// tmpl := template.Must(template.ParseFiles("templates/admin.html"))
	// tmpl.Execute(w, data)

	tmpl := template.Must(template.New("base.html").Funcs(template.FuncMap{
		"inc": func(i int) int { return i + 1 },
	}).ParseFiles(
		"templates/base.html",
		"templates/admin.html",
	))

	tmpl.Execute(w, PageData{
		Title: "Admin Dashboard",
		Data:  data,
	})

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
		WHERE student_id = :1
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

	// Same fallback story as the admin handler - keeping the old version
	// commented so we can revert quickly if something goes sideways.
	// tmpl := template.Must(template.ParseFiles("templates/student.html"))
	// tmpl.Execute(w, complaints)

	tmpl := template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/student.html",
	))

	tmpl.Execute(w, PageData{
		Title: "Student Dashboard",
		Data:  complaints,
	})
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

	fmt.Printf("addStudent: name=%q roll=%q room=%q username=%q\n", name, roll, room, username)

	_, err := db.Exec(
		`BEGIN hostel_pkg.add_student(:1, :2, :3, :4, :5); END;`,
		name, roll, room, username, password,
	)

	if err != nil {
		fmt.Println("addStudent error:", err)
		w.Write([]byte("Error adding student: " + err.Error()))
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
		 VALUES (:1, :2, :3)`,
		studentID, title, description,
	)

	if err != nil {
		w.Write([]byte("Error submitting complaint"))
		return
	}

	http.Redirect(w, r, "/student", http.StatusSeeOther)
}
func deleteStudentComplaint(w http.ResponseWriter, r *http.Request) {
	// POST-only - don't want this firing on a stray GET from a link prefetch.
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/student", http.StatusSeeOther)
		return
	}

	// Need a logged-in student before we touch anything.
	cookie, err := r.Cookie("student_id")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	complaintID := r.FormValue("id")
	if complaintID == "" {
		http.Redirect(w, r, "/student", http.StatusSeeOther)
		return
	}

	// The student_id check in the WHERE is on purpose - stops one student
	// from deleting another student's complaint by guessing IDs.
	_, err = db.Exec(
		"DELETE FROM complaints WHERE id = :1 AND student_id = :2",
		complaintID,
		cookie.Value,
	)
	if err != nil {
		w.Write([]byte("Error deleting complaint"))
		return
	}

	http.Redirect(w, r, "/student", http.StatusSeeOther)
}

func updateStudentComplaintStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/student", http.StatusSeeOther)
		return
	}

	cookie, err := r.Cookie("student_id")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	complaintID := r.FormValue("id")
	status := r.FormValue("status")
	if complaintID == "" || status == "" {
		http.Redirect(w, r, "/student", http.StatusSeeOther)
		return
	}

	_, err = db.Exec(
		"UPDATE complaints SET status = :1 WHERE id = :2 AND student_id = :3",
		status, complaintID, cookie.Value,
	)

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
		"BEGIN hostel_pkg.update_complaint_status(:1, :2); END;",
		id, status,
	)

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func deleteComplaint(w http.ResponseWriter, r *http.Request) {
	// POST only.
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Don't skip this - without it, anyone with the URL can wipe complaints.
	if _, err := r.Cookie("admin"); err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id := r.FormValue("id")

	_, err := db.Exec("DELETE FROM complaints WHERE id = :1", id)
	if err != nil {
		w.Write([]byte("Error deleting complaint"))
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func deleteStudent(w http.ResponseWriter, r *http.Request) {
	// POST only.
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	// Admins only.
	if _, err := r.Cookie("admin"); err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	studentID := r.FormValue("id")

	// Hand off to the PL/SQL proc - it deletes the student and their
	// complaints in one transaction so we don't end up with orphan rows.
	db.Exec("BEGIN hostel_pkg.delete_student(:1); END;", studentID)

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
func auditLogHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("admin"); err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	rows, err := db.Query(`
		SELECT a.id, a.complaint_id,
		       COALESCE(s.name, 'Deleted Student') AS student_name,
		       a.old_status, a.new_status, a.changed_at
		FROM complaint_audit a
		LEFT JOIN complaints c ON a.complaint_id = c.id
		LEFT JOIN students   s ON c.student_id   = s.id
		ORDER BY a.changed_at DESC
	`)
	if err != nil {
		w.Write([]byte("Error loading audit log"))
		return
	}
	defer rows.Close()

	type AuditEntry struct {
		ID          int
		ComplaintID int
		StudentName string
		OldStatus   string
		NewStatus   string
		ChangedAt   string
	}

	var entries []AuditEntry
	for rows.Next() {
		var e AuditEntry
		rows.Scan(&e.ID, &e.ComplaintID, &e.StudentName,
			&e.OldStatus, &e.NewStatus, &e.ChangedAt)
		entries = append(entries, e)
	}

	tmpl := template.Must(template.ParseFiles(
		"templates/base.html",
		"templates/audit.html",
	))
	tmpl.Execute(w, PageData{
		Title: "Audit Log",
		Data:  entries,
	})
}

func exportStudentsCSV(w http.ResponseWriter, r *http.Request) {
	// Admins only - this dumps the whole roster.
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

	// Header row first, then the rows.
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
