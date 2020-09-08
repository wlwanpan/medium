package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/wlwanpan/medium/minecraft-server-wrapper/msw"
)

func main() {
	cmd := msw.JavaExecCmd("server.jar", 1024, 1024)
	console := msw.NewConsole(cmd)
	wrapper := msw.NewWrapper(console, msw.LogParserFunc)
	wrapper.Start()

	go func() {
		time.Sleep(15 * time.Second)
		wrapper.Stop()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	console.Kill()
}
