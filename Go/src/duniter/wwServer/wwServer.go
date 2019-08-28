package main

import (
	
	"duniter/run"
	"runtime"

)

func main () {
	runtime.GOMAXPROCS(runtime.NumCPU())
	run.Start()
}
