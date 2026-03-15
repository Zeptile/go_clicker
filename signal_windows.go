//go:build windows

package main

import (
	"fmt"

	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

func setupSignals() (<-chan struct{}, <-chan struct{}) {
	toggleCh := make(chan struct{}, 1)
	sampleCh := make(chan struct{}, 1)

	go mainthread.Init(func() {
		go func() {
			for {
				hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyP)
				if err := hk.Register(); err != nil {
					continue
				}
				<-hk.Keydown()
				<-hk.Keyup()
				hk.Unregister()
				toggleCh <- struct{}{}
			}
		}()

		go func() {
			for {
				hk := hotkey.New([]hotkey.Modifier{hotkey.ModCtrl, hotkey.ModShift}, hotkey.KeyO)
				if err := hk.Register(); err != nil {
					continue
				}
				<-hk.Keydown()
				<-hk.Keyup()
				hk.Unregister()
				sampleCh <- struct{}{}
			}
		}()

		select {}
	})

	return toggleCh, sampleCh
}

func printControlInfo() {
	fmt.Println("Toggle:       Ctrl+Shift+P")
	fmt.Println("Sample color: Ctrl+Shift+O")
}
