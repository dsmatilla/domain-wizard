package main

import (
	"sync"
)

const (
	CertsDir = "/etc/redirect/certs/"
	ConfigDir = "/etc/redirect/config/"
)

func main() {
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	go httpServer()
	go httpsServer()
	waitGroup.Wait()
}