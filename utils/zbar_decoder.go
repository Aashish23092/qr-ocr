package utils

/*
#cgo CFLAGS: -IC:/msys64/mingw64/include
#cgo LDFLAGS: -LC:/msys64/mingw64/lib -lzbar

#include <stdlib.h>

// Declaration only — actual implementation lives in zbar_wrapper.c
int decode_qr_zbar(unsigned char *gray_data, int width, int height,
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

func DecodeWithZBar(img image.Image) ([]byte, error) {
	log.Println("[ZBar] STEP A: Starting ZBar decode")

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	log.Printf("[ZBar] STEP B: Image size %dx%d\n", width, height)

	grayData := make([]byte, width*height)
	i := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			grayData[i] = gray.Y
			i++
		}
	}

	output := make([]byte, 10240)
	var outLen C.int

	result := C.decode_qr_zbar(
		(*C.uchar)(unsafe.Pointer(&grayData[0])),
		C.int(width),
		C.int(height),
		(*C.uchar)(unsafe.Pointer(&output[0])),
		&outLen,
	)

	if result != 0 {
		return nil, fmt.Errorf("zbar decode failed: code %d", result)
	}

	log.Printf("[ZBar] STEP D SUCCESS: Decoded %d bytes\n", outLen)
	return output[:outLen], nil
}
