# landrun

A lightweight, secure sandbox for running Linux processes using Landlock LSM. Think firejail, but with kernel-level security and minimal overhead.

## Features

- 🔒 Kernel-level security using Landlock LSM
- 🚀 Lightweight and fast execution
- 🛡️ Fine-grained access control for directories
- 🔄 Support for read and write paths
- ⚡ Optional execution permissions for allowed paths
- 🌐 TCP network access control (binding and connecting)

## Demo

<p align="center">
  <img src="demo.gif" alt="landrun demo" width="700"/>
</p>

## Requirements

- Linux kernel 5.13 or later with Landlock LSM enabled
- Linux kernel 6.8 or later for network restrictions (TCP bind/connect)
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
go build -o landrun cmd/landrun/main.go
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
- `--bind-tcp <port>`: Allow binding to specified TCP port (can be specified multiple times)
- `--connect-tcp <port>`: Allow connecting to specified TCP port (can be specified multiple times)
- `--best-effort`: Use best effort mode, falling back to less restrictive sandbox if necessary [default: enabled]
- `--log-level <level>`: Set logging level (error, info, debug) [default: "error"]

### Important Notes

- You must explicitly add the path to the command you want to run with the `--ro` flag
- For system commands, you typically need to include `/usr/bin`, `/usr/lib`, and other system directories
- When using `--exec`, you still need to specify the directories containing executables with `--ro`
- Network restrictions require Linux kernel 6.8 or later with Landlock ABI v5
- The `--best-effort` flag allows graceful degradation on older kernels that don't support all requested restrictions

### Environment Variables

- `LANDRUN_LOG_LEVEL`: Set logging level (error, info, debug)

### Examples

1. Run a command with read-only access to a directory:

```bash
landrun --ro /usr/bin --ro /lib --ro /lib64 --ro /path/to/dir ls /path/to/dir
```

2. Run a command with write access to a directory:

```bash
landrun --ro /usr/bin --ro /lib --ro /lib64 --rw /path/to/dir touch /path/to/dir/newfile
```

3. Run a command with execution permissions:

```bash
landrun --ro /usr/bin --ro /lib --ro /lib64 --exec /usr/bin/bash
```

4. Run with debug logging:

```bash
landrun --log-level debug --ro /usr/bin --ro /lib --ro /lib64 --ro /path/to/dir ls
```

5. Run with network restrictions:

```bash
landrun --ro /usr/bin --ro /lib --ro /lib64 --bind-tcp 8080 --connect-tcp 53 /usr/bin/my-server
```

This will allow the program to only bind to TCP port 8080 and connect to TCP port 53.

6. Run a DNS client with appropriate permissions:

```bash
landrun --ro /usr/bin --ro /lib --ro /lib64 --ro /etc/resolv.conf --connect-tcp 53 dig example.com
```

This allows DNS resolution by granting access to /etc/resolv.conf and permitting connections to port 53 (DNS).

7. Run a web server with selective network permissions:

```bash
landrun --ro /usr/bin --ro /lib --ro /lib64 --ro /var/www --rw /var/log --bind-tcp 80 --bind-tcp 443 /usr/bin/nginx
```

## Security

landrun uses Linux's Landlock LSM to create a secure sandbox environment. It provides:

- File system access control
- Directory access restrictions
- Execution control
- TCP network restrictions
- Process isolation

Landlock is an access-control system that enables processes to securely restrict themselves and their future children. As a stackable Linux Security Module (LSM), it creates additional security layers on top of existing system-wide access controls, helping to mitigate security impacts from bugs or malicious behavior in applications.

### Landlock Access Control Rights

landrun leverages Landlock's fine-grained access control mechanisms, which include:

**File-specific rights:**

- Execute files (`LANDLOCK_ACCESS_FS_EXECUTE`)
- Write to files (`LANDLOCK_ACCESS_FS_WRITE_FILE`)
- Read files (`LANDLOCK_ACCESS_FS_READ_FILE`)
- Truncate files (`LANDLOCK_ACCESS_FS_TRUNCATE`) - Available since Landlock ABI v3

**Directory-specific rights:**

- Read directory contents (`LANDLOCK_ACCESS_FS_READ_DIR`)
- Remove directories (`LANDLOCK_ACCESS_FS_REMOVE_DIR`)
- Remove files (`LANDLOCK_ACCESS_FS_REMOVE_FILE`)
- Create various filesystem objects (char devices, directories, regular files, sockets, etc.)
- Refer/reparent files across directories (`LANDLOCK_ACCESS_FS_REFER`) - Available since Landlock ABI v2

**Network-specific rights** (requires Linux 6.8+ with Landlock ABI v5):

- Bind to specific TCP ports (`LANDLOCK_ACCESS_NET_BIND_TCP`)
- Connect to specific TCP ports (`LANDLOCK_ACCESS_NET_CONNECT_TCP`)

### Limitations

- Landlock must be supported by your kernel
- Network restrictions require Linux kernel 6.8+ with Landlock ABI v5
- Some operations may require additional permissions
- Files or directories opened before sandboxing are not subject to Landlock restrictions

## Kernel Compatibility Table

| Feature                            | Minimum Kernel Version | Landlock ABI Version |
| ---------------------------------- | ---------------------- | -------------------- |
| Basic filesystem sandboxing        | 5.13                   | 1                    |
| File referring/reparenting control | 5.17                   | 2                    |
| File truncation control            | 6.1                    | 3                    |
| Network TCP restrictions           | 6.8                    | 5                    |

## Troubleshooting

If you receive "permission denied" or similar errors:

1. Ensure you've added all necessary paths with `--ro` or `--rw`
2. Try running with `--log-level debug` to see detailed permission information
3. Check that Landlock is supported and enabled on your system:
   ```bash
   grep -E 'landlock|lsm=' /boot/config-$(uname -r)
   ```
   You should see `CONFIG_SECURITY_LANDLOCK=y` and `lsm=landlock,...` in the output
4. For network restrictions, verify your kernel version is 6.8+ with Landlock ABI v5:
   ```bash
   uname -r
   ```

## Technical Details

### Implementation

This project uses the `landlock-lsm/go-landlock` package for sandboxing, which provides both filesystem and network restrictions. The current implementation supports:

- Read/write/execute restrictions for files and directories
- TCP port binding restrictions
- TCP port connection restrictions
- Best-effort mode for graceful degradation on older kernels

### Best-Effort Mode

When using `--best-effort` (enabled by default), landrun will gracefully degrade to using the best available Landlock version on the current kernel. This means:

- On Linux 6.8+: Full filesystem and network restrictions
- On Linux 6.1-6.7: Filesystem restrictions including truncation, but no network restrictions
- On Linux 5.17-6.0: Basic filesystem restrictions including file reparenting, but no truncation control or network restrictions
- On Linux 5.13-5.16: Basic filesystem restrictions without file reparenting, truncation control, or network restrictions
- On older Linux: No restrictions (sandbox disabled)

## Future Features

Based on the Linux Landlock API capabilities, we plan to add:

- 🔒 Enhanced filesystem controls with more fine-grained permissions
- 🌐 Support for UDP and other network protocol restrictions (when supported by Linux kernel)
- 🔄 Process scoping and resource controls
- 🛡️ Additional security features as they become available in the Landlock API

## License

This project is licensed under the GNU General Public License v2

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
