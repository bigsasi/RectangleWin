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

package w32ex

import (
	"syscall"
	"unsafe"

	"github.com/gonutz/w32/v2"
)

const (
	GA_PARENT    = 1
	GA_ROOT      = 2
	GA_ROOTOWNER = 3
)

var user32 = syscall.NewLazyDLL("user32.dll")

func RegisterHotKey(hwnd w32.HWND, id, mod, vk int) bool {
	r1, _, _ := user32.NewProc("RegisterHotKey").Call(uintptr(hwnd), uintptr(id), uintptr(mod), uintptr(vk))
	return r1 != 0
}

func GetDpiForWindow(hwnd w32.HWND) int32 {
	r1, _, _ := user32.NewProc("GetDpiForWindow").Call(uintptr(hwnd))
	return int32(r1)
}

func GetWindowModuleFileName(hwnd w32.HWND) string {
	var path [32768]uint16
	ret, _, _ := user32.NewProc("GetWindowModuleFileNameW").Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(&path[0])),
		uintptr(len(path)),
	)
	if ret == 0 {
		return ""
	}
	return syscall.UTF16ToString(path[:])
}

func GetAncestor(hwnd w32.HWND, gaFlags uint) w32.HWND {
	r1, _, _ := user32.NewProc("GetAncestor").Call(uintptr(hwnd), uintptr(gaFlags))
	return w32.HWND(r1)
}

func GetShellWindow() (hwnd w32.HWND) {
	r1, _, _ := user32.NewProc("GetShellWindow").Call()
	return w32.HWND(r1)
}

func UnregisterHotKey(hwnd w32.HWND, id int) bool {
	r1, _, _ := user32.NewProc("UnregisterHotKey").Call(uintptr(hwnd), uintptr(id))
	return r1 != 0
}

func EnableModernDPIAwareness() bool {
	// Try SetProcessDpiAwarenessContext (Per-Monitor v2) first (Windows 10 1703+)
	procSetContext := user32.NewProc("SetProcessDpiAwarenessContext")
	if procSetContext.Find() == nil {
		// DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 = -4
		v2Context := ^uintptr(3)
		r1, _, _ := procSetContext.Call(v2Context)
		if r1 != 0 {
			return true
		}
	}

	// Fallback to legacy SetProcessDPIAware
	return SetProcessDPIAware()
}

func SetProcessDPIAware() bool {
	r1, _, _ := user32.NewProc("SetProcessDPIAware").Call()
	return r1 != 0
}

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	comctl32 = syscall.NewLazyDLL("comctl32.dll")
)

func GetCurrentThreadId() uint32 {
	r1, _, _ := kernel32.NewProc("GetCurrentThreadId").Call()
	return uint32(r1)
}

func PostThreadMessage(idThread uint32, msg uint32, wParam, lParam uintptr) bool {
	r1, _, _ := user32.NewProc("PostThreadMessageW").Call(
		uintptr(idThread),
		uintptr(msg),
		wParam,
		lParam,
	)
	return r1 != 0
}

func IsDialogMessage(hwnd w32.HWND, msg *w32.MSG) bool {
	r1, _, _ := user32.NewProc("IsDialogMessageW").Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(msg)),
	)
	return r1 != 0
}

func CreateSegoeUIFont(pointSize int) w32.HFONT {
	// Point size to height: -MulDiv(pointSize, DPI, 72)
	height := -(pointSize * 96 / 72)
	r1, _, _ := gdi32.NewProc("CreateFontW").Call(
		uintptr(height),
		0, 0, 0,
		400, // FW_NORMAL
		0, 0, 0,
		1, // DEFAULT_CHARSET
		0, 0, 0, 0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Segoe UI"))),
	)
	if r1 == 0 {
		return w32.HFONT(w32.GetStockObject(17)) // DEFAULT_GUI_FONT
	}
	return w32.HFONT(r1)
}

func InitCommonControls() {
	proc := comctl32.NewProc("InitCommonControls")
	if proc.Find() == nil {
		proc.Call()
	}
}

func RegisterWindowMessage(name string) uint32 {
	r1, _, _ := user32.NewProc("RegisterWindowMessageW").Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(name))),
	)
	return uint32(r1)
}

func CreateMutex(name string) (syscall.Handle, bool, error) {
	r1, _, err := kernel32.NewProc("CreateMutexW").Call(
		0,
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(name))),
	)
	if r1 == 0 {
		return 0, false, err
	}
	const ERROR_ALREADY_EXISTS = 183
	if errno, ok := err.(syscall.Errno); ok && errno == ERROR_ALREADY_EXISTS {
		return syscall.Handle(r1), false, nil
	}
	return syscall.Handle(r1), true, nil
}
