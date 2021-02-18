package nexus

import (
	"DaruBot/pkg/logger"
	"DaruBot/pkg/tools"
	"context"
	"fmt"
	"golang.org/x/sync/singleflight"
)

type Nexus interface {
	logger.Hook

	Register(Module) error
	Send(Message) error
}

type ModuleName string

type Module interface {
	logger.Hook
	Send(Message) error

	Init(CommandHandlerFunc) error
	Stop() error

	Name() ModuleName
	ListenForType(PayloadType) bool
}

type PayloadType string

type Message interface {
	GetType() PayloadType
	GetPayload() interface{}
}

type Command interface {
	GetType() PayloadType
	GetPayload() interface{}
}

type Response interface {
	GetPayload() interface{}
}

type CommandHandlerFunc func(context.Context, Command) (Response, error)

type nexus struct {
	modules      []Module
	handler      CommandHandlerFunc
	singleCaller singleflight.Group
}

func NewNexus(cl CommandHandlerFunc) Nexus {
	n := &nexus{
		modules:      nil,
		handler:      cl,
		singleCaller: singleflight.Group{},
	}

	return n
}

func (n *nexus) receiveCommand(ctx context.Context, cmd Command) (Response, error) {
	if n.handler != nil {
		req := func() (interface{}, error) {
			defer tools.Recover(nil)

			return n.handler(ctx, cmd)
		}

		result, err, _ := n.singleCaller.Do(string(cmd.GetType()), req)
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, fmt.Errorf("result is nil")
		}
		rs, ok := result.(Response)
		if !ok {
			return nil, fmt.Errorf("responce has wrong type")
		}

		return rs, nil
	}
	return nil, fmt.Errorf("commands are not listening")
}

func (n *nexus) Fire(hd *logger.HookData) error {
	return n.Fire(hd)
}

func (n *nexus) Register(module Module) error {
	if err := module.Init(n.receiveCommand); err != nil {
		return err
	}

	n.modules = append(n.modules, module)

	return nil
}

func (n *nexus) Send(msg Message) error {
	for _, module := range n.modules {
		if module.ListenForType(msg.GetType()) {
			return module.Send(msg)
		}
	}

	return nil
}

func (n *nexus) getModule(name ModuleName) Module {
	for _, module := range n.modules {
		if module.Name() == name {
			return module
		}
	}
	return nil
}
