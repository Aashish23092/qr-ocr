package handlers

import (
	"bytes"
	"crypto/rsa"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"math/big"
	"net/http"

	"github.com/Aashish23092/aadhaar-qr-service/services"
	"github.com/Aashish23092/aadhaar-qr-service/utils"
	"github.com/gin-gonic/gin"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

type QRHandler struct {
	PublicKey *rsa.PublicKey
}

func NewQRHandler(pub *rsa.PublicKey) *QRHandler {
	return &QRHandler{PublicKey: pub}
}

func decodeImage(fileBytes []byte) (image.Image, error) {
	if img, err := png.Decode(bytes.NewReader(fileBytes)); err == nil {
		return img, nil
	}
	if img, err := jpeg.Decode(bytes.NewReader(fileBytes)); err == nil {
		return img, nil
	}
	return nil, fmt.Errorf("unsupported image format")
}

func decodeQRZXing(img image.Image) ([]byte, error) {
	log.Println("[ZX] STEP A: Starting decodeQRZXing")
	log.Printf("[ZX] Image bounds → %v\n", img.Bounds())

	// 1. Create luminance source
	source := gozxing.NewLuminanceSourceFromImage(img)
	if source == nil {
		log.Println("[ZX] ERROR: Luminance source is nil")
		return nil, fmt.Errorf("luminance source is nil")
	}
	log.Printf("[ZX] STEP B: Luminance source created (%dx%d)\n",
		source.GetWidth(), source.GetHeight())

	// 2. Binarizer
	binarizer := gozxing.NewHybridBinarizer(source)
	if binarizer == nil {
		log.Println("[ZX] ERROR: HybridBinarizer returned nil")
		return nil, fmt.Errorf("hybrid binarizer nil")
	}
	log.Println("[ZX] STEP C: Hybrid binarizer OK")

	// 3. Binary bitmap
	bmp, err := gozxing.NewBinaryBitmap(binarizer)
	if err != nil {
		log.Printf("[ZX] STEP D ERROR: Binary bitmap creation failed → %v\n", err)
		return nil, fmt.Errorf("binary bitmap error: %v", err)
	}
	log.Println("[ZX] STEP D: Binary bitmap created")

	// 4. ZXing QR decode
	reader := qrcode.NewQRCodeReader()
	log.Println("[ZX] STEP E: Attempting ZXing decode…")

	result, err := reader.Decode(bmp, nil)
	if err != nil {
		log.Printf("[ZX] STEP F ERROR: ZXing decode failed → %v\n", err)
		return nil, fmt.Errorf("QR decode error: %v", err)
	}

	if result == nil {
		log.Println("[ZX] STEP F ERROR: ZXing returned nil result")
		return nil, fmt.Errorf("nil result from reader.Decode()")
	}

	text := result.GetText()
	log.Printf("[ZX] STEP G: ZXing decode SUCCESS, length=%d\n", len(text))

	return []byte(text), nil
}

func (h *QRHandler) Decode(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Println("STEP 1 ERROR: No file in request:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "file missing"})
		return
	}
	log.Println("STEP 1: File received successfully")
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Println("STEP 2 ERROR: Unable to read file bytes:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read file"})
		return
	}
	log.Println("STEP 2: File read OK, size:", len(fileBytes))

	img, err := decodeImage(fileBytes)
	if err != nil {
		log.Println("STEP 3 ERROR: decodeImage failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid image"})
		return
	}
	log.Println("STEP 3: Image decoded successfully")

	// ========================================================
	// STEP 4: Multi-stage QR Decoding Pipeline
	// ========================================================
	var qrBytes []byte
	var decodeErr error

	// Stage 1: OpenCV QR Detection & Cropping
	log.Println("STEP 4A: Attempting OpenCV QR detection and cropping...")
	croppedImg, detectErr := utils.DetectAndCropQR(img)
	if detectErr != nil {
		log.Printf("STEP 4A WARNING: OpenCV detection failed: %v, using original image\n", detectErr)
		croppedImg = img // Fallback to original
	} else {
		log.Println("STEP 4A SUCCESS: QR detected and cropped")
	}

	// Stage 2: Try quirc decoder first
	log.Println("STEP 4B: Attempting quirc decode...")
	qrBytes, decodeErr = utils.DecodeWithQuirc(croppedImg)
	if decodeErr == nil && len(qrBytes) > 0 {
		log.Printf("STEP 4B SUCCESS: quirc decoded %d bytes\n", len(qrBytes))
		goto DECODE_SUCCESS
	}
	log.Printf("STEP 4B FAILED: quirc decode error: %v\n", decodeErr)

	// Stage 3: Try ZBar decoder
	log.Println("STEP 4C: Attempting ZBar decode...")
	qrBytes, decodeErr = utils.DecodeWithZBar(croppedImg)
	if decodeErr == nil && len(qrBytes) > 0 {
		log.Printf("STEP 4C SUCCESS: ZBar decoded %d bytes\n", len(qrBytes))
		goto DECODE_SUCCESS
	}
	log.Printf("STEP 4C FAILED: ZBar decode error: %v\n", decodeErr)

	// Stage 4: Final fallback to ZXing
	log.Println("STEP 4D: Attempting ZXing decode (final fallback)...")
	qrBytes, decodeErr = decodeQRZXing(croppedImg)
	if decodeErr != nil {
		log.Println("STEP 4D ERROR: All decoders failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "QR not detected by any decoder"})
		return
	}
	log.Printf("STEP 4D SUCCESS: ZXing decoded %d bytes\n", len(qrBytes))

DECODE_SUCCESS:
	log.Println("STEP 4: QR decoded successfully, byte-length:", len(qrBytes))

	fmt.Println("QR bytes extracted:", len(qrBytes))

	//---------------------------------------------------------
	// 1️⃣ Try Secure QR v2 first
	//---------------------------------------------------------
	log.Println("STEP 5: Attempting ParseSecureQR (v2)...")
	if secureQR, err := services.ParseSecureQR(qrBytes, h.PublicKey); err == nil {
		log.Println("STEP 5 SUCCESS: Secure QR v2 decoded")
		c.JSON(http.StatusOK, gin.H{
			"type": "secure_qr_v2",
			"data": secureQR,
		})
		return
	} else {
		log.Println("STEP 5 FAILED: v2 parse error:", err)
	}

	//---------------------------------------------------------
	// 2️⃣ Check Secure QR v1
	//---------------------------------------------------------
	log.Println("STEP 6: Checking if QR is numeric:", isNumeric(qrBytes), "length:", len(qrBytes))
	//---------------------------------------------------------
	// 3️⃣ Check Secure QR v5
	//---------------------------------------------------------
	// 2️⃣ Secure QR v5 / v1 → numeric payload
	if isNumeric(qrBytes) && len(qrBytes) > 500 {

		// First try V5 (gzip + V5...)
		log.Println("STEP 7: Attempting ParseSecureQRV5 (v5)...")
		if v5, err := services.ParseSecureQRV5(qrBytes); err == nil {
			log.Println("STEP 7 SUCCESS: Secure QR v5 decoded")
			c.JSON(http.StatusOK, gin.H{
				"type": "secure_qr_v5",
				"data": v5,
			})
			return
		}
		log.Println("STEP 7 FAILED: v5 parse error")

		// Fallback: try old V1 (if you still need it)
		if v1, err := services.ParseSecureQRV1(qrBytes, nil); err == nil {
			c.JSON(http.StatusOK, gin.H{
				"type": "secure_qr_v1",
				"data": v1,
			})
			return
		}

		// If both fail
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "secure_qr_vx decode failed",
		})
		return
	}
	//---------------------------------------------------------
	// 3️⃣ Old QR (plain text)
	//---------------------------------------------------------
	if len(qrBytes) < 500 {
		c.JSON(http.StatusOK, gin.H{
			"type":     "old_qr",
			"raw_text": string(qrBytes),
		})
		return
	}

	//---------------------------------------------------------
	// 4️⃣ Unknown format
	//---------------------------------------------------------
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "unrecognized Aadhaar QR format",
	})

	// Debug fallback
	bi := new(big.Int)
	bi.SetString(string(qrBytes), 10)
	b := bi.Bytes()

	limit := 32
	if len(b) < limit {
		limit = len(b)
	}

	fmt.Printf("HEX: %x\n", b[:limit])
	fmt.Printf("ASCII: %s\n", b[:limit])
}

func isNumeric(b []byte) bool {
	for _, ch := range b {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
