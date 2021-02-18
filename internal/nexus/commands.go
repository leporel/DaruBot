package nexus

import (
	"DaruBot/pkg/nexus"
	"google.golang.org/protobuf/proto"
)

type Command struct {
	Cmd proto.Message
}

func (c *Command) GetType() nexus.PayloadType {
	return newType(c.Cmd)
}

func (c *Command) GetPayload() interface{} {
	return c.Cmd
}

type Response struct {
	Rsp proto.Message
}

func (r *Response) GetPayload() interface{} {
	return r.Rsp
}
