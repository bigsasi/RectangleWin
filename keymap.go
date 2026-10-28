// Copyright 2022 Ahmet Alp Balkan
// Portions Copyright 2026 bigsasi
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

	"github.com/gonutz/w32/v2"
)

// Win32 HotKey Modifiers
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-registerhotkey
const (
	MOD_ALT      = 0x0001
	MOD_CONTROL  = 0x0002
	MOD_SHIFT    = 0x0004
	MOD_WIN      = 0x0008
	MOD_NOREPEAT = 0x4000
)

var modKeyNames = map[int]string{
	MOD_WIN:     "Win",
	MOD_CONTROL: "Ctrl",
	MOD_ALT:     "Alt",
	MOD_SHIFT:   "Shift",
}

// namedVKs maps normalized lowercase string aliases to virtual key codes
var namedVKs = map[string]int{
	// Navigation & Editing
	"left":        w32.VK_LEFT,
	"right":       w32.VK_RIGHT,
	"up":          w32.VK_UP,
	"down":        w32.VK_DOWN,
	"enter":       w32.VK_RETURN,
	"return":      w32.VK_RETURN,
	"space":       w32.VK_SPACE,
	"spacebar":    w32.VK_SPACE,
	"tab":         w32.VK_TAB,
	"backspace":   w32.VK_BACK,
	"esc":         w32.VK_ESCAPE,
	"escape":      w32.VK_ESCAPE,
	"delete":      w32.VK_DELETE,
	"del":         w32.VK_DELETE,
	"insert":      w32.VK_INSERT,
	"ins":         w32.VK_INSERT,
	"home":        w32.VK_HOME,
	"end":         w32.VK_END,
	"pageup":      w32.VK_PRIOR,
	"pgup":        w32.VK_PRIOR,
	"pagedown":    w32.VK_NEXT,
	"pgdn":        w32.VK_NEXT,
	"pgdown":      w32.VK_NEXT,
	"capslock":    w32.VK_CAPITAL,
	"printscreen": w32.VK_SNAPSHOT,
	"prtsc":       w32.VK_SNAPSHOT,
	"pause":       w32.VK_PAUSE,

	// Common symbols
	"plus":         0xBB, // VK_OEM_PLUS
	"+":            0xBB,
	"minus":        0xBD, // VK_OEM_MINUS
	"-":            0xBD,
	"equal":        0xBB,
	"=":            0xBB,
	"comma":        0xBC, // VK_OEM_COMMA
	",":            0xBC,
	"period":       0xBE, // VK_OEM_PERIOD
	".":            0xBE,
	"slash":        0xBF, // VK_OEM_2 (/?)
	"/":            0xBF,
	"backslash":    0xDC, // VK_OEM_5 (\|)
	"\\":           0xDC,
	"semicolon":    0xBA, // VK_OEM_1 (;:)
	";":            0xBA,
	"quote":        0xDE, // VK_OEM_7 ('")
	"'":            0xDE,
	"bracketleft":  0xDB, // VK_OEM_4 ([{)
	"[":            0xDB,
	"bracketright": 0xDD, // VK_OEM_6 (]})
	"]":            0xDD,
	"backquote":    0xC0, // VK_OEM_3 (`~)
	"`":            0xC0,

	// Numpad
	"numpad0":     w32.VK_NUMPAD0,
	"numpad1":     w32.VK_NUMPAD1,
	"numpad2":     w32.VK_NUMPAD2,
	"numpad3":     w32.VK_NUMPAD3,
	"numpad4":     w32.VK_NUMPAD4,
	"numpad5":     w32.VK_NUMPAD5,
	"numpad6":     w32.VK_NUMPAD6,
	"numpad7":     w32.VK_NUMPAD7,
	"numpad8":     w32.VK_NUMPAD8,
	"numpad9":     w32.VK_NUMPAD9,
	"num0":        w32.VK_NUMPAD0,
	"num1":        w32.VK_NUMPAD1,
	"num2":        w32.VK_NUMPAD2,
	"num3":        w32.VK_NUMPAD3,
	"num4":        w32.VK_NUMPAD4,
	"num5":        w32.VK_NUMPAD5,
	"num6":        w32.VK_NUMPAD6,
	"num7":        w32.VK_NUMPAD7,
	"num8":        w32.VK_NUMPAD8,
	"num9":        w32.VK_NUMPAD9,
	"numadd":      w32.VK_ADD,
	"numsubtract": w32.VK_SUBTRACT,
	"nummultiply": w32.VK_MULTIPLY,
	"numdivide":   w32.VK_DIVIDE,
	"numdecimal":  w32.VK_DECIMAL,
}

func init() {
	// Add A-Z
	for ch := 'a'; ch <= 'z'; ch++ {
		namedVKs[string(ch)] = int('A' + (ch - 'a'))
	}
	// Add 0-9
	for ch := '0'; ch <= '9'; ch++ {
		namedVKs[string(ch)] = int(ch)
	}
	// Add F1 - F24
	for i := 1; i <= 24; i++ {
		namedVKs[fmt.Sprintf("f%d", i)] = w32.VK_F1 + (i - 1)
	}
}

// ParseKeyCombo parses a string like "Win+Alt+Left" or "Ctrl+Shift+F" into mod and vk.
// Returns mod=0, vk=0, err=nil if combo is empty or disabled ("none", "disabled", "").
func ParseKeyCombo(comboStr string) (int, int, error) {
	trimmed := strings.TrimSpace(comboStr)
	if trimmed == "" || strings.EqualFold(trimmed, "none") || strings.EqualFold(trimmed, "disabled") {
		return 0, 0, nil
	}

	parts := strings.Split(trimmed, "+")
	var mod int
	var vk int
	var keyFound bool

	for _, part := range parts {
		token := strings.ToLower(strings.TrimSpace(part))
		if token == "" {
			continue
		}

		switch token {
		case "win", "windows", "super", "cmd":
			mod |= MOD_WIN
		case "ctrl", "control":
			mod |= MOD_CONTROL
		case "alt", "option", "opt":
			mod |= MOD_ALT
		case "shift":
			mod |= MOD_SHIFT
		default:
			if keyFound {
				return 0, 0, fmt.Errorf("multiple non-modifier keys specified in '%s'", comboStr)
			}
			code, ok := namedVKs[token]
			if !ok {
				return 0, 0, fmt.Errorf("unrecognized key '%s' in '%s'", part, comboStr)
			}
			vk = code
			keyFound = true
		}
	}

	if !keyFound {
		return 0, 0, fmt.Errorf("no primary key specified in combo '%s'", comboStr)
	}

	// Always add MOD_NOREPEAT to avoid repeat key spamming
	mod |= MOD_NOREPEAT

	return mod, vk, nil
}

// FormatKeyCombo formats mod and vk into a clean canonical string like "Win + Alt + Left"
func FormatKeyCombo(mod, vk int) string {
	if vk == 0 {
		return "Disabled"
	}
	var out []string
	if mod&MOD_WIN != 0 {
		out = append(out, "Win")
	}
	if mod&MOD_CONTROL != 0 {
		out = append(out, "Ctrl")
	}
	if mod&MOD_ALT != 0 {
		out = append(out, "Alt")
	}
	if mod&MOD_SHIFT != 0 {
		out = append(out, "Shift")
	}

	keyName := getKeyDisplayName(vk)
	out = append(out, keyName)
	return strings.Join(out, " + ")
}

func getKeyDisplayName(vk int) string {
	if vk >= 'A' && vk <= 'Z' {
		return string(rune(vk))
	}
	if vk >= '0' && vk <= '9' {
		return string(rune(vk))
	}
	if vk >= w32.VK_F1 && vk <= w32.VK_F24 {
		return fmt.Sprintf("F%d", vk-w32.VK_F1+1)
	}
	if vk >= w32.VK_NUMPAD0 && vk <= w32.VK_NUMPAD9 {
		return fmt.Sprintf("Numpad%d", vk-w32.VK_NUMPAD0)
	}

	switch vk {
	case w32.VK_LEFT:
		return "Left"
	case w32.VK_RIGHT:
		return "Right"
	case w32.VK_UP:
		return "Up"
	case w32.VK_DOWN:
		return "Down"
	case w32.VK_RETURN:
		return "Enter"
	case w32.VK_SPACE:
		return "Space"
	case w32.VK_TAB:
		return "Tab"
	case w32.VK_BACK:
		return "Backspace"
	case w32.VK_ESCAPE:
		return "Esc"
	case w32.VK_DELETE:
		return "Delete"
	case w32.VK_INSERT:
		return "Insert"
	case w32.VK_HOME:
		return "Home"
	case w32.VK_END:
		return "End"
	case w32.VK_PRIOR:
		return "PageUp"
	case w32.VK_NEXT:
		return "PageDown"
	case 0xBB:
		return "+"
	case 0xBD:
		return "-"
	case 0xBC:
		return ","
	case 0xBE:
		return "."
	case 0xBF:
		return "/"
	case 0xDC:
		return "\\"
	case 0xBA:
		return ";"
	case 0xDE:
		return "'"
	case 0xDB:
		return "["
	case 0xDD:
		return "]"
	case 0xC0:
		return "`"
	default:
		return fmt.Sprintf("Key(0x%X)", vk)
	}
}
