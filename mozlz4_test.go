// Copyright (c) 2024 cions
// Licensed under the MIT License. See LICENSE for details.

package mozlz4

import (
	"bytes"
	"encoding/hex"
	"testing"
)

var tests = []struct {
	uncompressed, compressed string
}{
	{"", "6d6f7a4c7a34300000000000"},
	{"{}", "6d6f7a4c7a34300002000000207b7d"},
	{`{"data": "..."}`, "6d6f7a4c7a3430000f000000f0007b2264617461223a20222e2e2e227d"},
}

func unhex(s string) []byte {
	decoded, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return decoded
}

func TestCompress(t *testing.T) {
	for _, tt := range tests {
		if got, err := Compress([]byte(tt.uncompressed)); err != nil {
			t.Errorf("Compress(%q): unexpected error: %v", tt.uncompressed, err)
		} else if !bytes.Equal(got, unhex(tt.compressed)) {
			t.Errorf("Compress(%q): expected %v, but got %x", tt.uncompressed, tt.compressed, got)
		}
	}
}

func TestDecompress(t *testing.T) {
	for _, tt := range tests {
		if got, err := Decompress(unhex(tt.compressed)); err != nil {
			t.Errorf("Decompress(%q): unexpected error: %v", tt.compressed, err)
		} else if !bytes.Equal(got, []byte(tt.uncompressed)) {
			t.Errorf("Decompress(%q): expected %q, but got %q", tt.compressed, tt.uncompressed, got)
		}
	}

	errorTests := []string{
		"",
		"6d6f7a4c7a343000",
		"6d6f7a4c7a34300001",
		"6d6f7a4c7a34300001000000",
		"6d6f7a4c7a34300001000000207b7d",
		"6d6f7a4c7a34300003000000207b7d",
	}

	for _, tt := range errorTests {
		if _, err := Decompress(unhex(tt)); err == nil {
			t.Errorf("expected non-nil error")
		}
	}
}
