package request

type ParticipantAddRequest struct {
	Id        string `json:"chat_id"`
	NewUserId string `json:"user_id"`
}
