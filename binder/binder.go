package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

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

func generateIndexHtml(u []*UrlIndex) string {
	result := "<html><head><title>kttm index</title></head><body>\n"
	for _, v := range u {
		result += "<div>"
		for _, i := range v.images() {
			result += "<img src='" + i + "' />\n"
		}
		result += "</div>\n"
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
	fmt.Print(generateIndexHtml(urlIndexes))
}
