# Saafsafai 🧹

> *"Saafsafai"* is a Hindi/Urdu word meaning "cleanliness" or "cleanup"

A lightweight, automated system cleanup utility written in Go that helps keep your Linux system tidy by organizing downloads and removing old dependencies.

## ✨ Features

- **🗂️ Smart Downloads Organization**: Automatically categorizes and moves files in your Downloads folder into organized subdirectories
- **🗑️ Temporary File Cleanup**: Removes browser temp files, partial downloads, and other temporary files
- **📦 Node.js Cleanup**: Finds and removes old `node_modules` directories (30+ days old) to free up disk space
- **⚙️ Systemd Integration**: Runs automatically at boot or can be executed manually
- **📋 Detailed Logging**: Maintains daily logs of all cleanup activities
- **🎛️ Interactive Setup**: Easy configuration through command-line prompts

## 📁 File Organization

Saafsafai automatically organizes your downloads into these categories:

| Category | File Types |
|----------|------------|
| **Documents** | `.pdf`, `.txt`, `.docx`, `.doc`, `.rtf`, `.odt`, `.pages` |
| **Images** | `.png`, `.jpg`, `.jpeg`, `.gif`, `.bmp`, `.svg`, `.webp`, `.tiff` |
| **Videos** | `.mp4`, `.mkv`, `.avi`, `.mov`, `.wmv`, `.flv`, `.webm`, `.m4v` |
| **Audio** | `.mp3`, `.wav`, `.flac`, `.aac`, `.ogg`, `.wma`, `.m4a` |
| **Archives** | `.zip`, `.tar`, `.gz`, `.rar`, `.7z`, `.bz2`, `.xz`, `.tar.gz` |
| **Installers** | `.deb`, `.rpm`, `.dmg`, `.exe`, `.msi`, `.appimage`, `.sh`, `.pkg` |
| **Code** | `.py`, `.js`, `.go`, `.java`, `.cpp`, `.c`, `.html`, `.css`, `.json` |
| **Others** | All other file types |

## 🚀 Installation

### Prerequisites

- Go 1.19 or later
- Linux system with systemd
- Write access to home directory

### Build from Source

```bash
# Clone the repository
git clone github.com/prabalesh/saafsafai
cd saafsafai

# Build the binary
go build -o saafsafai main.go

# Make it executable
chmod +x saafsafai
```

## ⚙️ Setup

Run the interactive setup to configure saafsafai:

```bash
./saafsafai --setup
```

The setup will:
1. Ask for your cleanup preferences
2. Save configuration to `~/.config/saafsafai.json`
3. Install the binary to `~/.local/bin/saafsafai`
4. Create and enable a systemd service for automatic execution

## 🎮 Usage

### Commands

```bash
# Run cleanup with current configuration
saafsafai

# Interactive setup/reconfiguration
saafsafai --setup

# Show help
saafsafai --help

# Show version
saafsafai --version
```

### Manual Systemd Control

```bash
# Check service status
systemctl --user status saafsafai.service

# Run service manually
systemctl --user start saafsafai.service

# Disable automatic execution
systemctl --user disable saafsafai.service

# Re-enable automatic execution
systemctl --user enable saafsafai.service
```

## 📁 File Locations

```
~/.config/saafsafai.json              # Configuration file
~/.local/bin/saafsafai                # Installed binary
~/.config/systemd/user/saafsafai.service  # Systemd service file
~/.local/share/saafsafai/logs/        # Daily log files
```

## ⚙️ Configuration

The configuration file (`~/.config/saafsafai.json`) contains:

```json
{
  "clean_downloads": true,
  "delete_node_modules": true
}
```

### Configuration Options

- `clean_downloads`: Enable Downloads folder organization and temp file cleanup
- `delete_node_modules`: Enable removal of old node_modules directories (30+ days)

## 📊 Example Output

```
🧹 Saafsafai Cleanup Report — 2024-01-15 09:30:45

🗑️ Deleted temp files:
   - download.part
   - temp_file.tmp
   - incomplete.crdownload

📁 Moved files to category folders:
   - report.pdf
   - photo.jpg
   - presentation.pptx

📦 Deleted old node_modules folders:
   - /home/user/old-project/node_modules
   - /home/user/archived-app/node_modules

✨ Cleaned up 7 items total.
```

## 🛡️ Safety Features

- **Duplicate Handling**: Automatically renames files if destinations already exist
- **Error Recovery**: Continues operation even if individual file operations fail
- **Comprehensive Logging**: All actions are logged with timestamps
- **Conservative Age Limits**: Only removes node_modules older than 30 days
- **Non-Destructive**: Moves files rather than deleting them (except temp files)

## 🔧 Development

### Project Structure

```
.
├── main.go          # Main application code
├── README.md        # This file
└── go.mod           # Go module file
```

### Key Improvements Made

1. **Better Error Handling**: Comprehensive error handling with descriptive messages
2. **Structured Code**: Organized into methods with clear separation of concerns
3. **Enhanced File Categories**: More comprehensive file type categorization
4. **Duplicate File Handling**: Automatically handles filename conflicts
5. **Improved Logging**: Better formatted logs with timestamps and summaries
6. **Interactive Detection**: Better detection of interactive vs automated execution
7. **Help System**: Added `--help` and `--version` commands
8. **Constants**: Used constants for magic numbers and configuration values

### Building

```bash
# Build for current platform
go build -o saafsafai

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o saafsafai-linux-amd64
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🐛 Troubleshooting

### Common Issues

**Service not running automatically:**
```bash
# Check if service is enabled
systemctl --user is-enabled saafsafai.service

# Check service logs
journalctl --user -u saafsafai.service
```

**Permission errors:**
- Ensure the binary has execute permissions: `chmod +x ~/.local/bin/saafsafai`
- Check that `~/.local/bin` is in your PATH

**Config file issues:**
- Delete the config file and run `--setup` again: `rm ~/.config/saafsafai.json`

## 🙏 Acknowledgments

- Inspired by the need for automated system maintenance
- Built with Go's excellent standard library
- Uses systemd for reliable service management

---

*Keep your system clean with सफ़साफ़ाई! 🧹✨*
