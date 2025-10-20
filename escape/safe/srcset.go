package safe

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	_ "unsafe" // for go:linkname
)

// Srcset represents a known safe (partial) srcset attribute, safe to be
// embedded between double quotes and used as a srcset attribute or part of a
// srcset attribute.
type Srcset struct{ val string }

// ConstantSrcset creates a new Srcset wrapper from the given string constant.
func ConstantSrcset(c constant) Srcset {
	return trustedSrcset(string(c))
}

type ImageCandidate struct {
	URL                    URL
	WidthDescriptor        int
	PixelDensityDescriptor float64
}

// FormatSrcset takes a list of image candidates and formats them into a safe
// Srcset.
//
// Either or no descriptor must be set for all candidates.
// FormatSrcset panics if both descriptors are set, or if any descriptor is set
// to a negative value, NaN, or Inf.
func FormatSrcset(candidates []ImageCandidate) Srcset {
	var b strings.Builder
	for i, c := range candidates {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString(c.URL.Get())

		switch {
		case c.WidthDescriptor < 0:
			panic(fmt.Sprintf("image candidate %d: negative width descriptor: %d", i, c.WidthDescriptor))
		case c.PixelDensityDescriptor < 0:
			panic(fmt.Sprintf("image candidate %d: negative pixel density descriptor: %g", i, c.PixelDensityDescriptor))
		case math.IsNaN(c.PixelDensityDescriptor):
			panic(fmt.Sprintf("image candidate %d: pixel density descriptor is NaN", i))
		case math.IsInf(c.PixelDensityDescriptor, +1):
			panic(fmt.Sprintf("image candidate %d: infinite pixel density descriptor", i))
		case c.WidthDescriptor > 0 && c.PixelDensityDescriptor > 0:
			panic(fmt.Sprintf("image candidate %d: both descriptors set", i))
		case c.WidthDescriptor > 0:
			b.WriteString(" ")
			b.WriteString(strconv.Itoa(c.WidthDescriptor))
			b.WriteByte('w')
		case c.PixelDensityDescriptor > 0:
			b.WriteString(" ")
			b.WriteString(strconv.FormatFloat(c.PixelDensityDescriptor, 'f', -1, 64))
			b.WriteByte('x')
		}
	}

	return trustedSrcset(b.String())
}

//go:linkname trustedSrcset
func trustedSrcset(s string) Srcset { return Srcset{val: s} }

func (s Srcset) Get() string { return s.val }
