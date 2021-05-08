package network

import (
	"errors"
	"gameserver/common/utils"
	"io"
	"math"
)

// --------------
// | len | data |
// --------------
type MsgParser struct {
	lenMsgLen    int
	minMsgLen    uint32
	maxMsgLen    uint32
	littleEndian bool
}

func NewMsgParser() *MsgParser {
	p := new(MsgParser)
	p.lenMsgLen = 2
	p.minMsgLen = 1
	p.maxMsgLen = 4096
	p.littleEndian = false

	return p
}

// It's dangerous to call the method on reading or writing
func (this *MsgParser) SetMsgLen(lenMsgLen int, minMsgLen uint32, maxMsgLen uint32) {
	if lenMsgLen == 1 || lenMsgLen == 2 || lenMsgLen == 4 {
		this.lenMsgLen = lenMsgLen
	}
	if minMsgLen != 0 {
		this.minMsgLen = minMsgLen
	}
	if maxMsgLen != 0 {
		this.maxMsgLen = maxMsgLen
	}

	var max uint32
	switch this.lenMsgLen {
	case 1:
		max = math.MaxUint8
	case 2:
		max = math.MaxUint16
	case 4:
		max = math.MaxUint32
	}
	if this.minMsgLen > max {
		this.minMsgLen = max
	}
	if this.maxMsgLen > max {
		this.maxMsgLen = max
	}
}

// It's dangerous to call the method on reading or writing
func (this *MsgParser) SetByteOrder(littleEndian bool) {
	this.littleEndian = littleEndian
}

// goroutine safe
func (this *MsgParser) Read(conn *TCPConn) ([]byte, error) {
	var b [4]byte
	bufMsgLen := b[:this.lenMsgLen]

	// read len
	if _, err := io.ReadFull(conn, bufMsgLen); err != nil {
		return nil, err
	}

	// parse len
	var msgLen uint32
	switch this.lenMsgLen {
	case 1:
		msgLen = uint32(bufMsgLen[0])
	case 2:
		msgLen = uint32(utils.ByteToUint16(bufMsgLen, this.littleEndian))
	case 4:
		msgLen = uint32(utils.ByteToUint32(bufMsgLen, this.littleEndian))
	}

	// check len
	if msgLen > this.maxMsgLen {
		return nil, errors.New("message too long")
	} else if msgLen < this.minMsgLen {
		return nil, errors.New("message too short")
	}

	// data
	msgData := make([]byte, msgLen)
	if _, err := io.ReadFull(conn, msgData[:msgLen]); err != nil {
		return nil, err
	}

	return msgData[:msgLen], nil
}

// goroutine safe
func (this *MsgParser) Write(conn *TCPConn, args ...[]byte) error {
	// get len
	var msgLen uint32
	for i := 0; i < len(args); i++ {
		msgLen += uint32(len(args[i]))
	}

	// check len
	if msgLen > this.maxMsgLen {
		return errors.New("message too long")
	} else if msgLen < this.minMsgLen {
		return errors.New("message too short")
	}

	msg := make([]byte, this.lenMsgLen+int(msgLen))
	// write len
	switch this.lenMsgLen {
	case 1:
		msg[0] = byte(msgLen)
	case 2:
		utils.PutUint16ToByte(msg, uint16(msgLen), this.littleEndian)
	case 4:
		utils.PutUint32ToByte(msg, uint32(msgLen), this.littleEndian)
	}

	// write data
	l := this.lenMsgLen
	for i := 0; i < len(args); i++ {
		copy(msg[l:], args[i])
		l += len(args[i])
	}

	conn.Write(msg)

	return nil
}
