package ffmpeg

import (
	"bytes"
	"image"
	"image/png"
)

func BufToPng(buf []byte, width int, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// RGB → RGBA 변환
	j := 0
	for i := 0; i < len(buf); i += 3 {
		img.Pix[j+0] = buf[i+2] // R
		img.Pix[j+1] = buf[i+1] // G
		img.Pix[j+2] = buf[i+0] // B
		img.Pix[j+3] = 255
		j += 4
	}

	return img
}

func EncodePng(img *image.RGBA) ([]byte, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
