package ct

import (
	"encoding/base64"
	"errors"
	"fmt"
)

// RFC 6962 §3.4 and §4.6 structures.
const (
	entryX509    = 0
	entryPrecert = 1

	leafVersionV1       = 0
	leafTypeTimestamped = 0
	leafHeaderLen       = 1 + 1 + 8 + 2 // version, leaf type, timestamp, entry type
	issuerHashLen       = 32
	uint16Len           = 2
	uint24Len           = 3

	// maxEncodedLen bounds the base64 text of leaf_input and extra_data
	// (1 MiB). Real certificates and chains stay far below.
	maxEncodedLen = 1 << 20
)

var (
	errShort    = errors.New("truncated structure")
	errTooLarge = errors.New("larger than the supported size")
	errLength   = errors.New("inconsistent length")
)

// ParseEntry returns the DER certificate of a log entry. For a precertificate
// the leaf only carries the TBSCertificate, so the full pre-certificate is
// read from extra_data (RFC 6962 §4.6); both embed the same public key.
// Every length field is checked against the data actually present.
func ParseEntry(e RawEntry) (der []byte, precert bool, err error) {
	if len(e.LeafInput) > maxEncodedLen || len(e.ExtraData) > maxEncodedLen {
		return nil, false, errTooLarge
	}
	leaf, err := base64.StdEncoding.DecodeString(e.LeafInput)
	if err != nil {
		return nil, false, fmt.Errorf("leaf_input: %w", err)
	}
	if len(leaf) < leafHeaderLen {
		return nil, false, fmt.Errorf("leaf_input: %w", errShort)
	}
	if leaf[0] != leafVersionV1 || leaf[1] != leafTypeTimestamped {
		return nil, false, fmt.Errorf("leaf_input: unsupported version %d / leaf type %d", leaf[0], leaf[1])
	}
	body := leaf[leafHeaderLen:]
	switch entryType := int(leaf[leafHeaderLen-2])<<8 | int(leaf[leafHeaderLen-1]); entryType {
	case entryX509:
		der, err = parseX509Leaf(body)
		return der, false, err
	case entryPrecert:
		der, err = parsePrecertEntry(body, e.ExtraData)
		return der, true, err
	default:
		return nil, false, fmt.Errorf("unknown entry type %d", entryType)
	}
}

// parseX509Leaf reads ASN.1Cert<1..2^24-1> then CtExtensions, and nothing else.
func parseX509Leaf(body []byte) ([]byte, error) {
	der, rest, err := readUint24Prefixed(body)
	if err != nil {
		return nil, fmt.Errorf("x509 entry: %w", err)
	}
	if err := checkExtensions(rest); err != nil {
		return nil, fmt.Errorf("x509 entry: %w", err)
	}
	return der, nil
}

// parsePrecertEntry validates the leaf (issuer hash, TBS, extensions) and
// returns the pre-certificate found in extra_data.
func parsePrecertEntry(body []byte, extraB64 string) ([]byte, error) {
	if len(body) < issuerHashLen {
		return nil, fmt.Errorf("precert entry: %w", errShort)
	}
	_, rest, err := readUint24Prefixed(body[issuerHashLen:])
	if err != nil {
		return nil, fmt.Errorf("precert entry: tbs: %w", err)
	}
	if err := checkExtensions(rest); err != nil {
		return nil, fmt.Errorf("precert entry: %w", err)
	}

	extra, err := base64.StdEncoding.DecodeString(extraB64)
	if err != nil {
		return nil, fmt.Errorf("extra_data: %w", err)
	}
	der, chain, err := readUint24Prefixed(extra)
	if err != nil {
		return nil, fmt.Errorf("precert extra_data: %w", err)
	}
	if _, trailing, err := readUint24Prefixed(chain); err != nil || len(trailing) != 0 {
		return nil, fmt.Errorf("precert extra_data: chain: %w", errLength)
	}
	return der, nil
}

// checkExtensions verifies that rest is exactly a uint16-prefixed CtExtensions.
func checkExtensions(rest []byte) error {
	if len(rest) < uint16Len {
		return fmt.Errorf("extensions: %w", errShort)
	}
	if n := int(rest[0])<<8 | int(rest[1]); n != len(rest)-uint16Len {
		return fmt.Errorf("extensions: %w", errLength)
	}
	return nil
}

// readUint24Prefixed reads a TLS opaque<0..2^24-1> and returns it with the rest.
func readUint24Prefixed(b []byte) (val, rest []byte, err error) {
	if len(b) < uint24Len {
		return nil, nil, errShort
	}
	n := int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	if len(b) < uint24Len+n {
		return nil, nil, errShort
	}
	return b[uint24Len : uint24Len+n], b[uint24Len+n:], nil
}
