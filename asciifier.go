package main

import (
	"fmt"

	"io"
	"os"
	"errors"
	"strings"
	"net/http"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/png"
	_ "image/jpeg"
	"golang.org/x/image/draw"
	"github.com/disintegration/imaging"
)

// Asciifier ...
type Asciifier struct {
	Options			*Options
}

// NewAsciifier ...
func NewAsciifier(options *Options) *Asciifier {
	return &Asciifier{
		Options:		options,
	}
}

// Asciify ...
func (a *Asciifier) Asciify() error {
	var file io.ReadCloser
	if strings.HasPrefix(a.Options.Args[0], "http://") || strings.HasPrefix(a.Options.Args[0], "https://") {
		// Open URL.
		response, err := http.Get(a.Options.Args[0])
		if err != nil {
			return err
		}
		file = response.Body
	} else {
		// Open local file.
		f, err := os.Open(a.Options.Args[0])
		if err != nil {
			return err
		}
		file = f
	}
	defer file.Close()

	// Load image.
	src, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	// Invert?
	if a.Options.Invert {
		src = imaging.Invert(src)
	}

	// We don't need the terminal in all cases.
	var terminal *Terminal
	termWidth := 0
	if a.Options.Pixel || (!a.Options.HTML && a.Options.Width == 0 && a.Options.Height == 0) {
		terminal = NewTerminal()
		if terminal.Width == 0 || terminal.Height == 0 {
			return errors.New("Cannot determine terminal size")
		}

		termWidth = terminal.Width
	}

	// Determine how to scale the image.
	width, height := 0, 0
	if a.Options.Width > 0 && a.Options.Height > 0 {
		// Use provided values
		width, height = a.Options.Width, a.Options.Height
	} else if a.Options.Width > 0 {
		// Calculate height from width.
	} else if a.Options.Height > 0 {
		// Calculate width from height.
	} else if !a.Options.HTML {
		// Infer size from terminal.
		if a.Options.Pixel {
			// Scale to Width:Height*2.
			prop := float64(src.Bounds().Dy()) / float64(terminal.Height * 2 - 2)
			width = roundInt(float64(src.Bounds().Dx()) / prop)
			height = terminal.Height * 2 - 2

			// Fit width.
			if width > terminal.Width {
				prop = float64(src.Bounds().Dx()) / float64(terminal.Width - 1)
				width = terminal.Width - 1
				height = roundInt(float64(src.Bounds().Dy()) / prop)
			}
		} else {
			// Scale to Width/2:Height.
			prop := float64(src.Bounds().Dy()) / float64(terminal.Height - 1) / 2.0
			width = roundInt(float64(src.Bounds().Dx()) / prop)
			height = terminal.Height - 1

			// Fit width.
			if width > terminal.Width {
				prop = float64(src.Bounds().Dx()) / float64(terminal.Width - 1) * 2.0
				width = terminal.Width - 1
				height = roundInt(float64(src.Bounds().Dy()) / prop)
			}
		}
	}

	// Scale the image.
	if width > 0 && height > 0 {
		dst := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.Draw(dst, dst.Bounds(), image.Transparent, image.ZP, draw.Src)
		draw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)
		src = dst
	}

	width, height = src.Bounds().Dx(), src.Bounds().Dy()

	// Allocate runes & pixel buffers.
	runes := make([]rune, height * width)
	colors := make([]uint8, height * width)

	// Walk the image.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Get pixel.
			col := src.At(x, y)

			// Grayscale it.
			r, g, b, _ := col.RGBA()
			v := uint16((float64(r) * a.Options.RedWeight +
				float64(g) * a.Options.GreenWeight +
				float64(b) * a.Options.BlueWeight))
			grayscale := &color.RGBA64{R: v, G: v, B: v, A: 0}

			// Find nearest grayscale terminal color to assign character.
			minDistance := float64(0xFFFFFFFF)
			idx := 0
			for i, c := range ColorsG {
				if distance := colorDistance(c, grayscale); distance < minDistance {
					minDistance = distance
					idx = i
				}
			}

			// Assign character.
			runes[y * width + x] = a.Options.Charset[idx % len(a.Options.Charset)]

			if a.Options.Grayscale {
				// Assign color index as well.
				if idx == 1 {
					idx = 15
				} else if idx != 0 {
					idx += 230
				}

				colors[y * width + x] = uint8(idx)
			} else {
				// Find nearest color from 256-color.
				minDistance := float64(0xFFFFFFFF)
				idx := 0

				for i, c := range ColorsT {
					if distance := colorDistance(c, col); distance < minDistance {
						minDistance = distance
						idx = i
					}
				}

				colors[y * width + x] = uint8(idx)
			}
		}
	}

	// Header.
	a.PrintHeader()

	// Print the buffer.
	for y := 0; y < height; y++ {
		a.BeginLine(termWidth, width)

		for x := 0; x < width; x++ {
			if a.Options.Pixel {
				// Pixel mode - box drawing characters, 2 lines at a time.
				idx1 := colors[y * width + x]
				idx2 := colors[y * width + width + x]
				a.PrintPixel(idx1, idx2)
			} else {
				idx := colors[y * width + x]
				r := runes[y * width + x]
				a.PrintRune(idx, r)
			}
		}

		a.EndLine()

		if a.Options.Pixel {
			y++
		}
	}

	// Footer.
	a.PrintFooter()

	return nil
}

// PrintHeader ...
func (a *Asciifier) PrintHeader() {
	if a.Options.HTML {
		fmt.Println("<!DOCTYPE html>")
		fmt.Println("<html>")
		fmt.Println("<head>")
		fmt.Println("  <meta http-equiv=\"Content-Type\" content=\"text/html; charset=UTF-8\" />")
		fmt.Println("  <title>im2a asciified image</title>")
		fmt.Println("  <style type=\"text/css\">")
		fmt.Println("    body { background: #000000; }")
		fmt.Println("    pre { font: normal 12px/9px Menlo, monospace; }")
		if a.Options.Center {
			fmt.Println("    pre { text-aling: center; }")
		}
		if a.Options.Grayscale {
			for idx, color := range ColorsGG {
				fmt.Printf("    .c_%d { color: #%06x }\n", idx, color)
			}
		} else {
			for idx, color := range ColorsTT {
				fmt.Printf("    .c_%d { color: #%06x }\n", idx, color)
			}
		}
		fmt.Println("  </style>")
		fmt.Println("</head>")
		fmt.Println("<body>")
		fmt.Println("<pre>")
	}
}

// PrintFooter ...
func (a *Asciifier) PrintFooter() {
	if a.Options.HTML {
		fmt.Println("</pre>")
		fmt.Println("</body>")
		fmt.Println("</html>")
	}
}

// BeginLine ...
func (a *Asciifier) BeginLine(termWidth int, imageWidth int) {
	if a.Options.Center && !a.Options.HTML {
		fmt.Print(strings.Repeat(" ", (termWidth - imageWidth) / 2))
	}
}

// EndLine ...
func (a *Asciifier) EndLine() {
	if a.Options.HTML {
		fmt.Println("")
	} else {
		fmt.Println("\x1b[0;0m")
	}
}

// PrintRune ...
func (a *Asciifier) PrintRune(idx uint8, r rune) {
	if a.Options.HTML {
		if a.Options.Grayscale {
			if idx == 1 {
				idx = 15
			} else if idx != 0 {
				idx -= 230
			}
		}

		fmt.Printf("<span class=\"c_%d\">%c</span>", idx, r)
	} else {
		fmt.Printf("\x1b[38;5;%dm%c", idx, r)
	}
}

// PrintPixel ...
func (a *Asciifier) PrintPixel(idx1 uint8, idx2 uint8) {
	fmt.Printf("\x1b[48;5;%dm\x1b[38;5;%dm▄", idx1, idx2)
}
