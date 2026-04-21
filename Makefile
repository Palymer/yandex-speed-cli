# Сборка yandex-speed-cli (x64=amd64, x86=386). VERSION=1.2.3 make all
# macOS: только amd64 (Intel) и arm64 (Apple Silicon) — порт darwin/386 в Go удалён.
.PHONY: all clean native \
	linux-amd64 linux-386 linux-arm64 \
	darwin-amd64 darwin-arm64 \
	windows-amd64 windows-386

DIST ?= dist
LDFLAGS := -s -w
VERSION ?= dev
LDX := -X main.version=$(VERSION)

all: linux-amd64 linux-386 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64 windows-386

native:
	mkdir -p $(DIST)
	go build -trimpath -ldflags="$(LDFLAGS) $(LDX)" -o $(DIST)/yandex-speed-cli .

linux-amd64:
	mkdir -p $(DIST)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS) $(LDX)" -o $(DIST)/yandex-speed-cli-linux-amd64 .

linux-386:
	mkdir -p $(DIST)
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS) $(LDX)" -o $(DIST)/yandex-speed-cli-linux-386 .

linux-arm64:
	mkdir -p $(DIST)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS) $(LDX)" -o $(DIST)/yandex-speed-cli-linux-arm64 .

darwin-amd64:
	mkdir -p $(DIST)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS) $(LDX)" -o $(DIST)/yandex-speed-cli-darwin-amd64 .

darwin-arm64:
	mkdir -p $(DIST)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS) $(LDX)" -o $(DIST)/yandex-speed-cli-darwin-arm64 .

windows-amd64:
	mkdir -p $(DIST)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS) $(LDX)" -o $(DIST)/yandex-speed-cli-windows-amd64.exe .

windows-386:
	mkdir -p $(DIST)
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -trimpath -ldflags="$(LDFLAGS) $(LDX)" -o $(DIST)/yandex-speed-cli-windows-386.exe .

clean:
	rm -rf $(DIST)
