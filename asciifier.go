package main

import (
	"fmt"

	"errors"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/disintegration/imaging"
	"golang.org/x/image/draw"
)

// Pixel ...
type Pixel struct {
	Color       int
	Rune        rune
	Transparent bool
}

// Asciifier ...
type Asciifier struct {
	Options *Options
}

// NewAsciifier ...
func NewAsciifier(options *Options) *Asciifier {
	return &Asciifier{
		Options: options,
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
		prop := float64(src.Bounds().Dx()) / float64(a.Options.Width)
		width = a.Options.Width
		height = roundInt(float64(src.Bounds().Dy()) / prop)

		if !a.Options.HTML && !a.Options.Pixel {
			height /= 2
		}
	} else if a.Options.Height > 0 {
		// Calculate width from height.
		prop := float64(src.Bounds().Dy()) / float64(a.Options.Height)
		width = roundInt(float64(src.Bounds().Dx()) / prop)
		height = a.Options.Height

		if !a.Options.HTML && !a.Options.Pixel {
			width *= 2
		}
	} else if !a.Options.HTML {
		// Infer size from terminal.
		if a.Options.Pixel {
			// Scale to Width:Height*2.
			prop := float64(src.Bounds().Dy()) / float64(terminal.Height*2-2)
			width = roundInt(float64(src.Bounds().Dx()) / prop)
			height = terminal.Height*2 - 2

			// Fit width.
			if width > terminal.Width {
				prop = float64(src.Bounds().Dx()) / float64(terminal.Width-1)
				width = terminal.Width - 1
				height = roundInt(float64(src.Bounds().Dy()) / prop)
			}
		} else {
			// Scale to Width/2:Height.
			prop := float64(src.Bounds().Dy()) / float64(terminal.Height-1) / 2.0
			width = roundInt(float64(src.Bounds().Dx()) / prop)
			height = terminal.Height - 1

			// Fit width.
			if width > terminal.Width {
				prop = float64(src.Bounds().Dx()) / float64(terminal.Width-1) * 2.0
				width = terminal.Width - 1
				height = roundInt(float64(src.Bounds().Dy()) / prop)
			}
		}
	}

	// In pixel mode we need an even amount of rows.
	if a.Options.Pixel && height&1 == 1 {
		height++
	}

	// Scale the image.
	if width > 0 && height > 0 {
		dst := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.Draw(dst, dst.Bounds(), image.Transparent, image.ZP, draw.Src)
		draw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)
		src = dst
	}

	width, height = src.Bounds().Dx(), src.Bounds().Dy()

	// Allocate pixel buffer.
	pixels := make([]*Pixel, height*width)

	// Minimum opacity to consider a pixel fully transparent.
	minOpacity := uint32((1.0 - a.Options.TransparencyThreshold) * 0xFFFF)

	// Walk the image.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Get pixel.
			col := src.At(x, y)
			pixel := &Pixel{}

			// Grayscale it.
			r, g, b, aa := col.RGBA()
			v := uint16((float64(r)*a.Options.RedWeight +
				float64(g)*a.Options.GreenWeight +
				float64(b)*a.Options.BlueWeight))
			grayscale := &color.RGBA64{R: v, G: v, B: v, A: 0}

			// Find nearest grayscale terminal color to assign character.
			minDistance := math.Inf(0)
			idx := 0
			for i, c := range ColorsG {
				if distance := colorDistance(c, grayscale); distance < minDistance {
					minDistance = distance
					idx = i
				}
			}

			// Assign character.
			pixel.Rune = a.Options.Charset[idx%len(a.Options.Charset)]

			if a.Options.Transparent && aa <= minOpacity {
				// Pixel is transparent.
				pixel.Transparent = true
			} else if a.Options.Grayscale {
				// Assign color index from character index.
				if idx == 1 {
					idx = 15
				} else if idx != 0 {
					idx += 230
				}

				pixel.Color = idx
			} else {
				// Find nearest color from the 256-color map.
				minDistance := math.Inf(0)
				idx := 0

				for i, c := range ColorsT {
					if distance := colorDistance(c, col); distance < minDistance {
						minDistance = distance
						idx = i
					}
				}

				pixel.Color = idx
			}

			// Store the pixel.
			pixels[y*width+x] = pixel
		}
	}

	// Header.
	a.PrintHeader()

	// Print the buffer.
	for y := 0; y < height; y++ {
		a.BeginLine(termWidth, width)

		// We can only optimize colors on the same line.
		prev1 := &Pixel{Color: -1}
		prev2 := &Pixel{Color: -1}

		for x := 0; x < width; x++ {
			if a.Options.Pixel {
				// Pixel mode - box drawing characters, 2 lines at a time.
				current1 := pixels[y*width+x]
				current2 := pixels[y*width+width+x]
				a.PrintPixel(current1, current2, prev1, prev2)
				prev1 = current1
				prev2 = current2
			} else {
				current1 := pixels[y*width+x]
				a.PrintRune(current1, prev1)
				prev1 = current1
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
		fmt.Printf("<!-- im2a v%s -->\n", VERSION)
	}
}

// BeginLine ...
func (a *Asciifier) BeginLine(termWidth int, imageWidth int) {
	if a.Options.Center && !a.Options.HTML {
		fmt.Print(strings.Repeat(" ", (termWidth-imageWidth)/2))
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
func (a *Asciifier) PrintRune(current *Pixel, prev *Pixel) {
	if a.Options.HTML {
		idx := current.Color

		if a.Options.Grayscale {
			if idx == 1 {
				idx = 15
			} else if idx != 0 {
				idx -= 230
			}
		}

		if current.Transparent {
			fmt.Print(" ")
		} else {
			fmt.Printf("<span class=\"c_%d\">%c</span>", idx, current.Rune)
		}
	} else {
		if current.Transparent {
			if !prev.Transparent {
				fmt.Print("\x1b[49m")
			}

			fmt.Print(" ")
		} else {
			if current.Color != prev.Color {
				fmt.Printf("\x1b[38;5;%dm", current.Color)
			}

			fmt.Printf("%c", current.Rune)
		}
	}
}

// PrintPixel ...
func (a *Asciifier) PrintPixel(current1 *Pixel, current2 *Pixel, prev1 *Pixel, prev2 *Pixel) {
	if current1.Color != prev1.Color || current1.Transparent != prev1.Transparent {
		if current1.Transparent {
			fmt.Print("\x1b[49m")
		} else {
			fmt.Printf("\x1b[48;5;%dm", current1.Color)
		}
	}
	if current2.Color != prev2.Color || current2.Transparent != prev2.Transparent {
		if current2.Transparent {
			fmt.Print("\x1b[39m")
		} else {
			fmt.Printf("\x1b[38;5;%dm", current2.Color)
		}
	}

	if current1.Color == current2.Color || (current1.Transparent && current2.Transparent) {
		fmt.Print(" ")
	} else if current1.Transparent {
		fmt.Print("▀")
	} else {
		fmt.Print("▄")
	}
}
