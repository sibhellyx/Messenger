package chaterrors

import "errors"

var (
	ErrFailedCreateChat         = errors.New("failed create chat")
	ErrUserNotFound             = errors.New("user not found")
	ErrChatNotFound             = errors.New("chat not found")
	ErrAlreadyParticipant       = errors.New("user is already a participant of this chat")
	ErrCreatingChatWithYourself = errors.New("failed creating a chat with yourself")
	ErrDeletingAllParticipants  = errors.New("failed delete all participants")
	ErrDeletingChat             = errors.New("failed delete chat")
	ErrFullChat                 = errors.New("chat is full, there are no free spots for participants")
	ErrCheckTwoUsersNotFound    = errors.New("one or both users not found")
	ErrFailedCheckDirectedChat  = errors.New("failed to check directed chat")
	ErrFailedGetChat            = errors.New("failed to get chat")
	ErrFailedUpdateChat         = errors.New("failed to update chat")
	ErrFailedGetParticipant     = errors.New("failed get participant")
)
