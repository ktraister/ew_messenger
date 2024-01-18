package main

import (
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"time"
)

func main() {
	// Specify the domain for which you want to check the certificate
	domain := "endlesswaltz.xyz"

	ts := &tls.Config{
		InsecureSkipVerify: true,
	}

	// Make a TLS connection to the specified domain
	conn, err := tls.Dial("tcp", domain+":443", ts)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer conn.Close()

	// Get the peer certificate from the TLS connection
	cert := conn.ConnectionState().PeerCertificates[0]

	// Convert certificate to PEM format
	pemBlock := &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}
	pemEncoded := pem.EncodeToMemory(pemBlock)

	// Print PEM encoded certificate
	fmt.Println(string(pemEncoded))

	// Print individual fields
	fmt.Println("Serial Number:", cert.SerialNumber)
	fmt.Println("Subject:", cert.Subject)
	fmt.Println("Issuer:", cert.Issuer)
	fmt.Println("Not Before:", cert.NotBefore.Format(time.RFC3339))
	fmt.Println("Not After:", cert.NotAfter.Format(time.RFC3339))
	fmt.Println("Key Usage:", cert.KeyUsage)
	fmt.Println("Extended Key Usage:", cert.ExtKeyUsage)
	fmt.Println("DNS Names:", cert.DNSNames)
	fmt.Println("IP Addresses:", cert.IPAddresses)
	fmt.Println("Signature Algorithm:", cert.SignatureAlgorithm)
	fmt.Println("Public Key Algorithm:", cert.PublicKeyAlgorithm)
	fmt.Println("Public Key:", cert.PublicKey)
}
