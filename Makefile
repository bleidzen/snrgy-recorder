.PHONY: all windows darwin clean

BINARY_NAME=snrgy-recorder
VERSION=1.0.0

all: windows darwin

deps:
	go mod tidy

windows: deps
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -ldflags="-H windowsgui -s -w" -o $(BINARY_NAME).exe .

darwin: deps
	@echo "Building for macOS (Intel)..."
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-mac-intel .
	@echo "Building for macOS (Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BINARY_NAME)-mac-arm .

linux: deps
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-linux .

clean:
	rm -f $(BINARY_NAME).exe $(BINARY_NAME)-mac-intel $(BINARY_NAME)-mac-arm $(BINARY_NAME)-linux
