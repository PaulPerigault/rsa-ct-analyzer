// Package mocklog serves a local, read-only RFC 6962 log from a fixed list of
// certificates. It lets the crawler run with no network and no real log:
// tests, demos and CI use it as a simulated CT feed.
package mocklog

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// Entry is one certificate to publish.
type Entry struct {
	DER     []byte
	Precert bool // publish as a precertificate entry instead of an X.509 one
}

// Handler returns the log's HTTP handler. pageSize caps entries per
// get-entries answer, as real logs do (0 means 256).
func Handler(entries []Entry, pageSize int) http.Handler {
	if pageSize <= 0 {
		pageSize = 256
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ct/v1/get-sth", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"tree_size":           len(entries),
			"timestamp":           time.Now().UnixMilli(),
			"sha256_root_hash":    base64.StdEncoding.EncodeToString(make([]byte, 32)),
			"tree_head_signature": "",
		})
	})
	mux.HandleFunc("GET /ct/v1/get-entries", func(w http.ResponseWriter, r *http.Request) {
		start, err1 := strconv.Atoi(r.URL.Query().Get("start"))
		end, err2 := strconv.Atoi(r.URL.Query().Get("end"))
		if err1 != nil || err2 != nil || start < 0 || end < start || start >= len(entries) {
			http.Error(w, `{"error_message":"invalid range"}`, http.StatusBadRequest)
			return
		}
		end = min(end, start+pageSize-1, len(entries)-1)
		out := make([]map[string]string, 0, end-start+1)
		for _, e := range entries[start : end+1] {
			out = append(out, map[string]string{
				"leaf_input": base64.StdEncoding.EncodeToString(leaf(e)),
				"extra_data": base64.StdEncoding.EncodeToString(extra(e)),
			})
		}
		writeJSON(w, map[string]any{"entries": out})
	})
	return mux
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// leaf encodes a MerkleTreeLeaf (RFC 6962 §3.4), with empty extensions.
func leaf(e Entry) []byte {
	b := []byte{0, 0}                     // version v1, leaf type timestamped_entry
	b = append(b, 0, 0, 0, 0, 0, 0, 0, 1) // timestamp
	if e.Precert {
		b = append(b, 0, 1)
		b = append(b, make([]byte, 32)...) // issuer_key_hash (not checked by the tool)
		b = append(b, uint24Prefixed(e.DER)...)
		return append(b, 0, 0)
	}
	b = append(b, 0, 0)
	b = append(b, uint24Prefixed(e.DER)...)
	return append(b, 0, 0)
}

// extra encodes extra_data: for a precert the pre-certificate then an empty
// chain; for an X.509 entry just an empty chain.
func extra(e Entry) []byte {
	if e.Precert {
		return append(uint24Prefixed(e.DER), 0, 0, 0)
	}
	return []byte{0, 0, 0}
}

func uint24Prefixed(b []byte) []byte {
	n := len(b)
	return append([]byte{byte(n >> 16), byte(n >> 8), byte(n)}, b...)
}
