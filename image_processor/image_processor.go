package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
)

type EmptyLinesRange struct {
	index  int
	length int
}

func columnGaps(img image.Image) []EmptyLinesRange {
	bounds := img.Bounds()
	width := bounds.Size().X
	var result []bool
	currentLineValue := 0
	index := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			var mono int
			if float64(r+g+b)/3.0 < 128 {
				mono = 0
			} else {
				mono = 255
			}

			if x == 0 {
				currentLineValue = 0
				//				result[index] = false
				result = append(result, false)
			}

			if mono == 255 {
				currentLineValue += 1
			}
			if x == width-1 && currentLineValue == width {
				result[y] = true
			}
			index += 1
		}
	}

	prevLineIsEmpty := false
	firstEmptyLine := 0
	var emptyRanges []EmptyLinesRange
	for i, isEmpty := range result {
		if isEmpty {
			if !prevLineIsEmpty {
				firstEmptyLine = i
			}
			prevLineIsEmpty = true
		} else {
			if prevLineIsEmpty {
				emptyRanges = append(emptyRanges, EmptyLinesRange{firstEmptyLine, i - firstEmptyLine})
			}
			prevLineIsEmpty = false
		}
	}

	return emptyRanges
}

func decodeImage(filePath string) (img image.Image, err error) {
	reader, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	img, _, err = image.Decode(reader)
	return
}

func guessDivideColumns(emptyRanges []EmptyLinesRange) (r []EmptyLinesRange, err error) {
	var results []EmptyLinesRange
	for _, r := range emptyRanges {
		if r.length > 30 {
			results = append(results, r)
		}
	}
	if len(results) == 0 {
		return nil, errors.New("not found")
	}
	return results, nil
}

func drawHorizontalRedLine(img *image.RGBA, y int) {
	fmt.Printf("y = %d\n", y)
	x1 := 0
	x2 := img.Bounds().Size().X
	col := color.RGBA{0xff, 0x00, 0x00, 0xff}
	for ; x1 <= x2; x1++ {
		//img.Set(x1, y-1, col)
		img.Set(x1, y, col)
		//img.Set(x1, y+1, col)
	}
}

func main() {
	path := "../tmp/keepingtwo01.gif"
	img, err := decodeImage(path)
	if err != nil {
		log.Fatal(err)
	}
	emptyLinesRanges := columnGaps(img)
	for i, r := range emptyLinesRanges {
		fmt.Printf("%d: EmptyLinesRange(index=%d, length=%d)\n", i, r.index, r.length)
	}

	divideRanges, err := guessDivideColumns(emptyLinesRanges)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("guessed column: \v\n\n", divideRanges)

	dstImage := image.NewRGBA(img.Bounds())
	draw.Draw(dstImage, dstImage.Bounds(), img, image.ZP, draw.Src)
	for _, dr := range divideRanges {
		y := dr.index + int(dr.length/2)
		drawHorizontalRedLine(dstImage, y)
	}
	path, err = filepath.Abs(path + ".result.jpg")
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = jpeg.Encode(file, dstImage, nil)
	if err != nil {
		log.Fatal(err)
	}
}
