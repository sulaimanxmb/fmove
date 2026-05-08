# FMove: Native Go + FFmpeg Concatenation & Space Recovery

FMove is a high-performance native Go CLI utility designed specifically to solve storage bottlenecks when concatenating large video files (like GoPro clips).

Instead of wrapping the `ffmpeg` CLI binary, FMove uses **CGO** to interface directly with FFmpeg's native C libraries (`libavformat`, `libavcodec`, `libavutil`). This allows the tool to:
1. **Automatically sort clips** by their internal metadata creation time.
2. **Stream copy packets natively** without re-encoding, preserving exact original quality.
3. **Dynamically recover space** by instantly deleting each source clip the millisecond its stream is safely written to the output file.

## Requirements
- Go 1.21+
- FFmpeg development libraries installed via Homebrew on macOS (`brew install ffmpeg`)

## Building
```bash
# Ensure pkg-config can find the Homebrew FFmpeg libraries
export PKG_CONFIG_PATH="/opt/homebrew/lib/pkgconfig"
export CGO_CFLAGS="-I/opt/homebrew/include"
export CGO_LDFLAGS="-L/opt/homebrew/lib -lavformat -lavcodec -lavutil"

go build -o fmove .
```

## Usage
Simply run the binary in a directory containing `.MP4` files:
```bash
./fmove
```
