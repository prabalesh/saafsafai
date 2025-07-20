package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	CleanDownloads    bool `json:"clean_downloads"`
	DeleteNodeModules bool `json:"delete_node_modules"`
}

var (
	homeDir, _     = os.UserHomeDir()
	downloadsDir   = filepath.Join(homeDir, "Downloads")
	configPath     = filepath.Join(homeDir, ".config", "saafsafai.json")
	systemdUnitDir = filepath.Join(homeDir, ".config", "systemd", "user")
)

var summary = struct {
	DeletedFiles   []string
	MovedFiles     []string
	RemovedModules []string
}{}

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "--setup" {
		runSetup()
		return
	}

	config := loadConfig()

	if config.CleanDownloads {
		cleanDownloads()
	}

	if config.DeleteNodeModules {
		cleanOldNodeModules()
	}

	printSummary()
}

func runSetup() {
	reader := bufio.NewReader(os.Stdin)
	var config Config

	fmt.Println("⚙️  Welcome to saafsafai setup!")

	fmt.Print("Do you want to clean the Downloads folder? (y/n): ")
	input, _ := reader.ReadString('\n')
	config.CleanDownloads = strings.TrimSpace(input) == "y"

	fmt.Print("Do you want to delete unused node_modules folders (30+ days)? (y/n): ")
	input, _ = reader.ReadString('\n')
	config.DeleteNodeModules = strings.TrimSpace(input) == "y"

	saveConfig(config)
	installSystemdService()

	fmt.Println("✅ Setup complete. saafsafai will run at each boot.")
}

func loadConfig() Config {
	var config Config
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Println("🟡 Config not found. Run with --setup to configure.")
		os.Exit(0)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	return config
}

func saveConfig(cfg Config) {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	os.MkdirAll(filepath.Dir(configPath), 0755)
	os.WriteFile(configPath, data, 0644)
}

func cleanDownloads() {
	entries, _ := os.ReadDir(downloadsDir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(downloadsDir, entry.Name())
		ext := strings.ToLower(filepath.Ext(entry.Name()))

		if ext == ".tmp" || ext == ".part" || ext == ".crdownload" {
			os.Remove(path)
			summary.DeletedFiles = append(summary.DeletedFiles, entry.Name())
		} else {
			moveToCategory(path, ext)
		}
	}
}

func moveToCategory(filePath, ext string) {
	categories := map[string][]string{
		"Documents":  {".pdf", ".txt", ".docx"},
		"Images":     {".png", ".jpg", ".jpeg"},
		"Videos":     {".mp4", ".mkv"},
		"Audio":      {".mp3", ".wav"},
		"Archives":   {".zip", ".tar", ".gz"},
		"Installers": {".deb", ".AppImage", ".sh"},
	}

	category := "Others"
	for cat, exts := range categories {
		for _, x := range exts {
			if x == ext {
				category = cat
			}
		}
	}

	destDir := filepath.Join(downloadsDir, category)
	os.MkdirAll(destDir, 0755)
	dest := filepath.Join(destDir, filepath.Base(filePath))
	if err := os.Rename(filePath, dest); err == nil {
		summary.MovedFiles = append(summary.MovedFiles, filepath.Base(filePath))
	}
}

func cleanOldNodeModules() {
	cutoff := time.Now().AddDate(0, 0, -30)
	err := filepath.WalkDir(homeDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if d.IsDir() && d.Name() == "node_modules" {
			info, err := os.Stat(path)
			if err != nil {
				return nil
			}
			if info.ModTime().Before(cutoff) {
				if err := os.RemoveAll(path); err == nil {
					summary.RemovedModules = append(summary.RemovedModules, path)
				}
			}
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		log.Printf("Error while scanning for node_modules: %v", err)
	}
}

func installSystemdService() {
	os.MkdirAll(systemdUnitDir, 0755)

	// Install binary to ~/.local/bin/saafsafai
	localBinDir := filepath.Join(homeDir, ".local", "bin")
	os.MkdirAll(localBinDir, 0755)

	execPath, _ := os.Executable()
	targetPath := filepath.Join(localBinDir, "saafsafai")

	if execPath != targetPath {
		input, err := os.ReadFile(execPath)
		if err != nil {
			log.Fatalf("Failed to read current binary: %v", err)
		}
		err = os.WriteFile(targetPath, input, 0755)
		if err != nil {
			log.Fatalf("Failed to install binary: %v", err)
		}
		fmt.Println("✅ Installed binary to:", targetPath)
	}

	// Install systemd service
	serviceFile := filepath.Join(systemdUnitDir, "saafsafai.service")
	content := `[Unit]
Description=Saafsafai Cleanup Service
After=default.target

[Service]
ExecStart=` + targetPath + `
Restart=no

[Install]
WantedBy=default.target
`

	os.WriteFile(serviceFile, []byte(content), 0644)

	exec.Command("systemctl", "--user", "daemon-reexec").Run()
	exec.Command("systemctl", "--user", "daemon-reload").Run()
	exec.Command("systemctl", "--user", "enable", "saafsafai.service").Run()
}

func printSummary() {
	var lines []string

	logHeader := fmt.Sprintf("🧹 Saafsafai Cleanup Report — %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	lines = append(lines, logHeader)

	if len(summary.DeletedFiles) > 0 {
		lines = append(lines, "🗑️ Deleted temp files:")
		for _, f := range summary.DeletedFiles {
			lines = append(lines, " - "+f)
		}
		lines = append(lines, "")
	}

	if len(summary.MovedFiles) > 0 {
		lines = append(lines, "📁 Moved files to category folders:")
		for _, f := range summary.MovedFiles {
			lines = append(lines, " - "+f)
		}
		lines = append(lines, "")
	}

	if len(summary.RemovedModules) > 0 {
		lines = append(lines, "📦 Deleted old node_modules folders:")
		for _, m := range summary.RemovedModules {
			lines = append(lines, " - "+m)
		}
		lines = append(lines, "")
	}

	if len(summary.DeletedFiles)+len(summary.MovedFiles)+len(summary.RemovedModules) == 0 {
		lines = append(lines, "📭 Nothing to clean today.")
	}

	logText := strings.Join(lines, "\n")

	// Log to file
	logDir := filepath.Join(homeDir, ".local", "share", "saafsafai", "logs")
	os.MkdirAll(logDir, 0755)
	logFile := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")
	os.WriteFile(logFile, []byte(logText), 0644)

	// Print to terminal if interactive
	if isInteractive() {
		fmt.Println(logText)
	}
}

func isInteractive() bool {
	return os.Getenv("TERM") != "" && os.Getenv("DISPLAY") != ""
}
