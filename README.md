# FMove

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
Due to Native CGO compilation requirements, native Windows building is extremely complex. Windows users are strongly recommended to run FMove inside **Windows Subsystem for Linux (WSL)** using the Linux instructions above.

### Method 1: Go Install (Recommended)
If you have Go installed, you can download, compile, and install the tool globally in a single command:
```bash
go install github.com/sulaimaneksambi/fmove@latest
```

### Method 2: Homebrew Tap
If you prefer Homebrew for managing binaries:
```bash
brew tap sulaimaneksambi/fmove
brew install fmove
```
*(Note to dev: Make sure you create a `homebrew-fmove` repository on GitHub and place your `fmove.rb` formula in it for this to work!)*

---

## 🛠️ Local Development & Building

If you are cloning this repository to build it yourself, you don't have to worry about configuring C-compiler flags. We have provided a `Makefile`.

```bash
# Clone the repo
git clone https://github.com/sulaimaneksambi/fmove.git
cd fmove

# Build the binary automatically
make

# Run the UI Demonstration (No files required)
make animation

# Build and run the tool in the current directory
make run
```

## 🎮 Usage
Simply navigate to any directory containing `.MP4` or `.MOV` files and run:
```bash
fmove
```
You will be prompted with an interactive menu to either perform a **Normal Concatenation** (keeping your files) or engage **Space Recovery Mode** (dynamically freeing space).
