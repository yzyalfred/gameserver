package network

import (
	"fmt"
	"gameserver/core/log"
	"net"
	"sync"
	"time"
)

type ConnSet map[net.Conn]struct{}

type TCPConn struct {
	sync.Mutex
	conn      net.Conn
	writeChan chan []byte
	closeFlag bool
	msgParser *MsgParser
}

func newTCPConn(conn net.Conn, pendingWriteNum int, msgParser *MsgParser) *TCPConn {
	tcpConn := new(TCPConn)
	tcpConn.conn = conn
	tcpConn.writeChan = make(chan []byte, pendingWriteNum)
	tcpConn.msgParser = msgParser
	go func() {
		for b := range tcpConn.writeChan {
			if b == nil {
				break
			}
			_, err := conn.Write(b)
			if err != nil {
				break
			}
		}
		conn.Close()
		tcpConn.Lock()
		tcpConn.closeFlag = true
		tcpConn.Unlock()
	}()

	return tcpConn
}

func (this *TCPConn) doDestroy() {
	this.conn.(*net.TCPConn).SetLinger(0)
	this.conn.Close()

	if !this.closeFlag {
		close(this.writeChan)
		this.closeFlag = true
	}
}

func (this *TCPConn) Destroy() {
	this.Lock()
	defer this.Unlock()

	this.doDestroy()
}

func (this *TCPConn) Close() {
	this.Lock()
	defer this.Unlock()
	if this.closeFlag {
		return
	}

	this.doWrite(nil)
	this.closeFlag = true
}

func (this *TCPConn) doWrite(b []byte) {
	if len(this.writeChan) == cap(this.writeChan) {
		log.Warn("close conn: channel full")
		this.doDestroy()
		return
	}

	this.writeChan <- b
}

// b must not be modified by the others goroutines
func (this *TCPConn) Write(b []byte) {
	this.Lock()
	defer this.Unlock()
	if this.closeFlag || b == nil {
		return
	}

	this.doWrite(b)
}

func (this *TCPConn) Read(b []byte) (int, error) {
	return this.conn.Read(b)
}

func (this *TCPConn) LocalAddr() net.Addr {
	return this.conn.LocalAddr()
}

func (this *TCPConn) RemoteAddr() net.Addr {
	return this.conn.RemoteAddr()
}

func (this *TCPConn) ReadMsg() ([]byte, error) {
	return this.msgParser.Read(this)
}

func (this *TCPConn) WriteMsg(args ...[]byte) error {
	if this.closeFlag == true {
		return fmt.Errorf("conn is close")
	}
	return this.msgParser.Write(this, args...)
}

func (this *TCPConn) IsConnected() bool {
	return this.closeFlag == false
}

func (this *TCPConn) SetReadDeadline(d time.Duration)  {
	this.conn.SetReadDeadline(time.Now().Add(d))
}

func (this *TCPConn) SetWriteDeadline(d time.Duration)  {
	this.conn.SetWriteDeadline(time.Now().Add(d))
}