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
	"testing"
)

func TestValidateConfig_ValidDefaults(t *testing.T) {
	cfg := DefaultConfig()
	errs := ValidateConfig(&cfg)
	if len(errs) != 0 {
		t.Fatalf("expected 0 errors for default config, got: %v", errs)
	}
}

func TestValidateConfig_Duplicates(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Bindings.RightHalf = cfg.Bindings.LeftHalf // duplicate!

	errs := ValidateConfig(&cfg)
	if len(errs) == 0 {
		t.Fatalf("expected duplicate error, got 0 errors")
	}

	found := false
	for _, e := range errs {
		if e == `Duplicate shortcut "Win + Alt + Left" used for both Left Half and Right Half` {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected duplicate error message, got: %v", errs)
	}
}

func TestValidateConfig_InvalidKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Bindings.LeftHalf = "Win+Alt+SuperKeyDoesNotExist"

	errs := ValidateConfig(&cfg)
	if len(errs) == 0 {
		t.Fatalf("expected syntax error, got 0 errors")
	}
}

func TestValidateConfig_OutOfRangeSettings(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Settings.AlmostMaximizeMarginPercent = 90
	cfg.Settings.StepResizePercent = 0

	errs := ValidateConfig(&cfg)
	if len(errs) != 2 {
		t.Fatalf("expected 2 settings errors, got: %v", errs)
	}
}
