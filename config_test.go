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
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Bindings.LeftHalf != "Win+Alt+Left" {
		t.Errorf("expected default LeftHalf to be Win+Alt+Left, got %s", cfg.Bindings.LeftHalf)
	}
	if !cfg.Settings.CycleSubdivisions {
		t.Errorf("expected default CycleSubdivisions to be true")
	}
}

func TestConfigSerialization(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "config.json")

	cfg := DefaultConfig()
	cfg.Bindings.Center = "Ctrl+Alt+C"
	cfg.Settings.AlmostMaximizeMarginPercent = 10

	if err := SaveConfig(testPath, &cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	data, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("failed to read written config: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to unmarshal saved config: %v", err)
	}

	if loaded.Bindings.Center != "Ctrl+Alt+C" {
		t.Errorf("expected Center to be Ctrl+Alt+C, got %s", loaded.Bindings.Center)
	}
	if loaded.Settings.AlmostMaximizeMarginPercent != 10 {
		t.Errorf("expected AlmostMaximizeMarginPercent to be 10, got %d", loaded.Settings.AlmostMaximizeMarginPercent)
	}
}
