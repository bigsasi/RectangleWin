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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Bindings KeybindingsConfig `json:"bindings"`
	Settings SettingsConfig    `json:"settings"`
}

type KeybindingsConfig struct {
	// Halves & cycling
	LeftHalf   string `json:"left_half"`
	RightHalf  string `json:"right_half"`
	TopHalf    string `json:"top_half"`
	BottomHalf string `json:"bottom_half"`

	// Corners & cycling
	TopLeft     string `json:"top_left"`
	TopRight    string `json:"top_right"`
	BottomLeft  string `json:"bottom_left"`
	BottomRight string `json:"bottom_right"`

	// Direct Thirds
	FirstThird     string `json:"first_third"`
	CenterThird    string `json:"center_third"`
	LastThird      string `json:"last_third"`
	FirstTwoThirds string `json:"first_two_thirds"`
	LastTwoThirds  string `json:"last_two_thirds"`

	// Sizing & Placement
	Maximize       string `json:"maximize"`
	AlmostMaximize string `json:"almost_maximize"`
	Center         string `json:"center"`
	Restore        string `json:"restore"`
	Minimize       string `json:"minimize"`
	MakeLarger     string `json:"make_larger"`
	MakeSmaller    string `json:"make_smaller"`

	// Multi-Monitor & State
	NextMonitor string `json:"next_monitor"`
	PrevMonitor string `json:"prev_monitor"`
	AlwaysOnTop string `json:"always_on_top"`
}

type SettingsConfig struct {
	CycleSubdivisions           bool `json:"cycle_subdivisions"`
	AlmostMaximizeMarginPercent int  `json:"almost_maximize_margin_percent"`
	StepResizePercent           int  `json:"step_resize_percent"`
}

func DefaultConfig() Config {
	return Config{
		Bindings: KeybindingsConfig{
			LeftHalf:       "Win+Alt+Left",
			RightHalf:      "Win+Alt+Right",
			TopHalf:        "Win+Alt+Up",
			BottomHalf:     "Win+Alt+Down",
			TopLeft:        "Win+Ctrl+Alt+Left",
			TopRight:       "Win+Ctrl+Alt+Up",
			BottomLeft:     "Win+Ctrl+Alt+Down",
			BottomRight:    "Win+Ctrl+Alt+Right",
			FirstThird:     "",
			CenterThird:    "",
			LastThird:      "",
			FirstTwoThirds: "",
			LastTwoThirds:  "",
			Maximize:       "Win+Shift+F",
			AlmostMaximize: "Win+Alt+F",
			Center:         "Win+Alt+C",
			Restore:        "",
			Minimize:       "",
			MakeLarger:     "",
			MakeSmaller:    "",
			NextMonitor:    "Win+Alt+Ctrl+PageUp",
			PrevMonitor:    "Win+Alt+Ctrl+PageDown",
			AlwaysOnTop:    "Win+Alt+A",
		},
		Settings: SettingsConfig{
			CycleSubdivisions:           true,
			AlmostMaximizeMarginPercent: 6,
			StepResizePercent:           10,
		},
	}
}

// GetConfigPath returns the active configuration file path.
// If a config.json exists in the current directory, it is used (portable mode).
// Otherwise %APPDATA%\RectangleWin\config.json is used.
func GetConfigPath() (string, error) {
	// Check portable
	if _, err := os.Stat("config.json"); err == nil {
		abs, err := filepath.Abs("config.json")
		if err == nil {
			return abs, nil
		}
	}

	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.New("cannot locate APPDATA or UserHomeDir")
		}
		appData = filepath.Join(home, "AppData", "Roaming")
	}

	dir := filepath.Join(appData, "RectangleWin")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	return filepath.Join(dir, "config.json"), nil
}

// LoadConfig reads the configuration file or generates a default if missing.
func LoadConfig() (*Config, string, error) {
	path, err := GetConfigPath()
	if err != nil {
		cfg := DefaultConfig()
		return &cfg, "", err
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		cfg := DefaultConfig()
		if saveErr := SaveConfig(path, &cfg); saveErr != nil {
			fmt.Printf("warn: failed to write default config to %s: %v\n", path, saveErr)
		}
		return &cfg, path, nil
	} else if err != nil {
		cfg := DefaultConfig()
		return &cfg, path, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &cfg, path, fmt.Errorf("failed to parse JSON in %s: %w", path, err)
	}

	return &cfg, path, nil
}

// SaveConfig writes the configuration to disk with formatted indentation.
func SaveConfig(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
