package tcp_server

import (
	"container/list"
	"context"
	"errors"
	"net"
	"sync"
	"time"
)


// SocketService struct
type SocketService struct {
	_onReceivePackage 	func(*Session, *Package)
	_onReceiveResponse 	func(*Session, *Package)
	_onConnect    		func(*Session)
	_onDisconnect 		func(*Session, error)
	_routes				*sync.Map
	_sessions     		*sync.Map
	_interval      time.Duration
	_timeout       time.Duration
	_listenAddress string
	_status        int
	_listener      net.Listener
	_stop          chan error
}

// NewSocketService create a new socket service
func NewSocketService(listenAddress string) (*SocketService, error) {
	l, err := net.Listen("tcp", listenAddress)

	if err != nil {
		return nil, err
	}

	s := &SocketService{
		_sessions:     &sync.Map{},
		_stop:          make(chan error),
		_interval:      0 * time.Second,
		_timeout:       0 * time.Second,
		_listenAddress: listenAddress,
		_status:        StateInitialized,
		_listener:      l,
	}

	return s, nil
}
// RegMessageHandler register message handler
func (s *SocketService) onReceivePackage(session *Session, pkg *Package) {
	v, _ := s._routes.Load(pkg._kind)
	if v == nil {
		return
	}

	routes := v.(*list.List)
	if routes.Len() == 0 {
		return
	}

	for e := routes.Front(); e != nil; e = e.Next() {
		e.Value.(*Route)._handler(session, pkg)
	}
}

// RegMessageHandler register message handler
func (s *SocketService) addRoute(kind uint32, handler func(*Session, *Package)) {
	route := NewRoute(kind, handler)

	v, _ := s._routes.Load(kind)
	if v == nil {
		routes := list.New()
		routes.PushBack(route)

		s._routes.Store(kind, routes)
		return
	}

	routes := v.(*list.List)
	routes.PushBack(route)
}

// RegConnectHandler register connect handler
func (s *SocketService) RegisterReceiveResponseHnadler(handler func(*Session, *Package)) {
	s._onReceiveResponse = handler
}

func (s *SocketService) RegisterReceivePackageHnadler(handler func(*Session, *Package)) {
	s._onReceivePackage = handler
}

// RegConnectHandler register connect handler
func (s *SocketService) RegisterConnectHandler(handler func(*Session,)) {
	s._onConnect = handler
}

// RegDisconnectHandler register disconnect handler
func (s *SocketService) RegisterDisconnectHandler(handler func(*Session, error)) {
	s._onDisconnect = handler
}

// Run socket service
func (s *SocketService) run() {
	s._status = StateRunning
	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		s._status = StateStop
		cancel()
		_ = s._listener.Close()
	}()

	go s.acceptHandler(ctx)

	for {
		select {
		case <-s._stop:
			return
		}
	}
}

func (s *SocketService) acceptHandler(ctx context.Context) {
	for {
		c, err := s._listener.Accept()
		if err != nil {
			s._stop <- err
			return
		}

		go s.connectHandler(ctx, c)
	}
}

func (s *SocketService) connectHandler(ctx context.Context, c net.Conn) {
	conn := NewConn(c, s._interval, s._timeout)
	session := NewSession(conn)
	s._sessions.Store(session.GetSessionID(), session)

	connctx, cancel := context.WithCancel(ctx)

	defer func() {
		cancel()
		_ = conn.Close()
		s._sessions.Delete(session.GetSessionID())
	}()

	go conn.readCoroutine(connctx)
	go conn.writeCoroutine(connctx)

	if s._onConnect != nil {
		s._onConnect(session)
	}

	for {
		select {
		case err := <-conn._done:
			if s._onDisconnect != nil {
				s._onDisconnect(session, err)
			}
			return
		case pkg := <-conn._pkg:
			if pkg._kind == KindResponse {
				s._onReceiveResponse(session, pkg)
			} else {
				s.onReceivePackage(session, pkg)
			}
		}
	}
}

// GetStatus get socket service status
func (s *SocketService) GetStatus() int {
	return s._status
}

// Stop stop socket service with reason
func (s *SocketService) Stop(reason string) {
	s._stop <- errors.New(reason)
}

// SetHeartBeat set heart beat
func (s *SocketService) SetHeartBeat(hbInterval time.Duration, hbTimeout time.Duration) error {
	if s._status == StateRunning {
		return errors.New("Can't set heart beat on service running")
	}

	s._interval = hbInterval
	s._timeout = hbTimeout

	return nil
}

// GetConnsCount get connect count
func (s *SocketService) GetConnectionsCount() int {
	var count int
	s._sessions.Range(func(k, v interface{}) bool {
		count++
		return true
	})
	return count
}

// Unicast Unicast with session ID
func (s *SocketService) Unicast(sid string, pkg *Package) {
	v, ok := s._sessions.Load(sid)
	if ok {
		session := v.(*Session)
		err := session.GetConn().SendPackage(pkg)
		if err != nil {
			return
		}
	}
}

// Broadcast Broadcast to all connections
func (s *SocketService) Broadcast(pkg *Package) {
	s._sessions.Range(func(k, v interface{}) bool {
		s := v.(*Session)
		if err := s.GetConn().SendPackage(pkg); err != nil {
			// log.Println(err)
		}
		return true
	})
}
