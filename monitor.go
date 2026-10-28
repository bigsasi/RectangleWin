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
	"syscall"

	"github.com/gonutz/w32/v2"
)

func EnumMonitors(f func(d w32.HMONITOR) bool) bool {
	callback := syscall.NewCallback(func(h, _, _, _ uintptr) uintptr {
		if f(w32.HMONITOR(h)) {
			return 1
		}
		return 0
	})
	return w32.EnumDisplayMonitors(0, nil, callback, 0)
}

func GetAllMonitors() []w32.HMONITOR {
	var list []w32.HMONITOR
	EnumMonitors(func(d w32.HMONITOR) bool {
		list = append(list, d)
		return true
	})
	return list
}

func MoveWindowToMonitor(hwnd w32.HWND, next bool) error {
	if !isZonableWindow(hwnd) {
		return fmt.Errorf("window is not zonable")
	}

	monitors := GetAllMonitors()
	if len(monitors) <= 1 {
		return nil // single monitor, nothing to move
	}

	curMon := w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST)
	curIdx := -1
	for i, m := range monitors {
		if m == curMon {
			curIdx = i
			break
		}
	}

	if curIdx == -1 {
		curIdx = 0
	}

	var targetIdx int
	if next {
		targetIdx = (curIdx + 1) % len(monitors)
	} else {
		targetIdx = (curIdx - 1 + len(monitors)) % len(monitors)
	}

	var curMonInfo w32.MONITORINFO
	if !w32.GetMonitorInfo(curMon, &curMonInfo) {
		return fmt.Errorf("failed to get current monitor info")
	}

	targetMon := monitors[targetIdx]
	var targetMonInfo w32.MONITORINFO
	if !w32.GetMonitorInfo(targetMon, &targetMonInfo) {
		return fmt.Errorf("failed to get target monitor info")
	}

	curWork := curMonInfo.RcWork
	targetWork := targetMonInfo.RcWork
	winRect := w32.GetWindowRect(hwnd)
	if winRect == nil {
		return fmt.Errorf("failed to get window rect")
	}

	// Calculate relative position within current work area
	relX := float64(winRect.Left-curWork.Left) / float64(curWork.Width())
	relY := float64(winRect.Top-curWork.Top) / float64(curWork.Height())
	relW := float64(winRect.Width()) / float64(curWork.Width())
	relH := float64(winRect.Height()) / float64(curWork.Height())

	newW := int32(float64(targetWork.Width()) * relW)
	newH := int32(float64(targetWork.Height()) * relH)
	newLeft := targetWork.Left + int32(float64(targetWork.Width())*relX)
	newTop := targetWork.Top + int32(float64(targetWork.Height())*relY)

	// Clamp within target bounds
	if newW > targetWork.Width() {
		newW = targetWork.Width()
	}
	if newH > targetWork.Height() {
		newH = targetWork.Height()
	}
	if newLeft+newW > targetWork.Right {
		newLeft = targetWork.Right - newW
	}
	if newTop+newH > targetWork.Bottom {
		newTop = targetWork.Bottom - newH
	}
	if newLeft < targetWork.Left {
		newLeft = targetWork.Left
	}
	if newTop < targetWork.Top {
		newTop = targetWork.Top
	}

	if !w32.ShowWindow(hwnd, w32.SW_SHOWNORMAL) {
		// normalization if maximized
	}

	if !w32.SetWindowPos(hwnd, 0, int(newLeft), int(newTop), int(newW), int(newH), w32.SWP_NOZORDER|w32.SWP_NOACTIVATE) {
		return fmt.Errorf("failed to SetWindowPos: %d", w32.GetLastError())
	}

	return nil
}
