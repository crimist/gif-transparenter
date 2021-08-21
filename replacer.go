package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"image/color"
	"image/gif"
	"io/ioutil"
	"os"
	"runtime/debug"
	"syscall"
)

func check(err error) {
	if err != nil {
		debug.PrintStack()
		panic(err)
	}
}

func hexColorToRGBA(colorStr string) color.RGBA {
	b, err := hex.DecodeString(colorStr)
	check(err)

	return color.RGBA{b[0], b[1], b[2], 0xFF}
}

func main() {
	inputFile := os.Args[1]
	outputFile := os.Args[2]

	data, err := ioutil.ReadFile(inputFile)
	check(err)
	gifData, err := gif.DecodeAll(bytes.NewReader(data))
	check(err)

	for i := 0; i < len(gifData.Image); i++ {
		img := gifData.Image[i]

		var xMin, xMax = img.Bounds().Min.X, img.Bounds().Max.X
		var yMin, yMax = img.Bounds().Min.Y, img.Bounds().Max.Y

		fmt.Printf("[%-3v] [%vx%v] ([%v->%v] [%v->%v]): Replaced... ", i, xMax, yMax, xMin, xMax, yMin, yMax)

		changed := 0
		for y := yMin; y < yMax; y++ {
			for x := xMin; x < xMax; x++ {
				pix := img.At(x, y).(color.RGBA)

				allMatchN := func(c color.RGBA, n uint8) bool {
					return c.R == n && c.G == n && c.B == n
				}

				allSame := func(c color.RGBA) bool {
					return ((pix.R == pix.G) && (pix.R == pix.B) && (pix.G == pix.B))
				}

				allSameWithin := func(c color.RGBA, within uint8) bool {
					cmpWithin := func(a, b, within uint8) bool {
						if a > b {
							return a-within <= b
						} else {
							return a+within >= b
						}
					}

					return cmpWithin(pix.R, pix.G, within) && cmpWithin(pix.R, pix.B, within) && cmpWithin(pix.G, pix.B, within)
				}

				// if white than continue
				if allSame(pix) && pix.R >= 0xF0 { // all same and above 0xF0
					continue
				}

				// black or white
				if allMatchN(pix, 0) || allMatchN(pix, 0xFF) {
					continue
				}

				// if grey(ish) and all match
				if allSameWithin(pix, 3) && pix.R >= 100 && pix.R <= 170 {
					// replace pixel with null for transparency
					img.Set(x, y, color.RGBA{0, 0, 0, 0})

					changed++
				}
			}
		}

		fmt.Println(changed)
		gifData.Image[i] = img
	}

	var b []byte
	buf := bytes.NewBuffer(b)

	err = gif.EncodeAll(buf, gifData)
	check(err)

	syscall.Umask(0)
	ioutil.WriteFile(outputFile, buf.Bytes(), 0777)
}
