package nexus

type NexusModuleName string

type MessageType string

const (
	MessageTypeNotification MessageType = "notification"
	MessageTypeEvent        MessageType = "event"
)

type CommandType uint8

type Command struct {
	Type    CommandType
	Payload []byte
}

type NotifyKind string

const (
	NotifyKindLog     = "log"
	NotifyKindError   = "error"
	NotifyKindWarning = "warning"
	NotifyKindAlert   = "alert"
)

type Notification struct {
	Kind NotifyKind
	Msg  string
	Raw  interface{}
}

func (n *Notification) Type() MessageType {
	return MessageTypeNotification
}

type Response struct {
}
