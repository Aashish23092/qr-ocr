package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/Aashish23092/aadhaar-qr-service/handlers"
	"github.com/Aashish23092/aadhaar-qr-service/utils"
)

func main() {
	pub, err := utils.LoadUIDAIPublicKey("certs/uidai_public_cert.pem")
	if err != nil {
		log.Fatal("Failed loading public key:", err)
	}

	r := gin.Default()
	handler := handlers.NewQRHandler(pub)

	r.POST("/decode", handler.Decode)

	r.Run(":8080")
}
