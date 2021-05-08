package service

import (
	"fmt"
	"gameserver/core/log"
	"gameserver/core/processor"
	"runtime"
)

type CallIO struct {
	ClienId uint64

	Buff []byte
}

type Service struct {
	callChan chan *CallIO
	name     string

	processor processor.PBProcessor
}

func (this *Service) Send(clientId uint64, buff []byte) {
	this.callChan <- &CallIO{
		ClienId: clientId,
		Buff: buff,
	}
}

func (this *Service) call(io *CallIO) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			err := fmt.Errorf("%v: %s", r, buf[:l])
			log.Error("call error, clientId: %s, err: %v", io.ClienId, err)
		}
	}()

	this.processor.Route(io.ClienId, io.Buff)
}

func (this *Service) run() {
	for {
		select {
		case io := <- this.callChan:
			this.call(io)
		}
	}
}