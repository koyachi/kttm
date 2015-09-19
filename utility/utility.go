package utility

import (
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
)

type Config struct {
	SrcDir                  string
	ImageProcessorOutputDir string
	BinderOutputDir         string
}

func NewConfig() Config {
	return Config{
		SrcDir:                  "../data/src/",
		ImageProcessorOutputDir: "../data/image_processor_output/",
		BinderOutputDir:         "../data/binder_output",
	}
}

func DecodeImage(filePath string) (img image.Image, err error) {
	reader, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	img, _, err = image.Decode(reader)
	return
}

func SaveImage(img image.Image, path string) error {
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
