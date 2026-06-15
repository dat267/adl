BIN      := $(notdir $(CURDIR))
VERSION  ?= dev
LDFLAGS  := -s -w -X main.version=$(VERSION)
DIST     := dist

# OUT is overridden by CI: make build OUT=/path/to/binary
OUT ?= $(BIN)

# Platforms used by `make dist` for local cross-compilation
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	linux/arm \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64

.PHONY: build dist clean help

## build: Build for current platform (CI sets OUT= and env GOOS/GOARCH/VERSION)
build:
	go build -ldflags "$(LDFLAGS)" -o $(OUT) .

## dist: Cross-compile all platforms into ./dist/
dist: clean
	@mkdir -p $(DIST)
	@$(foreach P,$(PLATFORMS), \
		$(eval GOOS   := $(word 1,$(subst /, ,$(P)))) \
		$(eval GOARCH := $(word 2,$(subst /, ,$(P)))) \
		$(eval EXT    := $(if $(filter windows,$(GOOS)),.exe,)) \
		$(eval _OUT   := $(DIST)/$(BIN)-$(VERSION)-$(GOOS)-$(GOARCH)$(EXT)) \
		echo "  Building $(_OUT)..."; \
		GOOS=$(GOOS) GOARCH=$(GOARCH) \
			go build -ldflags "$(LDFLAGS)" -o $(_OUT) . || exit 1; \
	)
	@echo "Done. Artifacts in ./$(DIST)/"

## clean: Remove build artifacts
clean:
	@rm -rf $(DIST) $(BIN) $(BIN).exe

## help: List available targets
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
