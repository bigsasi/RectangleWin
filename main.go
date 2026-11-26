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
	"errors"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/getlantern/systray"
	"github.com/gonutz/w32/v2"

	"github.com/bigsasi/RectangleWin/w32ex"
)

const WM_RELOAD_HOTKEYS = 0x0400 + 100 // WM_USER + 100

var (
	mainThreadID            uint32
	lastResized             w32.HWND
	activeConfig            *Config
	activeConfigLock        sync.Mutex
	activeConfigPath        string
	singleInstanceMutex     syscall.Handle
	singleInstanceMutexLock sync.Mutex
)

const singleInstanceMutexName = "Local\\RectangleWin_SingleInstance_Mutex"

func releaseSingleInstanceMutex() {
	singleInstanceMutexLock.Lock()
	defer singleInstanceMutexLock.Unlock()
	if singleInstanceMutex != 0 {
		_ = syscall.CloseHandle(singleInstanceMutex)
		singleInstanceMutex = 0
	}
}

func main() {
	runtime.LockOSThread() // Win32 hotkeys and messages require locked OS thread

	// Enable Per-Monitor V2 DPI awareness before any user32 calls so all
	// window/monitor coordinates across screens share the same physical pixel space.
	if !w32ex.EnableModernDPIAwareness() {
		fmt.Println("warn: failed to set modern Per-Monitor V2 DPI awareness, fallback active")
	}

	showSettingsMsgID = w32ex.RegisterWindowMessage("RectangleWin_ShowSettings_Msg")

	// Single instance enforcement
	mutex, isFirst, err := w32ex.CreateMutex(singleInstanceMutexName)
	if err != nil {
		fmt.Printf("warn: single instance check: %v\n", err)
	}
	if !isFirst {
		fmt.Println("Another instance of RectangleWin is already running. Signaling settings...")
		w32.PostMessage(w32.HWND_BROADCAST, showSettingsMsgID, 0, 0)
		os.Exit(0)
	}
	singleInstanceMutexLock.Lock()
	singleInstanceMutex = mutex
	singleInstanceMutexLock.Unlock()
	defer releaseSingleInstanceMutex()

	CleanupOldBinary()

	mainThreadID = w32ex.GetCurrentThreadId()

	cfg, path, err := LoadConfig()
	if err != nil {
		fmt.Printf("warn: error loading config: %v\n", err)
	}
	activeConfig = cfg
	activeConfigPath = path
	fmt.Printf("loaded configuration from %s\n", activeConfigPath)

	autorun, err := AutoRunEnabled()
	if err != nil {
		fmt.Printf("warn: checking autorun: %v\n", err)
	}
	fmt.Printf("autorun enabled=%v\n", autorun)

	setupHotkeys(activeConfig)

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt)
	go func() {
		<-exitCh
		fmt.Println("exit signal received")
		systray.Quit()
	}()

	initTray()
	CheckForUpdatesOnStartup()
	if err := msgLoop(); err != nil {
		panic(err)
	}
}

func reloadConfig() {
	if w32ex.GetCurrentThreadId() == mainThreadID {
		reloadConfigInternal()
	} else {
		w32ex.PostThreadMessage(mainThreadID, WM_RELOAD_HOTKEYS, 0, 0)
	}
}

func reloadConfigInternal() {
	activeConfigLock.Lock()
	defer activeConfigLock.Unlock()

	cfg, path, err := LoadConfig()
	if err != nil {
		showMessageBox("Failed to load config: " + err.Error())
		return
	}
	activeConfig = cfg
	activeConfigPath = path

	UnregisterAllHotKeys()
	setupHotkeys(activeConfig)
	fmt.Println("hotkeys successfully reloaded on main thread")
}

func setupHotkeys(cfg *Config) {
	edgeFuncs := [][]resizeFunc{
		{leftHalf, leftTwoThirds, leftOneThirds},
		{rightHalf, rightTwoThirds, rightOneThirds},
		{topHalf, topTwoThirds, topOneThirds},
		{bottomHalf, bottomTwoThirds, bottomOneThirds},
	}
	edgeFuncTurn := make([]int, len(edgeFuncs))

	cornerFuncs := [][]resizeFunc{
		{topLeftHalf, topLeftTwoThirds, topLeftOneThirds},
		{topRightHalf, topRightTwoThirds, topRightOneThirds},
		{bottomLeftHalf, bottomLeftTwoThirds, bottomLeftOneThirds},
		{bottomRightHalf, bottomRightTwoThirds, bottomRightOneThirds},
	}
	cornerFuncTurn := make([]int, len(cornerFuncs))

	cycleFuncs := func(funcs [][]resizeFunc, turns *[]int, i int) {
		hwnd := w32.GetForegroundWindow()
		if hwnd == 0 {
			return
		}
		if lastResized != hwnd {
			*turns = make([]int, len(funcs))
		}
		if _, err := resize(hwnd, funcs[i][(*turns)[i]%len(funcs[i])]); err != nil {
			fmt.Printf("warn: resize: %v\n", err)
			return
		}
		(*turns)[i]++
		for j := 0; j < len(*turns); j++ {
			if j != i {
				(*turns)[j] = 0
			}
		}
	}

	cycleEdge := func(i int) {
		if !cfg.Settings.CycleSubdivisions {
			hwnd := w32.GetForegroundWindow()
			if hwnd != 0 {
				_, _ = resize(hwnd, edgeFuncs[i][0])
			}
			return
		}
		cycleFuncs(edgeFuncs, &edgeFuncTurn, i)
	}

	cycleCorner := func(i int) {
		if !cfg.Settings.CycleSubdivisions {
			hwnd := w32.GetForegroundWindow()
			if hwnd != 0 {
				_, _ = resize(hwnd, cornerFuncs[i][0])
			}
			return
		}
		cycleFuncs(cornerFuncs, &cornerFuncTurn, i)
	}

	directSnap := func(f resizeFunc) {
		hwnd := w32.GetForegroundWindow()
		if hwnd != 0 {
			lastResized = 0
			_, _ = resize(hwnd, f)
		}
	}

	type actionDef struct {
		id       int
		keyStr   string
		callback func()
	}

	b := cfg.Bindings
	actions := []actionDef{
		// Halves & Cycling
		{1, b.LeftHalf, func() { cycleEdge(0) }},
		{2, b.RightHalf, func() { cycleEdge(1) }},
		{3, b.TopHalf, func() { cycleEdge(2) }},
		{4, b.BottomHalf, func() { cycleEdge(3) }},

		// Corners & Cycling
		{5, b.TopLeft, func() { cycleCorner(0) }},
		{6, b.TopRight, func() { cycleCorner(1) }},
		{7, b.BottomLeft, func() { cycleCorner(2) }},
		{8, b.BottomRight, func() { cycleCorner(3) }},

		// Direct Thirds
		{9, b.FirstThird, func() { directSnap(leftOneThirds) }},
		{10, b.CenterThird, func() { directSnap(centerThird) }},
		{11, b.LastThird, func() { directSnap(rightOneThirds) }},
		{12, b.FirstTwoThirds, func() { directSnap(leftTwoThirds) }},
		{13, b.LastTwoThirds, func() { directSnap(rightTwoThirds) }},

		// Sizing & Placement
		{14, b.Maximize, func() {
			lastResized = 0
			if err := maximize(); err != nil {
				fmt.Printf("warn: maximize: %v\n", err)
			}
		}},
		{15, b.AlmostMaximize, func() { directSnap(almostMaximize) }},
		{16, b.Center, func() { directSnap(center) }},
		{17, b.Restore, func() {
			hwnd := w32.GetForegroundWindow()
			if isZonableWindow(hwnd) {
				w32.ShowWindow(hwnd, w32.SW_RESTORE)
			}
		}},
		{18, b.Minimize, func() {
			hwnd := w32.GetForegroundWindow()
			if isZonableWindow(hwnd) {
				w32.ShowWindow(hwnd, w32.SW_MINIMIZE)
			}
		}},
		{19, b.MakeLarger, func() { directSnap(makeLarger) }},
		{20, b.MakeSmaller, func() { directSnap(makeSmaller) }},

		// Multi-Monitor & State
		{21, b.NextMonitor, func() {
			hwnd := w32.GetForegroundWindow()
			if err := MoveWindowToMonitor(hwnd, true); err != nil {
				fmt.Printf("warn: move to next monitor: %v\n", err)
			}
		}},
		{22, b.PrevMonitor, func() {
			hwnd := w32.GetForegroundWindow()
			if err := MoveWindowToMonitor(hwnd, false); err != nil {
				fmt.Printf("warn: move to prev monitor: %v\n", err)
			}
		}},
		{23, b.AlwaysOnTop, func() {
			hwnd := w32.GetForegroundWindow()
			if err := toggleAlwaysOnTop(hwnd); err != nil {
				fmt.Printf("warn: toggleAlwaysOnTop: %v\n", err)
				return
			}
			fmt.Printf("> toggled always on top: %v\n", hwnd)
		}},
	}

	var failedHotKeys []HotKey
	for _, act := range actions {
		if strings.TrimSpace(act.keyStr) == "" {
			continue
		}
		mod, vk, err := ParseKeyCombo(act.keyStr)
		if err != nil {
			fmt.Printf("warn: invalid key combo %q for action %d: %v\n", act.keyStr, act.id, err)
			continue
		}
		if vk == 0 {
			continue
		}

		hk := HotKey{
			id:       act.id,
			mod:      mod,
			vk:       vk,
			callback: act.callback,
		}

		if !RegisterHotKey(hk) {
			failedHotKeys = append(failedHotKeys, hk)
		}
	}

	if len(failedHotKeys) > 0 {
		msg := "The following hotkey(s) are in use by another process:\n\n"
		for _, hk := range failedHotKeys {
			msg += "  - " + hk.Describe() + "\n"
		}
		msg += "\nTo use these hotkeys, rebind them in config.json or close the conflicting application."
		showMessageBox(msg)
	}
}

func showMessageBox(text string) {
	w32.MessageBox(w32.GetActiveWindow(), text, "RectangleWin", w32.MB_ICONINFORMATION|w32.MB_OK)
}

type resizeFunc func(disp, cur w32.RECT) w32.RECT

func center(disp, cur w32.RECT) w32.RECT {
	w := (disp.Width() - cur.Width()) / 2
	h := (disp.Height() - cur.Height()) / 2
	return w32.RECT{
		Left:   disp.Left + w,
		Right:  disp.Left + w + cur.Width(),
		Top:    disp.Top + h,
		Bottom: disp.Top + h + cur.Height(),
	}
}

// windowBorderOffsets returns the invisible Win10/11 border widths (left, top, right, bottom)
// between the full window rect (GetWindowRect) and the visible frame (DWMWA_EXTENDED_FRAME_BOUNDS).
func windowBorderOffsets(winRect, visibleFrame w32.RECT) (left, top, right, bottom int32) {
	return visibleFrame.Left - winRect.Left,
		visibleFrame.Top - winRect.Top,
		winRect.Right - visibleFrame.Right,
		winRect.Bottom - visibleFrame.Bottom
}

// expandRectForBorders expands a target visible rect outward by the invisible window borders
// so that the visible frame aligns with targetVisible.
func expandRectForBorders(targetVisible w32.RECT, left, top, right, bottom int32) w32.RECT {
	return w32.RECT{
		Left:   targetVisible.Left - left,
		Top:    targetVisible.Top - top,
		Right:  targetVisible.Right + right,
		Bottom: targetVisible.Bottom + bottom,
	}
}

func getWindowFrame(hwnd w32.HWND, fallback w32.RECT) w32.RECT {
	if ok, frame := w32.DwmGetWindowAttributeEXTENDED_FRAME_BOUNDS(hwnd); ok {
		return frame
	}
	return fallback
}

// setWindowVisibleRect positions hwnd so its visible DWM frame matches targetVisible.
// Under Per-Monitor V2 DPI awareness, GetWindowRect, DWMWA_EXTENDED_FRAME_BOUNDS,
// GetMonitorInfo, and SetWindowPos all use physical virtual-screen pixels.
// If moving the window changes its monitor DPI (triggering WM_DPICHANGED in the target app),
// it re-measures the new invisible borders and re-applies SetWindowPos.
func setWindowVisibleRect(hwnd w32.HWND, targetVisible w32.RECT) error {
	dpiBefore := w32ex.GetDpiForWindow(hwnd)

	apply := func() error {
		winRect := w32.GetWindowRect(hwnd)
		if winRect == nil {
			return fmt.Errorf("failed to GetWindowRect:%d", w32.GetLastError())
		}
		frame := getWindowFrame(hwnd, *winRect)
		lExtra, tExtra, rExtra, bExtra := windowBorderOffsets(*winRect, frame)
		pos := expandRectForBorders(targetVisible, lExtra, tExtra, rExtra, bExtra)
		if !w32.SetWindowPos(hwnd, 0, int(pos.Left), int(pos.Top), int(pos.Width()), int(pos.Height()), w32.SWP_NOZORDER|w32.SWP_NOACTIVATE) {
			return fmt.Errorf("failed to SetWindowPos:%d", w32.GetLastError())
		}
		return nil
	}

	if err := apply(); err != nil {
		return err
	}

	dpiAfter := w32ex.GetDpiForWindow(hwnd)
	if dpiBefore != 0 && dpiAfter != 0 && dpiAfter != dpiBefore {
		if err := apply(); err != nil {
			return err
		}
	}
	return nil
}

func resize(hwnd w32.HWND, f resizeFunc) (bool, error) {
	if !isZonableWindow(hwnd) {
		fmt.Printf("warn: non-zonable window: %s\n", w32.GetWindowText(hwnd))
		return false, nil
	}

	// Capture the current monitor BEFORE unmaximizing so a maximized window on a
	// secondary monitor stays on that monitor even if its restored placement was elsewhere.
	mon := w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST)
	var monInfo w32.MONITORINFO
	if !w32.GetMonitorInfo(mon, &monInfo) {
		return false, fmt.Errorf("failed to GetMonitorInfo:%d", w32.GetLastError())
	}

	wasZoomed := w32ex.IsZoomed(hwnd)
	if wasZoomed {
		w32.ShowWindow(hwnd, w32.SW_SHOWNORMAL)
	}

	rect := w32.GetWindowRect(hwnd)
	if rect == nil {
		return false, fmt.Errorf("failed to GetWindowRect:%d", w32.GetLastError())
	}
	frame := getWindowFrame(hwnd, *rect)

	// Under Per-Monitor V2 DPI awareness, monInfo.RcWork, rect, and frame are all in
	// physical virtual-screen pixels across every monitor.
	newVisible := f(monInfo.RcWork, frame)
	lExtra, tExtra, rExtra, bExtra := windowBorderOffsets(*rect, frame)
	newPos := expandRectForBorders(newVisible, lExtra, tExtra, rExtra, bExtra)

	lastResized = hwnd
	if !wasZoomed && sameRect(rect, &newPos) {
		return false, nil
	}

	if err := setWindowVisibleRect(hwnd, newVisible); err != nil {
		return false, err
	}
	return true, nil
}

func maximize() error {
	hwnd := w32.GetForegroundWindow()
	if !isZonableWindow(hwnd) {
		return errors.New("foreground window is not zonable")
	}
	if !w32.ShowWindow(hwnd, w32.SW_MAXIMIZE) {
		return fmt.Errorf("failed to ShowWindow:%d", w32.GetLastError())
	}
	return nil
}

func toggleAlwaysOnTop(hwnd w32.HWND) error {
	if !isZonableWindow(hwnd) {
		return errors.New("foreground window is not zonable")
	}

	if w32.GetWindowLong(hwnd, w32.GWL_EXSTYLE)&w32.WS_EX_TOPMOST != 0 {
		if !w32.SetWindowPos(hwnd, w32.HWND_NOTOPMOST, 0, 0, 0, 0, w32.SWP_NOMOVE|w32.SWP_NOSIZE) {
			return fmt.Errorf("failed to SetWindowPos(HWND_NOTOPMOST): %v", w32.GetLastError())
		}
	} else {
		if !w32.SetWindowPos(hwnd, w32.HWND_TOPMOST, 0, 0, 0, 0, w32.SWP_NOMOVE|w32.SWP_NOSIZE) {
			return fmt.Errorf("failed to SetWindowPos(HWND_TOPMOST) :%v", w32.GetLastError())
		}
	}
	return nil
}

func sameRect(a, b *w32.RECT) bool {
	return a != nil && b != nil && reflect.DeepEqual(*a, *b)
}
