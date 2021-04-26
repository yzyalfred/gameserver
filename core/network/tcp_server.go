package network

import (
	"gameserver/core/log"
	"net"
	"sync"
	"time"
)

type TCPServer struct {
	Addr            string
	MaxConnNum      int
	PendingWriteNum int
	NewAgent        func(*TCPConn) Agent
	ln              net.Listener
	conns           ConnSet
	mutexConns      sync.Mutex
	wgLn            sync.WaitGroup
	wgConns         sync.WaitGroup

	// msg parser
	LenMsgLen    int
	MinMsgLen    uint32
	MaxMsgLen    uint32
	LittleEndian bool
	msgParser    *MsgParser
}

func (this *TCPServer) Start() {
	this.init()
	go this.run()
}

func (this *TCPServer) init() {
	ln, err := net.Listen("tcp", this.Addr)
	if err != nil {
		log.Fatal("%v", err)
	}

	if this.MaxConnNum <= 0 {
		this.MaxConnNum = 100
		log.Info("invalid MaxConnNum, reset to %v", this.MaxConnNum)
	}
	if this.PendingWriteNum <= 0 {
		this.PendingWriteNum = 100
		log.Info("invalid PendingWriteNum, reset to %v", this.PendingWriteNum)
	}
	if this.NewAgent == nil {
		log.Fatal("NewAgent must not be nil")
	}

	this.ln = ln
	this.conns = make(ConnSet)

	// msg parser
	msgParser := NewMsgParser()
	msgParser.SetMsgLen(this.LenMsgLen, this.MinMsgLen, this.MaxMsgLen)
	msgParser.SetByteOrder(this.LittleEndian)
	this.msgParser = msgParser
}

func (this *TCPServer) run() {
	this.wgLn.Add(1)
	defer this.wgLn.Done()

	var tempDelay time.Duration
	for {
		conn, err := this.ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Warn("accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		tempDelay = 0

		this.mutexConns.Lock()
		if len(this.conns) >= this.MaxConnNum {
			this.mutexConns.Unlock()
			conn.Close()
			log.Debug("too many connections")
			continue
		}
		this.conns[conn] = struct{}{}
		this.mutexConns.Unlock()

		this.wgConns.Add(1)

		tcpConn := newTCPConn(conn, this.PendingWriteNum, this.msgParser)
		agent := this.NewAgent(tcpConn)
		go func() {
			agent.Run()

			// cleanup
			tcpConn.Close()
			this.mutexConns.Lock()
			delete(this.conns, conn)
			this.mutexConns.Unlock()
			agent.OnClose()

			this.wgConns.Done()
		}()
	}
}

func (this *TCPServer) Close() {
	this.ln.Close()
	this.wgLn.Wait()

	this.mutexConns.Lock()
	for conn := range this.conns {
		conn.Close()
	}
	this.conns = nil
	this.mutexConns.Unlock()
	this.wgConns.Wait()
}
