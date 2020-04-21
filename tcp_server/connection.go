package tcp_server

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	levels "github.com/go-kit/kit/log/deprecated_levels"
	"io"
	"net"
	"time"
)

// Connection wrap net.Connection
type Connection struct {
	_sid      string
	_raw      net.Conn
	_data     chan []byte
	_done     chan error
	_timer    *time.Timer
	_name     string
	_pkg      chan *Package
	_interval time.Duration
	_timeout  time.Duration
}

// GetName Get conn name
func (c *Connection) GetName() string {
	return c._name
}

// NewConn create new conn
func NewConn(c net.Conn, interval time.Duration, timeout time.Duration) *Connection {
	conn := &Connection{
		_raw:      c,
		_data:     make(chan []byte, 1000),
		_done:     make(chan error),
		_pkg:      make(chan *Package, 1000),
		_interval: interval,
		_timeout:  timeout,
	}

	conn._name = c.RemoteAddr().String()
	conn._timer = time.NewTimer(conn._interval)

	if conn._interval == 0 {
		conn._timer.Stop()
	}

	return conn
}

// Close close connection
func (c *Connection) Close() error {
	if !c._timer.Stop() {
		return errors.New(fmt.Sprintf("failed to stop timer of connection: %s", c._raw.RemoteAddr().String()))
	}
	return c._raw.Close()
}

// SendPackage send Package
func (c *Connection) SendPackage(pkg *Package) error {
	data, err := Encode(pkg)
	if err != nil {
		return err
	}

	c._data <- data
	return nil
}

// SendPackage send Package
func (c *Connection) SendResponse(id []byte, body []byte) error {
	pkg := NewResponse(id, KindResponse, body)
	data, err := Encode(pkg)
	if err != nil {
		return err
	}

	c._data <- data
	return nil
}

// writeCoroutine write coroutine
func (c *Connection) writeCoroutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case data := <-c._data:
			if data == nil {
				continue
			}

			if _, err := c._raw.Write(data); err != nil {
				c._done <- err
			}

		case <-c._timer.C:
			data := make([]byte, 0)

			pkg := NewPackage(KindHeartbeat, data)
			err := c.SendPackage(pkg)
			if err != nil {
				levels.WarnValue("failed to send heartbeat for connection: " + c._raw.RemoteAddr().String())
			}

			// 设置心跳timer
			if c._interval > 0 {
				c._timer.Reset(c._interval)
			}
		}
	}
}

// readCoroutine read coroutine
func (c *Connection) readCoroutine(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		default:
			// 设置超时
			if c._interval > 0 {
				err := c._raw.SetReadDeadline(time.Now().Add(c._timeout))
				if err != nil {
					c._done <- err
					continue
				}
			}
			// 读取长度
			header := make([]byte, 4)
			_, err := io.ReadFull(c._raw, header)
			if err != nil {
				c._done <- err
				continue
			}

			reader := bytes.NewReader(header)

			var size int32
			err = binary.Read(reader, binary.LittleEndian, &size)
			if err != nil {
				c._done <- err
				continue
			}

			// 读取数据
			data := make([]byte, size)
			_, err = io.ReadFull(c._raw, data)
			if err != nil {
				c._done <- err
				continue
			}

			// 解码
			pkg, err := Decode(data)
			if err != nil {
				c._done <- err
				continue
			}

			if pkg._kind == KindHeartbeat {
				continue
			}

			c._pkg <- pkg
		}
	}
}
