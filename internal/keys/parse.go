package keys

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
)

// ParseBytes extracts every RSA key from data, which is either a PEM bundle
// (any number of CERTIFICATE / PUBLIC KEY / RSA PUBLIC KEY / *PRIVATE KEY
// blocks) or a single DER structure. source labels the keys in reports.
func ParseBytes(data []byte, source string) Result {
	var res Result
	if !bytes.Contains(data, []byte("-----BEGIN")) {
		res.addDER(data, source, "")
		return res
	}
	var blocks []*pem.Block
	for rest := data; ; {
		var block *pem.Block
		if block, rest = pem.Decode(rest); block == nil {
			break
		}
		blocks = append(blocks, block)
	}
	for i, block := range blocks {
		label := source
		if len(blocks) > 1 { // bundle: tell the entries apart
			label = fmt.Sprintf("%s#%d", source, i+1)
		}
		res.addDER(block.Bytes, label, block.Type)
	}
	if len(res.Keys) == 0 && res.Skipped == 0 && len(res.Errors) == 0 {
		res.Errors = append(res.Errors, fmt.Errorf("%s: no PEM block found", source))
	}
	return res
}

// addDER tries the known DER encodings. pemType narrows the guess when known.
func (r *Result) addDER(der []byte, source, pemType string) {
	switch pemType {
	case "CERTIFICATE", "TRUSTED CERTIFICATE", "X509 CERTIFICATE":
		r.addCertificate(der, source)
		return
	case "CERTIFICATE REQUEST", "NEW CERTIFICATE REQUEST", "X509 CRL":
		r.Skipped++
		return
	}

	if cert, err := x509.ParseCertificate(der); err == nil {
		r.addParsedCertificate(cert, source)
		return
	}
	if pub, err := x509.ParsePKIXPublicKey(der); err == nil {
		r.addPublic(pub, source, KindPublicKey)
		return
	}
	if pub, err := x509.ParsePKCS1PublicKey(der); err == nil {
		r.addPublic(pub, source, KindPublicKey)
		return
	}
	if priv, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		r.addPublic(&priv.PublicKey, source, KindPrivateKey)
		return
	}
	if priv, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		r.addPrivate(priv, source)
		return
	}
	r.Errors = append(r.Errors, fmt.Errorf("%s: unrecognised key or certificate encoding", source))
}

func (r *Result) addCertificate(der []byte, source string) {
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		r.Errors = append(r.Errors, fmt.Errorf("%s: parse certificate: %w", source, err))
		return
	}
	r.addParsedCertificate(cert, source)
}

func (r *Result) addParsedCertificate(cert *x509.Certificate, source string) {
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		r.Skipped++
		return
	}
	sum := sha256.Sum256(cert.Raw)
	r.Keys = append(r.Keys, &Key{
		N:      new(big.Int).Set(pub.N),
		E:      pub.E,
		Source: source,
		Kind:   KindCertificate,
		Raw:    cert.Raw,
		Cert: &CertInfo{
			Subject:            cert.Subject.String(),
			Issuer:             cert.Issuer.String(),
			Serial:             cert.SerialNumber.Text(16),
			NotBefore:          cert.NotBefore,
			NotAfter:           cert.NotAfter,
			SignatureAlgorithm: cert.SignatureAlgorithm.String(),
			DNSNames:           cert.DNSNames,
			SHA256:             hex.EncodeToString(sum[:]),
			Precert:            isPrecert(cert),
		},
	})
}

// isPrecert detects the RFC 6962 poison extension.
func isPrecert(cert *x509.Certificate) bool {
	for _, ext := range cert.Extensions {
		if ext.Id.String() == "1.3.6.1.4.1.11129.2.4.3" {
			return true
		}
	}
	return false
}

func (r *Result) addPrivate(priv any, source string) {
	if rk, ok := priv.(*rsa.PrivateKey); ok {
		r.addPublic(&rk.PublicKey, source, KindPrivateKey)
		return
	}
	r.Skipped++
}

func (r *Result) addPublic(pub any, source string, kind Kind) {
	rk, ok := pub.(*rsa.PublicKey)
	if !ok {
		r.Skipped++
		return
	}
	r.Keys = append(r.Keys, &Key{N: new(big.Int).Set(rk.N), E: rk.E, Source: source, Kind: kind})
}
