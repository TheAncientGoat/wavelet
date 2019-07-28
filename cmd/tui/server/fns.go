package server

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
)

// FunctionCall is the struct containing parameters to call a function.
type FunctionCall struct {
	Name   string
	Params []FunctionParameter
}

// FunctionParameter is an interface with the Encode method to encode the
// function parameter.
type FunctionParameter interface {
	Encode() ([]byte, error)
}

// Jesus Christ why did I do this
// I could've just used a []byte and functions to convert them, instead of
// doing this stupid OOP pattern

/*
	HEX TYPES
*/

type fnParamHex struct {
	s string
}

// NewHexParameter returns a Hex function parameter
func NewHexParameter(s string) FunctionParameter {
	return &fnParamHex{s}
}

func (h *fnParamHex) Encode() ([]byte, error) {
	return hex.DecodeString(h.s)
}

/*
	BYTES/STRING TYPES
*/

type fnParamBytes struct {
	b []byte
}

func NewStringParameter(s string) FunctionParameter {
	return NewBytesParameter([]byte(s))
}

func NewBytesParameter(b []byte) FunctionParameter {
	return &fnParamBytes{b}
}

func (b *fnParamBytes) Encode() ([]byte, error) {
	var ibuf = make([]byte, 4)
	binary.LittleEndian.PutUint32(ibuf, uint32(len(b.b)))

	var buf bytes.Buffer
	buf.Write(ibuf)
	buf.Write(b.b)

	return buf.Bytes(), nil
}

/*
	UINTx/BYTE TYPES
*/

type fnParamByte struct {
	b byte
}

// NewByteParameter makes a new byte/uint8 function parameter
func NewByteParameter(b byte) FunctionParameter {
	return &fnParamByte{b}
}

func (b *fnParamByte) Encode() ([]byte, error) {
	return []byte{b.b}, nil
}

type fnParamUint16 struct {
	u uint16
}

// NewUint16Parameter makes a new uint16 function parameter
func NewUint16Parameter(u uint16) FunctionParameter {
	return &fnParamUint16{u}
}

func (u *fnParamUint16) Encode() ([]byte, error) {
	var buf = make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, u.u)
	return buf, nil
}

type fnParamUint32 struct {
	u uint32
}

// NewUint32Parameter makes a new uint32 function parameter
func NewUint32Parameter(u uint32) FunctionParameter {
	return &fnParamUint32{u}
}

func (u *fnParamUint32) Encode() ([]byte, error) {
	var buf = make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, u.u)
	return buf, nil
}

type fnParamUint64 struct {
	u uint64
}

// NewUint64Parameter makes a new uint64 function parameter
func NewUint64Parameter(u uint64) FunctionParameter {
	return &fnParamUint64{u}
}

func (u *fnParamUint64) Encode() ([]byte, error) {
	var buf = make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, u.u)
	return buf, nil
}
