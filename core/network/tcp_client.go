package network

import (
	"gameserver/core/log"
	"net"
	"sync"
	"time"
)

type TCPClient struct {
	sync.Mutex
	Addr            string
	ConnNum         int
	ConnectInterval time.Duration
	PendingWriteNum int
	AutoReconnect   bool
	NewAgent        func(*TCPConn) Agent
	cons            ConnSet
	wg              sync.WaitGroup
	closeFlag       bool

	// msg parser
	LenMsgLen    int
	MinMsgLen    uint32
	MaxMsgLen    uint32
	LittleEndian bool
	msgParser    *MsgParser
}

func (this *TCPClient) Start() {
	this.init()

	for i := 0; i < this.ConnNum; i++ {
		this.wg.Add(1)
		go this.connect()
	}
}

func (this *TCPClient) init() {
	this.Lock()
	defer this.Unlock()

	if this.ConnNum <= 0 {
		this.ConnNum = 1
		log.Info("invalid ConnNum, reset to %v", this.ConnNum)
	}
	if this.ConnectInterval <= 0 {
		this.ConnectInterval = 3 * time.Second
		log.Info("invalid ConnectInterval, reset to %v", this.ConnectInterval)
	}
	if this.PendingWriteNum <= 0 {
		this.PendingWriteNum = 1000
		log.Info("invalid PendingWriteNum, reset to %v", this.PendingWriteNum)
	}
	if this.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}
	if this.cons != nil {
		log.Fatal("client is running")
	}

	this.cons = make(ConnSet)
	this.closeFlag = false

	// msg parser
	msgParser := NewMsgParser()
	msgParser.SetMsgLen(this.LenMsgLen, this.MinMsgLen, this.MaxMsgLen)
	msgParser.SetByteOrder(this.LittleEndian)
	this.msgParser = msgParser
}

func (this *TCPClient) dial() net.Conn {
	for {
		conn, err := net.Dial("tcp", this.Addr)
		if this.closeFlag {
			return conn
		} else if err == nil && conn != nil {
			conn.(*net.TCPConn).SetNoDelay(true)
			return conn
		}

		log.Info("connect to %v error: %v", this.Addr, err)
		time.Sleep(this.ConnectInterval)
		continue
	}
}

func (this *TCPClient) connect() {
	defer this.wg.Done()

reconnect:
	conn := this.dial()
	if conn == nil {
		return
	}

	this.Lock()
	if this.closeFlag {
		this.Unlock()
		conn.Close()
		return
	}
	this.cons[conn] = struct{}{}
	this.Unlock()

	tcpConn := newTCPConn(conn, this.PendingWriteNum, this.msgParser)
	agent := this.NewAgent(tcpConn)
	agent.Run()

	// cleanup
	tcpConn.Close()
	this.Lock()
	delete(this.cons, conn)
	this.Unlock()
	agent.OnClose()

	if this.AutoReconnect {
		time.Sleep(this.ConnectInterval)
		goto reconnect
	}
}

func (this *TCPClient) Close(waitDone bool) {
	this.Lock()
	this.closeFlag = true
	for conn := range this.cons {
		conn.Close()
	}
	this.cons = nil
	this.Unlock()

	if waitDone == true {
		this.wg.Wait()
	}
}
