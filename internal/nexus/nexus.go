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
	SetListener(CommandListener)
}

type Module interface {
	Init(CommandListener) error
	Name() models.NexusModuleName
	ListenForType(models.MessageType) bool
	Send(Message) error
}

type Message interface {
	Type() models.MessageType
}

type CommandListener func(*models.Command) (*models.Response, error)

type nexus struct {
	modules  []Module
	listener CommandListener
	log      logger.Logger
	cmdLock  *sync.Mutex
}

func NewNexus(lg logger.Logger) Nexus {
	n := &nexus{
		modules:  nil,
		listener: nil,
		log:      lg,
		cmdLock:  &sync.Mutex{},
	}

	return n
}

func (n *nexus) ReceiveCommand(cmd *models.Command) (*models.Response, error) {
	if n.listener != nil {
		n.cmdLock.Lock()
		defer n.cmdLock.Unlock()
		return n.listener(cmd)
	}
	return nil, fmt.Errorf("commands are not listening")
}

func (n *nexus) SetListener(cl CommandListener) {
	n.listener = cl
}

func (n *nexus) Register(module Module) error {
	if err := module.Init(n.ReceiveCommand); err != nil {
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
