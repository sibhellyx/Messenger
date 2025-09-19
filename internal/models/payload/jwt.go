package payload

type JwtPayload struct {
	UserId string `json:"user_id"`
	Uuid   string `json:"uuid"`
}
