package main

import (
	"log"
	"net/http"
)

func main() {
	connectDB()
	defer db.Close()

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "templates/login.html")
	})

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/admin", adminDashboard)
	http.HandleFunc("/student", studentDashboard)
	http.HandleFunc("/add-student", addStudent)
	http.HandleFunc("/add-complaint", addComplaint)
	http.HandleFunc("/student/delete-complaint", deleteStudentComplaint)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/update-status", updateComplaintStatus)
	http.HandleFunc("/delete-complaint", deleteComplaint)
	http.HandleFunc("/delete-student", deleteStudent)
	http.HandleFunc("/export-students", exportStudentsCSV)
	http.HandleFunc("/audit-log", auditLogHandler)

	log.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
