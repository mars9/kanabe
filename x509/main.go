package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"os"
	"time"
)

var (
	host     = flag.String("host", "localhost", "host to generate a certificate for")
	org      = flag.String("org", "Acme Co.", "organization to generate a certificate for")
	certFile = flag.String("cert", "cert.pem", "where to save the certificate file")
	keyFile  = flag.String("key", "key.pem", "where to save the private key file")
	validFor = flag.Int("valid", 1, "certificate is for x years")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprint(os.Stderr, usageMsg)
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()

	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate private RSA key: %v\n", err)
		os.Exit(1)
	}

	now := time.Now()
	tmpl := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(0),
		Subject: pkix.Name{
			CommonName:   *host,
			Organization: []string{*org},
		},
		NotBefore: now.Add(-5 * time.Minute).UTC(),
		NotAfter:  now.AddDate(*validFor, 0, 0).UTC(),

		SubjectKeyId: []byte{1, 2, 3, 4},
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	b, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create certificate: %v\n", err)
		os.Exit(1)
	}

	certOut, err := os.Create(*certFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create %s: %v\n", *certFile, err)
		os.Exit(1)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: b})
	certOut.Close()
	fmt.Printf("written %s\n", *certFile)

	keyOut, err := os.OpenFile(*keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open %s: %v\n", *keyFile, err)
		os.Exit(1)
	}
	pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
	keyOut.Close()
	fmt.Printf("written %s\n", *keyFile)
}

const usageMsg = `
	Creates self-signed X.509-encoded keys and certificates.
`
