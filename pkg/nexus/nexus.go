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

type Module interface {
	Init(CommandHandlerFunc) error
	Name() NexusModuleName
	ListenForType(MessageType) bool
	Send(Message) error
	Stop() error
}

type Message interface {
	Type() MessageType
}

type CommandHandlerFunc func(context.Context, *Command) (*Response, error)

type nexus struct {
	modules      []Module
	handler      CommandHandlerFunc
	singleCaller singleflight.Group
	log          logger.Logger
}

func NewNexus(cl CommandHandlerFunc, lg logger.Logger) Nexus {
	n := &nexus{
		modules:      nil,
		handler:      cl,
		singleCaller: singleflight.Group{},
		log:          lg,
	}

	return n
}

func (n *nexus) receiveCommand(ctx context.Context, cmd *Command) (*Response, error) {
	if n.handler != nil {

		req := func() (interface{}, error) {
			defer tools.Recover(n.log)

			return n.handler(ctx, cmd)
		}

		result, err, _ := n.singleCaller.Do(string(cmd.Type), req)
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, fmt.Errorf("result is nil")
		}
		rs, ok := result.(*Response)
		if !ok {
			return nil, fmt.Errorf("responce has wrong type")
		}

		return rs, nil
	}
	return nil, fmt.Errorf("commands are not listening")
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
		if module.ListenForType(msg.Type()) {
			return module.Send(msg)
		}
	}

	return nil
}

func (n *nexus) Fire(hd *logger.HookData) error {

	msg := &Notification{
		Msg: fmt.Sprintf("[%s] [%s]\n%s",
			hd.Level, hd.Time.Format("01.02 15:04:05"), hd.Message),
		Raw: hd,
	}

	switch {
	case hd.Level > logger.WarnLevel:
		msg.Msg = NotifyKindLog
	case hd.Level == logger.WarnLevel:
		msg.Msg = NotifyKindWarning
	default:
		msg.Msg = NotifyKindError
	}

	return n.Send(msg)
}

func (n *nexus) getModule(name NexusModuleName) Module {
	for _, module := range n.modules {
		if module.Name() == name {
			return module
		}
	}
	return nil
}
