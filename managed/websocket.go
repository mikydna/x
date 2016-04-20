package managed

import (
	"sync"
	"time"
)

import (
	"golang.org/x/net/websocket"
)

var (
	ZeroBytes            = make([]byte, 0)
	DefaultTimeout       = 100 * time.Millisecond
	DefaultCheckInterval = 1 * time.Second
)

type Websockets struct {
	conns     map[string]*websocket.Conn
	callbacks map[string][]chan bool
	done      chan bool

	*sync.Mutex
}

func NewWebsockets() *Websockets {
	managed := &Websockets{
		conns:     make(map[string]*websocket.Conn),
		callbacks: make(map[string][]chan bool),
		done:      make(chan bool),

		Mutex: &sync.Mutex{},
	}

	go managed.check(DefaultCheckInterval)

	return managed
}

func (m *Websockets) Set(id string, conn *websocket.Conn) {
	m.Lock()
	defer m.Unlock()

	if existingConn, exists := m.conns[id]; exists {
		existingConn.Close()
		delete(m.conns, id)
	}

	m.conns[id] = conn
}

func (m *Websockets) Broadcast(b []byte) {
	m.Lock()
	defer m.Unlock()

	for _, conn := range m.conns {
		deadline := time.Now().Add(DefaultTimeout)
		conn.SetWriteDeadline(deadline)
		conn.Write(b)
	}
}

func (m *Websockets) Wait(id string) {
	callback := make(chan bool)
	m.callbacks[id] = append(m.callbacks[id], callback)
	<-callback
}

func (m *Websockets) check(interval time.Duration) {
	ticker := time.NewTicker(interval)
	alive := true
	for alive {
		select {
		case <-ticker.C:
			for connId, conn := range m.conns {
				go func() {
					m.Lock()
					defer m.Unlock()

					deadline := time.Now().Add(DefaultTimeout)
					conn.SetWriteDeadline(deadline)
					if _, err := conn.Write(ZeroBytes); err != nil {
						delete(m.conns, connId)
						m.notify(connId)
						conn.Close() // just in case
					}
				}()
			}

		case <-m.done:
			alive = false
		}
	}
}

func (m *Websockets) notify(id string) {
	m.Lock()
	defer m.Unlock()

	if callbacks, exists := m.callbacks[id]; exists {
		for i, callback := range callbacks {
			select {
			case callback <- true:
			default:
				m.callbacks[id] = append(m.callbacks[id][:i], m.callbacks[id][i+1:]...)
			}
		}
	}
}
