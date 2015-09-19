package main

import (
	"errors"
	"fmt"
	"github.com/koyachi/kttm/utility"
	"image"
	"image/color"
	"image/draw"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	DivideColumnsNotFoundError = errors.New("divide columns not found.")
)

type EmptyLinesRange struct {
	index  int
	length int
}

type ColumnDivider interface {
	DivideColumns(emptyRanges []EmptyLinesRange) (r []EmptyLinesRange, err error)
}

type GapSize int

func (g GapSize) DivideColumns(emptyRanges []EmptyLinesRange) (r []EmptyLinesRange, err error) {
	var results []EmptyLinesRange
	for _, r := range emptyRanges {
		if r.length > int(g) {
			results = append(results, r)
		}
	}
	if len(results) == 0 {
		return nil, DivideColumnsNotFoundError
	}
	return results, nil
}

type FixedGapInfo struct {
	threshold GapSize
	fixedGaps []int
}

func (f FixedGapInfo) DivideColumns(emptyRanges []EmptyLinesRange) (r []EmptyLinesRange, err error) {
	var results []EmptyLinesRange
	currentGapIndex := 0
	currentGapCount := -1
	for _, r := range emptyRanges {
		if r.length > int(f.threshold) {
			currentGapCount++
		}
		if f.fixedGaps[currentGapIndex] == currentGapCount {
			results = append(results, r)
			if currentGapIndex == len(f.fixedGaps)-1 {
				break
			}
			currentGapIndex++
			currentGapCount = 0
		}
	}
	if len(results) == 0 {
		return nil, DivideColumnsNotFoundError
	}
	return results, nil
}

type FileColumnDividerMap map[string]ColumnDivider

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
			if float64(r+g+b)/3.0 < float64(0xffff/2) {
				mono = 0
			} else {
				mono = 255
			}

			if x == 0 {
				currentLineValue = 0
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

func drawHorizontalRedLine(img *image.RGBA, y int) {
	fmt.Printf("y = %d\n", y)
	x1 := 0
	x2 := img.Bounds().Size().X
	col := color.RGBA{0xff, 0x00, 0x00, 0xff}
	for ; x1 <= x2; x1++ {
		img.Set(x1, y-1, col)
		img.Set(x1, y, col)
		img.Set(x1, y+1, col)
	}
}

func divideImageVertically(img image.Image, top int, height int) image.Image {
	h := height
	if top+height > img.Bounds().Max.Y {
		h = img.Bounds().Max.Y - top
	}
	srcPoint := image.Point{0, top}
	dstRect := image.Rect(0, 0, img.Bounds().Max.X, h)
	dstImage := image.NewRGBA(dstRect)
	draw.Draw(dstImage, dstImage.Bounds(), img, srcPoint, draw.Src)
	return dstImage
}

func process(config utility.Config, path string, columnDivider ColumnDivider) error {
	img, err := utility.DecodeImage(path)
	if err != nil {
		return err
	}
	emptyLinesRanges := columnGaps(img)
	for i, r := range emptyLinesRanges {
		fmt.Printf("%d: EmptyLinesRange(index=%d, length=%d)\n", i, r.index, r.length)
	}

	divideRanges, err := columnDivider.DivideColumns(emptyLinesRanges)
	if err != nil {
		return err
	}
	fmt.Printf("guessed column: \v\n\n", divideRanges)

	_, fileName := filepath.Split(path)
	outputDir := config.ImageProcessorOutputDir
	path, err = filepath.Abs(outputDir + fileName + ".result.jpg")
	if err != nil {
		return err
	}

	dstImage := image.NewRGBA(img.Bounds())
	draw.Draw(dstImage, dstImage.Bounds(), img, image.ZP, draw.Src)
	top := 0
	i := 0
	for _, dr := range divideRanges {
		y := dr.index + int(dr.length/2)
		height := y - top
		divImage := divideImageVertically(dstImage, top, height)
		divPath := outputDir + fileName + ".div_" + strconv.Itoa(i) + ".jpg"
		err := utility.SaveImage(divImage, divPath)
		if err != nil {
			return err
		}
		top = y
		i++
	}
	if top != dstImage.Bounds().Max.Y {
		height := dstImage.Bounds().Max.Y - top
		divImage := divideImageVertically(dstImage, top, height)
		divPath := outputDir + fileName + ".div_" + strconv.Itoa(i) + ".jpg"
		err := utility.SaveImage(divImage, divPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func processDir(config utility.Config, defaultGapSize GapSize, m FileColumnDividerMap) error {
	err := filepath.Walk(config.SrcDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		fmt.Println(path)
		rel, err := filepath.Rel(config.SrcDir, path)
		ext := filepath.Ext(rel)
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
			return nil
		}
		var divider ColumnDivider
		for k, v := range m {
			if strings.Contains(path, k) {
				divider = v
			}
		}
		if divider == nil {
			divider = ColumnDivider(defaultGapSize)
		}
		p := config.SrcDir + "/" + rel
		err = process(config, p, divider)
		if err == nil || err == DivideColumnsNotFoundError {
			return nil
		} else {
			return err
		}
	})
	if err != nil {
		return err
	}
	return nil
}

func main() {
	config := utility.NewConfig()
	//err := process("../data/src/keepingtwo15.gif") ok
	//err := process("../data/src/keepingtwo16.gif") !
	//err := process("../data/src/keeptwo_37a.gif") divide columns not found
	// TODO:
	// - 20,21
	//err := process("../data/src/keepingtwo20.gif", GapSize(59))
	// - 22: 50では一部大きい
	//err := process("../data/src/keepingtwo22.gif")
	// - 24,
	// x 25: >48では一部大きい
	//       3つめが狭い。狭さだけで判断できない
	//err := process("../data/src/keepingtwo25.gif", ColumnDivider(gapSize)) x
	/*
		var f FixedGapInfo
		f.threshold = 19
		f.fixedGaps = []int{3, 3, 3, 3}
		err := process("../data/src/keepingtwo25.gif", ColumnDivider(f))
	*/

	// - 33a
	//g33a := GapSize(44)
	//err := process("../data/src/keeptwo_33a.gif", ColumnDivider(g33a))
	// - 41
	//g41 := GapSize(44)
	//err := process("../data/src/keeptwo_41.gif", ColumnDivider(g41))
	// - 42a
	//g42a := GapSize(44)
	//err := process("../data/src/keeptwo_42a.gif", ColumnDivider(g42a))
	// - 56c
	//g56c := GapSize(42)
	//err := process("../data/src/keeptwo_56c.gif", ColumnDivider(g56c))

	// - 44a: (1,3,3)だが43bが2なのであってる
	// - 44b: (3,1)だが45aが2なのであってる
	// x 51a: (1,2,3)。最初の1,2は合わせて3になるべき。
	/*
		var f FixedGapInfo
		f.threshold = 17
		f.fixedGaps = []int{3, 3}
		err := process("../data/src/keeptwo_51a.gif", ColumnDivider(f))
	*/
	gap59 := GapSize(59)
	gap44 := GapSize(44)
	gap42 := GapSize(42)
	var f25 FixedGapInfo
	f25.threshold = 19
	f25.fixedGaps = []int{3, 3, 3, 3}
	var f51a FixedGapInfo
	f51a.threshold = 17
	f51a.fixedGaps = []int{3, 3}
	m := FileColumnDividerMap{
		"keepingtwo20.gif": gap59,
		"keeptwo_33a.gif":  gap44,
		"keeptwo_41.gif":   gap44,
		"keeptwo_42a.gif":  gap44,
		"keeptwo_56c.gif":  gap42,
		"keepingtwo25.gif": ColumnDivider(f25),
		"keeptwo_51a.gif":  ColumnDivider(f51a),
	}
	defaultGapSize := GapSize(48)
	err := processDir(config, defaultGapSize, m)
	if err != nil {
		log.Fatal(err)
	}
}
