package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
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

func generateIndexHtml(p []*Page) string {
	result := "<html><head><title>kttm index</title></head><body>\n"
	for _, v := range p {
		result += "<center>"
		// debug
		//result += "<p>" + strconv.Itoa(v.index) + "</p>"
		for _, i := range v.divImages {
			result += "<img src='" + i + "' /><br/>\n"
		}
		result += "</center>\n"
	}
	result += "</body></html>"
	return result
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
	fmt.Print(generateIndexHtml(pages))
}
