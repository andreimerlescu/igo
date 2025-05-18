# Variables
APP_NAME=igo
BUILD_DIR=bin
MAIN_PATH=.

# Go build flags
# -s: Strip symbols (reduces binary size)
# -w: Omit DWARF debugging information
LDFLAGS=-ldflags "-s -w"

.PHONY: all mac-intel mac-silicon linux-arm linux clean test

# Create build directory if it doesn't exist
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Build for all platforms
all: mac-intel mac-silicon linux linux-arm

# Build for macOS Intel (amd64)
mac-intel: $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_PATH)

# Build for macOS Silicon (arm64)
mac-silicon: $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_PATH)

# Build for Linux ARM64
linux-arm: $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(MAIN_PATH)

# Build for Linux AMD64
linux: $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PATH)

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Run tests
test:
	./test.sh
