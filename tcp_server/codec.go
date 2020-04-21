package tcp_server

import (
	"bytes"
	"encoding/binary"
	"errors"
)

//1. size  	uint32
//2. kind  	uint32
//3. signature uint32
//4. body  	[]byte

// Encode from Package to []byte
func Encode(pkg *Package) ([]byte, error) {
	buffer := new(bytes.Buffer)

	err := binary.Write(buffer, binary.LittleEndian, pkg._size)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.LittleEndian, pkg._id)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.LittleEndian, pkg._kind)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.LittleEndian, pkg._checksum)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buffer, binary.LittleEndian, pkg._body)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

//1. kind  	uint32
//2. signature uint32
//3. body  	[]byte
func Decode(data []byte) (*Package, error) {
	reader := bytes.NewReader(data)

	inherentLength := uint32(32 + 4 + 4)
	size := uint32(len(data))
	if size < inherentLength {
		return nil, errors.New("package is invalid")
	}

	id := make([]byte, 32)
	err := binary.Read(reader, binary.LittleEndian, &id)
	if err != nil {
		return nil, err
	}

	var kind uint32
	err = binary.Read(reader, binary.LittleEndian, &kind)
	if err != nil {
		return nil, err
	}

	var signature uint32
	err = binary.Read(reader, binary.LittleEndian, &signature)
	if err != nil {
		return nil, err
	}

	body := make([]byte, size)
	err = binary.Read(reader, binary.LittleEndian, &body)
	if err != nil {
		return nil, err
	}

	pkg := &Package{}
	pkg._id = id
	pkg._size = size
	pkg._kind = kind
	pkg._checksum = signature
	pkg._body = body

	if !pkg.Verify() {
		return nil, errors.New("package is changed")
	}
	return pkg, nil
}
