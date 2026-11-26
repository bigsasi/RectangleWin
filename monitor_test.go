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

func TestSecondaryMonitorBorderOffsets(t *testing.T) {
	// Secondary monitor to the right (Left=1920) with typical Win10/11 7px invisible borders on L/R/B
	winRect := w32.RECT{Left: 1913, Top: 0, Right: 2887, Bottom: 1047}
	visibleFrame := w32.RECT{Left: 1920, Top: 0, Right: 2880, Bottom: 1040}

	l, top, r, b := windowBorderOffsets(winRect, visibleFrame)
	if l != 7 || top != 0 || r != 7 || b != 7 {
		t.Fatalf("unexpected border offsets: got (%d, %d, %d, %d), want (7, 0, 7, 7)", l, top, r, b)
	}

	target := w32.RECT{Left: 2880, Top: 0, Right: 3840, Bottom: 1040}
	pos := expandRectForBorders(target, l, top, r, b)
	want := w32.RECT{Left: 2873, Top: 0, Right: 3847, Bottom: 1047}
	if pos != want {
		t.Fatalf("expandRectForBorders got %+v, want %+v", pos, want)
	}
}

func TestSecondaryMonitorSnapGeometry(t *testing.T) {
	monitors := []struct {
		name string
		work w32.RECT
	}{
		{"right_secondary", w32.RECT{Left: 1920, Top: 0, Right: 3840, Bottom: 1040}},
		{"left_secondary_negative", w32.RECT{Left: -2560, Top: 0, Right: 0, Bottom: 1400}},
		{"top_secondary_negative", w32.RECT{Left: 0, Top: -1080, Right: 1920, Bottom: -40}},
	}

	for _, m := range monitors {
		t.Run(m.name, func(t *testing.T) {
			lh := leftHalf(m.work, w32.RECT{})
			if lh.Left != m.work.Left || lh.Top != m.work.Top || lh.Width() != m.work.Width()/2 || lh.Height() != m.work.Height() {
				t.Errorf("leftHalf(%+v) = %+v", m.work, lh)
			}

			rh := rightHalf(m.work, w32.RECT{})
			if rh.Right != m.work.Right || rh.Top != m.work.Top || rh.Width() != m.work.Width()/2 || rh.Height() != m.work.Height() {
				t.Errorf("rightHalf(%+v) = %+v", m.work, rh)
			}

			ct := centerThird(m.work, w32.RECT{})
			if ct.Left != m.work.Left+m.work.Width()/3 || ct.Width() != m.work.Width()/3 || ct.Bottom != m.work.Bottom {
				t.Errorf("centerThird(%+v) = %+v", m.work, ct)
			}
		})
	}
}

func TestMapRectToMonitor(t *testing.T) {
	primary := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1040}
	secondaryRight := w32.RECT{Left: 1920, Top: 0, Right: 4480, Bottom: 1400} // 2560x1400
	secondaryLeft := w32.RECT{Left: -1920, Top: 0, Right: 0, Bottom: 1040}

	// Left half of primary -> left half of 2560x1400 secondary
	win := w32.RECT{Left: 0, Top: 0, Right: 960, Bottom: 1040}
	got := mapRectToMonitor(win, primary, secondaryRight)
	want := w32.RECT{Left: 1920, Top: 0, Right: 3200, Bottom: 1400}
	if got != want {
		t.Errorf("primary->secondaryRight got %+v, want %+v", got, want)
	}

	// Right half of secondaryRight -> right half of negative-coordinate secondaryLeft
	win2 := w32.RECT{Left: 3200, Top: 0, Right: 4480, Bottom: 1400}
	got2 := mapRectToMonitor(win2, secondaryRight, secondaryLeft)
	want2 := w32.RECT{Left: -960, Top: 0, Right: 0, Bottom: 1040}
	if got2 != want2 {
		t.Errorf("secondaryRight->secondaryLeft got %+v, want %+v", got2, want2)
	}
}

func TestSortMonitorEntries(t *testing.T) {
	entries := []monitorEntry{
		{handle: 3, rect: w32.RECT{Left: 1920, Top: 0, Right: 3840, Bottom: 1080}},
		{handle: 1, rect: w32.RECT{Left: -1920, Top: 0, Right: 0, Bottom: 1080}},
		{handle: 4, rect: w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}},
		{handle: 2, rect: w32.RECT{Left: 0, Top: -1080, Right: 1920, Bottom: 0}},
	}
	sortMonitorEntries(entries)

	wantOrder := []w32.HMONITOR{1, 2, 4, 3}
	for i, want := range wantOrder {
		if entries[i].handle != want {
			t.Fatalf("sorted index %d: got handle %d, want %d", i, entries[i].handle, want)
		}
	}
}
