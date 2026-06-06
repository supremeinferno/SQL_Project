package main

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	RollSearch string
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"inc": func(i int) int { return i + 1 },
	}
}

func renderPage(w http.ResponseWriter, title string, data interface{}, files ...string) {
	allFiles := append([]string{"templates/base.html"}, files...)
	tmpl, err := template.New("base.html").Funcs(templateFuncs()).ParseFiles(allFiles...)
	if err != nil {
		log.Printf("template parse error: %v", err)
		http.Error(w, "Unable to load page", http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base.html", PageData{Title: title, Data: data}); err != nil {
		log.Printf("template execute error: %v", err)
		http.Error(w, "Unable to render page", http.StatusInternalServerError)
	}
}

func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("admin")
	if err != nil || cookie.Value != "true" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return false
	}
	return true
}

func currentStudentID(w http.ResponseWriter, r *http.Request) (int, bool) {
	cookie, err := r.Cookie("student_id")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return 0, false
	}

	studentID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		clearCookie(w, "student_id")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return 0, false
	}

	return studentID, true
}

func setSessionCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func nullString(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

func formatDBTime(value interface{}) string {
	switch v := value.(type) {
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	role := r.FormValue("role")

	switch role {
	case "admin":
		var id int
		err := db.QueryRow(
			"SELECT id FROM admins WHERE username = :1 AND password = :2",
			username,
			password,
		).Scan(&id)
		if err == nil {
			clearCookie(w, "student_id")
			setSessionCookie(w, "admin", "true")
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("admin login query error: %v", err)
		}

	case "student":
		var id int
		err := db.QueryRow(
			"SELECT id FROM students WHERE username = :1 AND password = :2",
			username,
			password,
		).Scan(&id)
		if err == nil {
			clearCookie(w, "admin")
			setSessionCookie(w, "student_id", strconv.Itoa(id))
			http.Redirect(w, r, "/student", http.StatusSeeOther)
			return
		}
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("student login query error: %v", err)
		}
	}

	http.Error(w, "Invalid login credentials", http.StatusUnauthorized)
}

func adminDashboard(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	cRows, err := db.Query(`
		SELECT complaint_id, student_name, roll_no, room_no,
		       title, description, status
		FROM admin_complaint_view
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Printf("load complaints: %v", err)
		http.Error(w, "Error loading complaints", http.StatusInternalServerError)
		return
	}
	defer cRows.Close()

	var complaints []ComplaintView
	for cRows.Next() {
		var c ComplaintView
		var name, roll, room sql.NullString

		if err := cRows.Scan(
			&c.ID,
			&name,
			&roll,
			&room,
			&c.Title,
			&c.Description,
			&c.Status,
		); err != nil {
			log.Printf("scan complaint: %v", err)
			http.Error(w, "Error loading complaints", http.StatusInternalServerError)
			return
		}

		c.StudentName = nullString(name)
		c.RollNo = nullString(roll)
		c.RoomNo = nullString(room)
		complaints = append(complaints, c)
	}
	if err := cRows.Err(); err != nil {
		log.Printf("iterate complaints: %v", err)
		http.Error(w, "Error loading complaints", http.StatusInternalServerError)
		return
	}

	rollSearch := strings.TrimSpace(r.URL.Query().Get("roll_search"))
	var sRows *sql.Rows
	if rollSearch != "" {
		sRows, err = db.Query(`
			SELECT id, name, roll_no, room_no, username, password
			FROM students
			WHERE UPPER(roll_no) LIKE UPPER(:1)
			ORDER BY id
		`, "%"+rollSearch+"%")
	} else {
		sRows, err = db.Query(`
			SELECT id, name, roll_no, room_no, username, password
			FROM students
			ORDER BY id
		`)
	}
	if err != nil {
		log.Printf("load students: %v", err)
		http.Error(w, "Error loading students", http.StatusInternalServerError)
		return
	}
	defer sRows.Close()

	var students []StudentView
	for sRows.Next() {
		var s StudentView
		if err := sRows.Scan(
			&s.ID,
			&s.Name,
			&s.RollNo,
			&s.RoomNo,
			&s.Username,
			&s.Password,
		); err != nil {
			log.Printf("scan student: %v", err)
			http.Error(w, "Error loading students", http.StatusInternalServerError)
			return
		}
		students = append(students, s)
	}
	if err := sRows.Err(); err != nil {
		log.Printf("iterate students: %v", err)
		http.Error(w, "Error loading students", http.StatusInternalServerError)
		return
	}

	renderPage(w, "Admin Dashboard", AdminPageData{
		Complaints: complaints,
		Students:   students,
		RollSearch: rollSearch,
	}, "templates/admin.html")
}

func studentDashboard(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("admin"); err == nil {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	studentID, ok := currentStudentID(w, r)
	if !ok {
		return
	}

	rows, err := db.Query(`
		SELECT id, title, description, status
		FROM complaints
		WHERE student_id = :1
		ORDER BY created_at DESC
	`, studentID)
	if err != nil {
		log.Printf("load student complaints: %v", err)
		http.Error(w, "Error loading complaints", http.StatusInternalServerError)
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
		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.Status); err != nil {
			log.Printf("scan student complaint: %v", err)
			http.Error(w, "Error loading complaints", http.StatusInternalServerError)
			return
		}
		complaints = append(complaints, c)
	}
	if err := rows.Err(); err != nil {
		log.Printf("iterate student complaints: %v", err)
		http.Error(w, "Error loading complaints", http.StatusInternalServerError)
		return
	}

	renderPage(w, "Student Dashboard", complaints, "templates/student.html")
}

func addStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	if !requireAdmin(w, r) {
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	roll := strings.TrimSpace(r.FormValue("roll"))
	room := strings.TrimSpace(r.FormValue("room"))
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	if name == "" || roll == "" || room == "" || username == "" || password == "" {
		http.Error(w, "All student fields are required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(
		`BEGIN hostel_pkg.add_student(:1, :2, :3, :4, :5); END;`,
		name,
		roll,
		room,
		username,
		password,
	)
	if err != nil {
		log.Printf("add student: %v", err)
		http.Error(w, "Error adding student: "+err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func addComplaint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/student", http.StatusSeeOther)
		return
	}

	studentID, ok := currentStudentID(w, r)
	if !ok {
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	if title == "" || description == "" {
		http.Error(w, "Complaint title and description are required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec(
		`INSERT INTO complaints (student_id, title, description)
		 VALUES (:1, :2, :3)`,
		studentID,
		title,
		description,
	)
	if err != nil {
		log.Printf("add complaint: %v", err)
		http.Error(w, "Error submitting complaint", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/student", http.StatusSeeOther)
}

func deleteStudentComplaint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/student", http.StatusSeeOther)
		return
	}

	studentID, ok := currentStudentID(w, r)
	if !ok {
		return
	}

	complaintID := strings.TrimSpace(r.FormValue("id"))
	if complaintID == "" {
		http.Redirect(w, r, "/student", http.StatusSeeOther)
		return
	}

	_, err := db.Exec(
		"DELETE FROM complaints WHERE id = :1 AND student_id = :2",
		complaintID,
		studentID,
	)
	if err != nil {
		log.Printf("delete student complaint: %v", err)
		http.Error(w, "Error deleting complaint", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/student", http.StatusSeeOther)
}

func logout(w http.ResponseWriter, r *http.Request) {
	clearCookie(w, "student_id")
	clearCookie(w, "admin")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func updateComplaintStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	if !requireAdmin(w, r) {
		return
	}

	id := strings.TrimSpace(r.FormValue("id"))
	status := strings.TrimSpace(r.FormValue("status"))
	if id == "" || status == "" {
		http.Error(w, "Complaint ID and status are required", http.StatusBadRequest)
		return
	}

	if _, err := db.Exec(
		"BEGIN hostel_pkg.update_complaint_status(:1, :2); END;",
		id,
		status,
	); err != nil {
		log.Printf("update complaint status: %v", err)
		http.Error(w, "Error updating complaint status: "+err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func deleteComplaint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	if !requireAdmin(w, r) {
		return
	}

	id := strings.TrimSpace(r.FormValue("id"))
	if id == "" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	_, err := db.Exec("DELETE FROM complaints WHERE id = :1", id)
	if err != nil {
		log.Printf("delete complaint: %v", err)
		http.Error(w, "Error deleting complaint", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func deleteStudent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	if !requireAdmin(w, r) {
		return
	}

	studentID := strings.TrimSpace(r.FormValue("id"))
	if studentID == "" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	if _, err := db.Exec("BEGIN hostel_pkg.delete_student(:1); END;", studentID); err != nil {
		log.Printf("delete student: %v", err)
		http.Error(w, "Error deleting student: "+err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func auditLogHandler(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	rows, err := db.Query(`
		SELECT a.id,
		       a.complaint_id,
		       COALESCE(a.student_name, s.name, 'Deleted or unavailable') AS student_name,
		       a.old_status,
		       a.new_status,
		       a.changed_at
		FROM complaint_audit a
		LEFT JOIN complaints c ON a.complaint_id = c.id
		LEFT JOIN students   s ON c.student_id   = s.id
		ORDER BY a.changed_at DESC
	`)
	if err != nil {
		log.Printf("load audit log: %v", err)
		http.Error(w, "Error loading audit log", http.StatusInternalServerError)
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
		var changedAt interface{}
		if err := rows.Scan(
			&e.ID,
			&e.ComplaintID,
			&e.StudentName,
			&e.OldStatus,
			&e.NewStatus,
			&changedAt,
		); err != nil {
			log.Printf("scan audit log: %v", err)
			http.Error(w, "Error loading audit log", http.StatusInternalServerError)
			return
		}
		e.ChangedAt = formatDBTime(changedAt)
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		log.Printf("iterate audit log: %v", err)
		http.Error(w, "Error loading audit log", http.StatusInternalServerError)
		return
	}

	renderPage(w, "Audit Log", entries, "templates/audit.html")
}

func exportStudentsCSV(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}

	rows, err := db.Query(`
		SELECT id, name, roll_no, room_no, username, password
		FROM students
		ORDER BY id
	`)
	if err != nil {
		log.Printf("export students: %v", err)
		http.Error(w, "Error exporting students", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=students.csv")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	if err := writer.Write([]string{"ID", "Name", "Roll No", "Room", "Username", "Password"}); err != nil {
		log.Printf("write csv header: %v", err)
		return
	}

	for rows.Next() {
		var s StudentView
		if err := rows.Scan(&s.ID, &s.Name, &s.RollNo, &s.RoomNo, &s.Username, &s.Password); err != nil {
			log.Printf("scan csv student: %v", err)
			return
		}

		if err := writer.Write([]string{
			strconv.Itoa(s.ID),
			s.Name,
			s.RollNo,
			s.RoomNo,
			s.Username,
			s.Password,
		}); err != nil {
			log.Printf("write csv row: %v", err)
			return
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("iterate csv students: %v", err)
	}
}
