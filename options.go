package main

import (
	"github.com/go2c/optparse"
)

// Options ...
type Options struct {
	Args					[]string
	Help					bool
	Version					bool
	Invert					bool
	Center					bool
	Grayscale				bool
	HTML					bool
	Pixel					bool
	Transparent				bool
	TransparencyThreshold	float64
	Width					int
	Height					int
	Charset					[]rune
	RedWeight				float64
	GreenWeight				float64
	BlueWeight				float64
}

// NewOptions ...
func NewOptions() *Options {
	return &Options{
		Args:					[]string{},
		Help:					false,
		Version:				false,
		Invert:					false,
		Center:					false,
		Grayscale:				false,
		HTML:					false,
		Pixel:					false,
		Transparent:			false,
		TransparencyThreshold:	1.0,
		Width:					0,
		Height:					0,
		Charset:				[]rune(" M   ...',;:clodxkO0KXNWMM"),
		RedWeight:				0.2989,
		GreenWeight:			0.5866,
		BlueWeight:				0.1145,
	}
}

// Parse ...
func (o *Options) Parse() error {
	charset := string(o.Charset)

	optparse.BoolVar(&o.Help, "help", 'h', o.Help)
	optparse.BoolVar(&o.Version, "version", 'v', o.Version)

	optparse.BoolVar(&o.Invert, "invert", 'i', o.Invert)
	optparse.BoolVar(&o.Center, "center", 't', o.Center)
	optparse.BoolVar(&o.Grayscale, "grayscale", 'g', o.Grayscale)
	optparse.BoolVar(&o.HTML, "html", 'm', o.HTML)
	optparse.BoolVar(&o.Pixel, "pixel", 'p', o.Pixel)

	optparse.BoolVar(&o.Transparent, "transparent", 'T', o.Transparent)
	optparse.FloatVar(&o.TransparencyThreshold, "transparency-threshold", 'X', o.TransparencyThreshold)

	optparse.IntVar(&o.Width, "width", 'W', o.Width)
	optparse.IntVar(&o.Height, "height", 'H', o.Height)

	optparse.StringVar(&charset, "charset", 'c', charset)

	optparse.FloatVar(&o.RedWeight, "float", 'R', o.RedWeight)
	optparse.FloatVar(&o.GreenWeight, "float", 'G', o.GreenWeight)
	optparse.FloatVar(&o.BlueWeight, "float", 'B', o.BlueWeight)

	args, err := optparse.Parse()
	if err != nil {
		return err
	}

	o.Charset = []rune(charset)
	o.Args = args

	return nil
}
