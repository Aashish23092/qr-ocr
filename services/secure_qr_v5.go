package services

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"math/big"
)

type SecureQRV5 struct {
	Version      string `json:"version"`
	ReferenceID  string `json:"reference_id"`
	Name         string `json:"name"`
	DOB          string `json:"dob"`
	Gender       string `json:"gender"`
	CareOf       string `json:"care_of"`
	District     string `json:"district"`
	PostOffice   string `json:"post_office"`
	Pincode      string `json:"pincode"`
	VTC          string `json:"vtc"`
	State        string `json:"state"`
	SubDistrict  string `json:"sub_district"`
	Location     string `json:"location"`
	MaskedMobile string `json:"masked_mobile"`
	MaskedEmail  string `json:"masked_email,omitempty"`
	RawText      string `json:"raw_text"`
}

func ParseSecureQRV5(raw []byte) (*SecureQRV5, error) {

	// 1️⃣ Convert decimal → big.Int → bytes
	bi := new(big.Int)
	if _, ok := bi.SetString(string(raw), 10); !ok {
		return nil, errors.New("V5: input is not decimal")
	}
	zipped := bi.Bytes()

	// Must begin with GZIP magic
	if len(zipped) < 2 || zipped[0] != 0x1f || zipped[1] != 0x8b {
		return nil, errors.New("V5: not gzip data")
	}

	// 2️⃣ GZIP decompress
	gz, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("V5: gunzip error: %w", err)
	}
	defer gz.Close()

	unzipped, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("V5: read error: %w", err)
	}

	// 3️⃣ Split by 0xFF (UIDAI field delimiter)
	parts := bytes.Split(unzipped, []byte{0xFF})

	// Clean empty entries (UIDAI often leaves empty fields)
	clean := make([]string, 0)
	for _, p := range parts {
		clean = append(clean, string(p))
	}

	// Must have at least 18 meaningful fields
	if len(clean) < 18 {
		return nil, fmt.Errorf("V5: insufficient fields (%d)", len(clean))
	}

	// 4️⃣ Map fields based on official UIDAI V5 layout
	model := &SecureQRV5{
		Version:      clean[0],
		ReferenceID:  safe(clean, 2),
		Name:         safe(clean, 3),
		DOB:          safe(clean, 4),
		Gender:       safe(clean, 5),
		CareOf:       safe(clean, 6),
		District:     safe(clean, 7),
		PostOffice:   safe(clean, 9),
		Pincode:      safe(clean, 11),
		VTC:          safe(clean, 12),
		State:        safe(clean, 13),
		SubDistrict:  safe(clean, 14),
		Location:     safe(clean, 16),
		MaskedMobile: safe(clean, 17),
		RawText:      string(unzipped),
	}

	// 5️⃣ Optional masked email (not always present)
	if len(clean) > 18 {
		model.MaskedEmail = clean[18]
	}

	return model, nil
}

// protects from index out of range
func safe(arr []string, idx int) string {
	if idx >= len(arr) {
		return ""
	}
	return arr[idx]
}
