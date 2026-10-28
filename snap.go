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

import "github.com/gonutz/w32/v2"

func toLeft(d w32.RECT, mul, div int32) w32.RECT {
	return w32.RECT{
		Left:   d.Left,
		Top:    d.Top,
		Right:  d.Left + (d.Width()*mul)/div,
		Bottom: d.Top + d.Height(),
	}
}

func toRight(d w32.RECT, mul, div int32) w32.RECT {
	return w32.RECT{
		Left:   d.Left + d.Width() - (d.Width()*mul)/div,
		Top:    d.Top,
		Right:  d.Left + d.Width(),
		Bottom: d.Top + d.Height(),
	}
}

func toTop(d w32.RECT, mul, div int32) w32.RECT {
	return w32.RECT{
		Left:   d.Left,
		Top:    d.Top,
		Right:  d.Left + d.Width(),
		Bottom: d.Top + (d.Height()*mul)/div,
	}
}

func toBottom(d w32.RECT, mul, div int32) w32.RECT {
	return w32.RECT{
		Left:   d.Left,
		Top:    d.Top + d.Height() - (d.Height()*mul)/div,
		Right:  d.Left + d.Width(),
		Bottom: d.Top + d.Height(),
	}
}

// Halves & Thirds
func leftHalf(disp, _ w32.RECT) w32.RECT      { return toLeft(disp, 1, 2) }
func leftOneThirds(disp, _ w32.RECT) w32.RECT { return toLeft(disp, 1, 3) }
func leftTwoThirds(disp, _ w32.RECT) w32.RECT { return toLeft(disp, 2, 3) }

func rightHalf(disp, _ w32.RECT) w32.RECT      { return toRight(disp, 1, 2) }
func rightOneThirds(disp, _ w32.RECT) w32.RECT { return toRight(disp, 1, 3) }
func rightTwoThirds(disp, _ w32.RECT) w32.RECT { return toRight(disp, 2, 3) }

func topHalf(disp, _ w32.RECT) w32.RECT      { return toTop(disp, 1, 2) }
func topOneThirds(disp, _ w32.RECT) w32.RECT { return toTop(disp, 1, 3) }
func topTwoThirds(disp, _ w32.RECT) w32.RECT { return toTop(disp, 2, 3) }

func bottomHalf(disp, _ w32.RECT) w32.RECT      { return toBottom(disp, 1, 2) }
func bottomOneThirds(disp, _ w32.RECT) w32.RECT { return toBottom(disp, 1, 3) }
func bottomTwoThirds(disp, _ w32.RECT) w32.RECT { return toBottom(disp, 2, 3) }

// Center Third
func centerThird(disp, _ w32.RECT) w32.RECT {
	w := disp.Width() / 3
	return w32.RECT{
		Left:   disp.Left + w,
		Top:    disp.Top,
		Right:  disp.Left + w + w,
		Bottom: dBottom(disp),
	}
}

// Corners
func topLeftHalf(disp, _ w32.RECT) w32.RECT      { return merge(toLeft(disp, 1, 2), toTop(disp, 1, 2)) }
func topLeftTwoThirds(disp, _ w32.RECT) w32.RECT { return merge(toLeft(disp, 2, 3), toTop(disp, 1, 2)) }
func topLeftOneThirds(disp, _ w32.RECT) w32.RECT { return merge(toLeft(disp, 1, 3), toTop(disp, 1, 2)) }

func topRightHalf(disp, _ w32.RECT) w32.RECT      { return merge(toRight(disp, 1, 2), toTop(disp, 1, 2)) }
func topRightTwoThirds(disp, _ w32.RECT) w32.RECT { return merge(toRight(disp, 2, 3), toTop(disp, 1, 2)) }
func topRightOneThirds(disp, _ w32.RECT) w32.RECT { return merge(toRight(disp, 1, 3), toTop(disp, 1, 2)) }

func bottomLeftHalf(disp, _ w32.RECT) w32.RECT      { return merge(toLeft(disp, 1, 2), toBottom(disp, 1, 2)) }
func bottomLeftTwoThirds(disp, _ w32.RECT) w32.RECT { return merge(toLeft(disp, 2, 3), toBottom(disp, 1, 2)) }
func bottomLeftOneThirds(disp, _ w32.RECT) w32.RECT { return merge(toLeft(disp, 1, 3), toBottom(disp, 1, 2)) }

func bottomRightHalf(disp, _ w32.RECT) w32.RECT      { return merge(toRight(disp, 1, 2), toBottom(disp, 1, 2)) }
func bottomRightTwoThirds(disp, _ w32.RECT) w32.RECT { return merge(toRight(disp, 2, 3), toBottom(disp, 1, 2)) }
func bottomRightOneThirds(disp, _ w32.RECT) w32.RECT { return merge(toRight(disp, 1, 3), toBottom(disp, 1, 2)) }

// Almost Maximize
func almostMaximize(disp, _ w32.RECT) w32.RECT {
	marginW := disp.Width() * 5 / 100
	marginH := disp.Height() * 5 / 100
	return w32.RECT{
		Left:   disp.Left + marginW,
		Top:    disp.Top + marginH,
		Right:  disp.Right - marginW,
		Bottom: disp.Bottom - marginH,
	}
}

// Make window larger / smaller by step
func makeLarger(disp, cur w32.RECT) w32.RECT {
	dw := disp.Width() * 5 / 100
	dh := disp.Height() * 5 / 100
	newW := cur.Width() + dw*2
	newH := cur.Height() + dh*2
	if newW > disp.Width() {
		newW = disp.Width()
	}
	if newH > disp.Height() {
		newH = disp.Height()
	}
	newLeft := cur.Left - dw
	newTop := cur.Top - dh
	if newLeft < disp.Left {
		newLeft = disp.Left
	}
	if newTop < disp.Top {
		newTop = disp.Top
	}
	if newLeft+newW > disp.Right {
		newLeft = disp.Right - newW
	}
	if newTop+newH > disp.Bottom {
		newTop = disp.Bottom - newH
	}
	return w32.RECT{
		Left:   newLeft,
		Top:    newTop,
		Right:  newLeft + newW,
		Bottom: newTop + newH,
	}
}

func makeSmaller(disp, cur w32.RECT) w32.RECT {
	dw := disp.Width() * 5 / 100
	dh := disp.Height() * 5 / 100
	newW := cur.Width() - dw*2
	newH := cur.Height() - dh*2
	minW := disp.Width() * 20 / 100
	minH := disp.Height() * 20 / 100
	if newW < minW {
		newW = minW
	}
	if newH < minH {
		newH = minH
	}
	newLeft := cur.Left + dw
	newTop := cur.Top + dh
	return w32.RECT{
		Left:   newLeft,
		Top:    newTop,
		Right:  newLeft + newW,
		Bottom: newTop + newH,
	}
}

func merge(x, y w32.RECT) w32.RECT {
	return w32.RECT{Left: x.Left, Right: x.Right, Top: y.Top, Bottom: y.Bottom}
}

func dBottom(d w32.RECT) int32 {
	return d.Top + d.Height()
}
