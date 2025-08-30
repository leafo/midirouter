# MIDI Router

A command-line MIDI router with interactive configuration that routes MIDI from a single physical input to multiple virtual outputs with filtering capabilities.

## Features

- **Interactive Configuration**: Step-by-step setup wizard
- **Flexible Output Creation**: Route to any number of virtual MIDI outputs (1-16)
- **Filtering**:
  - Channel filtering (route only specific MIDI channels)
  - Note range filtering with live MIDI note detection
- **Real-time Routing Display**: Shows which outputs receive each MIDI message
- **Configuration Persistence**: Saves settings to JSON for future use
- **Live Note Range Selection**: Play notes on your MIDI device to set filtering ranges

## Building

```bash
go build -o midirouter
```

## Usage

### Basic Usage

1. **Connect your MIDI input device**

2. **Run the interactive configuration:**
   ```bash
   ./midirouter
   ```
   This will run the configuration wizard and then start the router.

### Save Configuration Only

To save a configuration without running the router:
```bash
./midirouter --save-config my-config.json
```
This runs the interactive configuration and saves it to the specified file, then exits.

### Load Existing Configuration

To run the router with a previously saved configuration:
```bash
./midirouter --config my-config.json
```
This loads the specified configuration file and immediately starts the router without interactive setup.

### Command-Line Options

- `midirouter` - Run interactive configuration and start router
- `midirouter --save-config FILE` - Run interactive configuration, save result to FILE and exit
- `midirouter --config FILE` - Load configuration from FILE and start router
- `midirouter --quiet` - Suppress MIDI message logging during operation
- `midirouter -h` - Show usage information

**Note**: Flags can be combined, e.g., `midirouter --config my-setup.json --quiet`

3. **Follow the configuration wizard:**
   - **Step 1**: Select MIDI input device from the list
   - **Step 2**: Enter base name for outputs (default: "MIDI Router")
   - **Step 3**: Specify number of virtual outputs (1-16)
   - **Step 4**: Configure each output:
     - **Channel Filter**: Optionally filter by MIDI channel (1-16)
     - **Note Range Filter**: Optionally set note range using live detection:
       - Play the lowest note you want to include
       - Play the highest note you want to include
       - Confirm the detected range

4. **Router starts with real-time display:**
   - Shows which outputs receive each MIDI message
   - Displays `[DROPPED]` for filtered messages
   - Press Ctrl+C to stop

## Configuration Examples

### Basic Setup (No Filters)
```
Step 1: Select MIDI Input Device
Available MIDI Input Devices:
1: USB MIDI Device:USB MIDI Device Port-0 128:0

Select input device (1-1): 1

Step 2: Output Configuration
Enter base name for outputs (default: 'MIDI Router'): My Piano
Number of virtual outputs to create: 2

Configuring Output 1 of 2
Name: My Piano Out 1
Enable channel filter? (y/N): n
Enable note range filter? (y/N): n

Configuring Output 2 of 2
Name: My Piano Out 2
Enable channel filter? (y/N): n
Enable note range filter? (y/N): n
```

### Advanced Setup (With Filters)
```
Configuring Output 1 of 2
Name: Bass Out 1
Enable channel filter? (y/N): y
Channel number (1-16): 1
Enable note range filter? (y/N): y

Note Range Configuration
========================
Play the LOWEST note you want to include in this output.
Press Enter when ready to capture lowest note...
Listening for lowest note... (play a note within 30 seconds)
Captured note: C2 (MIDI note 36) on channel 1
Lowest note captured: C2 (MIDI note 36)

Now play the HIGHEST note you want to include in this output.
Press Enter when ready to capture highest note...
Listening for highest note... (play a note within 30 seconds)
Captured note: B3 (MIDI note 59) on channel 1
Highest note captured: B3 (MIDI note 59)
Confirm note range C2 (36) to B3 (59)? (Y/n): y
```

### Save Configuration Example

```
$ ./midirouter --save-config bass-split.json
MIDI Router - Interactive Configuration
======================================

Step 1: Select MIDI Input Device
Available MIDI Input Devices:
1: USB MIDI Device:USB MIDI Device Port-0 128:0

Select input device (1-1): 1
Selected input: USB MIDI Device:USB MIDI Device Port-0 128:0

Step 2: Output Configuration
Enter base name for outputs (default: 'MIDI Router'): Bass Split
Number of virtual outputs to create: 2

[... configuration continues ...]

Configuration saved to bass-split.json

Router configuration complete. Use './midirouter --config bass-split.json' to run with saved config.
```

### Complete Workflow Example

```bash
# 1. Create and save a configuration
./midirouter --save-config piano-split.json

# 2. Later, run the router with the saved configuration
./midirouter --config piano-split.json

# Output:
# MIDI Router - Loading Configuration
# ==================================
# Loaded configuration from piano-split.json
#
# Creating 2 virtual MIDI outputs:
# - Piano Split Out 1 (Channel 1)
# - Piano Split Out 2 (Channel 2)
#
# Configuration Summary:
# Input: USB Piano:USB Piano Port-0 128:0
# Outputs: 2 virtual ports
#
# MIDI router is running. Press Ctrl+C to stop...
```

### Quiet Mode

Use `--quiet` to suppress real-time MIDI message logging for background operation:

```bash
./midirouter --config my-setup.json --quiet
```

**Normal Mode Output:**
```
[Piano Out 1] NoteOn channel: 0, key: 48, velocity: 80
[Piano Out 1] NoteOff channel: 0, key: 48, velocity: 64
[DROPPED] NoteOn channel: 1, key: 72, velocity: 90
```

**Quiet Mode Output:**
```
(no MIDI message logging - only startup/shutdown messages)
```

## Dependencies

- `gitlab.com/gomidi/midi/v2` - MIDI library for Go
- `gitlab.com/gomidi/midi/v2/drivers/rtmididrv` - Cross-platform MIDI driver

## Runtime Display

```
Creating 2 virtual MIDI outputs:
- My Piano Out 1 (Channel 1) (Notes C2-B3)
- My Piano Out 2

Configuration Summary:
Input: USB MIDI Device:USB MIDI Device Port-0 128:0
Outputs: 2 virtual ports

MIDI router is running. Press Ctrl+C to stop...
Routing display format: [Output names] Message

[My Piano Out 1] NoteOn channel: 0, key: 48, velocity: 80
[My Piano Out 1] NoteOff channel: 0, key: 48, velocity: 64
[DROPPED] NoteOn channel: 1, key: 72, velocity: 90
[My Piano Out 2] ControlChange channel: 2, controller: 64, value: 127
```

## Configuration File

Settings are automatically saved to `config.json`:

```json
{
  "input_device": "USB MIDI Device:USB MIDI Device Port-0 128:0",
  "output_base": "My Piano",
  "outputs": [
    {
      "name": "My Piano Out 1",
      "channel_filter": {
        "enabled": true,
        "channel": 1
      },
      "note_range_filter": {
        "enabled": true,
        "min_note": 36,
        "max_note": 59
      }
    },
    {
      "name": "My Piano Out 2",
      "channel_filter": {
        "enabled": false,
        "channel": 0
      },
      "note_range_filter": {
        "enabled": false,
        "min_note": 0,
        "max_note": 0
      }
    }
  ]
}
```
