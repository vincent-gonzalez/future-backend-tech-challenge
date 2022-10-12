package types

type Appointment struct {
	Id        uint `json:"id"`
	TrainerId uint `json:"trainer_id"`
	UserId    uint `json:"user_id"`
	AppointmentTime
}
