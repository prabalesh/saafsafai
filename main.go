package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	configFileName    = "saafsafai.json"
	serviceName       = "saafsafai.service"
	binaryName        = "saafsafai"
	nodeModulesMaxAge = 30 // days
)

type Config struct {
	CleanDownloads    bool `json:"clean_downloads"`
	DeleteNodeModules bool `json:"delete_node_modules"`
}

type Summary struct {
	DeletedFiles   []string `json:"deleted_files"`
	MovedFiles     []string `json:"moved_files"`
	RemovedModules []string `json:"removed_modules"`
}

type App struct {
	homeDir        string
	downloadsDir   string
	configPath     string
	systemdUnitDir string
	logDir         string
	summary        Summary
}

func NewApp() (*App, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	app := &App{
		homeDir:        homeDir,
		downloadsDir:   filepath.Join(homeDir, "Downloads"),
		configPath:     filepath.Join(homeDir, ".config", configFileName),
		systemdUnitDir: filepath.Join(homeDir, ".config", "systemd", "user"),
		logDir:         filepath.Join(homeDir, ".local", "share", "saafsafai", "logs"),
		summary:        Summary{},
	}

	return app, nil
}

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	args := os.Args[1:]

	switch {
	case len(args) > 0 && args[0] == "--setup":
		if err := app.runSetup(); err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
		return
	case len(args) > 0 && args[0] == "--help":
		app.printHelp()
		return
	case len(args) > 0 && args[0] == "--version":
		fmt.Println("saafsafai v1.0.0")
		return
	}

	if err := app.run(); err != nil {
		log.Fatalf("Cleanup failed: %v", err)
	}
}

func (app *App) run() error {
	config, err := app.loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if config.CleanDownloads {
		if err := app.cleanDownloads(); err != nil {
			log.Printf("Error cleaning downloads: %v", err)
		}
	}

	if config.DeleteNodeModules {
		if err := app.cleanOldNodeModules(); err != nil {
			log.Printf("Error cleaning node_modules: %v", err)
		}
	}

	return app.printSummary()
}

func (app *App) printHelp() {
	fmt.Println(`saafsafai - A system cleanup utility

Usage:
  saafsafai           Run cleanup based on configuration
  saafsafai --setup   Run interactive setup
  saafsafai --help    Show this help message
  saafsafai --version Show version information

Configuration file location: ~/.config/saafsafai.json
Logs location: ~/.local/share/saafsafai/logs/`)
}

func (app *App) runSetup() error {
	reader := bufio.NewReader(os.Stdin)
	var config Config

	fmt.Println("⚙️  Welcome to saafsafai setup!")
	fmt.Println()

	cleanDownloads, err := app.askYesNo(reader, "Do you want to clean the Downloads folder?")
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	config.CleanDownloads = cleanDownloads

	deleteNodeModules, err := app.askYesNo(reader, "Do you want to delete unused node_modules folders (30+ days old)?")
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	config.DeleteNodeModules = deleteNodeModules

	if err := app.saveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if err := app.installSystemdService(); err != nil {
		return fmt.Errorf("failed to install systemd service: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Setup complete! saafsafai will run at each boot.")
	fmt.Println("📁 Config saved to:", app.configPath)
	fmt.Println("🔧 To manually run: saafsafai")
	fmt.Println("📋 To see logs: ls", app.logDir)

	return nil
}

func (app *App) askYesNo(reader *bufio.Reader, question string) (bool, error) {
	fmt.Printf("%s (y/n): ", question)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	response := strings.TrimSpace(strings.ToLower(input))
	return response == "y" || response == "yes", nil
}

func (app *App) loadConfig() (Config, error) {
	var config Config

	if _, err := os.Stat(app.configPath); os.IsNotExist(err) {
		return config, fmt.Errorf("config file not found at %s. Run 'saafsafai --setup' to configure", app.configPath)
	}

	data, err := os.ReadFile(app.configPath)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return config, nil
}

func (app *App) saveConfig(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(app.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(app.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (app *App) cleanDownloads() error {
	if _, err := os.Stat(app.downloadsDir); os.IsNotExist(err) {
		log.Printf("Downloads directory does not exist: %s", app.downloadsDir)
		return nil
	}

	entries, err := os.ReadDir(app.downloadsDir)
	if err != nil {
		return fmt.Errorf("failed to read downloads directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(app.downloadsDir, entry.Name())
		ext := strings.ToLower(filepath.Ext(entry.Name()))

		// Delete temporary files
		if app.isTempFile(ext) {
			if err := os.Remove(filePath); err != nil {
				log.Printf("Failed to delete temp file %s: %v", entry.Name(), err)
				continue
			}
			app.summary.DeletedFiles = append(app.summary.DeletedFiles, entry.Name())
		} else {
			// Move to category folder
			if err := app.moveToCategory(filePath, ext); err != nil {
				log.Printf("Failed to move file %s: %v", entry.Name(), err)
			}
		}
	}

	return nil
}

func (app *App) isTempFile(ext string) bool {
	tempExts := []string{".tmp", ".part", ".crdownload", ".download"}
	for _, tempExt := range tempExts {
		if ext == tempExt {
			return true
		}
	}
	return false
}

func (app *App) moveToCategory(filePath, ext string) error {
	categories := map[string][]string{
		"Documents":  {".pdf", ".txt", ".docx", ".doc", ".rtf", ".odt", ".pages"},
		"Images":     {".png", ".jpg", ".jpeg", ".gif", ".bmp", ".svg", ".webp", ".tiff"},
		"Videos":     {".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm", ".m4v"},
		"Audio":      {".mp3", ".wav", ".flac", ".aac", ".ogg", ".wma", ".m4a"},
		"Archives":   {".zip", ".tar", ".gz", ".rar", ".7z", ".bz2", ".xz", ".tar.gz"},
		"Installers": {".deb", ".rpm", ".dmg", ".exe", ".msi", ".appimage", ".sh", ".pkg"},
		"Code":       {".py", ".js", ".go", ".java", ".cpp", ".c", ".html", ".css", ".json"},
	}

	category := "Others"
	for cat, exts := range categories {
		for _, x := range exts {
			if x == ext {
				category = cat
				break
			}
		}
		if category != "Others" {
			break
		}
	}

	destDir := filepath.Join(app.downloadsDir, category)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create category directory: %w", err)
	}

	fileName := filepath.Base(filePath)
	dest := filepath.Join(destDir, fileName)

	// Handle duplicate filenames
	counter := 1
	for {
		if _, err := os.Stat(dest); os.IsNotExist(err) {
			break
		}

		base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		ext := filepath.Ext(fileName)
		dest = filepath.Join(destDir, fmt.Sprintf("%s_%d%s", base, counter, ext))
		counter++
	}

	if err := os.Rename(filePath, dest); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	app.summary.MovedFiles = append(app.summary.MovedFiles, fileName)
	return nil
}

func (app *App) cleanOldNodeModules() error {
	cutoff := time.Now().AddDate(0, 0, -nodeModulesMaxAge)

	err := filepath.WalkDir(app.homeDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip directories we can't read
			return nil
		}

		if d.IsDir() && d.Name() == "node_modules" {
			info, err := os.Stat(path)
			if err != nil {
				return nil
			}

			if info.ModTime().Before(cutoff) {
				if err := os.RemoveAll(path); err != nil {
					log.Printf("Failed to remove node_modules at %s: %v", path, err)
				} else {
					app.summary.RemovedModules = append(app.summary.RemovedModules, path)
				}
			}

			return filepath.SkipDir // Don't descend into node_modules
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("error scanning for node_modules: %w", err)
	}

	return nil
}

func (app *App) installSystemdService() error {
	if err := os.MkdirAll(app.systemdUnitDir, 0755); err != nil {
		return fmt.Errorf("failed to create systemd directory: %w", err)
	}

	// Install binary to ~/.local/bin/saafsafai
	localBinDir := filepath.Join(app.homeDir, ".local", "bin")
	if err := os.MkdirAll(localBinDir, 0755); err != nil {
		return fmt.Errorf("failed to create local bin directory: %w", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	targetPath := filepath.Join(localBinDir, binaryName)

	if execPath != targetPath {
		if err := app.copyFile(execPath, targetPath); err != nil {
			return fmt.Errorf("failed to install binary: %w", err)
		}
		fmt.Printf("✅ Installed binary to: %s\n", targetPath)
	}

	// Create systemd service file
	serviceFile := filepath.Join(app.systemdUnitDir, serviceName)
	serviceContent := fmt.Sprintf(`[Unit]
Description=Saafsafai Cleanup Service
After=default.target

[Service]
Type=oneshot
ExecStart=%s
Environment=HOME=%s

[Install]
WantedBy=default.target
`, targetPath, app.homeDir)

	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write systemd service file: %w", err)
	}

	// Enable and start the service
	commands := [][]string{
		{"systemctl", "--user", "daemon-reload"},
		{"systemctl", "--user", "enable", serviceName},
	}

	for _, cmd := range commands {
		if err := exec.Command(cmd[0], cmd[1:]...).Run(); err != nil {
			log.Printf("Warning: Failed to run %v: %v", cmd, err)
		}
	}

	fmt.Printf("✅ Systemd service installed: %s\n", serviceFile)
	return nil
}

func (app *App) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	// Make executable
	return os.Chmod(dst, 0755)
}

func (app *App) printSummary() error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	var lines []string

	lines = append(lines, fmt.Sprintf("🧹 Saafsafai Cleanup Report — %s", timestamp))
	lines = append(lines, "")

	if len(app.summary.DeletedFiles) > 0 {
		lines = append(lines, "🗑️ Deleted temp files:")
		for _, f := range app.summary.DeletedFiles {
			lines = append(lines, "   - "+f)
		}
		lines = append(lines, "")
	}

	if len(app.summary.MovedFiles) > 0 {
		lines = append(lines, "📁 Moved files to category folders:")
		for _, f := range app.summary.MovedFiles {
			lines = append(lines, "   - "+f)
		}
		lines = append(lines, "")
	}

	if len(app.summary.RemovedModules) > 0 {
		lines = append(lines, "📦 Deleted old node_modules folders:")
		for _, m := range app.summary.RemovedModules {
			lines = append(lines, "   - "+m)
		}
		lines = append(lines, "")
	}

	totalItems := len(app.summary.DeletedFiles) + len(app.summary.MovedFiles) + len(app.summary.RemovedModules)
	if totalItems == 0 {
		lines = append(lines, "📭 Nothing to clean today.")
	} else {
		lines = append(lines, fmt.Sprintf("✨ Cleaned up %d items total.", totalItems))
	}

	logText := strings.Join(lines, "\n")

	// Ensure log directory exists
	if err := os.MkdirAll(app.logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Write to daily log file
	logFile := filepath.Join(app.logDir, time.Now().Format("2006-01-02")+".log")
	if err := os.WriteFile(logFile, []byte(logText+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write log file: %w", err)
	}

	// Print to stdout if running interactively
	if app.isInteractive() {
		fmt.Println(logText)
	}

	return nil
}

func (app *App) isInteractive() bool {
	return os.Getenv("TERM") != "" && (os.Getenv("DISPLAY") != "" || os.Getenv("SSH_CLIENT") != "")
}
