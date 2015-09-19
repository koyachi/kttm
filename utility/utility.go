package utility

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
