# adl (ASMR Downloader)

A fast, lightweight, dependency-free CLI tool for downloading ASMR audio tracks and metadata from `asmr-200.com` (asmr.one).

It automatically organizes downloads by circle and work title, fetches all high-quality artwork, saves the work's metadata (`info.json`), and intelligently skips files you've already downloaded.

## Installation

### 1. Pre-compiled Binaries (Recommended)

You can download ready-to-run executables for your operating system directly from our **[Releases Page](https://github.com/dat267/adl/releases)**.

**For Windows:**
1. Download the `adl-windows-amd64.exe` file.
2. Open your Command Prompt or PowerShell in the folder where you downloaded it.
3. Run: `.\adl-windows-amd64.exe [RJ-ID]`

**For macOS/Linux:**
1. Download the appropriate binary (e.g., `adl-darwin-amd64` for Intel Macs, `adl-darwin-arm64` for Apple Silicon, `adl-linux-amd64` for Linux).
2. Open your terminal.
3. Make it executable: `chmod +x adl-darwin-arm64`
4. Run it: `./adl-darwin-arm64 [RJ-ID]`

*(Tip: Rename the downloaded file to `adl` and move it to a directory in your PATH, such as `/usr/local/bin`, to run it from anywhere by just typing `adl`.)*

### 2. Install using Go

If you already have Go installed, you can build and install the latest version directly from the source code.

```bash
go install github.com/dat267/adl@latest
```

This will automatically place the `adl` binary in your `$GOPATH/bin` directory.

## Usage

```bash
Usage: adl [options] <ID or URL>...

Options:
  -dir string
        Custom base download directory (default: ~/Audio/ASMR)
  -concurrency int
        Number of concurrent file downloads (default: 1)
  -exclude string
        Regex pattern to exclude tracks by title
  -prefer-flac
        Skip downloading WAV if a matching FLAC file already exists locally
```

### Examples

**Download a single work:**
```bash
adl RJ01570159
```

**Download multiple works concurrently (speed up downloads with `-concurrency`):**
```bash
adl -concurrency 4 RJ123456 RJ654321
```

**Save files to a specific directory:**
```bash
adl -dir "/Volumes/External/ASMR" RJ123456
```

**Exclude specific files (e.g., skip all `.pdf` documents or specific tracks):**
```bash
adl -exclude "\.pdf$" RJ123456
```

## Features

- **Cross-platform**: Runs natively on Windows, macOS, and Linux without any dependencies.
- **Smart Resuming**: Automatically skips files that have already been downloaded.
- **Metadata Archival**: Automatically saves the API metadata as `info.json` inside the downloaded work's root directory.
- **FLAC Support**: If you use the `-prefer-flac` flag, it won't waste bandwidth downloading `.wav` files if you've already converted them to `.flac` locally.
- **Concurrent Downloading**: Speed up the process by downloading multiple tracks at the same time.
