/*
   Copyright 2022 Josh Deprez

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// The ui program provides a user interface to the csirac implementation.
package main

import (
	"embed"
	"image/color"
	"image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed embed
var embeds embed.FS

var (
	crtsym = mustLoadImage("embed/crtsym.png")
)

func main() {
	ebiten.SetWindowResizable(true)
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("TODO")

	ebiten.RunGame(csiracUI{})
}

type csiracUI struct{}

func (csiracUI) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{69, 69, 69, 255})

	screen.DrawImage(crtsym, nil)
}

func (csiracUI) Layout(int, int) (int, int) { return 640, 480 }

func (csiracUI) Update() error { return nil }

func mustLoadImage(name string) *ebiten.Image {
	f, err := embeds.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	im, err := png.Decode(f)
	if err != nil {
		panic(err)
	}
	return ebiten.NewImageFromImage(im)
}
