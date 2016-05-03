package main

import (
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/cloudfoundry-incubator/ducati-daemon/lib/namespace"
	"github.com/pivotal-golang/lager"
)

func main() {
	runtime.LockOSThread()

	logger := lager.NewLogger("netns-exec")

	if len(os.Args) < 2 {
		log.Fatalf("first arg: netns path")
	}

	netnsPath := os.Args[1]

	if len(os.Args) < 3 {
		log.Fatalf("provide a command")
	}

	pathOpener := &namespace.PathOpener{
		Logger: logger,
	}
	netns, err := pathOpener.OpenPath(netnsPath)
	if err != nil {
		log.Fatalf("%s", err)
	}

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = netns.Execute(func(f *os.File) error {
		return cmd.Run()
	})
	if err != nil {
		log.Fatalf("%s", err)
	}
}
