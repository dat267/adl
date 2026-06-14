# adl

A lightweight, dependency-free CLI tool for fetching and managing remote audio assets.

It automatically organizes downloads, retrieves associated metadata (`info.json`), and intelligently skips previously synced files to save bandwidth.

## Installation

### 1. Pre-compiled Binaries (Recommended)

You can download ready-to-run executables for your operating system directly from our **[Releases Page](https://github.com/dat267/adl/releases)**.

**For Windows:**
1. Download the `adl-windows-amd64.exe` file.
2. Open your Command Prompt or PowerShell in the folder where you downloaded it.
3. Run: `.\adl-windows-amd64.exe [ID]`

**For macOS/Linux:**
1. Download the appropriate binary (e.g., `adl-darwin-arm64`).
2. Open your terminal.
3. Make it executable: `chmod +x adl-darwin-arm64`
4. Run it: `./adl-darwin-arm64 [ID]`

*(Tip: Rename the downloaded file to `adl` and move it to a directory in your PATH, such as `/usr/local/bin`, to run it from anywhere by just typing `adl`.)*

### 2. Install using Go

If you already have Go installed, you can build and install the latest version directly from the source code.

```bash
go install github.com/dat267/adl@latest
```

## Usage

```bash
Usage: adl [options] <ID or URL>...

Options:
  -dir string
        Custom base download directory
  -concurrency int
        Number of concurrent file downloads (default: 1)
  -exclude string
        Regex pattern to exclude files by title
  -exclude-ext string
        Comma-separated list of file extensions to exclude (e.g., 'mp4,pdf,txt')
  -prefer-flac
        Skip downloading WAV if a matching FLAC file already exists locally
```

### Examples

**Fetch a single target:**
```bash
adl 01570159
```

**Fetch multiple targets concurrently (speed up downloads with `-concurrency`):**
```bash
adl -concurrency 4 123456 654321
```

**Save files to a specific directory:**
```bash
adl -dir "/Volumes/External/Data" 123456
```

**Exclude specific file extensions (e.g., skip videos and PDFs):**
```bash
adl -exclude-ext "mp4,pdf" 123456
```

**Exclude specific files by title (using regex):**
```bash
adl -exclude "bonus" 123456
```

## Features

- **Cross-platform**: Runs natively on Windows, macOS, and Linux without any dependencies.
- **Smart Resuming**: Automatically skips files that have already been downloaded.
- **Metadata Archival**: Automatically saves the API metadata as `info.json` inside the downloaded root directory.
- **FLAC Support**: If you use the `-prefer-flac` flag, it won't waste bandwidth downloading `.wav` files if you've already converted them to `.flac` locally.
- **Concurrent Downloading**: Speed up the process by downloading multiple files at the same time.
