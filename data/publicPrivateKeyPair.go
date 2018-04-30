package data

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"fmt"
)

// To encode publicKey use:
// publicKeyBytes, _ = x509.MarshalPKIXPublicKey(&private_key.PublicKey)

func main() {
	p384 := elliptic.P384()
	priv1, _ := ecdsa.GenerateKey(p384, rand.Reader)

	privateKeyBytes, _ := x509.MarshalECPrivateKey(priv1)

	encodedBytes := hex.EncodeToString(privateKeyBytes)
	fmt.Println("Private key:")
	fmt.Printf("%s\n", encodedBytes)

	privateKeyBytesRestored, _ := hex.DecodeString(encodedBytes)
	priv2, _ := x509.ParseECPrivateKey(privateKeyBytesRestored)

	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(&priv1.PublicKey)
	encodedPubBytes := hex.EncodeToString(publicKeyBytes)
	fmt.Println("Public key:")
	fmt.Printf("%s\n", encodedPubBytes)

	data := []byte("data")
	// Signing by priv1
	r, s, _ := ecdsa.Sign(rand.Reader, priv1, data)

	// Verifying against priv2 (restored from priv1)
	if !ecdsa.Verify(&priv2.PublicKey, data, r, s) {
		fmt.Printf("Error")
		return
	}

	fmt.Printf("Key was restored from string successfully\n")
}
