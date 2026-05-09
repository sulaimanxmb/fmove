# Makefile for FMove (Native CGO Build)

# Apple Silicon (M2) Homebrew paths for FFmpeg C-libraries
export PKG_CONFIG_PATH=/opt/homebrew/lib/pkgconfig
export CGO_CFLAGS=-I/opt/homebrew/include
export CGO_LDFLAGS=-L/opt/homebrew/lib -lavformat -lavcodec -lavutil

.PHONY: all build run demo clean

all: build

build:
	@echo "==> Building FMove Native Engine..."
	go build -ldflags="-X main.AppVersion=$$(git describe --tags --abbrev=0 2>/dev/null || echo dev)" -o fmove .
	@echo "==> Build successful! Run ./fmove to start."

run: build
	@echo "==> Running FMove..."
	./fmove

animation: build
	@echo "==> Running FMove in Animation Mode..."
	./fmove --animation

clean:
	@echo "==> Cleaning up..."
	rm -f fmove fmove_output.mp4
