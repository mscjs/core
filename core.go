package core

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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

	// Parse proxy if present
	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err == nil && proxyURL.User != nil {
			// Proxy has auth, create extension
			pwd, _ := proxyURL.User.Password()
			extPath, err := createProxyAuthExtension(proxyURL.Hostname(), proxyURL.Port(), proxyURL.User.Username(), pwd)
			if err != nil {
				log.Printf("Failed to create proxy extension: %v", err)
			} else {
				opts = append(opts, cu.WithExtensions(extPath))
				// Use sanitized URL for the proxy flag
				proxy = fmt.Sprintf("%s://%s:%s", proxyURL.Scheme, proxyURL.Hostname(), proxyURL.Port())
			}
		}
		opts = append(opts, cu.WithChromeFlags(chromedp.ProxyServer(proxy)))
	}

	ctx, cancel, err := cu.New(cu.NewConfig(opts...))
	if err != nil {
		log.Fatalf("Failed to create stealth context: %v", err)
	}
	return ctx, cancel
}

func createProxyAuthExtension(host, port, user, pass string) (string, error) {
	dir, err := os.MkdirTemp("", "proxy-auth-ext")
	if err != nil {
		return "", err
	}

	manifest := `{
    "version": "1.0.0",
    "manifest_version": 2,
    "name": "Chrome Proxy",
    "permissions": [
        "proxy",
        "tabs",
        "unlimitedStorage",
        "storage",
        "<all_urls>",
        "webRequest",
        "webRequestBlocking"
    ],
    "background": {
        "scripts": ["background.js"]
    },
    "minimum_chrome_version": "22.0.0"
}`

	background := fmt.Sprintf(`
var config = {
    mode: "fixed_servers",
    rules: {
        singleProxy: {
            scheme: "http",
            host: "%s",
            port: parseInt(%s)
        },
        bypassList: ["foobar.com"]
    }
};

chrome.proxy.settings.set({value: config, scope: "regular"}, function() {});

chrome.webRequest.onAuthRequired.addListener(
    function(details) {
        return {
            authCredentials: {
                username: "%s",
                password: "%s"
            }
        };
    },
    {urls: ["<all_urls>"]},
    ["blocking"]
);
`, host, port, user, pass)

	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0644); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dir, "background.js"), []byte(background), 0644); err != nil {
		return "", err
	}

	return dir, nil
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
