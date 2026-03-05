package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

const (
	certFile = "certs/server.crt"
	keyFile  = "certs/server.key"
)

// TLSCertFiles returns paths to the TLS cert and key.
// If they don't exist yet, it generates a self-signed certificate
// valid for localhost + all private-network IPs.
func TLSCertFiles() (cert, key string) {
	cert, key = certFile, keyFile

	// ถ้ามีไฟล์อยู่แล้ว ไม่ต้อง generate ใหม่
	if _, err := os.Stat(cert); err == nil {
		if _, err := os.Stat(key); err == nil {
			return
		}
	}

	log.Println("[TLS] กำลังสร้าง self-signed certificate สำหรับทดสอบ...")

	// ── สร้าง private key (ECDSA P-256) ─────────────────────────────────
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("[TLS] สร้าง key ไม่ได้: %v", err)
	}

	// ── สร้าง certificate template ─────────────────────────────────────
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	tmpl := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      pkix.Name{Organization: []string{"LoanApp Dev"}, CommonName: "localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour), // 1 ปี

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		// SAN: localhost + private IPs
		DNSNames:    []string{"localhost"},
		IPAddresses: localIPs(),
	}

	// ── Sign ───────────────────────────────────────────────────────────
	certDER, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("[TLS] sign cert ไม่ได้: %v", err)
	}

	// ── เขียนไฟล์ cert ─────────────────────────────────────────────────
	os.MkdirAll("certs", 0755)

	cFile, err := os.Create(cert)
	if err != nil {
		log.Fatalf("[TLS] สร้างไฟล์ cert ไม่ได้: %v", err)
	}
	pem.Encode(cFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	cFile.Close()

	// ── เขียนไฟล์ key ──────────────────────────────────────────────────
	kFile, err := os.Create(key)
	if err != nil {
		log.Fatalf("[TLS] สร้างไฟล์ key ไม่ได้: %v", err)
	}
	privBytes, _ := x509.MarshalECPrivateKey(priv)
	pem.Encode(kFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	kFile.Close()

	log.Println("[TLS] ✅ สร้าง cert สำเร็จ:", cert)
	return
}

// localIPs returns all non-loopback IPv4 addresses + 127.0.0.1
func localIPs() []net.IP {
	ips := []net.IP{net.ParseIP("127.0.0.1")}
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		if ipNet, ok := a.(*net.IPNet); ok && ipNet.IP.To4() != nil {
			ips = append(ips, ipNet.IP)
		}
	}
	return ips
}
