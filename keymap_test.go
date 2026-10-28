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

	"github.com/gonutz/w32/v2"
)

func TestParseKeyCombo(t *testing.T) {
	tests := []struct {
		input       string
		expectedMod int
		expectedVK  int
		expectErr   bool
	}{
		{"Win+Alt+Left", MOD_WIN | MOD_ALT | MOD_NOREPEAT, w32.VK_LEFT, false},
		{"ctrl + alt + right", MOD_CONTROL | MOD_ALT | MOD_NOREPEAT, w32.VK_RIGHT, false},
		{"Win+Shift+F", MOD_WIN | MOD_SHIFT | MOD_NOREPEAT, int('F'), false},
		{"Win+Alt+Ctrl+Down", MOD_WIN | MOD_ALT | MOD_CONTROL | MOD_NOREPEAT, w32.VK_DOWN, false},
		{"Win+Alt+Enter", MOD_WIN | MOD_ALT | MOD_NOREPEAT, w32.VK_RETURN, false},
		{"Win+Alt+Space", MOD_WIN | MOD_ALT | MOD_NOREPEAT, w32.VK_SPACE, false},
		{"Win+Alt+F12", MOD_WIN | MOD_ALT | MOD_NOREPEAT, w32.VK_F12, false},
		{"Win+Alt+Numpad5", MOD_WIN | MOD_ALT | MOD_NOREPEAT, w32.VK_NUMPAD5, false},
		{"", 0, 0, false},
		{"none", 0, 0, false},
		{"disabled", 0, 0, false},
		{"Win+Alt", 0, 0, true},             // missing key
		{"Win+Alt+Left+Right", 0, 0, true},   // multiple keys
		{"Win+Alt+InvalidKey", 0, 0, true},   // invalid key
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			mod, vk, err := ParseKeyCombo(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tt.input, err)
			}
			if mod != tt.expectedMod {
				t.Errorf("input %q: expected mod 0x%X, got 0x%X", tt.input, tt.expectedMod, mod)
			}
			if vk != tt.expectedVK {
				t.Errorf("input %q: expected vk %d, got %d", tt.input, tt.expectedVK, vk)
			}
		})
	}
}

func TestFormatKeyCombo(t *testing.T) {
	formatted := FormatKeyCombo(MOD_WIN|MOD_ALT, w32.VK_LEFT)
	if formatted != "Win + Alt + Left" {
		t.Errorf("expected 'Win + Alt + Left', got %q", formatted)
	}

	disabled := FormatKeyCombo(0, 0)
	if disabled != "Disabled" {
		t.Errorf("expected 'Disabled', got %q", disabled)
	}
}
