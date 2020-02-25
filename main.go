package main

import (
	"fmt"
	"io"
	"os"
)

// VERSION ...
const VERSION = "0.6.1"

// Usage ...
func usage(f io.Writer, name string) {
	fmt.Fprintf(f, "im2a %s\n", VERSION)
	fmt.Fprintln(f)

	fmt.Fprintln(f, "Convert image files to ASCII art")
	fmt.Fprintln(f, "Copyright (C) 2013 Latchezar Tzvetkoff")
	fmt.Fprintln(f, "Distributed under The Beerware License")
	fmt.Fprintln(f)

	fmt.Fprintln(f, "Usage:")
	fmt.Fprintf(f, "  %s [options] [arguments]\n", name)
	fmt.Fprintln(f)

	fmt.Fprintln(f, "Common options:")
	fmt.Fprintln(f, "  -h, --help                Print help and exit")
	fmt.Fprintln(f, "  -v, --version             Print version and exit")
	fmt.Fprintln(f)

	fmt.Fprintln(f, "Specific options:")
	fmt.Fprintln(f, "  -i, --invert                      Invert the image")
	fmt.Fprintln(f, "  -t, --center                      Center the image")
	fmt.Fprintln(f, "  -g, --grayscale                   Grayscale output")
	fmt.Fprintln(f, "  -m, --html                        HTML mode")
	fmt.Fprintln(f, "  -p, --pixel                       Pixel mode")
	fmt.Fprintln(f, "  -T, --transparent                 Enable transparency")
	fmt.Fprintln(f, "  -X, --transparency-threshold=X    Set transparency threshold (default: 1.0)")
	fmt.Fprintln(f, "  -W, --width=N                     Set output width")
	fmt.Fprintln(f, "  -H, --height=N                    Set output height")
	fmt.Fprintln(f, "  -c, --charset=C                   Set output charset")
	fmt.Fprintln(f, "  -R, --red-weight=RW               Set red component weight (default: 0.2989)")
	fmt.Fprintln(f, "  -G, --green-weight=GW             Set green component weight (default: 0.5866)")
	fmt.Fprintln(f, "  -B, --blue-weight=BW              Set blue component weight (default: 0.1145)")

	if f == os.Stderr {
		os.Exit(1) // revive:disable-line:deep-exit
	}

	os.Exit(0) // revive:disable-line:deep-exit
}

// Main ...
func main() {
	options := NewOptions()
	if err := options.Parse(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n\n", os.Args[0], err.Error())
		usage(os.Stderr, os.Args[0])
	}

	if options.Help {
		usage(os.Stdout, os.Args[0])
	}
	if options.Version {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	// Sanitize command-line options.
	if options.HTML && options.Pixel {
		fmt.Fprintf(os.Stderr,
			"%s: cannot use --html and --pixel at the same time\n\n",
			os.Args[0])
		usage(os.Stderr, os.Args[0])
	}

	if len(options.Args) != 1 {
		fmt.Fprintf(os.Stderr,
			"%s: wrong number of arguments (given %d, expected %d)\n\n",
			os.Args[0], len(options.Args), 1)
		usage(os.Stderr, os.Args[0])
	}

	// Create asfiifier.
	asciifier := NewAsciifier(options)
	if err := asciifier.Asciify(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n\n", os.Args[0], err.Error())
		os.Exit(0)
	}
}
