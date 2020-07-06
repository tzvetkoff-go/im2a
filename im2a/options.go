package im2a

import (
	"fmt"

	"github.com/tzvetkoff-go/optparse"
)

// Options ...
type Options struct {
	Image                 string
	Help                  bool
	Version               bool
	Invert                bool
	Center                bool
	Grayscale             bool
	HTML                  bool
	Pixel                 bool
	Transparent           bool
	TransparencyThreshold float64
	Width                 int
	Height                int
	Charset               []rune
	RedWeight             float64
	GreenWeight           float64
	BlueWeight            float64
}

// NewOptions ...
func NewOptions() *Options {
	return &Options{
		Image:                 "",
		Help:                  false,
		Version:               false,
		Invert:                false,
		Center:                false,
		Grayscale:             false,
		HTML:                  false,
		Pixel:                 false,
		Transparent:           false,
		TransparencyThreshold: 1.0,
		Width:                 0,
		Height:                0,
		Charset:               []rune(" M   ...',;:clodxkO0KXNWMM"),
		RedWeight:             0.2989,
		GreenWeight:           0.5866,
		BlueWeight:            0.1145,
	}
}

// ParseCommandLine ...
func (o *Options) ParseCommandLine(args []string) error {
	charset := string(o.Charset)

	parser := optparse.New()
	parser.BoolVar(&o.Help, "help", 'h', o.Help)
	parser.BoolVar(&o.Version, "version", 'v', o.Version)

	parser.BoolVar(&o.Invert, "invert", 'i', o.Invert)
	parser.BoolVar(&o.Center, "center", 't', o.Center)
	parser.BoolVar(&o.Grayscale, "grayscale", 'g', o.Grayscale)
	parser.BoolVar(&o.HTML, "html", 'm', o.HTML)
	parser.BoolVar(&o.Pixel, "pixel", 'p', o.Pixel)

	parser.BoolVar(&o.Transparent, "transparent", 'T', o.Transparent)
	parser.FloatVar(&o.TransparencyThreshold, "transparency-threshold", 'X', o.TransparencyThreshold)

	parser.IntVar(&o.Width, "width", 'W', o.Width)
	parser.IntVar(&o.Height, "height", 'H', o.Height)

	parser.StringVar(&charset, "charset", 'c', charset)

	parser.FloatVar(&o.RedWeight, "float", 'R', o.RedWeight)
	parser.FloatVar(&o.GreenWeight, "float", 'G', o.GreenWeight)
	parser.FloatVar(&o.BlueWeight, "float", 'B', o.BlueWeight)

	args, err := parser.Parse(args)
	if err != nil {
		return err
	}

	if len(args) != 1 {
		return fmt.Errorf("wrong number of arguments (given %d, expected %d)", len(args), 1)
	}

	o.Charset = []rune(charset)
	fmt.Println(args)
	o.Image = args[0]

	return nil
}
