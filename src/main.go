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
	trainerId := params.ByName("id")
	if len(trainerId) < 1 {
		// return 400
	}
	// use trainer id to query data store
	// marshal list of appointments to json
	// return
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

func trainerAppointmentsHandler(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	trainerId := params.ByName("id")
	if len(trainerId) < 1 {
		// return 400
		responseStatus := Response{
			Status: "fail",
		}
		readError := &ErrorResponse{
			Response: responseStatus,
			Message: "Trainer ID is required",
		}
		errorJSON, _ := json.Marshal(readError)
		res.WriteHeader(400)
		res.Write([]byte(errorJSON))
		return
	}

	rows, err := database.Query(`SELECT * FROM appointments WHERE trainer_id = ?`, trainerId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var trainerAppointments []Appointment
	for rows.Next() {
		var appointment Appointment
		if err := rows.Scan(
			&appointment.Id,
			&appointment.TrainerId,
			&appointment.UserId,
			&appointment.AppointmentTime.StartsAt,
			&appointment.AppointmentTime.EndsAt,
			); err != nil {
				responseStatus := Response{
					Status: "error",
				}
				readError := &ErrorResponse{
					Response: responseStatus,
					Message: err.Error(),
				}
				errorJSON, _ := json.Marshal(readError)
				res.WriteHeader(500)
				res.Write([]byte(errorJSON))
				return
			}
		appointment.AppointmentTime.StartsAt = convertDateTimeToRFC3339(appointment.AppointmentTime.StartsAt)
		appointment.AppointmentTime.EndsAt = convertDateTimeToRFC3339(appointment.AppointmentTime.EndsAt)
		trainerAppointments = append(trainerAppointments, appointment)
	}
	if err = rows.Err(); err != nil {
		rowError := &ErrorResponse{
			Response: Response{
				Status: "error",
			},
			Message: err.Error(),
		}
		errorJSON, _ := json.Marshal(rowError)
		res.WriteHeader(500)
		res.Write([]byte(errorJSON))
		return
	}
	// marshal list of appointments to json
	if trainerAppointments == nil || len(trainerAppointments) < 1 {
		trainerAppointments = make([]Appointment, 0)
	}
	responseBody := &DataResponse{
		Response: Response{
			Status: "success",
		},
		Data: trainerAppointments,
	}
	responseBodyJSON, _ := json.Marshal(responseBody)

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(200)
	res.Write([]byte(responseBodyJSON))
}

func loadStartingData() {
	startingData, err := os.ReadFile("./appointments.json")
	if err != nil {
		log.Fatal(err)
	}

	var startingAppointments []Appointment
	err = json.Unmarshal(startingData, &startingAppointments)
	if err != nil {
		log.Fatal(err)
	}

	insertSQL := `INSERT INTO appointments(id, trainer_id, user_id, starts_at, ends_at) VALUES (?, ?, ?, ?, ?)`
	statement, err := database.Prepare(insertSQL)
	if err != nil {
		log.Println(err)
		return
	}

	for _, appointment := range startingAppointments {
		startsAtUtcTime, err := time.Parse(time.RFC3339, appointment.AppointmentTime.StartsAt)
		endsAtUtcTime, err := time.Parse(time.RFC3339, appointment.AppointmentTime.EndsAt)
		// log.Println(appointment.AppointmentTime.StartsAt)
		// log.Println(startsAtUtcTime)
		// log.Println(appointment)
		_, err = statement.Exec(appointment.Id, appointment.TrainerId, appointment.UserId, startsAtUtcTime, endsAtUtcTime)
		if err != nil {
			log.Println(err)
		}
	}
}

func initDB() error {
	os.Remove("./db/api-test-sqlite.db")

	file, err := os.Create("./db/api-test-sqlite.db")
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer file.Close()

	database, err = sql.Open("sqlite3", "./db/api-test-sqlite.db")
	if err != nil {
		log.Fatal(err)
		return err
	}
	//defer database.Close()

	createTableSQL := `CREATE TABLE appointments (
	"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
	"trainer_id" integer,
	"user_id" integer,
	"starts_at" integer,
	"ends_at" integer);`
	statement, err := database.Prepare(createTableSQL)
	if err != nil {
		log.Fatal(err)
		return err
	}
	statement.Exec()
	return nil
}

func main() {
	fmt.Println("Hello World!")
	//mux := http.NewServeMux()
	//mux.HandleFunc("/", indexHandler)
	err := initDB()
	if err != nil {
		log.Panicln(err)
	}
	loadStartingData()
	router := httprouter.New()
	router.GET("/", indexHandler)
	router.GET("/appointments/:id", appointmentsHandler)
	router.POST("/appointments", appointmentsPostHandler)
	router.GET("/trainers/appointments?starts_at=:startsAt;ends_at=:endsAt", trainerAppointmentsHandler)
	log.Fatal(http.ListenAndServe(":8081", router))
}
