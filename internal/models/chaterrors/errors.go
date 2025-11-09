package chaterrors

import "errors"

var (
	// repos layer
	ErrFailedCreateChat                = errors.New("failed create chat")
	ErrUserNotFound                    = errors.New("user not found")
	ErrChatNotFound                    = errors.New("chat not found")
	ErrChatIsDirected                  = errors.New("chat is directed, can't add participant")
	ErrChatIsDirectedEnter             = errors.New("chat is directed, you can't enter to direct chat")
	ErrChatIsPrivateEnter              = errors.New("chat is private, you can't enter to private chat")
	ErrAlreadyParticipant              = errors.New("user is already a participant of this chat")
	ErrNotParticipant                  = errors.New("user is not a participant of this chat")
	ErrCreatingChatWithYourself        = errors.New("failed creating a chat with yourself")
	ErrDeletingAllParticipants         = errors.New("failed delete all participants")
	ErrDeletingChat                    = errors.New("failed delete chat")
	ErrFullChat                        = errors.New("chat is full, there are no free spots for participants")
	ErrCheckTwoUsersNotFound           = errors.New("one or both users not found")
	ErrFailedCheckDirectedChat         = errors.New("failed to check directed chat")
	ErrFailedGetChat                   = errors.New("failed to get chat")
	ErrFailedUpdateChat                = errors.New("failed to update chat")
	ErrFailedGetParticipant            = errors.New("failed get participant")
	ErrFailedGetChats                  = errors.New("failed get chats")
	ErrFailedDeleteParticipant         = errors.New("failed delete participant")
	ErrUserIsOwner                     = errors.New("owner can't leave from chat")
	ErrFailedLeaveDirectedChat         = errors.New("failed leave from directed chat, delete chat")
	ErrFailedRemoveParticipantByMember = errors.New("failed remove user, member can't remove participants")
	ErrFailedRemoveAdminOrOwnerByAdmin = errors.New("failed remove user, admin can't remove another admins or owner")
	ErrFailedUpdateParticipant         = errors.New("failed update participant")
	ErrFailedCheckParticipant          = errors.New("failed to check participant")

	// service layer
	ErrInvalidUser             = errors.New("invalid user_id")
	ErrInvalidChat             = errors.New("invalid chat_id")
	ErrNotPermission           = errors.New("participant doesn't have permission")
	ErrCantUpdaeteDirect       = errors.New("failed update chat, direct chat not updated")
	ErrInvalidNameForSearch    = errors.New("invalid searching name")
	ErrInvalidIdNewParticipant = errors.New("invalid new_participant_id")
	ErrFailedGetParticipants   = errors.New("failed get chat participants")
)
