package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var (
	DebugMode  = true
	Ver        = ""
	Build      = ""
	GitBrahnch = ""
)

// TODO versioning https://blog.alexellis.io/inject-build-time-vars-golang/

func Run() {

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println(sig) // TODO logger INFO
		done <- true
	}()

	rootCtx := context.Background()
	_, cancelFn := context.WithCancel(rootCtx)

	fmt.Println("To close program correctly, use Ctrl+C")
	<-done
	fmt.Println("exiting")
	cancelFn()
}
