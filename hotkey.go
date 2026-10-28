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

	"github.com/bigsasi/RectangleWin/w32ex"
	"github.com/gonutz/w32/v2"
)

var (
	hotkeyRegistrations = make(map[int]*HotKey)
	showSettingsMsgID   uint32
)

type HotKey struct {
	id, mod, vk int
	callback    func()
}

func (h HotKey) String() string { return fmt.Sprintf("mod=0x%x,vk=%d", h.mod, h.vk) }

func (h HotKey) Describe() string {
	return FormatKeyCombo(h.mod, h.vk)
}

func RegisterHotKey(h HotKey) bool {
	if _, ok := hotkeyRegistrations[h.id]; ok {
		w32ex.UnregisterHotKey(0, h.id)
		delete(hotkeyRegistrations, h.id)
	}
	ok := w32ex.RegisterHotKey(0, h.id, h.mod, h.vk)
	if ok {
		hotkeyRegistrations[h.id] = &h
	}
	return ok
}

func UnregisterHotKey(id int) bool {
	if _, ok := hotkeyRegistrations[id]; !ok {
		return false
	}
	ok := w32ex.UnregisterHotKey(0, id)
	delete(hotkeyRegistrations, id)
	return ok
}

func UnregisterAllHotKeys() {
	for id := range hotkeyRegistrations {
		w32ex.UnregisterHotKey(0, id)
	}
	hotkeyRegistrations = make(map[int]*HotKey)
}

func msgLoop() error {
	defer fmt.Println("event loop finished")
	for {
		var m w32.MSG
		c := w32.GetMessage(&m, 0, 0, 0)
		if c == -1 {
			return fmt.Errorf("GetMessage failed: %d", c)
		} else if c == 0 {
			// WM_QUIT received
			return nil
		}
		if m.Message == w32.WM_HOTKEY {
			h, ok := hotkeyRegistrations[int(m.WParam)]
			if !ok {
				return fmt.Errorf("hotkey without callback: %#v", m)
			}
			fmt.Printf("trace: hotkey id=%d (%s)\n", m.WParam, h)
			h.callback()
		} else if m.Message == WM_RELOAD_HOTKEYS {
			fmt.Println("trace: received WM_RELOAD_HOTKEYS on main thread")
			reloadConfigInternal()
		} else if showSettingsMsgID != 0 && m.Message == showSettingsMsgID {
			fmt.Println("trace: received show settings signal from secondary instance")
			ShowSettingsWindow()
		} else {
			fmt.Printf("unhandled message received:0x%x %d\n", m.Message, m.Message)
			w32.TranslateMessage(&m)
			w32.DispatchMessage(&m)
		}
	}
}
