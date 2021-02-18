package nexus

import (
	"DaruBot/pkg/nexus"
	"fmt"
	"google.golang.org/protobuf/proto"
)

func newType(pm proto.Message) nexus.PayloadType {
	return nexus.PayloadType(fmt.Sprint(pm.ProtoReflect().Descriptor().Index()))
}
