package worker

import (
	"context"

	"github.com/hack-pad/safejs"
)

type GlobalSelf struct {
	self safejs.Value
	port *messagePort
}

func Self() (*GlobalSelf, error) {
	self, err := safejs.Global().Get("self")
	if err != nil {
		return nil, err
	}
	port, err := wrapMessagePort(self)
	if err != nil {
		return nil, err
	}
	return &GlobalSelf{
		self: self,
		port: port,
	}, nil
}

func (s *GlobalSelf) PostMessage(message safejs.Value, transfers []safejs.Value) error {
	return s.port.PostMessage(message, transfers)
}

func (s *GlobalSelf) Listen(ctx context.Context) (<-chan MessageEvent, error) {
	return s.port.Listen(ctx)
}

func (s *GlobalSelf) Close() error {
	_, err := s.self.Call("close")
	return err
}

func (s *GlobalSelf) Name() (string, error) {
	name, err := s.self.Get("name")
	if err != nil {
		return "", err
	}
	return name.String()
}
