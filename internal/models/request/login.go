package request

type LoginRequest struct {
	Tgname   string `json:"tg_username"`
	Password string `json:"password"`
}

type LoginParams struct {
	UserAgent string
	LastIp    string
}
