package main

type Appointment struct {
	TrainerId string `json:"trainer_id"`
	UserId string `json:"user_id"`
	StartsAt string `json:"starts_at"`
	EndsAt string `json:"ends_at"`
}
