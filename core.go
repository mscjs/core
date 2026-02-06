package core

import (
	"context"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	cu "github.com/Davincible/chromedp-undetected"
	"github.com/chromedp/chromedp"
)

// NewStealthContext creates a context with undetected-chromedp
func NewStealthContext(timeout time.Duration, headless bool, proxy string) (context.Context, context.CancelFunc) {
	chromePath := getChromePath()

	// Actually cu.NewConfig takes ...Option.
	// We'll build a slice of options.
	var opts []cu.Option

	if chromePath != "" {
		opts = append(opts, cu.WithChromeBinary(chromePath))
	} else {
		log.Println("Could not find Google Chrome installation. Chromedp will attempt to find it automatically.")
	}

	opts = append(opts, cu.WithTimeout(timeout))

	// Only add WithHeadless if true (it takes no args and enables it)
	if headless {
		// undetected-chromedp does not support headless on Darwin (macOS).
		// We auto-disable it to prevent crashes, but log a warning.
		if runtime.GOOS == "darwin" {
			log.Println("Warning: Headless mode is not supported on macOS by the undetected driver. Falling back to visible mode.")
		} else {
			opts = append(opts, cu.WithHeadless())
		}
	}

	if proxy != "" {
		opts = append(opts, cu.WithChromeFlags(chromedp.ProxyServer(proxy)))
	}

	ctx, cancel, err := cu.New(cu.NewConfig(opts...))
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
