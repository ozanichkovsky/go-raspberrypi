package main

import (
	"fmt"
	"github.com/warthog618/gpiod"
	"github.com/warthog618/gpiod/device/rpi"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	c, err := gpiod.NewChip("gpiochip0")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	values := map[int]string{0: "inactive", 1: "active"}
	ledPin := rpi.GPIO17
	buttonPin := rpi.GPIO18
	ledPinState := 0
	buttonState := 1
	lastButtonState := 1
	var lastChangeTime int64
	var captureTime int64 = 50

	l, err := c.RequestLine(ledPin, gpiod.AsOutput(ledPinState))
	if err != nil {
		panic(err)
	}
	defer func() {
		l.Reconfigure(gpiod.AsInput)
		l.Close()
	}()

	fmt.Printf("Set pin %d %s\n", ledPin, values[ledPinState])

	b, err := c.RequestLine(buttonPin, gpiod.AsInput)
	if err != nil {
		panic(err)
	}

	defer func() {
		b.Reconfigure(gpiod.AsInput)
		b.Close()
	}()

	// capture exit signals to ensure pin is reverted to input on exit.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	for {
		select {
		case <-quit:
			return
		default:
			r, err := b.Value()

			if err != nil {
				panic(err)
			}

			if r != lastButtonState {
				lastChangeTime = time.Now().UnixNano() / 1000000
			}

			if time.Now().UnixNano() / 1000000 - lastChangeTime > captureTime {
				if r != buttonState {
					buttonState = r
					if buttonState == 0 {
						print("Button is pressed\n")
						ledPinState ^= 1

						if ledPinState == 1 {
							print("Turn on LED...\n")
						} else {
							print("Turn off LED...\n")
						}
					} else {
						print("Button is released\n")
					}
				}
			}

			l.SetValue(ledPinState)
			lastButtonState = r
		}
	}
}

