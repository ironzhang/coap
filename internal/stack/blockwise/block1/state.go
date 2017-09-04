package block1

import (
	"fmt"

	"github.com/ironzhang/coap/internal/stack/base"
)

type state interface {
	Name() string
	OnStart()
	OnFinish()

	Update()
	OnAckTimeout(base.Message)
	Recv(base.Message) error
	Send(base.Message) error
}

type machine struct {
	current state
	states  map[string]state
}

func (m *machine) Init(states ...state) {
	m.states = make(map[string]state)
	for i, s := range states {
		if i == 0 {
			m.current = s
			m.current.OnStart()
		}
		m.states[s.Name()] = s
	}
}

func (m *machine) SetState(name string) {
	s, ok := m.states[name]
	if !ok {
		panic(fmt.Errorf("state(%s) not found", name))
	}
	m.current.OnFinish()
	m.current = s
	m.current.OnStart()
}

func (m *machine) Update() {
	m.current.Update()
}

func (m *machine) OnAckTimeout(msg base.Message) {
	m.current.OnAckTimeout(msg)
}

func (m *machine) Recv(msg base.Message) error {
	return m.current.Recv(msg)
}

func (m *machine) Send(msg base.Message) error {
	return m.current.Send(msg)
}
