package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/go-playground/colors"
	"github.com/pcelvng/task-tools/file"
	_ "golang.org/x/image/webp"
)

// Pixel struct example
type Pixel struct {
	R int
	G int
	B int
	A int
}

// Stats is the definition of a color in the image
type Stats struct {
	ID string // the id of the color
	Pixel
	Count   int
	Percent float64
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

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

	img {
		width: 500px;
	}

  </style>
  <meta http-equiv="X-UA-Compatible">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
</head>
<body>
<div>
`)

	htmlPixels := `<div><img src="%s"></div><p style="page-break-before: always"></p>
<div style='font-size: larger;margin:20px'>pixel count: %d pixels size (%d x %d) Colors: %d</div>
<div style='font-size: larger;margin:20px'>%s</div>
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

	colorRowHtml := `
  <tr>
    <td style="text-align:center">%s</td>
    <td style="background-color:%s;"></td>
    <td>%d</td>
    <td>%s</td>
    <td>%s</td>
    <td>%.2f%%</td>
  </tr>
`

	flag.Parse()
	input := flag.Arg(0)
	multiple, _ := strconv.ParseFloat(flag.Arg(1), 64)
	if multiple == 0 {
		multiple = 1
	}

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
	total := bounds.Max.X * bounds.Max.Y
	output := make([][]Pixel, bounds.Max.Y)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// get the rgba values of the pixel
			r, g, b, a := m.At(x, y).RGBA()
			// convert the mulitplied rgba values to an actual pixel
			p := rgbaToPixel(r, g, b, a)
			// add the pixel to the matrix
			output[y] = append(output[y], p)
			// add the pixel to the color count
			colorCount[p]++
		}
	}

	buf := bytes.NewBuffer(htmlHeader)

	buf.Write([]byte(
		fmt.Sprintf(htmlPixels,
			base64PNG(imgName),
			total,
			bounds.Max.X,
			bounds.Max.Y,
			len(colorCount),
			sizeStr(bounds.Max)),
	))

	colorList := []string{"Σ", "0", "∆", "1", "#", "2", "3", "4", "5", "6", "7", "8", "9", "C", "D", "F", "G", "H", "J", "K", "L", "M", "N", "P", "Q", "R", "U", "V", "W", "X", "Y", "Z", "@", "$", "*"}
	list := make([]Stats, 0)

	i := 0
	for k, v := range colorCount {
		var c string
		// if there are more colors than the color list, just use the number (can't make pattern)
		if i > len(colorList)-1 {
			c = strconv.Itoa(i)
		} else {
			c = colorList[i]
		}

		list = append(list, Stats{
			ID:      c,
			Pixel:   k,
			Count:   v,
			Percent: float64(v) / float64(total) * 100.0,
		})
		i++
	}

	buf.Write(htmlTable)

	//sort the list by number of pixels
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].Count > list[j].Count
	})

	// loop though the stats list and build a html table row for each item
	for i := range list {
		rgb := fmt.Sprintf("rgb(%d,%d,%d)", list[i].R, list[i].G, list[i].B)
		color, err := colors.Parse(rgb)
		if err != nil {
			log.Println("could not parse color", rgb)
		}

		if i > len(colorList)-1 {
			list[i].ID = strconv.Itoa(i)
		} else {
			list[i].ID = colorList[i]
		}

		id := list[i].ID
		if id == "Σ" {
			id = "&#931;"
		}

		if id == "∆" {
			id = "	&#8710;"
		}

		s := fmt.Sprintf(colorRowHtml,
			id,
			rgb,
			list[i].Count,
			color.ToHEX(),
			rgb,
			list[i].Percent)

		buf.Write([]byte(s))
	}

	buf.Write([]byte("</tbody>\n</table>\n"))
	buf.Write([]byte("</div>\n</body>\n</html>\n"))
	fmt.Print(buf.String())

	// write out a pattern.csv file that can be use to generate a pattern image
	names := strings.Split(imgName, ".")
	grid := Grid{
		Codes: make(map[Point]Position),
		Name:  names[0],
		MaxX:  float64(bounds.Max.X),
		MaxY:  float64(bounds.Max.Y),
	}

	os.Remove("pattern.csv")
	if len(list) <= len(colorList) {
		c, _ := file.NewWriter("pattern.csv", file.NewOptions())
		for iy, y := range output {
			//c.Write([]byte(""))
			for ix, x := range y {
				for _, s := range list {
					if s.Pixel == x {
						c.Write([]byte(s.ID))
						grid.Set(multiple, ix, iy, s.ID)
						break
					}
				}
			}
			c.Write([]byte("\n"))
		}
		c.Close()
	} else {
		log.Println("to many colors to create pattern", len(list), "limit", len(colorList))
		os.Exit(1)
	}

	newImageGrid(bounds.Max.X*30, bounds.Max.Y*30, grid)
}

// img.At(x, y).RGBA() returns four uint32 values; we want a Pixel
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{int(r / 257), int(g / 257), int(b / 257), int(a / 257)}
}

func sizeStr(p image.Point) string {
	mmX := p.X * 2
	mmY := p.Y * 2
	inX := float64(mmX) / 25.4
	inY := float64(mmY) / 25.4

	return fmt.Sprintf("size  (%dmm x %dmm)  (%.2fin, %.2fin)", mmX, mmY, inX, inY)
}

func base64PNG(imgPath string) string {
	b, err := ioutil.ReadFile(imgPath)
	if err != nil {
		log.Fatalln("cannot read image", err)
	}

	var base64Img string

	mimeType := http.DetectContentType(b)

	// Prepend the appropriate URI scheme header depending
	// on the MIME type
	switch mimeType {
	case "image/jpeg":
		base64Img += "data:image/jpeg;base64,"
	case "image/png":
		base64Img += "data:image/png;base64,"
	case "imge/webp":
		base64Img += "data:image/webp;base64,"
	}

	base64Img += base64.StdEncoding.EncodeToString(b)
	return base64Img
}
