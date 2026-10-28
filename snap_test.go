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

func TestSnapGeometry(t *testing.T) {
	disp := w32.RECT{Left: 0, Top: 0, Right: 1920, Bottom: 1080}
	cur := w32.RECT{Left: 100, Top: 100, Right: 900, Bottom: 700}

	tests := []struct {
		name string
		f    resizeFunc
		w    int32
		h    int32
		l    int32
		t    int32
	}{
		{"leftHalf", leftHalf, 960, 1080, 0, 0},
		{"rightHalf", rightHalf, 960, 1080, 960, 0},
		{"topHalf", topHalf, 1920, 540, 0, 0},
		{"bottomHalf", bottomHalf, 1920, 540, 0, 540},
		{"leftOneThirds", leftOneThirds, 640, 1080, 0, 0},
		{"leftTwoThirds", leftTwoThirds, 1280, 1080, 0, 0},
		{"rightOneThirds", rightOneThirds, 640, 1080, 1280, 0},
		{"rightTwoThirds", rightTwoThirds, 1280, 1080, 640, 0},
		{"centerThird", centerThird, 640, 1080, 640, 0},
		{"topLeftHalf", topLeftHalf, 960, 540, 0, 0},
		{"topRightHalf", topRightHalf, 960, 540, 960, 0},
		{"bottomLeftHalf", bottomLeftHalf, 960, 540, 0, 540},
		{"bottomRightHalf", bottomRightHalf, 960, 540, 960, 540},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := tt.f(disp, cur)
			if res.Width() != tt.w {
				t.Errorf("%s: expected width %d, got %d", tt.name, tt.w, res.Width())
			}
			if res.Height() != tt.h {
				t.Errorf("%s: expected height %d, got %d", tt.name, tt.h, res.Height())
			}
			if res.Left != tt.l {
				t.Errorf("%s: expected left %d, got %d", tt.name, tt.l, res.Left)
			}
			if res.Top != tt.t {
				t.Errorf("%s: expected top %d, got %d", tt.name, tt.t, res.Top)
			}
		})
	}
}

func TestAlmostMaximize(t *testing.T) {
	disp := w32.RECT{Left: 0, Top: 0, Right: 1000, Bottom: 1000}
	cur := w32.RECT{Left: 0, Top: 0, Right: 500, Bottom: 500}
	res := almostMaximize(disp, cur)

	if res.Left != 50 || res.Top != 50 || res.Right != 950 || res.Bottom != 950 {
		t.Errorf("almostMaximize unexpected result: %#v", res)
	}
}
