package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/vincent-gonzalez/future-backend-homework-project/src/middleware"
	"github.com/vincent-gonzalez/future-backend-homework-project/src/models"
	"github.com/vincent-gonzalez/future-backend-homework-project/src/types"
	"github.com/vincent-gonzalez/future-backend-homework-project/src/utils"

	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
)

// getAvailableAppointmentsHandler returns a list of available appointment
// times for a trainer between two dates
func getAvailableAppointmentsHandler(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	trainerId := req.URL.Query().Get("trainer_id")
	startsAtDate := req.URL.Query().Get("starts_at")
	endsAtDate := req.URL.Query().Get("ends_at")

	startsAtUtcTime, _ := time.Parse(time.RFC3339, startsAtDate)
	endsAtUtcTime, _ := time.Parse(time.RFC3339, endsAtDate)

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

// postAppointmentHandler inserts a new appointment into the database
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

	newAppointment.Id, err = models.InsertAppointment(*newAppointment)
	if err != nil {
		log.Println(err)
		utils.WriteErrorResponse(res, "failed to create appointment", http.StatusInternalServerError)
		return
	}

	utils.WriteDataResponse(res, "success", newAppointment, http.StatusCreated)
}

// getTrainerAppointmentsHandler returns a list of scheduled appointments for a trainer
func getTrainerAppointmentsHandler(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	trainerId := params.ByName("id")
	if len(trainerId) < 1 {
		utils.WriteErrorResponse(res, "trainer_id is required", http.StatusBadRequest)
		return
	}

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

	for _, appointment := range startingAppointments {
		startsAtUtcTime, _ := time.Parse(time.RFC3339, appointment.AppointmentTime.StartsAt)
		endsAtUtcTime, _ := time.Parse(time.RFC3339, appointment.AppointmentTime.EndsAt)
		appointment.AppointmentTime.StartsAt = startsAtUtcTime.Format(time.RFC3339)
		appointment.AppointmentTime.EndsAt = endsAtUtcTime.Format(time.RFC3339)

		_, err = models.InsertAppointment(appointment)
		if err != nil {
			log.Printf("Failed to save in DB starting appointment with ID: %v", appointment.Id)
		}
	}

	return nil
}

func main() {
	var err error
	fmt.Println("Starting API server...")

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
