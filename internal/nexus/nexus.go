package nexus

import (
	"DaruBot/internal/models"
	"DaruBot/pkg/logger"
	"fmt"
	"sync"
)

type Nexus interface {
	logger.Hook

	Register(Module) error
	Send(Message) error
}

type Module interface {
	Init(CommandHandlerFunc) error
	Name() models.NexusModuleName
	ListenForType(models.MessageType) bool
	Send(Message) error
	Stop() error
}

type Message interface {
	Type() models.MessageType
}

type CommandHandlerFunc func(*models.Command) (*models.Response, error)

type nexus struct {
	modules []Module
	handler CommandHandlerFunc
	log     logger.Logger
	cmdLock *sync.Mutex
}

func NewNexus(lg logger.Logger, cl CommandHandlerFunc) Nexus {
	n := &nexus{
		modules: nil,
		handler: cl,
		log:     lg,
		cmdLock: &sync.Mutex{},
	}

	return n
}

func (n *nexus) receiveCommand(cmd *models.Command) (*models.Response, error) {
	if n.handler != nil {
		n.cmdLock.Lock()
		defer n.cmdLock.Unlock()
		return n.handler(cmd)
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

	msg := &models.Notification{
		Msg: fmt.Sprintf("[%s] [%s]\n%s",
			hd.Level, hd.Time.Format("01.02 15:04:05"), hd.Message),
		Raw: hd,
	}

	switch {
	case hd.Level > logger.WarnLevel:
		msg.Msg = models.NotifyKindLog
	case hd.Level == logger.WarnLevel:
		msg.Msg = models.NotifyKindWarning
	default:
		msg.Msg = models.NotifyKindError
	}

	return n.Send(msg)
}

func (n *nexus) getModule(name models.NexusModuleName) Module {
	for _, module := range n.modules {
		if module.Name() == name {
			return module
		}
	}
	return nil
}
