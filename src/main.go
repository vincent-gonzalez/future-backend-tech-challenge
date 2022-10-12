package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/vincent-gonzalez/future-backend-homework-project/src/middleware"
	"github.com/vincent-gonzalez/future-backend-homework-project/src/types"
	"github.com/vincent-gonzalez/future-backend-homework-project/src/models"
	"github.com/vincent-gonzalez/future-backend-homework-project/src/utils"

	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
)

// const datetime_format_without_t = "2006-01-02 15:04:05-07:00"

var database *sql.DB

func getAvailableAppointmentsHandler(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	trainerId := req.URL.Query().Get("trainer_id")
	startsAtDate := req.URL.Query().Get("starts_at")
	endsAtDate := req.URL.Query().Get("ends_at")

	startsAtUtcTime,_ := time.Parse(time.RFC3339, startsAtDate)
	endsAtUtcTime,_ := time.Parse(time.RFC3339, endsAtDate)
	//log.Println(startsAtUtcTime)
	// get the trainer's currently scheduled appointments
	// rows, err := database.Query(`SELECT starts_at, ends_at FROM appointments
	//  WHERE trainer_id = ? AND starts_at >= ? AND ends_at <= ?`,
	// 	trainerId, startsAtUtcTime, endsAtUtcTime)
	// if err != nil {
	// 	log.Println(err)
	// 	WriteErrorResponse(res, "failed while retrieving appointment record data", http.StatusInternalServerError)
	// 	return
	// }
	// defer rows.Close()

	// // iterate over scheduled appointments
	// var appointments []types.AppointmentTime
	// for rows.Next() {
	// 	var appointment types.AppointmentTime
	// 	if err := rows.Scan(
	// 		&appointment.StartsAt,
	// 		&appointment.EndsAt,
	// 	); err != nil {
	// 		log.Println(err)
	// 		WriteErrorResponse(res, "failed while reading available appointments data", http.StatusInternalServerError)
	// 		return
	// 	}

	// 	appointment.StartsAt = ConvertDateTimeToRFC3339(appointment.StartsAt, datetime_format_without_t)
	// 	appointment.EndsAt = ConvertDateTimeToRFC3339(appointment.EndsAt, datetime_format_without_t)
	// 	appointments = append(appointments, appointment)
	// }

	// if err = rows.Err(); err != nil {
	// 	log.Panicln(err)
	// 	WriteErrorResponse(res, "failed while processing available appointments", http.StatusInternalServerError)
	// 	return
	// }
	appointments, err := models.GetScheduledAppointmentsBetweenDates(trainerId, startsAtUtcTime, endsAtUtcTime)
	if err != nil {
		log.Println(err)
		utils.WriteErrorResponse(res, "failed to retrieve available appointments", http.StatusInternalServerError)
	}

	availableAppointments := utils.FindTrainerAvailability(appointments, startsAtDate, endsAtDate)

	// if no available appointment times are found, send an empty list instead of nil/null
	if availableAppointments == nil || len(availableAppointments) < 1 {
		availableAppointments = make([]types.AppointmentTime, 0)
	}

	utils.WriteDataResponse(res, "success", availableAppointments, http.StatusOK)
}

func postAppointmentHandler(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	requestBody, err := io.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		log.Println(err)
		utils.WriteErrorResponse(res, "failed to process client request", http.StatusInternalServerError)
		return
	}

	var appointment types.AppointmentRequest
	if err = json.Unmarshal(requestBody, &appointment); err != nil {
		log.Println(err)
		utils.WriteErrorResponse(res, "failed to process client request", http.StatusInternalServerError)
		return
	}

	convertedTrainerId, err := strconv.ParseUint(appointment.TrainerId, 10, 32)
	if len(appointment.TrainerId) < 1 || err != nil {
		utils.WriteErrorResponse(res, "invalid trainer_id", http.StatusBadRequest)
		return
	}

	convertedUserId, err := strconv.ParseUint(appointment.UserId, 10, 32)
	if len(appointment.UserId) < 1 || err != nil {
		utils.WriteErrorResponse(res, "invalid user_id", http.StatusBadRequest)
		return
	}

	if !utils.IsValidAppointmentTime(appointment.StartsAt, appointment.EndsAt) {
		utils.WriteErrorResponse(res, "invalid appointment time", http.StatusBadRequest)
		return
	}

	// Appointment json tags are different from AppointmentRequest json tags
	// the difference means that a new struct instance is needed after unmarshaling the request
	newAppointment := &types.Appointment{
		TrainerId: uint(convertedTrainerId),
		UserId:    uint(convertedUserId),
		AppointmentTime: types.AppointmentTime{
			StartsAt: appointment.StartsAt,
			EndsAt:   appointment.EndsAt,
		},
	}

	// isAvailable := utils.CheckIfAppointmentTimeIsAvailable(newAppointment.TrainerId, newAppointment.StartsAt, newAppointment.EndsAt)
	isAvailable := models.IsAppointmentTimeAvailable(newAppointment.TrainerId, newAppointment.StartsAt, newAppointment.EndsAt)
	if !isAvailable {
		responseData := struct {
			StartsAt string `json:"starts_at"`
		}{
			StartsAt: fmt.Sprintf("An appointment is not available at %v", newAppointment.StartsAt),
		}
		utils.WriteDataResponse(res, "fail", responseData, http.StatusBadRequest)
		return
	}

	// // save request into DB
	// insertSQL := `INSERT INTO appointments(trainer_id, user_id, starts_at, ends_at) VALUES (?, ?, ?, ?)`
	// statement, err := database.Prepare(insertSQL)
	// if err != nil {
	// 	log.Println(err)
	// 	utils.WriteErrorResponse(res, "failed while creating new appointment", http.StatusInternalServerError)
	// 	return
	// }

	// startsAt, _ := time.Parse(time.RFC3339, appointment.StartsAt)
	// endsAt, _ := time.Parse(time.RFC3339, appointment.EndsAt)

	// insertResult, err := statement.Exec(appointment.TrainerId, appointment.UserId, startsAt, endsAt)
	// if err != nil {
	// 	log.Println(err)
	// 	utils.WriteErrorResponse(res, "failed while creating new appointment", http.StatusInternalServerError)
	// 	return
	// }
	// newAppointmentId, _ := insertResult.LastInsertId()
	// newAppointment.Id = uint(newAppointmentId)
	newAppointment.Id, err = models.InsertAppointment(*newAppointment)
	if err != nil {
		log.Println(err)
		utils.WriteErrorResponse(res, "failed to create appointment", http.StatusInternalServerError)
		return
	}

	utils.WriteDataResponse(res, "success", newAppointment, http.StatusCreated)
}

func getTrainerAppointmentsHandler(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	trainerId := params.ByName("id")
	if len(trainerId) < 1 {
		utils.WriteErrorResponse(res, "trainer_id is required", http.StatusBadRequest)
		return
	}

	// rows, err := database.Query(`SELECT * FROM appointments WHERE trainer_id = ?`, trainerId)
	// if err != nil {
	// 	log.Println(err)
	// 	utils.WriteErrorResponse(res, "failed to find appointments", http.StatusInternalServerError)
	// 	return
	// }
	// defer rows.Close()

	// iterate over result rows and create appointment instances from row data
	// var trainerAppointments []types.Appointment
	// for rows.Next() {
	// 	var appointment types.Appointment
	// 	if err := rows.Scan(
	// 		&appointment.Id,
	// 		&appointment.TrainerId,
	// 		&appointment.UserId,
	// 		&appointment.StartsAt,
	// 		&appointment.EndsAt,
	// 	); err != nil {
	// 		log.Println(err)
	// 		utils.WriteErrorResponse(res, "failed while reading appointment record data", http.StatusInternalServerError)
	// 		return
	// 	}

	// 	appointment.AppointmentTime.StartsAt = utils.ConvertDateTimeToRFC3339(appointment.AppointmentTime.StartsAt, datetime_format_without_t)
	// 	appointment.AppointmentTime.EndsAt = utils.ConvertDateTimeToRFC3339(appointment.AppointmentTime.EndsAt, datetime_format_without_t)
	// 	trainerAppointments = append(trainerAppointments, appointment)
	// }

	// if err = rows.Err(); err != nil {
	// 	log.Println(err)
	// 	utils.WriteErrorResponse(res, "failed while reading appointment record data", http.StatusInternalServerError)
	// 	return
	// }

	trainerAppointments, err := models.GetTrainerAppointments(trainerId)
	if err != nil {
		log.Println(err)
		utils.WriteErrorResponse(res, "failed to retrieve trainer's appointments", http.StatusInternalServerError)
		return
	}

	// if no appointments were found, return an empty list instead of nil/null
	if trainerAppointments == nil || len(trainerAppointments) < 1 {
		trainerAppointments = make([]types.Appointment, 0)
	}

	utils.WriteDataResponse(res, "success", trainerAppointments, http.StatusOK)
}

func loadStartingData() error {
	startingData, err := os.ReadFile("./appointments.json")
	if err != nil {
		return fmt.Errorf("unable to open starting appointments file. %w", err)
	}

	var startingAppointments []types.Appointment
	err = json.Unmarshal(startingData, &startingAppointments)
	if err != nil {
		return fmt.Errorf("failed to unmarshal starting appointment JSON. %w", err)
	}

	// insertSQL := `INSERT INTO appointments(id, trainer_id, user_id, starts_at, ends_at) VALUES (?, ?, ?, ?, ?)`
	// statement, err := database.Prepare(insertSQL)
	// if err != nil {
	// 	return fmt.Errorf("unable to prepare insert statement for starting data. %w", err)
	// }
	for _, appointment := range startingAppointments {
		startsAtUtcTime, _ := time.Parse(time.RFC3339, appointment.AppointmentTime.StartsAt)
		endsAtUtcTime, _ := time.Parse(time.RFC3339, appointment.AppointmentTime.EndsAt)
		appointment.AppointmentTime.StartsAt = startsAtUtcTime.Format(time.RFC3339)
		appointment.AppointmentTime.EndsAt = endsAtUtcTime.Format(time.RFC3339)
		// _, err = statement.Exec(appointment.Id, appointment.TrainerId, appointment.UserId, startsAtUtcTime, endsAtUtcTime)
		// if err != nil {
		// 	log.Printf("Failed to save in DB starting appointment with ID: %v", appointment.Id)
		// }
		_, err = models.InsertAppointment(appointment)
		if err != nil {
			log.Printf("Failed to save in DB starting appointment with ID: %v", appointment.Id)
		}
	}

	return nil
}

// func initializeDatabase() error {
// 	os.Remove("./db/api-test-sqlite.db")

// 	file, err := os.Create("./db/api-test-sqlite.db")
// 	if err != nil {
// 		return fmt.Errorf("unable to create test database. %w", err)
// 	}
// 	defer file.Close()

// 	database, err = sql.Open("sqlite3", "./db/api-test-sqlite.db")
// 	if err != nil {
// 		return fmt.Errorf("unable to open test database. %w", err)
// 	}

// 	createTableSQL := `CREATE TABLE appointments (
// 	"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
// 	"trainer_id" integer,
// 	"user_id" integer,
// 	"starts_at" integer,
// 	"ends_at" integer);`
// 	statement, err := database.Prepare(createTableSQL)
// 	if err != nil {
// 		return fmt.Errorf("unable to create appointments table. %w", err)
// 	}
// 	statement.Exec()

// 	return nil
// }

func main() {
	var err error
	fmt.Println("Starting API server...")

	// err = initializeDatabase()
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	err = models.InitDB()
	if err != nil {
		log.Fatalln(err)
	}

	err = loadStartingData()
	if err != nil {
		log.Fatalln(err)
	}

	headerMiddleware := middleware.HeaderMiddleware
	requiredParamsMiddleware := middleware.RequiredParameters

	router := httprouter.New()
	router.GET("/appointments", headerMiddleware(requiredParamsMiddleware(getAvailableAppointmentsHandler, "trainer_id", "starts_at", "ends_at")))
	router.POST("/appointments", headerMiddleware(postAppointmentHandler))
	router.GET("/trainers/:id/appointments", headerMiddleware(getTrainerAppointmentsHandler))
	log.Fatal(http.ListenAndServe(":8081", router))
}
