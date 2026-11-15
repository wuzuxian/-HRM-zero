package org

import (
	"bytes"
	"encoding/base64"
	"io"
	"testing"
)

func encodeBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func decodeBase64(t *testing.T, s string) io.Reader {
	t.Helper()
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("decode base64: %v", err)
	}
	return bytes.NewReader(b)
}
