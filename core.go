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

func NewStealthContext(duration time.Duration) (context.Context, context.CancelFunc) {
	chromePath := getChromePath()
	if chromePath == "" {
		log.Fatal("Could not find Google Chrome installation.")
	}
	// Note: We are using the default config here.
	// If you need more customization you can expose config struct.
	ctx, cancel, err := cu.New(cu.NewConfig(
		cu.WithChromeBinary(chromePath),
	))
	if err != nil {
		log.Fatalf("Failed to create undetected context: %v", err)
	}
	ctx, cancelTimeout := context.WithTimeout(ctx, duration)
	return ctx, func() {
		cancelTimeout()
		cancel()
	}
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
