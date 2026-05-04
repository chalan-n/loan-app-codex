package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
)

type PreprocessResult struct {
	Data     []byte
	MIMEType string
}

// PreprocessIDCard keeps the old public API while using the faster OCR pipeline.
func PreprocessIDCard(imageBytes []byte, mimeType string) (*PreprocessResult, error) {
	return PreprocessOCRImage(imageBytes, 1600, 1.12, 88)
}

// PreprocessOCRImage resizes large mobile photos and lightly boosts contrast.
// Returning JPEG keeps Gemini requests smaller and faster than PNG.
func PreprocessOCRImage(imageBytes []byte, maxDimension int, contrast float64, jpegQuality int) (*PreprocessResult, error) {
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, fmt.Errorf("preprocess: decode image: %w", err)
	}

	resized := resizeImage(img, maxDimension)
	bounds := resized.Bounds()
	processed := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			px := resized.At(bounds.Min.X+x, bounds.Min.Y+y)
			r, g, b, a := px.RGBA()
			processed.SetRGBA(x, y, color.RGBA{
				R: uint8(adjustContrast(float64(r>>8), contrast)),
				G: uint8(adjustContrast(float64(g>>8), contrast)),
				B: uint8(adjustContrast(float64(b>>8), contrast)),
				A: uint8(a >> 8),
			})
		}
	}

	if jpegQuality <= 0 || jpegQuality > 100 {
		jpegQuality = 88
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, processed, &jpeg.Options{Quality: jpegQuality}); err != nil {
		return nil, fmt.Errorf("preprocess: encode jpeg: %w", err)
	}
	return &PreprocessResult{Data: buf.Bytes(), MIMEType: "image/jpeg"}, nil
}

func resizeImage(src image.Image, maxDimension int) image.Image {
	if maxDimension <= 0 {
		return src
	}
	bounds := src.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()
	if srcW <= maxDimension && srcH <= maxDimension {
		return src
	}

	scale := float64(maxDimension) / float64(srcW)
	if srcH > srcW {
		scale = float64(maxDimension) / float64(srcH)
	}
	dstW := max(1, int(float64(srcW)*scale))
	dstH := max(1, int(float64(srcH)*scale))
	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))

	for y := 0; y < dstH; y++ {
		srcY := bounds.Min.Y + int(float64(y)*float64(srcH)/float64(dstH))
		for x := 0; x < dstW; x++ {
			srcX := bounds.Min.X + int(float64(x)*float64(srcW)/float64(dstW))
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}

func adjustContrast(v, factor float64) float64 {
	if factor <= 0 {
		factor = 1
	}
	return clampF((v-128)*factor+128, 0, 255)
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
