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
	"fmt"
	"strings"
)

// ValidateConfig checks if all hotkey bindings in the config are valid
// and ensures no two different actions share the same hotkey.
func ValidateConfig(cfg *Config) []string {
	var errs []string

	type actionEntry struct {
		name   string
		keyStr string
	}

	b := cfg.Bindings
	entries := []actionEntry{
		{"Left Half", b.LeftHalf},
		{"Right Half", b.RightHalf},
		{"Top Half", b.TopHalf},
		{"Bottom Half", b.BottomHalf},
		{"Top Left", b.TopLeft},
		{"Top Right", b.TopRight},
		{"Bottom Left", b.BottomLeft},
		{"Bottom Right", b.BottomRight},
		{"First Third", b.FirstThird},
		{"Center Third", b.CenterThird},
		{"Last Third", b.LastThird},
		{"First Two Thirds", b.FirstTwoThirds},
		{"Last Two Thirds", b.LastTwoThirds},
		{"Maximize", b.Maximize},
		{"Almost Maximize", b.AlmostMaximize},
		{"Center", b.Center},
		{"Restore", b.Restore},
		{"Minimize", b.Minimize},
		{"Make Larger", b.MakeLarger},
		{"Make Smaller", b.MakeSmaller},
		{"Next Monitor", b.NextMonitor},
		{"Prev Monitor", b.PrevMonitor},
		{"Always on Top", b.AlwaysOnTop},
	}

	seen := make(map[string]string) // normalized key combo -> action name

	for _, entry := range entries {
		trimmed := strings.TrimSpace(entry.keyStr)
		if trimmed == "" || strings.EqualFold(trimmed, "none") || strings.EqualFold(trimmed, "disabled") {
			continue
		}

		mod, vk, err := ParseKeyCombo(trimmed)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: invalid shortcut format %q (%v)", entry.name, entry.keyStr, err))
			continue
		}

		if vk == 0 {
			continue
		}

		canonical := FormatKeyCombo(mod, vk)
		if existingAction, found := seen[canonical]; found {
			errs = append(errs, fmt.Sprintf("Duplicate shortcut %q used for both %s and %s", canonical, existingAction, entry.name))
		} else {
			seen[canonical] = entry.name
		}
	}

	if cfg.Settings.AlmostMaximizeMarginPercent < 1 || cfg.Settings.AlmostMaximizeMarginPercent > 40 {
		errs = append(errs, "Almost Maximize Margin must be between 1% and 40%")
	}

	if cfg.Settings.StepResizePercent < 1 || cfg.Settings.StepResizePercent > 50 {
		errs = append(errs, "Step Resize must be between 1% and 50%")
	}

	return errs
}
