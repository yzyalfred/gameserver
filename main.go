package main

import (
	"gameserver/core/log"
)

func test1() {
	test2()
}


func test2() {
	test3()
}

func test3() {
	log.Error("test mul error for log")
}

func main()  {
	log.InitLog("../log", "info", true, 0)
	log.Info("test info for log")
	log.Error("test error for log")
	test3()
}
