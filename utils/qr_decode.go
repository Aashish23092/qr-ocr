package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

func DecodeQR(imgBytes []byte) ([]byte, error) {
	// Decode PNG or JPEG
	img, err := decodeImage(imgBytes)
	if err != nil {
		return nil, fmt.Errorf("image decode error: %v", err)
	}

	// Convert Go image → LuminanceSource
	source := gozxing.NewLuminanceSourceFromImage(img)

	// Create HybridBinarizer (OK, returns 1 value)
	binarizer := gozxing.NewHybridBinarizer(source)

	// Create BinaryBitmap (returns 2 values)
	bitmap, err := gozxing.NewBinaryBitmap(binarizer)
	if err != nil {
		return nil, fmt.Errorf("binary bitmap error: %v", err)
	}

	// QR reader
	reader := qrcode.NewQRCodeReader()

	// Decode QR
	result, err := reader.Decode(bitmap, nil)
	if err != nil {
		return nil, fmt.Errorf("QR decode error: %v", err)
	}

	return []byte(result.GetText()), nil
}

func decodeImage(b []byte) (image.Image, error) {
	if img, err := png.Decode(bytes.NewReader(b)); err == nil {
		return img, nil
	}
	if img, err := jpeg.Decode(bytes.NewReader(b)); err == nil {
		return img, nil
	}
	return nil, fmt.Errorf("unsupported image format")
}
