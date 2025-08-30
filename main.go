package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	fmt.Println("MIDI Router - Physical Input to Dual Virtual Outputs")
	fmt.Println("==================================================")

	drv, err := rtmididrv.New()
	if err != nil {
		log.Fatalf("Failed to create MIDI driver: %v", err)
	}
	defer drv.Close()

	err = runMIDIRouter(drv)
	if err != nil {
		log.Fatalf("MIDI router error: %v", err)
	}
}

func runMIDIRouter(drv *rtmididrv.Driver) error {
	ins, err := drv.Ins()
	if err != nil {
		return fmt.Errorf("failed to get MIDI inputs: %w", err)
	}

	if len(ins) == 0 {
		return fmt.Errorf("no MIDI input devices found")
	}

	fmt.Printf("\nAvailable MIDI Input Devices:\n")
	for i, in := range ins {
		fmt.Printf("%d: %s\n", i+1, in.String())
	}

	fmt.Print("\nSelect input device (1-", len(ins), "): ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	choice, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || choice < 1 || choice > len(ins) {
		return fmt.Errorf("invalid selection")
	}

	selectedInput := ins[choice-1]
	fmt.Printf("Selected input: %s\n", selectedInput.String())

	virtualOut1, err := drv.OpenVirtualOut("MIDI Router Out 1")
	if err != nil {
		return fmt.Errorf("failed to create virtual output 1: %w", err)
	}
	defer virtualOut1.Close()

	virtualOut2, err := drv.OpenVirtualOut("MIDI Router Out 2")
	if err != nil {
		return fmt.Errorf("failed to create virtual output 2: %w", err)
	}
	defer virtualOut2.Close()

	fmt.Println("\nVirtual MIDI outputs created:")
	fmt.Println("- MIDI Router Out 1")
	fmt.Println("- MIDI Router Out 2")

	send1, err := midi.SendTo(virtualOut1)
	if err != nil {
		return fmt.Errorf("failed to create sender for output 1: %w", err)
	}
	
	send2, err := midi.SendTo(virtualOut2)
	if err != nil {
		return fmt.Errorf("failed to create sender for output 2: %w", err)
	}

	stop, err := midi.ListenTo(selectedInput, func(msg midi.Message, timestampms int32) {
		err1 := send1(msg)
		err2 := send2(msg)
		
		if err1 != nil {
			log.Printf("Error sending to output 1: %v", err1)
		}
		if err2 != nil {
			log.Printf("Error sending to output 2: %v", err2)
		}
		
		fmt.Printf("Routed MIDI message: %v\n", msg)
	})

	if err != nil {
		return fmt.Errorf("failed to start listening: %w", err)
	}

	fmt.Println("\nMIDI router is running. Press Ctrl+C to stop...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	stop()
	
	return nil
}