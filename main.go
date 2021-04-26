package main

import (
	"fmt"
	"runtime"
	"sync"
)

var sp = sync.Pool{
	New: func() interface{} {
		return make([]byte, 16)
	},
}

func main()  {
	buf := sp.Get().([]byte)
	buf[0] = 1

	runtime.GC()
	runtime.GC()

	sp.Put(buf)
	buf2 := sp.Get().([]byte)
	buf2[1] = 2

	fmt.Println(buf)
	fmt.Println(buf2)
}
