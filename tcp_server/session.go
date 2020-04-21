package tcp_server

import (
	uuid "github.com/satori/go.uuid"
)

// Session struct
type Session struct {
	_id      string
	_uid      string
	_conn     *Connection
	_settings map[string]interface{}
}

// NewSession create a new session
func NewSession(conn *Connection) *Session {
	session := &Session{
		_id: uuid.NewV4().String(),
		_uid: "",
		_conn: conn,
		_settings: make(map[string]interface{}),
	}

	return session
}

// GetSessionID get session ID
func (s *Session) GetSessionID() string {
	return s._id
}

// BindUserID bind a user ID to session
func (s *Session) BindUserID(uid string) {
	s._uid = uid
}

// GetUserID get user ID
func (s *Session) GetUserID() string {
	return s._uid
}

// GetConn get zero.Connection pointer
func (s *Session) GetConn() *Connection {
	return s._conn
}

// SetConn set a zero.Connection to session
func (s *Session) SetConn(conn *Connection) {
	s._conn = conn
}

// GetSetting get setting
func (s *Session) GetSetting(key string) interface{} {
	if v, ok := s._settings[key]; ok {
		return v
	}

	return nil
}

// SetSetting set setting
func (s *Session) SetSetting(key string, value interface{}) {
	s._settings[key] = value
}
