package authhandler

import authservice "github.com/sibhellyx/Messenger/internal/services/authService"

type AuthHandler struct {
	service *authservice.AuthService
}

func NewAuthHandler(service *authservice.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}
