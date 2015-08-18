package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
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

func main() {
	img, err := decodeImage("../tmp/keepingtwo01.gif")
	if err != nil {
		fmt.Printf("err = %v\n", err)
		return
	}
	for i, r := range columnGaps(img) {
		fmt.Printf("%d: EmptyLinesRange(index=%d, length=%d)\n", i, r.index, r.length)
	}
}
