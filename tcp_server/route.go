package tcp_server

type Route struct {
	_kind uint32
	_handler func(*Session, *Package)
}

func NewRoute(kind uint32, handler func(*Session, *Package)) *Route {
	route := &Route{
		_kind: kind,
		_handler: handler,
	}
	return route
}