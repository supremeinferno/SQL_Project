package main

import (
	"log"
	"net/http"
)

func main() {
	connectDB()
	log.Println("Database connected")

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/login.html")
	})

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/admin", adminDashboard)
	http.HandleFunc("/student", studentDashboard)
	http.HandleFunc("/add-student", addStudent)
	http.HandleFunc("/add-complaint", addComplaint)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/update-status", updateComplaintStatus)
	http.HandleFunc("/delete-complaint", deleteComplaint)
	http.HandleFunc("/delete-student", deleteStudent)
	http.HandleFunc("/export-students", exportStudentsCSV)

	log.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
