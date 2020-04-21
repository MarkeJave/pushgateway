package tcp_server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/adler32"
	"net"

	. "github.com/satori/go.uuid"
)

type Package struct {
	_size  		uint32

	/// 36 bytes
	_id       []byte

	_kind     uint32
	_checksum uint32

	_body  		[]byte
	_error 		net.Error
}

// NewMessage create a new message
func NewResponse(id []byte, kind uint32, body []byte) *Package {
	pkg := &Package{
		_size:     uint32(len(body)) + 4 + 4,
		_id:       id,
		_kind:     kind,
		_checksum: checksum(id, kind, body),
		_body:     body,
	}

	return pkg
}

func NewStateResponse(id []byte, code int) (*Package, error) {
	body := new(bytes.Buffer)
	err := binary.Write(body, binary.LittleEndian, code)

	if err != nil {
		return nil, err
	}
	return NewResponse(id, KindResponse, body.Bytes()), nil
}

func NewSuccessResponse(id []byte) (*Package, error) {
	return NewStateResponse(id, CodeSuccess)
}

func NewFailureResponse(id []byte) (*Package, error) {
	return NewStateResponse(id, CodeFailed)
}

// NewMessage create a new message
func NewPackage(kind uint32, body []byte) *Package {
	id := NewV4().Bytes()
	pkg := &Package{
		_size:     uint32(len(body)) + 4 + 4,
		_id:       id,
		_kind:     kind,
		_checksum: checksum(id, kind, body),
		_body:     body,
	}

	return pkg
}

// GetData get message data
func (pkg *Package) GetId() []byte {
	return pkg._id
}

// GetData get message data
func (pkg *Package) GetBody() []byte {
	return pkg._body
}

// Verify verify checksum
func (pkg *Package) Verify() bool {
	return pkg._checksum == pkg.Checksum()
}

func (pkg *Package) Checksum() uint32 {
	return checksum(pkg._id, pkg._kind, pkg._body)
}

func checksum(id []byte, kind uint32, body []byte) uint32 {
	if body == nil  {
		return 0
	}

	data := new(bytes.Buffer)

	err := binary.Write(data, binary.LittleEndian, id)
	if err != nil {
		return 0
	}

	err = binary.Write(data, binary.LittleEndian, kind)
	if err != nil {
		return 0
	}

	err = binary.Write(data, binary.LittleEndian, body)
	if err != nil {
		return 0
	}

	checksum := adler32.Checksum(data.Bytes())
	return checksum
}

func (pkg *Package) String() string {
	return fmt.Sprintf("Size=%d Kind=%d DataLen=%d Checksum=%d", pkg._size, pkg._kind, len(pkg._body), pkg._checksum)
}
