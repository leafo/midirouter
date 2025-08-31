package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

// ChannelFilter represents a MIDI channel filter
type ChannelFilter struct {
	Enabled bool  `json:"enabled"`
	Channel uint8 `json:"channel"` // 1-16
}

// NoteRangeFilter represents a note range filter
type NoteRangeFilter struct {
	Enabled bool  `json:"enabled"`
	MinNote uint8 `json:"min_note"` // MIDI note number 0-127
	MaxNote uint8 `json:"max_note"` // MIDI note number 0-127
}

// OutputConfig represents the configuration for a single output
type OutputConfig struct {
	Name            string          `json:"name"`
	ChannelFilter   ChannelFilter   `json:"channel_filter"`
	NoteRangeFilter NoteRangeFilter `json:"note_range_filter"`
}

// Config represents the complete router configuration
type Config struct {
	InputDevice string         `json:"input_device"`
	OutputBase  string         `json:"output_base"`
	Outputs     []OutputConfig `json:"outputs"`
}

func main() {
	// Define command-line flags
	saveConfigFile := flag.String("save-config", "", "Save result of configuration to specified file and exit (does not run router)")
	configFile := flag.String("config", "", "Load configuration from specified file and start router")
	quiet := flag.Bool("quiet", false, "Suppress MIDI message logging during operation")
	flag.Parse()

	drv, err := rtmididrv.New()
	if err != nil {
		log.Fatalf("Failed to create MIDI driver: %v", err)
	}
	defer drv.Close()

	var config *Config

	// Check execution mode
	if *configFile != "" {
		// Config file mode: load existing config and run router

		config, err = loadConfigWithFallback(*configFile, drv)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

	} else {
		// Interactive mode

		config, err = interactiveConfig(drv)
		if err != nil {
			log.Fatalf("Configuration error: %v", err)
		}

		// Check if we're in save-only mode
		if *saveConfigFile != "" {
			err = saveConfig(config, *saveConfigFile)
			if err != nil {
				log.Fatalf("Failed to save config: %v", err)
			}
			fmt.Printf("Configuration saved to %s\n", *saveConfigFile)
			return
		}

		// Normal interactive mode: save to default location
		err = saveConfig(config, "config.json")
		if err != nil {
			log.Printf("Warning: Failed to save config: %v", err)
		}
	}

	// Run the router with the loaded/configured setup
	err = runMIDIRouter(drv, config, *quiet)
	if err != nil {
		log.Fatalf("MIDI router error: %v", err)
	}
}

// saveConfig saves the configuration to a JSON file or prints to stdout if filename is empty
func saveConfig(config *Config, filename string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if filename == "" {
		fmt.Print(string(data))
		return nil
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// loadConfig loads configuration from a JSON file
func loadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// validateConfigStructure validates the configuration structure (outputs, filters, etc.)
func validateConfigStructure(config *Config) error {
	if len(config.Outputs) == 0 {
		return fmt.Errorf("no outputs configured")
	}

	for i, output := range config.Outputs {
		if output.Name == "" {
			return fmt.Errorf("output %d has no name", i+1)
		}
		if output.ChannelFilter.Enabled && (output.ChannelFilter.Channel < 1 || output.ChannelFilter.Channel > 16) {
			return fmt.Errorf("output %d has invalid channel: %d (must be 1-16)", i+1, output.ChannelFilter.Channel)
		}
		if output.NoteRangeFilter.Enabled && output.NoteRangeFilter.MinNote > output.NoteRangeFilter.MaxNote {
			return fmt.Errorf("output %d has invalid note range: %d-%d", i+1, output.NoteRangeFilter.MinNote, output.NoteRangeFilter.MaxNote)
		}
	}

	return nil
}

// validateInputDevice checks if the input device exists in the available devices
func validateInputDevice(deviceName string, drv *rtmididrv.Driver) error {
	ins, err := drv.Ins()
	if err != nil {
		return fmt.Errorf("failed to get MIDI inputs: %w", err)
	}

	for _, in := range ins {
		if in.String() == deviceName {
			return nil
		}
	}

	return fmt.Errorf("configured input device not found: %s\nAvailable devices: %v",
		deviceName, getDeviceNames(ins))
}

// loadConfigWithFallback loads config and falls back to interactive input selection if device not found
func loadConfigWithFallback(filename string, drv *rtmididrv.Driver) (*Config, error) {
	config, err := loadConfig(filename)
	if err != nil {
		return nil, err
	}

	// Validate config structure first
	if err := validateConfigStructure(config); err != nil {
		return nil, err
	}

	// Check if input device exists
	if err := validateInputDevice(config.InputDevice, drv); err != nil {
		fmt.Printf("Warning: %s\n", err.Error())

		selectedInput, err := selectInputDevice(drv)
		if err != nil {
			return nil, fmt.Errorf("failed to select input device: %w", err)
		}

		config.InputDevice = selectedInput.String()
	}

	return config, nil
}

// loadAndValidateConfig loads configuration from file and validates it
func loadAndValidateConfig(filename string, drv *rtmididrv.Driver) (*Config, error) {
	config, err := loadConfig(filename)
	if err != nil {
		return nil, err
	}

	// Validate config structure
	if err := validateConfigStructure(config); err != nil {
		return nil, err
	}

	// Validate input device
	if err := validateInputDevice(config.InputDevice, drv); err != nil {
		return nil, err
	}

	return config, nil
}

// getDeviceNames extracts device names for error messages
func getDeviceNames(devices []drivers.In) []string {
	names := make([]string, len(devices))
	for i, device := range devices {
		names[i] = device.String()
	}
	return names
}

// selectInputDevice presents available MIDI input devices and lets user select one
func selectInputDevice(drv *rtmididrv.Driver) (drivers.In, error) {
	reader := bufio.NewReader(os.Stdin)

	// Get available input devices
	ins, err := drv.Ins()
	if err != nil {
		return nil, fmt.Errorf("failed to get MIDI inputs: %w", err)
	}

	if len(ins) == 0 {
		return nil, fmt.Errorf("no MIDI input devices found")
	}

	fmt.Printf("Select MIDI Input Device:\n")
	for i, in := range ins {
		fmt.Printf("  %d: %s\n", i+1, in.String())
	}

	fmt.Print("Select input device (1-", len(ins), "): ")
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	choice, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || choice < 1 || choice > len(ins) {
		return nil, fmt.Errorf("invalid selection")
	}

	selectedInput := ins[choice-1]
	return selectedInput, nil
}

// interactiveConfig guides the user through configuration setup
func interactiveConfig(drv *rtmididrv.Driver) (*Config, error) {
	reader := bufio.NewReader(os.Stdin)
	config := &Config{}

	fmt.Println("Starting interactive configuration...")

	// Select input device
	selectedInput, err := selectInputDevice(drv)
	if err != nil {
		return nil, err
	}
	config.InputDevice = selectedInput.String()

	// Get output base name
	fmt.Print("Enter base name for outputs (default: 'MIDI Router'): ")
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	outputBase := strings.TrimSpace(line)
	if outputBase == "" {
		outputBase = "MIDI Router"
	}
	config.OutputBase = outputBase

	// Get number of outputs
	fmt.Print("Number of virtual outputs to create: ")
	line, err = reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	numOutputs, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || numOutputs < 1 || numOutputs > 16 {
		return nil, fmt.Errorf("invalid number of outputs (must be 1-16)")
	}

	// Configure each output
	config.Outputs = make([]OutputConfig, numOutputs)
	for i := 0; i < numOutputs; i++ {
		fmt.Printf("Output %d: %s Out %d\n", i+1, outputBase, i+1)
		config.Outputs[i].Name = fmt.Sprintf("%s Out %d", outputBase, i+1)

		// Channel filter
		fmt.Print("Enable channel filter? (y/N): ")
		line, err = reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		if strings.ToLower(strings.TrimSpace(line)) == "y" {
			fmt.Print("Channel number (1-16): ")
			line, err = reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("failed to read input: %w", err)
			}

			channel, err := strconv.Atoi(strings.TrimSpace(line))
			if err != nil || channel < 1 || channel > 16 {
				return nil, fmt.Errorf("invalid channel number (must be 1-16)")
			}

			config.Outputs[i].ChannelFilter.Enabled = true
			config.Outputs[i].ChannelFilter.Channel = uint8(channel)
		}

		// Note range filter
		fmt.Print("Enable note range filter? (y/N): ")
		line, err = reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input: %w", err)
		}

		if strings.ToLower(strings.TrimSpace(line)) == "y" {
			noteRange, err := configureNoteRange(selectedInput)
			if err != nil {
				return nil, fmt.Errorf("failed to configure note range: %w", err)
			}
			config.Outputs[i].NoteRangeFilter = *noteRange
		}
	}

	return config, nil
}

// noteToName converts a MIDI note number to note name
func noteToName(note uint8) string {
	noteNames := []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}
	octave := int(note)/12 - 1
	noteName := noteNames[note%12]
	return fmt.Sprintf("%s%d", noteName, octave)
}

// configureNoteRange configures note range by listening to actual MIDI input
func configureNoteRange(inputPort drivers.In) (*NoteRangeFilter, error) {
	fmt.Printf("  Play the LOWEST note: ")

	minNote, err := captureNote(inputPort)
	if err != nil {
		return nil, fmt.Errorf("failed to capture min note: %w", err)
	}

	fmt.Printf("  Play the HIGHEST note: ")

	maxNote, err := captureNote(inputPort)
	if err != nil {
		return nil, fmt.Errorf("failed to capture max note: %w", err)
	}

	if minNote > maxNote {
		minNote, maxNote = maxNote, minNote
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Confirm range %s to %s? (Y/n): ",
		noteToName(minNote), noteToName(maxNote))
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	if strings.ToLower(strings.TrimSpace(line)) == "n" {
		return &NoteRangeFilter{Enabled: false}, nil
	}

	return &NoteRangeFilter{
		Enabled: true,
		MinNote: minNote,
		MaxNote: maxNote,
	}, nil
}

// captureNote listens for a single Note On event and returns the note number
func captureNote(inputPort drivers.In) (uint8, error) {
	noteChan := make(chan uint8, 1)
	errorChan := make(chan error, 1)

	// Start listening for MIDI input
	stop, err := midi.ListenTo(inputPort, func(msg midi.Message, timestampms int32) {
		var channel, key, velocity uint8
		if msg.GetNoteOn(&channel, &key, &velocity) && velocity > 0 {
			fmt.Printf("%s\n", noteToName(key))
			select {
			case noteChan <- key:
			default:
			}
		}
	})

	if err != nil {
		return 0, fmt.Errorf("failed to start listening: %w", err)
	}

	defer stop()

	// Wait for note capture with timeout
	select {
	case note := <-noteChan:
		return note, nil
	case err := <-errorChan:
		return 0, fmt.Errorf("error during note capture: %w", err)
	case <-time.After(30 * time.Second):
		return 0, fmt.Errorf("timeout: no note captured within 30 seconds")
	}
}

// shouldRouteMessage checks if a message should be routed to a specific output
func shouldRouteMessage(msg midi.Message, outputConfig *OutputConfig) bool {
	// Channel filter
	if outputConfig.ChannelFilter.Enabled {
		var channel, key, velocity uint8
		if msg.GetNoteOn(&channel, &key, &velocity) || msg.GetNoteOff(&channel, &key, &velocity) {
			if channel+1 != outputConfig.ChannelFilter.Channel {
				return false
			}
		} else {
			// For other message types, try to get channel
			if len(msg) >= 1 {
				msgChannel := (msg[0] & 0x0F) + 1
				if msgChannel != outputConfig.ChannelFilter.Channel {
					return false
				}
			}
		}
	}

	// Note range filter
	if outputConfig.NoteRangeFilter.Enabled {
		var channel, key, velocity uint8
		if msg.GetNoteOn(&channel, &key, &velocity) || msg.GetNoteOff(&channel, &key, &velocity) {
			if key < outputConfig.NoteRangeFilter.MinNote || key > outputConfig.NoteRangeFilter.MaxNote {
				return false
			}
		}
	}

	return true
}

func runMIDIRouter(drv *rtmididrv.Driver, config *Config, quiet bool) error {
	// Find the configured input device
	ins, err := drv.Ins()
	if err != nil {
		return fmt.Errorf("failed to get MIDI inputs: %w", err)
	}

	var selectedInput drivers.In
	for _, in := range ins {
		if in.String() == config.InputDevice {
			selectedInput = in
			break
		}
	}

	if selectedInput == nil {
		return fmt.Errorf("configured input device not found: %s", config.InputDevice)
	}

	// Create virtual outputs
	outputs := make([]drivers.Out, len(config.Outputs))
	senders := make([]func(midi.Message) error, len(config.Outputs))

	for i, outputConfig := range config.Outputs {
		virtualOut, err := drv.OpenVirtualOut(outputConfig.Name)
		if err != nil {
			return fmt.Errorf("failed to create virtual output %d: %w", i+1, err)
		}
		defer virtualOut.Close()

		sender, err := midi.SendTo(virtualOut)
		if err != nil {
			return fmt.Errorf("failed to create sender for output %d: %w", i+1, err)
		}

		outputs[i] = virtualOut
		senders[i] = sender
	}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}
	fmt.Printf("Running with configuration:\n%s\n", configJSON)
	fmt.Println("Press Ctrl+C to stop...")

	// Start routing
	stop, err := midi.ListenTo(selectedInput, func(msg midi.Message, timestampms int32) {
		routedTo := make([]string, 0, len(config.Outputs))

		for i, outputConfig := range config.Outputs {
			if shouldRouteMessage(msg, &outputConfig) {
				err := senders[i](msg)
				if err != nil {
					log.Printf("Error sending to %s: %v", outputConfig.Name, err)
				} else {
					routedTo = append(routedTo, outputConfig.Name)
				}
			}
		}

		if !quiet {
			if len(routedTo) > 0 {
				fmt.Printf("[%s] %v\n", strings.Join(routedTo, ", "), msg)
			} else {
				fmt.Printf("\033[2m[DROPPED] %v\033[0m\n", msg)
			}
		}
	})

	if err != nil {
		return fmt.Errorf("failed to start listening: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down...")
	stop()

	return nil
}
