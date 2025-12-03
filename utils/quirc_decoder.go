package utils

/*
#cgo CFLAGS: -IC:/ProgramData/mingw64/mingw64/include
#cgo LDFLAGS: -LC:/ProgramData/mingw64/mingw64/lib -lquirc

#include <stdlib.h>

// Declaration only — REAL implementation is in quirc_wrapper.c
int decode_qr_quirc(unsigned char *gray_data, int width, int height,
                    unsigned char *output, int *output_len);
*/
import "C"
import (
	"fmt"
	"image"
	"image/color"
	"log"
	"unsafe"
)

// DecodeWithQuirc decodes a QR using native C Quirc.
func DecodeWithQuirc(img image.Image) ([]byte, error) {
	log.Println("[quirc] STEP A: Starting quirc decode")

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	log.Printf("[quirc] STEP B: Image size %dx%d\n", width, height)

	// Convert to grayscale buffer (quirc expects 1 byte per pixel)
	grayData := make([]byte, width*height)
	i := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			grayData[i] = gray.Y
			i++
		}
	}

	log.Println("[quirc] STEP C: Converted to grayscale")

	// Output buffer (QR payload <= 8896 bytes max)
	output := make([]byte, 10*1024)
	var outLen C.int

	result := C.decode_qr_quirc(
		(*C.uchar)(unsafe.Pointer(&grayData[0])),
		C.int(width),
		C.int(height),
		(*C.uchar)(unsafe.Pointer(&output[0])),
		&outLen,
	)

	if result != 0 {
		return nil, fmt.Errorf("quirc decode failed: code %d", result)
	}

	log.Printf("[quirc] STEP D SUCCESS: Decoded %d bytes\n", outLen)
	return output[:outLen], nil
}
