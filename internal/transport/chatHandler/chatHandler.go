package chathandler

type ChatServiceInterface interface {
}

type ChatHandler struct {
	service ChatServiceInterface
}

func NewChatHandler(service ChatServiceInterface) *ChatHandler {
	return &ChatHandler{
		service: service,
	}
}
