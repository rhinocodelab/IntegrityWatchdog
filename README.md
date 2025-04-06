# IntegrityWatchdog

A lightweight file integrity monitoring CLI tool for Linux/Unix systems. It helps detect unauthorized or unexpected changes in your file system.

## Features

- Baseline creation for file system state
- File change detection (added, modified, deleted)
- Permission monitoring
- JSON output support
- Daemon mode for continuous monitoring
- Cross-platform support (Linux, macOS)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/rhinocodelab/IntegrityWatchdog.git
cd IntegrityWatchdog

# Build the binary
go build -o fim

# Move to a directory in your PATH
sudo mv fim /usr/local/bin/
```

### Using Go

```bash
go install github.com/rhinocodelab/IntegrityWatchdog@latest
```

## Configuration

The tool uses a configuration file located at `~/.fim/fim.conf`. The configuration file is created automatically when you run `fim init` for the first time.

Example configuration:

```ini
[monitor]
# Required: List of paths to monitor (comma-separated)
paths = /etc, /usr/bin

# Optional: Paths to exclude from monitoring (comma-separated)
exclude = /tmp, /var/log, /proc

[logging]
# Optional: Log file path
logfile = /var/log/fim.log

[output]
# Optional: Enable verbose output
verbose = true
```

## Usage

### Initialize

Create a new baseline of your file system:

```bash
fim init
```

This will:
1. Create a `.fim` directory in your home folder
2. Create a configuration file if it doesn't exist
3. Scan configured directories
4. Store file metadata and hashes
5. Create a baseline file at `~/.fim/baseline.json`

### Scan

Scan for changes since the last baseline:

```bash
fim scan
```

Output will show added, modified, and deleted files.

### Daemon Mode

Run the tool in daemon mode for continuous monitoring:

```bash
fim scan --daemon
```

You can specify a custom scan interval:

```bash
fim scan --daemon --interval 10m
```

To stop the daemon:

```bash
fim stop
```

### JSON Output

Get scan results in JSON format:

```bash
fim scan --json
```

## Development

### Building

```bash
go build -o fim
```

### Testing

```bash
go test ./...
```

## Security Considerations

- The tool requires appropriate permissions to access monitored directories
- Consider running with elevated privileges for system directories
- Be cautious with sensitive file paths in the configuration

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 