package safe

import (
	"math"
	"testing"

	"github.com/mavolin/corgi/v2/internal/should"
)

func TestConstantSrcset(t *testing.T) {
	t.Parallel()

	want := constant("image1x.jpg 1x, image2x.jpg 2x")
	got := ConstantSrcset(want).Get()
	should.Equal(t, got, string(want))
}

func TestFormatSrcset(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		want := "image1.jpg 1.4x,image2.jpg 3x"
		got := FormatSrcset([]ImageCandidate{
			{URL: ConstantURL("image1.jpg"), PixelDensityDescriptor: 1.4},
			{URL: ConstantURL("image2.jpg"), PixelDensityDescriptor: 3},
		})
		should.Equal(t, got.Get(), want)
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name       string
			candidates ImageCandidate
			wantPanic  string
		}{
			{
				name: "both descriptors",
				candidates: ImageCandidate{
					URL:                    ConstantURL("image1.jpg"),
					WidthDescriptor:        300,
					PixelDensityDescriptor: 2,
				},
				wantPanic: "image candidate 0: both descriptors set",
			}, {
				name: "negative width descriptor",
				candidates: ImageCandidate{
					URL:             ConstantURL("image1.jpg"),
					WidthDescriptor: -100,
				},
				wantPanic: "image candidate 0: negative width descriptor: -100",
			}, {
				name: "negative pixel density descriptor",
				candidates: ImageCandidate{
					URL:                    ConstantURL("image1.jpg"),
					PixelDensityDescriptor: -1.5,
				},
				wantPanic: "image candidate 0: negative pixel density descriptor: -1.5",
			}, {
				name: "-inf pixel density descriptor",
				candidates: ImageCandidate{
					URL:                    ConstantURL("image1.jpg"),
					PixelDensityDescriptor: math.Inf(-1),
				},
				wantPanic: "image candidate 0: negative pixel density descriptor: -Inf",
			}, {
				name: "+inf pixel density descriptor",
				candidates: ImageCandidate{
					URL:                    ConstantURL("image1.jpg"),
					PixelDensityDescriptor: math.Inf(+1),
				},
				wantPanic: "image candidate 0: infinite pixel density descriptor",
			}, {
				name: "NaN pixel density descriptor",
				candidates: ImageCandidate{
					URL:                    ConstantURL("image1.jpg"),
					PixelDensityDescriptor: math.NaN(),
				},
				wantPanic: "image candidate 0: pixel density descriptor is NaN",
			},
		}
		for _, c := range testCases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()
				should.Panic(t, func() {
					FormatSrcset([]ImageCandidate{c.candidates})
				}, c.wantPanic)
			})
		}
	})
}
