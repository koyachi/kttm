package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// TODO:move to utility package
func decodeImage(filePath string) (img image.Image, err error) {
	reader, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	img, _, err = image.Decode(reader)
	return
}

// TODO:move to utility package
func saveImage(img image.Image, path string) error {
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		if os.IsNotExist(err) {
			dir, _ := filepath.Split(path)
			err = os.Mkdir(dir, os.FileMode(0755))
			if err != nil {
				return err
			}
			file, err = os.Create(path)
		} else {
			return err
		}
	}
	err = jpeg.Encode(file, img, nil)
	if err != nil {
		return err
	}
	return nil
}

type UrlIndex struct {
	index     int
	url       string
	divImages []string
}

func (u *UrlIndex) images() []string {
	if len(u.divImages) > 0 {
		return u.divImages
	} else {
		_, fileName := filepath.Split(u.url)
		return []string{"../tmp/" + fileName}
	}
}

// for sort
type ByIndex []*UrlIndex

func (b ByIndex) Len() int {
	return len(b)
}
func (b ByIndex) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
func (b ByIndex) Less(i, j int) bool {
	return b[i].index < b[j].index
}

type Page struct {
	index     int
	divImages []string
}

func (p *Page) concatImages() error {
	width := 0
	height := 0
	for _, i := range p.divImages {
		img, err := decodeImage(i)
		if err != nil {
			return err
		}
		height += img.Bounds().Size().Y
		if width < img.Bounds().Size().X {
			width = img.Bounds().Size().X
		}
	}
	dstRect := image.Rect(0, 0, width, height)
	dstImage := image.NewRGBA(dstRect)
	top := 0
	for _, i := range p.divImages {
		img, err := decodeImage(i)
		if err != nil {
			return err
		}
		srcPoint := image.Point{0, 0}
		dstRect := image.Rect(0, top, img.Bounds().Size().X, top+img.Bounds().Size().Y)
		draw.Draw(dstImage, dstRect, img, srcPoint, draw.Src)
		top += img.Bounds().Size().Y
	}
	return saveImage(dstImage, p.imagePath())
}

func (p *Page) imagePath() string {
	return "../binder_output/page_" + strconv.Itoa(p.index) + ".jpg"
}

func parseJson(path string) (u []*UrlIndex, err error) {
	reader, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var m map[string]string
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	var urlIndexes []*UrlIndex
	for k, v := range m {
		index, err := strconv.Atoi(k)
		if err != nil {
			return nil, err
		}
		urlIndexes = append(urlIndexes, &UrlIndex{index, v, nil})
	}
	sort.Sort(ByIndex(urlIndexes))
	return urlIndexes, nil
}

func searchDivImages(urlIndexes []*UrlIndex) error {
	inputDir := "../image_processor_output/"
	for _, u := range urlIndexes {
		var divImages []string
		_, fileName := filepath.Split(u.url)
		for i := 0; i < 4; i++ {
			divFilePath := inputDir + fileName + ".div_" + strconv.Itoa(i) + ".jpg"
			f, err := os.Open(divFilePath)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				} else {
					return err
				}
			}
			f.Close()
			divImages = append(divImages, divFilePath)
		}
		u.divImages = divImages
	}
	return nil
}

func collectPages(urlIndexes []*UrlIndex) (pages []*Page, err error) {
	pageIndex := 0
	var tmpImages []string
	for _, u := range urlIndexes {
		for _, imgPath := range u.images() {
			//fmt.Printf("imgPath = %s\n", imgPath)
			img, err := decodeImage(imgPath)
			if err != nil {
				return nil, err
			}
			height := img.Bounds().Size().Y
			if math.Abs(float64(1600-height)) > 100 {
				tmpImages = append(tmpImages, imgPath)
				if len(tmpImages) == 2 {
					pages = append(pages, &Page{pageIndex, tmpImages})
					pageIndex++
				}
			} else {
				pages = append(pages, &Page{pageIndex, []string{imgPath}})
				pageIndex++
				tmpImages = []string{}
			}
		}
	}
	return pages, nil
}

func generateIndexHtml(p []*Page) (string, error) {
	result := "<html><head><title>kttm index</title></head><body>\n"
	for _, v := range p {
		result += "<center>"
		// debug
		//result += "<p>" + strconv.Itoa(v.index) + "</p>"
		err := v.concatImages()
		if err != nil {
			return "", err
		}
		result += "<img src='" + v.imagePath() + "' /><br/>\n"
		result += "</center>\n"
	}
	result += "</body></html>"
	return result, nil
}

func main() {
	urlIndexes, err := parseJson("../tmp/urlIndex.json")
	if err != nil {
		log.Fatal(err)
	}
	//	fmt.Printf("urlIndexes = %v", urlIndexes)
	err = searchDivImages(urlIndexes)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("urlIndexes = %v", urlIndexes)
	pages, err := collectPages(urlIndexes)
	if err != nil {
		log.Fatal(err)
	}
	pageHtml, err := generateIndexHtml(pages)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(pageHtml)
}
