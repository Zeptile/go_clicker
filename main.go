package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/spf13/pflag"

	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

var running = false
var config Configuration

func main() { mainthread.Init(app) }

func app() {
	config = parseConsoleArguments()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		for {

			if !running {
				continue
			}

			rand.Seed(time.Now().UnixNano())
			var waitForMs int

			if config.randomMode {
				waitForMs = rand.Intn(int(config.randomIntervalEnd))
			} else {
				waitForMs = int(config.intervalMs)
			}

			time.Sleep(time.Duration(waitForMs) * time.Millisecond)

			robotgo.Click()
			if config.debugMode {
				fmt.Printf("[Debug] Clicked after waiting %dms\n", waitForMs)
			}
	
		}
		
	}()
	go func() {
		defer wg.Done()

		for {
			err := listenHotkey(hotkey.KeyP, hotkey.ModCtrl, hotkey.ModShift)
			if err != nil {
				log.Println(err)
			}
	
			if !running {
				fmt.Println("Resumed")
			} else {
				fmt.Println("Paused")
			}
	
			running = !running
		}

	}()
	wg.Wait()
}

func listenHotkey(key hotkey.Key, mods ...hotkey.Modifier) (err error) {
	ms := []hotkey.Modifier{}
	ms = append(ms, mods...)
	hk := hotkey.New(ms, key)

	err = hk.Register()
	if err != nil {
		return
	}

	<-hk.Keydown()
	<-hk.Keyup()
	
	hk.Unregister()
	return
}

func parseConsoleArguments() Configuration {
	var config Configuration

	pflag.BoolVar(&config.randomMode, "random", false, "Flag to enable random mode, click interval is random between 0 and randomIntervalEnd")
	pflag.BoolVar(&config.debugMode, "debug", false, "Flag to enable debug mode, prints debug logs to console")

	pflag.Int64Var(&config.intervalMs, "intervalMs", 50, "Interval in milliseconds to click in normal mode")
	pflag.Int64Var(&config.randomIntervalEnd, "randomIntervalEnd", 100, "Interval treshold in milliseconds to click in random mode")

	pflag.Parse()

	if config.randomMode {
		if config.randomIntervalEnd < 0 {
			fmt.Println("Random interval end must be greater than 0")
			pflag.PrintDefaults()
			panic("Random interval end must be greater than 0")
		}
	} else {
		if config.intervalMs < 0 {
			fmt.Println("Interval must be greater than 0")
			pflag.PrintDefaults()
			panic("Interval must be greater than 0")
		}
	}

	pflag.Parse()

	return config
}