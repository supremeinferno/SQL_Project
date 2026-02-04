package main

import (
	"log"
	"net/http"

	"hostel-management/database"
)

func main() {

	// ✅ Connect Database
	database.ConnectDB()
	log.Println("Database connected")

	// ✅ Static files
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))

	// ✅ Default route → login page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/login.html")
	})

	// ✅ Auth routes
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logout)

	// ✅ Dashboards
	http.HandleFunc("/admin", adminDashboard)
	http.HandleFunc("/student", studentDashboard)

	// ✅ Student operations
	http.HandleFunc("/add-student", addStudent)
	http.HandleFunc("/delete-student", deleteStudent)
	http.HandleFunc("/export-students", exportStudentsCSV)

	// ✅ Complaint operations
	http.HandleFunc("/add-complaint", addComplaint)
	http.HandleFunc("/update-status", updateComplaintStatus)
	http.HandleFunc("/delete-complaint", deleteComplaint)

	// ✅ Start server
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
