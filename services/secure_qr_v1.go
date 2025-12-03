package services

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"math/big"
)

type SecureQRV1 struct {
	Name        string `json:"name"`
	DOB         string `json:"dob"`
	Gender      string `json:"gender"`
	ReferenceID string `json:"reference_id"`
	MobileHash  string `json:"mobile_hash"`
	EmailHash   string `json:"email_hash"`
}

func ParseSecureQRV1(raw []byte, _ interface{}) (*SecureQRV1, error) {

	// 1️⃣ Decimal → bytes
	bi := new(big.Int)
	if _, ok := bi.SetString(string(raw), 10); !ok {
		return nil, errors.New("invalid decimal QR")
	}
	compressed := bi.Bytes()

	// Must be GZIP
	if len(compressed) < 3 || compressed[0] != 0x1f || compressed[1] != 0x8b {
		return nil, errors.New("QR is not GZIP - not SecureQRV1")
	}

	// 2️⃣ GZIP decompress
	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	unzipped, err := io.ReadAll(gz)
	if err != nil {
		return nil, err
	}

	// 3️⃣ Split fields
	parts := bytes.Split(unzipped, []byte("\n"))
	if len(parts) < 6 {
		return nil, errors.New("invalid V1 text block")
	}

	return &SecureQRV1{
		Name:        string(parts[0]),
		DOB:         string(parts[1]),
		Gender:      string(parts[2]),
		ReferenceID: string(parts[3]),
		MobileHash:  string(parts[4]),
		EmailHash:   string(parts[5]),
	}, nil
}
