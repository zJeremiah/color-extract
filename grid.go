package main

import (
	"image/color"
	"log"

	"github.com/fogleman/gg"
)

type Point struct {
	X float64
	Y float64
}

type Rect struct {
	X0 float64
	Y0 float64
	X1 float64
	Y1 float64
}

type Grid struct {
	Name         string
	Codes        map[Point]Position
	ToManyColors bool
}

type Position struct {
	Area Rect
	Code string
}

func NewPoint(x, y float64) Point {
	return Point{X: x, Y: y}
}

// read the pattern file to generate an image grid
func newImageGrid(width, height int, grid Grid) {
	img := gg.NewContext(width, height)
	err := img.LoadFontFace("./fonts/UbuntuMono-Regular.ttf", 24)
	if err != nil {
		log.Fatalln("error setting font face", err.Error())
	}
	clr := color.RGBA{120, 120, 120, 255}
	img.SetColor(clr)
	img.SetStrokeStyle(gg.NewSolidPattern(clr))

	//img.SetLineWidth(1)
	var x, y float64
	var point, pn Point
	var p, n Position
	var found bool

	for {
		p, found = grid.Value(x, y)
		if !found {
			x = 0
			y++
			p, found = grid.Value(x, y)
			if !found {
				break
			}
		}

		point = Point{x, y}

		// this will draw t he code font onto the image
		img.DrawStringAnchored(p.Code, p.Area.X0, p.Area.Y0, -0.8, 1.3)
		// check to the right of the current pixel
		pn = NewPoint(point.X+1, point.Y)
		n, found = grid.Value(pn.X, pn.Y)
		if found && n.Code != p.Code {
			// if point.X == 0 && point.Y > 0 {
			// 	fmt.Printf("position (%.0f,%.0f)[%s] right (%.0f,%.0f)[%s] drawing line from (%.0f,%.0f) to (%.0f,%.0f)\n ", point.X, point.Y, p.Code, pr.X, pr.Y, r.Code, p.Area.X1, p.Area.Y0, p.Area.X1, p.Area.Y1)
			// }
			img.DrawLine(p.Area.X1, p.Area.Y0, p.Area.X1, p.Area.Y1)
			img.Stroke()
		}

		// check below the current pixel
		pn = NewPoint(point.X, point.Y+1)
		n, found = grid.Value(pn.X, pn.Y)
		if found && n.Code != p.Code {
			img.DrawLine(p.Area.X0, p.Area.Y1, p.Area.X1, p.Area.Y1)
			img.Stroke()
		}

		x++
	}

	// draw a rectangle around the whole grid
	img.DrawRectangle(0, 0, float64(width), float64(height))
	img.Stroke()

	err = img.SavePNG(grid.Name + ".grid.png")
	if err != nil {
		log.Fatalln("error saving png", err.Error())
	}
}

func (g *Grid) Value(x, y float64) (p Position, found bool) {
	p, ok := g.Codes[Point{X: x, Y: y}]
	return p, ok
}

func (g *Grid) Set(ix, iy int, code string) {
	x, y := float64(ix), float64(iy)
	g.Codes[Point{X: x, Y: y}] = Position{
		Code: code,
		Area: Rect{x * 30, y * 30, (x * 30) + 30, (y * 30) + 30},
	}
}
