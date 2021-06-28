package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"sort"
	"strconv"

	"github.com/go-playground/colors"
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

func main() {
	htmlHeader := []byte(`<!DOCTYPE html>
<html>
<head>
  <style>
  td, th {
	border: 1px solid #ddd;
	padding: 9px;
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
<div>
`)

	htmlPixels := `<div><img src="%s"></div>
<div style='font-size: larger;margin:20px'>pixels: %d size (%d x %d) Colors: %d</div>
`
	htmlTable := []byte(`
	<table class="table" style="width:70%;">
<thead><tr>
  <th style="width:5%">Number</th>
  <th style="width:25%">Color</th>
  <th>Pixel Count</th>
  <th>Color Code</th>
  <th>RGB</th>
  <th>Percentage</th>
</tr></thead>
<tbody>
`)

	htmlColors := `
  <tr>
    <td style="text-align:center">%d</td>
    <td style="background-color:%s;"></td>
	<td>%d</td>
	<td>%s</td>
	<td>%s</td>
	<td>%.2f%%</td>
  </tr>
`

	flag.Parse()
	input := flag.Arg(0)
	row := flag.Arg(1)
	rowNum, _ := strconv.Atoi(row)
	colorCount := make(map[Pixel]int)
	imgName := input

	reader, err := os.Open(imgName)
	if err != nil {
		log.Fatal(err)
	}

	m, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	// loops though all the pixels to get the color of each one
	// and to count each of the colors found
	bounds := m.Bounds()
	outputRow := make([]Pixel, 0)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := m.At(x, y).RGBA()
			p := rgbaToPixel(r, g, b, a)

			if row != "" && y == rowNum {
				outputRow = append(outputRow, p)
			}

			colorCount[p]++
		}
	}

	buf := bytes.NewBuffer(htmlHeader)

	total := bounds.Max.X * bounds.Max.Y
	buf.Write([]byte(
		fmt.Sprintf(htmlPixels,
			imgName,
			total,
			bounds.Max.X,
			bounds.Max.Y,
			len(colorCount)),
	))

	list := make([]Stats, 0)
	for k, v := range colorCount {
		list = append(list, Stats{
			Pixel:   k,
			Count:   v,
			Percent: float64(v) / float64(total) * 100,
		})
	}

	buf.Write(htmlTable)

	//sort the list by number of pixels
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Count > list[j].Count
	})

	// loop though the stats list and build a html table row for each item
	for i, l := range list {
		rgb := fmt.Sprintf("rgb(%d,%d,%d)", l.R, l.G, l.B)
		color, err := colors.Parse(rgb)
		if err != nil {
			log.Println("could not parse color", rgb)
		}
		s := fmt.Sprintf(htmlColors,
			i,
			rgb,
			l.Count,
			color.ToHEX(),
			rgb,
			l.Percent)

		buf.Write([]byte(s))
	}

	buf.Write([]byte("</tbody>\n</table>\n"))

	// Added a 2nd Arg to allow the printing of a list of the colors (by number)
	// in a selected row
	if len(outputRow) > 0 {
		buf.Write([]byte("<br><hr><div>Row:" + row + " Output Colors</div>\n<div>\n"))
		for _, v := range outputRow {
			for i, b := range list {
				if b.Pixel == v {
					buf.Write([]byte(fmt.Sprintf("%d, ", i)))
					break
				}
			}
		}
		buf.Write([]byte("</div>\n"))
	}

	buf.Write([]byte("</div>\n</body>\n</html>\n"))
	fmt.Print(buf.String())
}

// img.At(x, y).RGBA() returns four uint32 values; we want a Pixel
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{int(r / 257), int(g / 257), int(b / 257), int(a / 257)}
}
