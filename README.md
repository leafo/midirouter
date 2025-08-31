# MIDI Router

Routes MIDI from a single input to multiple virtual outputs with channel and note range filtering.

## Features

- Interactive configuration wizard
- Multiple virtual MIDI outputs (1-16) that can be filtered by channel and note range
- Save routing configuration to JSON to load quickly later

## Building

```bash
go build -o midirouter
```

## Usage

### Basic Commands

```bash
# Interactive setup and run
./midirouter

# Save configuration only
./midirouter --save-config my-config.json

# Load saved configuration
./midirouter --config my-config.json

# Suppress message logging
./midirouter --config my-config.json --quiet
```

### Interactive Configuration

1. Select MIDI input device
2. Set output base name (default: "MIDI Router") 
3. Choose number of virtual outputs (1-16)
4. Configure each output:
   - Set output name
   - Optional: Enable channel filter (1-16)
   - Optional: Enable note range filter (play notes to set range)

## Configuration File

Example JSON configuration:

```json
{
  "input_device": "USB MIDI Device:USB MIDI Device Port-0 128:0",
  "output_base": "MIDI Router", 
  "outputs": [
    {
      "name": "Out 1",
      "channel_filter": {
        "channel": 1
      },
      "note_range_filter": {
        "min_note": 36,
        "max_note": 59
      }
    },
    {
      "name": "Out 2",
      "channel_filter": null,
      "note_range_filter": null
    }
  ]
}
```
