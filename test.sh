#!/bin/bash

# Don't exit on error, we'll handle errors in the run_test function
set +e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Build the binary
print_status "Building landrun binary..."
go build -o landrun cmd/landrun/main.go
if [ $? -ne 0 ]; then
    print_error "Failed to build landrun binary"
    exit 1
fi
print_success "Binary built successfully"

# Create test directories
TEST_DIR="test_env"
RO_DIR="$TEST_DIR/ro"
RW_DIR="$TEST_DIR/rw"
EXEC_DIR="$TEST_DIR/exec"
NESTED_DIR="$TEST_DIR/nested/path/deep"

print_status "Setting up test environment..."
rm -rf "$TEST_DIR"
mkdir -p "$RO_DIR" "$RW_DIR" "$EXEC_DIR" "$NESTED_DIR"

# Create test files
echo "readonly content" > "$RO_DIR/test.txt"
echo "readwrite content" > "$RW_DIR/test.txt"
echo "nested content" > "$NESTED_DIR/test.txt"
echo "#!/bin/bash" > "$EXEC_DIR/test.sh"
echo "echo 'executable content'" >> "$EXEC_DIR/test.sh"
chmod +x "$EXEC_DIR/test.sh"

# Create a script in RW dir to test execution in RW dirs
echo "#!/bin/bash" > "$RW_DIR/rw_script.sh"
echo "echo 'this script is in a read-write directory'" >> "$RW_DIR/rw_script.sh"
chmod +x "$RW_DIR/rw_script.sh"

# Function to run a test case
run_test() {
    local name="$1"
    local cmd="$2"
    local expected_exit="$3"
    
    print_status "Running test: $name"
    eval "$cmd"
    local exit_code=$?
    
    if [ $exit_code -eq $expected_exit ]; then
        print_success "Test passed: $name"
        return 0
    else
        print_error "Test failed: $name (expected exit $expected_exit, got $exit_code)"
        exit 1
    fi
}

# Test cases
print_status "Starting test cases..."

# Basic access tests
run_test "Read-only access to file" \
    "./landrun --log-level debug --rox /usr --ro /lib --ro /lib64 --ro $RO_DIR -- cat $RO_DIR/test.txt" \
    0


run_test "Read-write access to file" \
    "./landrun --log-level debug --rox /usr --ro /lib --ro /lib64 --ro $RO_DIR --rw $RW_DIR touch $RW_DIR/new.txt" \
    0

run_test "No write access to read-only directory" \
    "./landrun --log-level debug --rox /usr --ro /lib --ro /lib64 --ro $RO_DIR --rw $RW_DIR touch $RO_DIR/new.txt" \
    1

# Executable permission tests
run_test "Execute access with rox flag" \
    "./landrun --log-level debug --rox /usr --ro /lib --ro /lib64 --rox $EXEC_DIR -- $EXEC_DIR/test.sh" \
    0

run_test "No execute access with just ro flag" \
    "./landrun --log-level debug --rox /usr --ro /lib --ro /lib64 --ro $EXEC_DIR -- $EXEC_DIR/test.sh" \
    1

run_test "Execute access in read-write directory" \
    "./landrun --log-level debug --rox /usr --ro /lib --ro /lib64 --rwx $RW_DIR -- $RW_DIR/rw_script.sh" \
    0

run_test "No execute access in read-write directory without rwx" \
    "./landrun --log-level debug --rox /usr --ro /lib --ro /lib64 --rw $RW_DIR -- $RW_DIR/rw_script.sh" \
    1

# Directory traversal tests
run_test "Directory traversal with root access" \
    "./landrun --log-level debug --rox / -- ls /usr" \
    0

run_test "Deep directory traversal" \
    "./landrun --log-level debug --rox / -- ls $NESTED_DIR" \
    0

# Multiple paths and complex specifications
run_test "Multiple read paths" \
    "./landrun --log-level debug --rox /usr --ro /lib --ro /lib64 --ro $RO_DIR --ro $NESTED_DIR -- cat $NESTED_DIR/test.txt" \
    0

run_test "Comma-separated paths" \
    "./landrun --log-level debug --rox /usr --ro /lib,/lib64,$RO_DIR -- cat $RO_DIR/test.txt" \
    0

# System command tests
run_test "Simple system command" \
    "./landrun --log-level debug --rox /usr --ro  /etc -- whoami" \
    0

run_test "System command with arguments" \
    "./landrun --log-level debug --rox / -- ls -la /usr/bin" \
    0

# Edge cases
run_test "Non-existent read-only path" \
    "./landrun --log-level debug --ro /usr --ro /lib --ro /lib64 --ro /nonexistent/path -- ls" \
    1

run_test "No configuration" \
    "./landrun --log-level debug -- ls /" \
    1

# Process creation and redirection tests
run_test "Process creation with pipe" \
    "./landrun --log-level debug --rox / -- bash -c 'ls /usr | grep bin'" \
    0

run_test "File redirection" \
    "./landrun --log-level debug --rox / --rw $RW_DIR -- bash -c 'ls /usr > $RW_DIR/output.txt && cat $RW_DIR/output.txt'" \
    0

# Network restrictions tests (if kernel supports it)
run_test "TCP connection without permission" \
    "./landrun --log-level debug --rox /usr --ro / -- curl -s --connect-timeout 2 https://example.com" \
    7

run_test "TCP connection with permission" \
    "./landrun --log-level debug --rox /usr --ro / --connect-tcp 443 -- curl -s --connect-timeout 2 https://example.com" \
    0

# Combining different permission types
run_test "Mixed permissions" \
    "./landrun --log-level debug --rox /usr --ro /lib --ro /lib64 --rox $EXEC_DIR --rwx $RW_DIR -- bash -c '$EXEC_DIR/test.sh > $RW_DIR/output.txt && cat $RW_DIR/output.txt'" \
    0

# Specific regression tests for bugs we fixed
run_test "Root path traversal regression test" \
    "./landrun --log-level debug --rox /usr -- /usr/bin/ls /usr" \
    0

run_test "Execute from read-only paths regression test" \
    "./landrun --log-level debug --rox /usr --ro /usr/bin -- /usr/bin/id" \
    0

# Cleanup
print_status "Cleaning up..."
rm -rf "$TEST_DIR"
rm -f landrun

print_success "All tests completed!" 