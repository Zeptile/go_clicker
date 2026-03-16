//go:build unix

package input

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func Setup() (<-chan struct{}, <-chan struct{}) {
	sigToggle := make(chan os.Signal, 1)
	sigSample := make(chan os.Signal, 1)
	signal.Notify(sigToggle, syscall.SIGUSR1)
	signal.Notify(sigSample, syscall.SIGUSR2)

	toggleCh := make(chan struct{}, 1)
	sampleCh := make(chan struct{}, 1)

	go func() {
		for range sigToggle {
			toggleCh <- struct{}{}
		}
	}()
	go func() {
		for range sigSample {
			sampleCh <- struct{}{}
		}
	}()

	return toggleCh, sampleCh
}

func PrintControlInfo() {
	pid := os.Getpid()
	fmt.Printf("PID: %d\n", pid)
	fmt.Println("Toggle:       kill -USR1", pid)
	fmt.Println("Sample color: kill -USR2", pid)
}
