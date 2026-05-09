# FMove

# FMove

[![Go Report Card](https://goreportcard.com/badge/github.com/sulaimanxmb/fmove)](https://goreportcard.com/report/github.com/sulaimanxmb/fmove)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sulaimanxmb/fmove)](https://github.com/sulaimanxmb/fmove)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-blue)](#)

> A blazing-fast, native Go CLI for lossless video concatenation and SSD space recovery.


```text
  ______  __       __   ______   ____    ____  ______
 |   ___||  \     /  | /  __  \  \   \  /   / |   ___|
 |  |__  |   \   /   ||  |  |  |  \   \/   /  |  |__
 |   __| |    \_/    ||  |  |  |   \      /   |   __|
 |  |    |  |\   /|  ||  '--'  |    \    /    |  |____
 |__|    |__| \_/ |__| \______/      \__/     |_______|

                                           by Sulaiman
                       Native Space Recovery & Concatenator
```

This is a tool which I created as I was having trouble in concatenating multiple video files from GoPro with `ffmpeg`

FMove is a high-performance, native Go CLI utility designed specifically to solve storage bottlenecks when concatenating large video files (like GoPro and iPhone clips). 

Instead of wrapping the `ffmpeg` CLI binary, FMove uses **CGO** to interface directly with FFmpeg's native C libraries (`libavformat`, `libavcodec`, `libavutil`). 

### Features
1. **Intelligent Sorting**: Automatically sorts clips by their internal metadata creation time.
2. **Lossless Stream Copying**: Copies packets natively without re-encoding, preserving exact original quality while handling complex Apple HEVC/Dolby Vision codecs.
3. **Space Recovery Mode**: Dynamically recovers SSD space by instantly deleting each source clip the millisecond its stream is safely written to the output file.
4. **Interactive UI**: A vibrant, animated, rainbow CLI experience built natively in Go.

---

## 🚀 Installation

FMove relies on native C-libraries, so make sure you have `ffmpeg` and `pkg-config` installed on your system first. 

### Prerequisites (macOS)
```bash
brew install ffmpeg pkg-config
```

### Prerequisites (Linux)
```bash
sudo apt install libavformat-dev libavcodec-dev libavutil-dev pkg-config
```

### Prerequisites (Windows)
FMove automatically cross-compiles Windows binaries. You **do not** need to install FFmpeg or a C-compiler on Windows.
1. Go to the [Releases Page](https://github.com/sulaimanxmb/fmove/releases).
2. Download `fmove-windows-amd64.zip`.
3. Extract the folder. It contains `fmove.exe` and the required FFmpeg `.dll` files.
4. Run `fmove.exe` in any directory containing your video files!

### Method 1: Go Install (Recommended)
If you have Go installed, you can download, compile, and install the tool globally in a single command:
```bash
go install github.com/sulaimanxmb/fmove@latest
```

### Method 2: Homebrew Tap
If you prefer Homebrew for managing binaries:
```bash
brew tap sulaimanxmb/fmove
brew install fmove
```

---

## 🛠️ Local Development & Building

If you are cloning this repository to build it yourself, you don't have to worry about configuring C-compiler flags. We have provided a `Makefile`.

```bash
# Clone the repo
git clone https://github.com/sulaimanxmb/fmove.git
cd fmove

# Build the binary automatically
make

# Build and run the tool in the current directory
make run
```

## 🎮 Usage
Simply navigate to any directory containing `.MP4` or `.MOV` files and run:
```bash
fmove -c .
```

To combine specific files explicitly, separate them with a comma:
```bash
fmove -c "clip1.mp4, clip2.mp4"
```

FMove has several advanced options for automation, output naming, and filtering. To see all available flags, run:
```bash
fmove --help
```
