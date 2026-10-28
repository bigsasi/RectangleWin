// Copyright 2026 bigsasi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	_ "embed"
	"fmt"
	"runtime"
	"sync"

	"github.com/gonutz/w32/v2"
	"github.com/jchv/go-webview2"
)

//go:embed ui/settings.html
var settingsHTML string

var (
	settingsView     webview2.WebView
	settingsViewLock sync.Mutex
)

// ShowSettingsWindow opens or focuses the modern Settings UI window.
func ShowSettingsWindow() {
	settingsViewLock.Lock()
	if settingsView != nil {
		hwnd := w32.HWND(uintptr(settingsView.Window()))
		if hwnd != 0 && w32.IsWindow(hwnd) {
			w32.ShowWindow(hwnd, w32.SW_RESTORE)
			w32.SetForegroundWindow(hwnd)
			settingsViewLock.Unlock()
			return
		}
		settingsView = nil
	}
	settingsViewLock.Unlock()

	go runSettingsWindow()
}

func runSettingsWindow() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug: false,
		WindowOptions: webview2.WindowOptions{
			Title:  "RectangleWin Preferences",
			Width:  880,
			Height: 680,
			Center: true,
		},
	})
	if w == nil {
		fmt.Println("warn: failed to initialize WebView2 for settings window")
		return
	}

	settingsViewLock.Lock()
	settingsView = w
	settingsViewLock.Unlock()

	defer func() {
		settingsViewLock.Lock()
		settingsView = nil
		settingsViewLock.Unlock()
	}()

	w.SetTitle("RectangleWin Preferences")
	w.SetSize(880, 680, webview2.HintNone)
	w.SetSize(720, 520, webview2.HintMin)

	// Bind Go methods callable from JavaScript
	w.Bind("loadConfig", func() (*Config, error) {
		activeConfigLock.Lock()
		defer activeConfigLock.Unlock()
		if activeConfig == nil {
			d := DefaultConfig()
			return &d, nil
		}
		return activeConfig, nil
	})

	w.Bind("saveConfig", func(newCfg Config) ([]string, error) {
		errs := ValidateConfig(&newCfg)
		if len(errs) > 0 {
			return errs, nil
		}

		if activeConfigPath == "" {
			p, err := GetConfigPath()
			if err != nil {
				return []string{err.Error()}, nil
			}
			activeConfigPath = p
		}

		if err := SaveConfig(activeConfigPath, &newCfg); err != nil {
			return []string{err.Error()}, nil
		}

		activeConfigLock.Lock()
		activeConfig = &newCfg
		activeConfigLock.Unlock()

		reloadConfig()
		return nil, nil
	})

	w.Bind("resetDefaults", func() (*Config, error) {
		d := DefaultConfig()
		return &d, nil
	})

	w.Bind("closeWindow", func() {
		w.Dispatch(func() {
			w.Destroy()
		})
	})

	// Autorun (Launch at Startup)
	w.Bind("getAutoRun", func() (bool, error) {
		return AutoRunEnabled()
	})

	w.Bind("setAutoRun", func(enable bool) error {
		if enable {
			return AutoRunEnable()
		}
		return AutoRunDisable()
	})

	w.Bind("openExternalUrl", func(url string) error {
		return w32.ShellExecute(0, "open", url, "", "", w32.SW_SHOWNORMAL)
	})

	w.SetHtml(settingsHTML)
	w.Run()
}
