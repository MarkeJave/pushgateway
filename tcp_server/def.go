package tcp_server

const (
	STUnknown = iota
	StateInitialized
	StateRunning
	StateStop
)

const (
	KindHeartbeat = iota
	KindResponse
)


const (
	CodeSuccess = iota
	CodeFailed
)
