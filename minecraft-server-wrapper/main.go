package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/wlwanpan/medium/minecraft-server-wrapper/msw"
)

func main() {
	wrapper := msw.NewDefaultWrapper("server.jar", 1024, 1024)
	wrapper.Start()

	go func() {
		time.Sleep(15 * time.Second)
		wrapper.Stop()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	wrapper.Kill()
}
