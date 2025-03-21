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

print_status "Setting up test environment..."
rm -rf "$TEST_DIR"
mkdir -p "$RO_DIR" "$RW_DIR" "$EXEC_DIR"

# Create test files
echo "readonly content" > "$RO_DIR/test.txt"
echo "readwrite content" > "$RW_DIR/test.txt"
echo "#!/bin/bash" > "$EXEC_DIR/test.sh"
echo "echo 'executable content'" >> "$EXEC_DIR/test.sh"
chmod +x "$EXEC_DIR/test.sh"

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
        return 1
    fi
}

# Test cases
print_status "Starting test cases..."

# Test 1: Basic read-only access
run_test "Read-only access" \
    "./landrun --ro /usr --ro /lib --ro /lib64 --ro $RO_DIR --rw $RW_DIR --ro $RW_DIR -- cat $RO_DIR/test.txt" \
    0

# Test 2: Read-write access
run_test "Read-write access" \
    "./landrun --ro /usr --ro /lib --ro /lib64 --ro $RO_DIR --rw $RW_DIR --ro $RW_DIR -- bash -c 'echo new content > $RW_DIR/test.txt && cat $RW_DIR/test.txt'" \
    0

# Test 3: No write access to read-only directory
run_test "No write access to read-only directory" \
    "./landrun --ro /usr --ro /lib --ro /lib64 --ro $RO_DIR --rw $RW_DIR --ro $RW_DIR -- bash -c 'echo new content > $RO_DIR/test.txt'" \
    1

# Test 4: No read access to read-write directory
run_test "No read access to read-write directory" \
    "./landrun --ro /usr --ro /lib --ro /lib64 --ro $RO_DIR --rw $RW_DIR -- cat $RW_DIR/test.txt" \
    1

# Test 5: Executable access (should fail without --exec)
run_test "No executable access without --exec" \
    "./landrun --ro /usr --ro /lib --ro /lib64 --ro $EXEC_DIR --rw $RW_DIR --ro $RW_DIR -- $EXEC_DIR/test.sh" \
    1

# Test 6: Executable access with --exec
run_test "Executable access with --exec" \
    "./landrun --ro /usr --ro /lib --ro /lib64 --ro $EXEC_DIR --rw $RW_DIR --ro $RW_DIR --exec -- $EXEC_DIR/test.sh" \
    0

# Test 7: Access to non-existent path
run_test "Access to non-existent path" \
    "./landrun --ro /usr --ro /lib --ro /lib64 --ro /nonexistent/path --rw $RW_DIR --ro $RW_DIR -- cat /nonexistent/path" \
    1

# Test 8: Empty configuration
run_test "Empty configuration" \
    "./landrun --ro /usr --ro /lib --ro /lib64 --ro /etc -- cat /etc/passwd" \
    0

# Test 9: Multiple read-only paths
run_test "Multiple read-only paths" \
    "./landrun --ro /usr --ro /lib --ro /lib64 --ro $RO_DIR --ro /etc --rw $RW_DIR --ro $RW_DIR -- cat $RO_DIR/test.txt" \
    0

# Test 10: Multiple read-write paths
run_test "Multiple read-write paths" \
    "./landrun --ro /usr --ro /lib --ro /lib64 --ro $RO_DIR --rw $RW_DIR --ro $RW_DIR --rw /tmp --ro /tmp -- bash -c 'echo test > $RW_DIR/test.txt'" \
    0

# Cleanup
print_status "Cleaning up..."
rm -rf "$TEST_DIR"
rm -f landrun

print_success "All tests completed!" 