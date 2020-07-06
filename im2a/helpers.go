package im2a

import (
	"image/color"
	"math"
)

func roundInt(f float64) int {
	i := int(f * 10)

	if i%10 > 4 {
		return i/10 + 1
	}

	return i / 10
}

func colorDistance(c1 color.Color, c2 color.Color) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	return math.Sqrt(math.Pow(float64(r1)-float64(r2), 2) +
		math.Pow(float64(g1)-float64(g2), 2) +
		math.Pow(float64(b1)-float64(b2), 2))
}
