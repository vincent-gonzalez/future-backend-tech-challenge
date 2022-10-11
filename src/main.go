package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB

func checkIfAppointmentTimeIsAvailable(trainer_id int, start_time string, ends_at string) bool {
	var appointment Appointment
	err := database.QueryRow(`SELECT * FROM appointments
	WHERE trainer_id = ? AND starts_at = ? AND ends_at = ?`,
	 trainer_id, start_time, ends_at).Scan(&appointment)
	if err != nil {
		if err == sql.ErrNoRows {
			return true
		}
		return false // some other error
	}

	return false // an appointment matching the criteria was found
}

func indexHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if req.URL.Path != "/" {
		http.NotFound(res, req)
		return
	}

	switch req.Method {
	case http.MethodGet:
		res.Write([]byte("<h1>Hello World 2!</h1>"))
	}
}

func appointmentsHandler(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	res.Write([]byte(fmt.Sprintf("<h1>Appointments for trainer: %v </h1>", params.ByName("id"))))
}

func appointmentsPostHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	requestBody, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		// write error response res.
	}
	var appointment Appointment
	if err = json.Unmarshal(requestBody, appointment); err != nil {
		// write error response
	}

	if isAvailable := checkIfAppointmentTimeIsAvailable(
		appointment.TrainerId, appointment.StartsAt, appointment.EndsAt); !isAvailable {
		// return bad request
	}

	// save request into DB
	insertSQL := `INSERT INTO appointments(trainer_id, user_id, starts_at, ends_at) VALUES (?, ?, ?, ?)`
	statement, err := database.Prepare(insertSQL)
	if err != nil {
		// 500 error
	}
	_, err = statement.Exec(appointment.TrainerId, appointment.UserId, appointment.StartsAt, appointment.EndsAt)
	if err != nil {
		// 500 error
	}
	res.WriteHeader(200)

	res.Write([]byte("<h1>Post successful!</h1>"))
}

func trainerAppointmentsHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	
}

func initDB() error {
	os.Remove("../db/api-test-sqlite.db")

	file, err := os.Create("../db/api-test-sqlite.db")
	if err != nil {
		return err
	}
	defer file.Close()

	database, err = sql.Open("sqlite3", "../db/api-test-sqlite.db")
	defer database.Close()

	createTableSQL := `CREATE TABLE appointments (
	"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT
	"trainer_id" integer
	"user_id" integer
	"starts_at" datetime
	"ends_at" datetime);`
	statement, err := database.Prepare(createTableSQL)
	if err != nil {
		//
	}
	statement.Exec()

}

func main() {
	fmt.Println("Hello World!")
	//mux := http.NewServeMux()
	//mux.HandleFunc("/", indexHandler)
	router := httprouter.New()
	router.GET("/", indexHandler)
	router.GET("/appointments/:id", appointmentsHandler)
	router.POST("/appointments", appointmentsPostHandler)
	router.GET("/trainers/appointments?starts_at=:startsAt;ends_at=:endsAt", trainerAppointmentsHandler)
	log.Fatal(http.ListenAndServe(":8081", router))
}
