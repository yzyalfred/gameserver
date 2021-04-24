package main

import "gameserver/core/log"

func main()  {
	log.InitLog("../log", "info", "debug", "error")
	log.Debug("test debug -------------> %s", "hello")
	log.Info("test info----------->")
}
