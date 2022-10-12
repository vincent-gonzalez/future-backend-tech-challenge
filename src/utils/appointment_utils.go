package utils

import (
	"github.com/vincent-gonzalez/future-backend-homework-project/src/types"
	"time"
)

func FindTrainerAvailability(scheduledAppointments []types.AppointmentTime, startDateTime string, endDateTime string) []types.AppointmentTime {
	unavailableAppointments := make(map[int64]bool, len(scheduledAppointments))

	// load scheduled appointments into map to check against for open slots
	for _, appointment := range scheduledAppointments {
		key, _ := time.Parse(time.RFC3339, appointment.StartsAt)
		unavailableAppointments[key.Unix()] = true
	}

	startTime, _ := time.Parse(time.RFC3339, startDateTime)
	endTime, _ := time.Parse(time.RFC3339, endDateTime)

	// gracefully ignore start times that begin before business hours
	if startTime.Hour() < 8 {
		startTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 8, 0, 0, 0, startTime.Location())
	}

	// find open availability based on already scheduled appointments
	var availableAppointments []types.AppointmentTime
	for currentTime := startTime; !currentTime.After(endTime); currentTime = currentTime.Add(time.Minute * 30) {
		// if not found in map, then time slot is available
		if _, ok := unavailableAppointments[currentTime.Unix()]; !ok {
			openTimeslotEnd := currentTime.Add(time.Minute * 30)
			openTimeslot := types.AppointmentTime{
				StartsAt: currentTime.Format(time.RFC3339),
				EndsAt:   openTimeslotEnd.Format(time.RFC3339),
			}
			availableAppointments = append(availableAppointments, openTimeslot)
		}

		// move to next day in date range after reaching the end of the business day
		if currentTime.Hour() >= 16 && currentTime.Minute() >= 30 {
			currentTime = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day()+1, 7, 30, 0, 0, currentTime.Location())
		}
	}

	return availableAppointments
}

func IsValidAppointmentTime(startsAt string, endsAt string) bool {
	startsAtTime, _ := time.Parse(time.RFC3339, startsAt)
	endsAtTime, _ := time.Parse(time.RFC3339, endsAt)

	if startsAtTime.Hour() < 8 || startsAtTime.Hour() >= 17 {
		return false
	}

	if endsAtTime.Hour() > 17 || endsAtTime.Hour() <= 8 {
		return false
	}

	appointmentDuration := endsAtTime.Sub(startsAtTime)
	if appointmentDuration.Minutes() != 30 {
		return false
	}

	return true
}
