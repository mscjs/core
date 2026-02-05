package core

import (
	"context"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	cu "github.com/Davincible/chromedp-undetected"
)

// NewStealthContext creates a context with undetected-chromedp
func NewStealthContext(timeout time.Duration, headless bool) (context.Context, context.CancelFunc) {
	chromePath := getChromePath()
	var options []cu.ConfigOption
	if chromePath != "" {
		options = append(options, cu.WithChromeBinary(chromePath))
	} else {
		log.Println("Could not find Google Chrome installation. Chromedp will attempt to find it automatically.")
	}

	options = append(options, cu.WithTimeout(timeout))
	options = append(options, cu.WithHeadless(headless))

	ctx, cancel, err := cu.New(cu.NewConfig(options...))
	if err != nil {
		log.Fatalf("Failed to create stealth context: %v", err)
	}
	return ctx, cancel
}

func getChromePath() string {
	if path := os.Getenv("CHROME_PATH"); path != "" {
		return path
	}
	if path := os.Getenv("GOOGLE_CHROME_BIN"); path != "" {
		return path
	}
	var paths []string
	switch runtime.GOOS {
	case "darwin":
		paths = []string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/usr/bin/google-chrome",
		}
	case "windows":
		paths = []string{
			os.Getenv("ProgramFiles") + "\\Google\\Chrome\\Application\\chrome.exe",
			os.Getenv("ProgramFiles(x86)") + "\\Google\\Chrome\\Application\\chrome.exe",
		}
	case "linux":
		paths = []string{"/usr/bin/google-chrome", "/usr/bin/chromium"}
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	if path, err := exec.LookPath("google-chrome"); err == nil {
		return path
	}
	return ""
}
