# landrun

A lightweight, secure sandbox for running Linux processes using Landlock. Think firejail, but with kernel-level security and minimal overhead.

Linux Landlock is a kernel-native security module that lets unprivileged processes sandbox themselves.

Landrun is designed to make it practical to sandbox any command with fine-grained filesystem and network access controls. No root. No containers. No SELinux/AppArmor configs.

It's lightweight, auditable, and wraps Landlock v5 features (file access + TCP restrictions).

## Features

- 🔒 Kernel-level security using Landlock
- 🚀 Lightweight and fast execution
- 🛡️ Fine-grained access control for directories
- 🔄 Support for read and write paths
- ⚡ Path-specific execution permissions
- 🌐 TCP network access control (binding and connecting)

## Demo

<p align="center">
  <img src="demo.gif" alt="landrun demo" width="700"/>
</p>

## Requirements

- Linux kernel 5.13 or later with Landlock enabled
- Linux kernel 6.7 or later for network restrictions (TCP bind/connect)
- Go 1.18 or later (for building from source)

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

### Distros

#### Arch (AUR)
maintained by [Vcalv](https://github.com/vcalv)

[AUR](https://aur.archlinux.org/packages/landrun-git)
```bash
yay -S landrun
```

#### Slackware
maintained by [r1w1s1](https://github.com/r1w1s1)

[Slackbuild](https://slackbuilds.org/repository/15.0/network/landrun/?search=landrun)
```bash
sudo sbopkg -i packagename
```

## Usage

Basic syntax:

```bash
landrun [options] <command> [args...]
```

### Options

- `--ro <path>`: Allow read-only access to specified path (can be specified multiple times or as comma-separated values)
- `--rox <path>`: Allow read-only access with execution to specified path (can be specified multiple times or as comma-separated values)
- `--rw <path>`: Allow read-write access to specified path (can be specified multiple times or as comma-separated values)
- `--rwx <path>`: Allow read-write access with execution to specified path (can be specified multiple times or as comma-separated values)
- `--bind-tcp <port>`: Allow binding to specified TCP port (can be specified multiple times or as comma-separated values)
- `--connect-tcp <port>`: Allow connecting to specified TCP port (can be specified multiple times or as comma-separated values)
- `--env <var>`: Environment variable to pass to the sandboxed command (format: KEY=VALUE or just KEY to pass current value)
- `--best-effort`: Use best effort mode, falling back to less restrictive sandbox if necessary [default: disabled]
- `--log-level <level>`: Set logging level (error, info, debug) [default: "error"]
- `--unrestricted-network`: allows unrestricted network access.
- `--unrestricted-filesystem`: allows unrestricted filesystem access.

### Important Notes

- You must explicitly add the directory to the command you want to run with `--rox` flag
- For system commands, you typically need to include `/usr/bin`, `/usr/lib`, and other system directories
- Use `--rwx` for directories where you need both write access and the ability to execute files
- Network restrictions require Linux kernel 6.7 or later with Landlock ABI v4
- By default, no environment variables are passed to the sandboxed command. Use `--env` to explicitly pass environment variables
- The `--best-effort` flag allows graceful degradation on older kernels that don't support all requested restrictions
- Paths can be specified either using multiple flags or as comma-separated values (e.g., `--ro /usr,/lib,/home`)

### Environment Variables

- `LANDRUN_LOG_LEVEL`: Set logging level (error, info, debug)

### Examples

1. Run a command with read-only access to a directory:

```bash
landrun --rox /usr/ --ro /path/to/dir ls /path/to/dir
```

2. Run a command with write access to a directory:

```bash
landrun --rox /usr/bin --ro /lib --rw /path/to/dir touch /path/to/dir/newfile
```

3. Run a command with execution permissions:

```bash
landrun --rox /usr/ --ro /lib,/lib64 /usr/bin/bash
```

4. Run with debug logging:

```bash
landrun --log-level debug --rox /usr/ --ro /lib,/lib64,/path/to/dir ls /path/to/dir
```

5. Run with network restrictions:

```bash
landrun --rox /usr/ --ro /lib,/lib64 --bind-tcp 8080 --connect-tcp 80 /usr/bin/my-server
```

This will allow the program to only bind to TCP port 8080 and connect to TCP port 80.

6. Run a DNS client with appropriate permissions:

```bash
landrun --log-level debug --ro /etc,/usr --rox /usr/ --connect-tcp 443 nc kernel.org 443
```

This allows connections to port 443, requires access to /etc/resolv.conf for resolving DNS.

7. Run a web server with selective network permissions:

```bash
landrun --rox /usr/bin --ro /lib,/lib64,/var/www --rwx /var/log --bind-tcp 80,443 /usr/bin/nginx
```

8. Running anything without providing parameters is... maximum security jail!

```bash
landrun ls
```

9. If you keep getting permission denied without knowing what exactly going on, best to use strace with it.

```bash
landrun --rox /usr strace -f -e trace=all ls
```

10. Run with specific environment variables:

```bash
landrun --rox /usr --ro /etc --env HOME --env PATH --env CUSTOM_VAR=my_value -- env
```

This example passes the current HOME and PATH variables, plus a custom variable named CUSTOM_VAR.

## Security

landrun uses Linux's Landlock to create a secure sandbox environment. It provides:

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

**Network-specific rights** (requires Linux 6.7+ with Landlock ABI v4):

- Bind to specific TCP ports (`LANDLOCK_ACCESS_NET_BIND_TCP`)
- Connect to specific TCP ports (`LANDLOCK_ACCESS_NET_CONNECT_TCP`)

### Limitations

- Landlock must be supported by your kernel
- Network restrictions require Linux kernel 6.7 or later with Landlock ABI v4
- Some operations may require additional permissions
- Files or directories opened before sandboxing are not subject to Landlock restrictions

## Kernel Compatibility Table

| Feature                            | Minimum Kernel Version | Landlock ABI Version |
| ---------------------------------- | ---------------------- | -------------------- |
| Basic filesystem sandboxing        | 5.13                   | 1                    |
| File referring/reparenting control | 5.19                   | 2                    |
| File truncation control            | 6.2                    | 3                    |
| Network TCP restrictions           | 6.7                    | 4                    |
| IOCTL on special files             | 6.10                   | 5                    |

## Troubleshooting

If you receive "permission denied" or similar errors:

1. Ensure you've added all necessary paths with `--ro` or `--rw`
2. Try running with `--log-level debug` to see detailed permission information
3. Check that Landlock is supported and enabled on your system:
   ```bash
   grep -E 'landlock|lsm=' /boot/config-$(uname -r)
   # alternatively, if there are no /boot/config-* files
   zgrep -iE 'landlock|lsm=' /proc/config.gz
   # another alternate method
   grep -iE 'landlock|lsm=' /lib/modules/$(uname -r)/config
   ```
   You should see `CONFIG_SECURITY_LANDLOCK=y` and `lsm=landlock,...` in the output
4. For network restrictions, verify your kernel version is 6.7+ with Landlock ABI v4:
   ```bash
   uname -r
   ```

## Technical Details

### Implementation

This project uses the [landlock-lsm/go-landlock](https://github.com/landlock-lsm/go-landlock) package for sandboxing, which provides both filesystem and network restrictions. The current implementation supports:

- Read/write/execute restrictions for files and directories
- TCP port binding restrictions
- TCP port connection restrictions
- Best-effort mode for graceful degradation on older kernels

### Best-Effort Mode

When using `--best-effort` (disabled by default), landrun will gracefully degrade to using the best available Landlock version on the current kernel. This means:

- On Linux 6.7+: Full filesystem and network restrictions
- On Linux 6.2-6.6: Filesystem restrictions including truncation, but no network restrictions
- On Linux 5.19-6.1: Basic filesystem restrictions including file reparenting, but no truncation control or network restrictions
- On Linux 5.13-5.18: Basic filesystem restrictions without file reparenting, truncation control, or network restrictions
- On older Linux: No restrictions (sandbox disabled)

### Tests

The project includes a comprehensive test suite that verifies:

- Basic filesystem access controls (read-only, read-write, execute)
- Directory traversal and path handling
- Network restrictions (TCP bind/connect)
- Environment variable isolation
- System command execution
- Edge cases and regression tests

Run the tests with:

```bash
./test.sh
```

Use `--keep-binary` to preserve the test binary after completion:

```bash
./test.sh --keep-binary
```

Use `--use-system` to test against the system-installed landrun binary:

```bash
./test.sh --use-system
```

## Future Features

Based on the Linux Landlock API capabilities, we plan to add:

- 🔒 Enhanced filesystem controls with more fine-grained permissions
- 🌐 Support for UDP and other network protocol restrictions (when supported by Linux kernel)
- 🔄 Process scoping and resource controls
- 🛡️ Additional security features as they become available in the Landlock API

## Acknowledgements

This project wouldn't exist without:

- [Landlock](https://landlock.io), the kernel security module enabling unprivileged sandboxing - maintained by [@l0kod](https://github.com/l0kod)
- [go-landlock](https://github.com/landlock-lsm/go-landlock), the Go bindings powering this tool - developed by [@gnoack](https://github.com/gnoack)

## License

This project is licensed under the GNU General Public License v2

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
