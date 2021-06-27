package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"sort"
	"strconv"

	_ "golang.org/x/image/webp"
)

// Pixel struct example
type Pixel struct {
	R int
	G int
	B int
	A int
}

type Stats struct {
	Pixel
	Count   int
	Percent float64
}

type Hex string

func hexModel(c color.Color) color.Color {
	if _, ok := c.(Hex); ok {
		return c
	}
	r, g, b, _ := c.RGBA()
	return RGBToHex(uint8(r>>8), uint8(g>>8), uint8(b>>8))
}

// HexToRGB converts an Hex string to a RGB triple.
func HexToRGB(h Hex) (uint8, uint8, uint8) {
	if len(h) > 0 && h[0] == '#' {
		h = h[1:]
	}
	if len(h) == 3 {
		h = h[:1] + h[:1] + h[1:2] + h[1:2] + h[2:] + h[2:]
	}
	if len(h) == 6 {
		if rgb, err := strconv.ParseUint(string(h), 16, 32); err == nil {
			return uint8(rgb >> 16), uint8((rgb >> 8) & 0xFF), uint8(rgb & 0xFF)
		}
	}
	return 0, 0, 0
}

// RGBA returns the alpha-premultiplied red, green, blue and alpha values
// for the Hex.
func (c Hex) RGBA() (uint32, uint32, uint32, uint32) {
	r, g, b := HexToRGB(c)
	return uint32(r) * 0x101, uint32(g) * 0x101, uint32(b) * 0x101, 0xffff
}

// RGBToHex converts an RGB triple to an Hex string.
func RGBToHex(r, g, b uint8) Hex {
	return Hex(fmt.Sprintf("#%X%X%X", r, g, b))
}

var HexModel = color.ModelFunc(hexModel)

func main() {

	htmlHeader := []byte(`<!DOCTYPE html>
<html>
<head>
  <style>
  td, th {
	border: 1px solid #ddd;
	padding: 9px;
	width: 15%
  }
  
  tr:nth-child(even){background-color: #f2f2f2;}
  
  tr:hover {background-color: #ddd;}

  th {
	padding-top: 12px;
	padding-bottom: 12px;
	text-align: left;
	background-color: #04AA6D;
	color: white;
  }

  </style>
  <meta http-equiv="X-UA-Compatible">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
</head>
<body>
<div style="">
`)

	htmlPixels := (`<div><img src="%s"></div>
<div style='font-size: larger;margin:20px'>pixels: %d size (%d x %d) Colors: %d</div>
`)
	htmlTable := []byte(`
	<table class="table" style="width:70%;">
<thead>
  <tr><th>Color</th><th>Pixel Count</th><th>Color Code</th><th>RGB</th><th>Percentage</th></tr>
</thead>
<tbody>
`)

	htmlColors := (`
  <tr>
    <td style="background-color:rgb(%d,%d,%d);"></td>
	<td>%d</td>
	<td>%s</td>
	<td>rgb(%d, %d, %d)</td>
	<td>%f%%</td>
  </tr>
`)

	flag.Parse()
	input := flag.Arg(0)

	colorCount := make(map[Pixel]int)
	imgName := input

	fmt.Println("[", imgName, "]")

	reader, err := os.Open(imgName)
	if err != nil {
		log.Fatal(err)
	}

	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	bounds := m.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := m.At(x, y).RGBA()
			p := rgbaToPixel(r, g, b, a)

			colorCount[p]++
		}
	}

	buf := bytes.NewBuffer(htmlHeader)

	total := bounds.Max.X * bounds.Max.Y
	buf.Write([]byte(fmt.Sprintf(htmlPixels, imgName, total, bounds.Max.X, bounds.Max.Y, len(colorCount))))

	list := make([]Stats, 0)
	for k, v := range colorCount {

		list = append(list, Stats{
			Pixel:   k,
			Count:   v,
			Percent: float64(v) / float64(total) * 100,
		})
	}

	buf.Write(htmlTable)

	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Count > list[j].Count
	})

	for _, l := range list {
		buf.Write([]byte(fmt.Sprintf(htmlColors,
			l.R, l.G, l.B, l.Count,
			RGBToHex(uint8(l.R), uint8(l.G), uint8(l.B)),
			l.R, l.G, l.B,
			l.Percent,
		)))
	}

	buf.Write([]byte("</tbody></table></div></body>\n</html>\n"))

	fmt.Print(buf.String())
}

// img.At(x, y).RGBA() returns four uint32 values; we want a Pixel
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{int(r / 257), int(g / 257), int(b / 257), int(a / 257)}
}
