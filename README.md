# landrun

A lightweight, secure sandbox for running Linux processes using Landlock LSM. Think firejail, but with kernel-level security and minimal overhead.

## Features

- 🔒 Kernel-level security using Landlock LSM
- 🚀 Lightweight and fast execution
- 🛡️ Fine-grained access control for files and directories
- 🔄 Support for read-only and read-write paths
- ⚡ Optional execution permissions for allowed paths
- 📝 Configurable logging levels

## Requirements

- Linux kernel 5.13 or later with Landlock LSM enabled
- Go 1.24.1 or later (for building from source)

## Installation

### Quick Install

```bash
go install github.com/zouuup/landrun/cmd/landrun@latest
```

### From Source

```bash
git clone https://github.com/zouuup/landrun.git
cd landrun
go build
sudo cp landrun /usr/local/bin/
```

## Usage

Basic syntax:

```bash
landrun [options] <command> [args...]
```

### Options

- `--ro <path>`: Allow read-only access to specified path (can be specified multiple times)
- `--rw <path>`: Allow read-write access to specified path (can be specified multiple times)
- `--exec`: Allow executing files in allowed paths
- `--log-level <level>`: Set logging level (error, info, debug) [default: "info"]

### Environment Variables

- `LANDRUN_LOG_LEVEL`: Set logging level (error, info, debug)

### Examples

1. Run a command with read-only access to a directory:

```bash
landrun --ro /path/to/dir ls /path/to/dir
```

2. Run a command with read-write access to a directory:

```bash
landrun --rw /path/to/dir touch /path/to/dir/newfile
```

3. Run a command with execution permissions:

```bash
landrun --ro /usr/bin --exec /usr/bin/bash
```

4. Run with debug logging:

```bash
landrun --log-level debug --ro /path/to/dir ls
```

## Security

landrun uses Linux's Landlock LSM to create a secure sandbox environment. It provides:

- File system access control
- Directory access restrictions
- Execution control
- Process isolation

## Future Features

Based on the Linux Landlock API capabilities, we plan to add:

- 🌐 Network access control

  - Port binding restrictions
  - TCP/UDP connection controls
  - Network protocol filtering

- 🔒 Enhanced filesystem controls

  - Truncate operation controls
  - File descriptor inheritance rules
  - Directory hierarchy restrictions

- 🔄 Process scoping

  - IPC (Inter-Process Communication) restrictions
  - Resource access limitations
  - Cross-domain communication controls

- 🛡️ Additional security features
  - Network namespace integration
  - User namespace support
  - Capability restrictions

## License

This project is licensed under the GNU General Public License v2

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
