package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/vincent-gonzalez/future-backend-homework-project/src/types"
	"github.com/vincent-gonzalez/future-backend-homework-project/src/utils"
)

var database *sql.DB
var dateFormat = utils.Datetime_format_without_t

func InitDB() error {
	os.Remove("./db/api-test-sqlite.db")

	file, err := os.Create("./db/api-test-sqlite.db")
	if err != nil {
		return fmt.Errorf("unable to create test database. %w", err)
	}
	defer file.Close()

	database, err = sql.Open("sqlite3", "./db/api-test-sqlite.db")
	if err != nil {
		return fmt.Errorf("unable to open test database. %w", err)
	}

	createTableSQL := `CREATE TABLE appointments (
	"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,
	"trainer_id" integer,
	"user_id" integer,
	"starts_at" integer,
	"ends_at" integer);`
	statement, err := database.Prepare(createTableSQL)
	if err != nil {
		return fmt.Errorf("unable to create appointments table. %w", err)
	}
	statement.Exec()

	return nil
}

func GetScheduledAppointmentsBetweenDates(trainerId string, startsAt time.Time, endsAt time.Time) ([]types.AppointmentTime, error) {
	rows, err := database.Query(`SELECT starts_at, ends_at FROM appointments
	 WHERE trainer_id = ? AND starts_at >= ? AND ends_at <= ?`,
		trainerId, startsAt, endsAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// iterate over scheduled appointments
	var appointments []types.AppointmentTime
	for rows.Next() {
		var appointment types.AppointmentTime
		if err := rows.Scan(
			&appointment.StartsAt,
			&appointment.EndsAt,
		); err != nil {
			return nil, err
		}

		appointment.StartsAt = utils.ConvertDateTimeToRFC3339(appointment.StartsAt, dateFormat)
		appointment.EndsAt = utils.ConvertDateTimeToRFC3339(appointment.EndsAt, dateFormat)
		appointments = append(appointments, appointment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return appointments, nil
}

func GetAppointmentByTrainerAndDate(trainerId uint, startsAt time.Time, endsAt time.Time) (types.Appointment, error) {
	var appointment types.Appointment
	err := database.QueryRow(`SELECT id FROM appointments
	WHERE trainer_id = ? AND starts_at = ? AND ends_at = ?`,
		trainerId, startsAt, endsAt).Scan(&appointment.Id)
	log.Println(appointment)
		return appointment, err
}

func GetTrainerAppointments(trainerId string) ([]types.Appointment, error) {
	rows, err := database.Query(`SELECT * FROM appointments WHERE trainer_id = ?`, trainerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// iterate over result rows and create appointment instances from row data
	var trainerAppointments []types.Appointment
	for rows.Next() {
		var appointment types.Appointment
		if err := rows.Scan(
			&appointment.Id,
			&appointment.TrainerId,
			&appointment.UserId,
			&appointment.StartsAt,
			&appointment.EndsAt,
		); err != nil {
			return nil, err
		}

		appointment.AppointmentTime.StartsAt = utils.ConvertDateTimeToRFC3339(appointment.AppointmentTime.StartsAt, dateFormat)
		appointment.AppointmentTime.EndsAt = utils.ConvertDateTimeToRFC3339(appointment.AppointmentTime.EndsAt, dateFormat)
		trainerAppointments = append(trainerAppointments, appointment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return trainerAppointments, nil
}

func InsertAppointment(appointment types.Appointment) (uint, error) {
	insertSQL := `INSERT INTO appointments(trainer_id, user_id, starts_at, ends_at) VALUES (?, ?, ?, ?)`
	statement, err := database.Prepare(insertSQL)
	if err != nil {
		return 0, err
	}

	startsAt, _ := time.Parse(time.RFC3339, appointment.StartsAt)
	endsAt, _ := time.Parse(time.RFC3339, appointment.EndsAt)

	insertResult, err := statement.Exec(appointment.TrainerId, appointment.UserId, startsAt, endsAt)
	if err != nil {
		return 0, err
	}

	newAppointmentId, _ := insertResult.LastInsertId()
	return uint(newAppointmentId), nil
}
