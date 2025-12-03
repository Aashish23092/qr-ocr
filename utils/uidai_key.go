// package utils

// import (
// 	"crypto/rsa"
// 	"crypto/x509"
// 	"encoding/pem"
// 	"errors"
// 	"os"
// )

// func LoadUIDAIPublicKey(path string) (*rsa.PublicKey, error) {
// 	pemBytes, err := os.ReadFile(path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	block, _ := pem.Decode(pemBytes)
// 	if block == nil {
// 		return nil, errors.New("invalid PEM")
// 	}

// 	cert, err := x509.ParseCertificate(block.Bytes)
// 	if err != nil {
// 		return nil, err
// 	}

// 	pub, ok := cert.PublicKey.(*rsa.PublicKey)
// 	if !ok {
// 		return nil, errors.New("not RSA key")
// 	}

// 	return pub, nil
// }

package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func LoadUIDAIPublicKey(path string) (*rsa.PublicKey, error) {
	pemBytes, err := os.ReadFile("C:\\aadhaar-qr-service\\certs\\uidai_public_cert.pem")
	if err != nil {
		return nil, fmt.Errorf("read err: %v", err)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse cert err: %v", err)
	}

	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not RSA public key")
	}

	return pub, nil
}
