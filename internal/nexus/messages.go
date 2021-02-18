package nexus

import (
	"DaruBot/internal/nexus/pb/schema/gen"
	"DaruBot/pkg/nexus"
	"google.golang.org/protobuf/proto"
)

var (
	MessageTypeLog = newType(&gen.Log{})
)

type Message struct {
	Type nexus.PayloadType
	Msg  proto.Message
}

func (m *Message) GetType() nexus.PayloadType {
	return m.Type
}

func (m *Message) GetPayload() interface{} {
	return m.Msg
}
