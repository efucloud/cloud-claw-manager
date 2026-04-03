package main

import (
	"github.com/efucloud/cloud-claw-manager/cmd/server"
	"math/rand"
	"os"
	"runtime"
	"time"
)

func main() {
	rand.NewSource(time.Now().UTC().UnixNano())
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	command := server.NewRunnerServerCommand()
	command.Flags().SortFlags = true
	command.Execute()
}
