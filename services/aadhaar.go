package services

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

type AadhaarOfflineKyc struct {
	XMLName     xml.Name `xml:"OfflinePaperlessKyc"`
	ReferenceID string   `xml:"referenceId,attr"`
	Name        string   `xml:"Poi name,attr"`
	Gender      string   `xml:"Poi gender,attr"`
	DOB         string   `xml:"Poi dob,attr"`
	Phone       string   `xml:"Poi phone,attr"`
	Email       string   `xml:"Poi email,attr"`

	CO          string `xml:"Poa co,attr"`
	House       string `xml:"Poa house,attr"`
	Street      string `xml:"Poa street,attr"`
	Landmark    string `xml:"Poa lm,attr"`
	Locality    string `xml:"Poa loc,attr"`
	VTC         string `xml:"Poa vtc,attr"`
	SubDistrict string `xml:"Poa subdist,attr"`
	District    string `xml:"Poa dist,attr"`
	State       string `xml:"Poa state,attr"`
	Pincode     string `xml:"Poa pc,attr"`
}

type AadhaarSecureQR struct {
	XML       AadhaarOfflineKyc `json:"xml"`
	Photo     []byte            `json:"photo"`
	Valid     bool              `json:"signature_valid"`
	RawXML    string            `json:"raw_xml"`
	Reference string            `json:"reference"`
	FullAddr  string            `json:"address"`
	Name      string            `json:"name"`
	Gender    string            `json:"gender"`
	DOB       string            `json:"dob"`
	Aadhaar   string            `json:"aadhaar_number,omitempty"`
}

func ParseSecureQR(data []byte, pub *rsa.PublicKey) (*AadhaarSecureQR, error) {
	if len(data) < (2 + 2 + 4 + 4 + 256) {
		return nil, fmt.Errorf("not secure QR: too small")
	}

	r := bytes.NewReader(data)

	// Version + Format
	var version, format uint16
	if err := binary.Read(r, binary.LittleEndian, &version); err != nil {
		return nil, fmt.Errorf("invalid header (version)")
	}
	if err := binary.Read(r, binary.LittleEndian, &format); err != nil {
		return nil, fmt.Errorf("invalid header (format)")
	}

	// Validate known UIDAI versions
	if version != 1 && version != 2 {
		return nil, fmt.Errorf("not secure QR: unknown version %d", version)
	}

	// XML length
	var xmlLen uint32
	if err := binary.Read(r, binary.LittleEndian, &xmlLen); err != nil {
		return nil, fmt.Errorf("invalid xml length")
	}

	if int(xmlLen) <= 0 || int(xmlLen) > len(data) {
		return nil, fmt.Errorf("xml length out of bounds")
	}

	xmlBytes := make([]byte, xmlLen)
	if _, err := io.ReadFull(r, xmlBytes); err != nil {
		return nil, fmt.Errorf("cannot read xml: %v", err)
	}

	// Photo length
	var photoLen uint32
	if err := binary.Read(r, binary.LittleEndian, &photoLen); err != nil {
		return nil, fmt.Errorf("invalid photo length")
	}

	if int(photoLen) < 0 || int(photoLen) > len(data) {
		return nil, fmt.Errorf("photo length out of bounds")
	}

	photoBytes := make([]byte, photoLen)
	if _, err := io.ReadFull(r, photoBytes); err != nil {
		return nil, fmt.Errorf("cannot read photo: %v", err)
	}

	// Signature must be the LAST 256 bytes
	if len(data) < 256 {
		return nil, fmt.Errorf("secure QR missing signature block")
	}
	signature := data[len(data)-256:]

	// Validate signature (SHA-256 over XML only)
	hash := sha256.Sum256(xmlBytes)
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hash[:], signature); err != nil {
		return nil, fmt.Errorf("signature verification failed")
	}

	// Parse XML
	var xmlKYC AadhaarOfflineKyc
	if err := xml.Unmarshal(xmlBytes, &xmlKYC); err != nil {
		return nil, fmt.Errorf("XML parse error: %v", err)
	}

	// Build address
	addr := fmt.Sprintf("%s %s %s %s %s %s %s %s %s %s",
		xmlKYC.CO, xmlKYC.House, xmlKYC.Street, xmlKYC.Landmark,
		xmlKYC.Locality, xmlKYC.SubDistrict, xmlKYC.District,
		xmlKYC.State, xmlKYC.Pincode, xmlKYC.VTC,
	)

	return &AadhaarSecureQR{
		XML:       xmlKYC,
		Photo:     photoBytes,
		Valid:     true,
		RawXML:    string(xmlBytes),
		Reference: xmlKYC.ReferenceID,
		Name:      xmlKYC.Name,
		Gender:    xmlKYC.Gender,
		DOB:       xmlKYC.DOB,
		FullAddr:  addr,
	}, nil
}

func ParseAadhaarQR(data []byte, pub *rsa.PublicKey) (interface{}, string, error) {
	// Try secure QR
	if sec, err := ParseSecureQR(data, pub); err == nil {
		return sec, "secure", nil
	}

	return nil, "", errors.New("unrecognized Aadhaar QR format")
}
