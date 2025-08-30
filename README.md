# MIDI Router

A command-line MIDI router that takes input from a single physical MIDI device and routes it to two virtual MIDI output devices.

## Features

- Discovers and lists available MIDI input devices
- Interactive device selection
- Creates two named virtual MIDI output ports:
  - "MIDI Router Out 1"
  - "MIDI Router Out 2"
- Real-time MIDI message routing from input to both outputs
- Graceful shutdown with Ctrl+C

## Prerequisites

- Go 1.24.2 or later
- MIDI input device connected to your system
- Operating system with MIDI support (Linux, Windows, macOS)

## Building

```bash
go build -o midirouter
```

## Usage

1. Connect your MIDI input device
2. Run the router:
   ```bash
   ./midirouter
   ```
3. Select the input device from the list
4. The router will create two virtual MIDI outputs that other applications can connect to
5. Press Ctrl+C to stop

## Virtual Output Ports

Once running, other MIDI applications can connect to these virtual ports:
- **MIDI Router Out 1**
- **MIDI Router Out 2**

All MIDI messages received from the selected input device will be forwarded to both virtual outputs simultaneously.

## Dependencies

- `gitlab.com/gomidi/midi/v2` - MIDI library for Go
- `gitlab.com/gomidi/midi/v2/drivers/rtmididrv` - Cross-platform MIDI driver

## Example

```
MIDI Router - Physical Input to Dual Virtual Outputs
==================================================

Available MIDI Input Devices:
1: Midi Through:Midi Through Port-0 14:0
2: USB MIDI Device:USB MIDI Device Port-0 128:0

Select input device (1-2): 2
Selected input: USB MIDI Device:USB MIDI Device Port-0 128:0

Virtual MIDI outputs created:
- MIDI Router Out 1
- MIDI Router Out 2

MIDI router is running. Press Ctrl+C to stop...
```