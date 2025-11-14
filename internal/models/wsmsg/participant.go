package wsmsg

type ParticipantAction string

const (
	Add         ParticipantAction = "added"
	Entered     ParticipantAction = "entered"
	Leaved      ParticipantAction = "leaved"
	Removed     ParticipantAction = "removed"
	ChatDeleted ParticipantAction = "deleted"
)

type ParticipantMsg struct {
	ChatID uint              `json:"chat_id"`
	UserID uint              `json:"user_id,omitempty"`
	Type   string            `json:"type"`
	Action ParticipantAction `json:"action"`
}
