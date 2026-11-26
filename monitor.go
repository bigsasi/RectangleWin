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
	"math"
	"sort"
	"sync"
	"syscall"

	"github.com/bigsasi/RectangleWin/w32ex"
	"github.com/gonutz/w32/v2"
)

var (
	enumMonitorsMu     sync.Mutex
	enumMonitorsOnce   sync.Once
	enumMonitorsWinCb  uintptr
	enumMonitorsActive func(w32.HMONITOR) bool
)

func initEnumMonitorsCallback() {
	enumMonitorsOnce.Do(func() {
		enumMonitorsWinCb = syscall.NewCallback(func(h, _, _, _ uintptr) uintptr {
			if enumMonitorsActive != nil && enumMonitorsActive(w32.HMONITOR(h)) {
				return 1
			}
			return 0
		})
	})
}

func EnumMonitors(f func(d w32.HMONITOR) bool) bool {
	initEnumMonitorsCallback()
	enumMonitorsMu.Lock()
	enumMonitorsActive = f
	defer func() {
		enumMonitorsActive = nil
		enumMonitorsMu.Unlock()
	}()
	return w32.EnumDisplayMonitors(0, nil, enumMonitorsWinCb, 0)
}

type monitorEntry struct {
	handle w32.HMONITOR
	rect   w32.RECT
}

func sortMonitorEntries(entries []monitorEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].rect.Left != entries[j].rect.Left {
			return entries[i].rect.Left < entries[j].rect.Left
		}
		return entries[i].rect.Top < entries[j].rect.Top
	})
}

func GetAllMonitors() []w32.HMONITOR {
	var entries []monitorEntry
	EnumMonitors(func(d w32.HMONITOR) bool {
		var info w32.MONITORINFO
		if w32.GetMonitorInfo(d, &info) {
			entries = append(entries, monitorEntry{handle: d, rect: info.RcMonitor})
		} else {
			entries = append(entries, monitorEntry{handle: d})
		}
		return true
	})
	sortMonitorEntries(entries)

	list := make([]w32.HMONITOR, len(entries))
	for i, e := range entries {
		list[i] = e.handle
	}
	return list
}

// mapRectToMonitor proportionally maps a window's visible rect from curWork into targetWork
// and clamps the result within targetWork.
func mapRectToMonitor(winFrame, curWork, targetWork w32.RECT) w32.RECT {
	if curWork.Width() <= 0 || curWork.Height() <= 0 {
		return targetWork
	}

	relX := float64(winFrame.Left-curWork.Left) / float64(curWork.Width())
	relY := float64(winFrame.Top-curWork.Top) / float64(curWork.Height())
	relW := float64(winFrame.Width()) / float64(curWork.Width())
	relH := float64(winFrame.Height()) / float64(curWork.Height())

	newW := int32(math.Round(float64(targetWork.Width()) * relW))
	newH := int32(math.Round(float64(targetWork.Height()) * relH))
	newLeft := targetWork.Left + int32(math.Round(float64(targetWork.Width())*relX))
	newTop := targetWork.Top + int32(math.Round(float64(targetWork.Height())*relY))

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

	return w32.RECT{
		Left:   newLeft,
		Top:    newTop,
		Right:  newLeft + newW,
		Bottom: newTop + newH,
	}
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
	wasMaximized := w32ex.IsZoomed(hwnd)

	lastResized = 0

	if wasMaximized {
		w32.ShowWindow(hwnd, w32.SW_SHOWNORMAL)
		if err := setWindowVisibleRect(hwnd, targetWork); err != nil {
			return err
		}
		w32.ShowWindow(hwnd, w32.SW_MAXIMIZE)
		return nil
	}

	winRect := w32.GetWindowRect(hwnd)
	if winRect == nil {
		return fmt.Errorf("failed to get window rect")
	}
	winFrame := getWindowFrame(hwnd, *winRect)

	targetRect := mapRectToMonitor(winFrame, curWork, targetWork)
	return setWindowVisibleRect(hwnd, targetRect)
}
