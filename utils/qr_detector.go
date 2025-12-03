package utils

import (
	"image"
	"log"
)

// DetectAndCropQR is a lightweight stub that simply returns
// the original image without using OpenCV.
// This keeps the rest of the code compiling cleanly.
func DetectAndCropQR(img image.Image) (image.Image, error) {
	log.Println("[Stub QR Detector] OpenCV disabled — returning original image")
	return img, nil
}
