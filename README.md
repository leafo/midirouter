# MIDI Router

Routes MIDI from a single input to multiple virtual outputs with channel and note range filtering.

## Features

- Interactive configuration wizard
- Multiple virtual MIDI outputs (1-16) that can be filtered by channel and note range
- Override output channel to remap MIDI messages to different channels
- Transpose note events by semitones (+/- 127 semitones)
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
   - Optional: Enable channel override (1-16)
   - Optional: Enable note transposition (-127 to +127 semitones)

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
      "note_range_filter": null,
      "override_channel": 5,
      "transpose_semitones": -12
    }
  ]
}
```

## Filters and Processing

### Channel Filter
Only routes MIDI messages from the specified channel (1-16).

### Note Range Filter
Only routes note on/off messages within the specified note range (0-127). Other message types pass through.

### Channel Override
Changes the channel number of forwarded MIDI messages to the specified channel (1-16). This happens after filtering, so you can filter on the original channel and then override to a different output channel.

### Note Transposition
Transposes note on/off messages by the specified number of semitones (-127 to +127). Positive values transpose up, negative values transpose down. If transposition would result in a note outside the MIDI range (0-127), the original message is sent unchanged. Only affects note messages - other MIDI messages pass through unmodified.
