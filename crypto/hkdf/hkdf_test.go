// Iris - Decentralized Messaging Framework
// Copyright 2013 Peter Szilagyi. All rights reserved.
//
// Iris is dual licensed: you can redistribute it and/or modify it under the
// terms of the GNU General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later
// version.
//
// The framework is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// Alternatively, the Iris framework may be used in accordance with the terms
// and conditions contained in a signed written agreement between you and the
// author(s).
//
// Author: peterke@gmail.com (Peter Szilagyi)
package hkdf

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"io"
	"testing"
)

type hkdfTest struct {
	hash   func() hash.Hash
	master []byte
	salt   []byte
	info   []byte
	out    []byte
}

var hkdfTests = []hkdfTest{
	// Tests from RFC 5869
	{
		sha256.New,
		[]byte{
			0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
			0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
			0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		},
		[]byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0a, 0x0b, 0x0c,
		},
		[]byte{
			0xf0, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7,
			0xf8, 0xf9,
		},
		[]byte{
			0x3c, 0xb2, 0x5f, 0x25, 0xfa, 0xac, 0xd5, 0x7a,
			0x90, 0x43, 0x4f, 0x64, 0xd0, 0x36, 0x2f, 0x2a,
			0x2d, 0x2d, 0x0a, 0x90, 0xcf, 0x1a, 0x5a, 0x4c,
			0x5d, 0xb0, 0x2d, 0x56, 0xec, 0xc4, 0xc5, 0xbf,
			0x34, 0x00, 0x72, 0x08, 0xd5, 0xb8, 0x87, 0x18,
			0x58, 0x65,
		},
	},
	{
		sha256.New,
		[]byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
			0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
			0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27,
			0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
			0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37,
			0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
			0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47,
			0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f,
		},
		[]byte{
			0x60, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67,
			0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f,
			0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77,
			0x78, 0x79, 0x7a, 0x7b, 0x7c, 0x7d, 0x7e, 0x7f,
			0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87,
			0x88, 0x89, 0x8a, 0x8b, 0x8c, 0x8d, 0x8e, 0x8f,
			0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97,
			0x98, 0x99, 0x9a, 0x9b, 0x9c, 0x9d, 0x9e, 0x9f,
			0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7,
			0xa8, 0xa9, 0xaa, 0xab, 0xac, 0xad, 0xae, 0xaf,
		},
		[]byte{
			0xb0, 0xb1, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7,
			0xb8, 0xb9, 0xba, 0xbb, 0xbc, 0xbd, 0xbe, 0xbf,
			0xc0, 0xc1, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7,
			0xc8, 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf,
			0xd0, 0xd1, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7,
			0xd8, 0xd9, 0xda, 0xdb, 0xdc, 0xdd, 0xde, 0xdf,
			0xe0, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5, 0xe6, 0xe7,
			0xe8, 0xe9, 0xea, 0xeb, 0xec, 0xed, 0xee, 0xef,
			0xf0, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7,
			0xf8, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff,
		},
		[]byte{
			0xb1, 0x1e, 0x39, 0x8d, 0xc8, 0x03, 0x27, 0xa1,
			0xc8, 0xe7, 0xf7, 0x8c, 0x59, 0x6a, 0x49, 0x34,
			0x4f, 0x01, 0x2e, 0xda, 0x2d, 0x4e, 0xfa, 0xd8,
			0xa0, 0x50, 0xcc, 0x4c, 0x19, 0xaf, 0xa9, 0x7c,
			0x59, 0x04, 0x5a, 0x99, 0xca, 0xc7, 0x82, 0x72,
			0x71, 0xcb, 0x41, 0xc6, 0x5e, 0x59, 0x0e, 0x09,
			0xda, 0x32, 0x75, 0x60, 0x0c, 0x2f, 0x09, 0xb8,
			0x36, 0x77, 0x93, 0xa9, 0xac, 0xa3, 0xdb, 0x71,
			0xcc, 0x30, 0xc5, 0x81, 0x79, 0xec, 0x3e, 0x87,
			0xc1, 0x4c, 0x01, 0xd5, 0xc1, 0xf3, 0x43, 0x4f,
			0x1d, 0x87,
		},
	},
	{
		sha256.New,
		[]byte{
			0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
			0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
			0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		},
		[]byte{},
		[]byte{},
		[]byte{
			0x8d, 0xa4, 0xe7, 0x75, 0xa5, 0x63, 0xc1, 0x8f,
			0x71, 0x5f, 0x80, 0x2a, 0x06, 0x3c, 0x5a, 0x31,
			0xb8, 0xa1, 0x1f, 0x5c, 0x5e, 0xe1, 0x87, 0x9e,
			0xc3, 0x45, 0x4e, 0x5f, 0x3c, 0x73, 0x8d, 0x2d,
			0x9d, 0x20, 0x13, 0x95, 0xfa, 0xa4, 0xb6, 0x1a,
			0x96, 0xc8,
		},
	},
	{
		sha1.New,
		[]byte{
			0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
			0x0b, 0x0b, 0x0b,
		},
		[]byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0a, 0x0b, 0x0c,
		},
		[]byte{
			0xf0, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7,
			0xf8, 0xf9,
		},
		[]byte{
			0x08, 0x5a, 0x01, 0xea, 0x1b, 0x10, 0xf3, 0x69,
			0x33, 0x06, 0x8b, 0x56, 0xef, 0xa5, 0xad, 0x81,
			0xa4, 0xf1, 0x4b, 0x82, 0x2f, 0x5b, 0x09, 0x15,
			0x68, 0xa9, 0xcd, 0xd4, 0xf1, 0x55, 0xfd, 0xa2,
			0xc2, 0x2e, 0x42, 0x24, 0x78, 0xd3, 0x05, 0xf3,
			0xf8, 0x96,
		},
	},
	{
		sha1.New,
		[]byte{
			0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
			0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
			0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
			0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27,
			0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
			0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37,
			0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
			0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47,
			0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f,
		},
		[]byte{
			0x60, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67,
			0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f,
			0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77,
			0x78, 0x79, 0x7a, 0x7b, 0x7c, 0x7d, 0x7e, 0x7f,
			0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87,
			0x88, 0x89, 0x8a, 0x8b, 0x8c, 0x8d, 0x8e, 0x8f,
			0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97,
			0x98, 0x99, 0x9a, 0x9b, 0x9c, 0x9d, 0x9e, 0x9f,
			0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7,
			0xa8, 0xa9, 0xaa, 0xab, 0xac, 0xad, 0xae, 0xaf,
		},
		[]byte{
			0xb0, 0xb1, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7,
			0xb8, 0xb9, 0xba, 0xbb, 0xbc, 0xbd, 0xbe, 0xbf,
			0xc0, 0xc1, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7,
			0xc8, 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf,
			0xd0, 0xd1, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7,
			0xd8, 0xd9, 0xda, 0xdb, 0xdc, 0xdd, 0xde, 0xdf,
			0xe0, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5, 0xe6, 0xe7,
			0xe8, 0xe9, 0xea, 0xeb, 0xec, 0xed, 0xee, 0xef,
			0xf0, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7,
			0xf8, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff,
		},
		[]byte{
			0x0b, 0xd7, 0x70, 0xa7, 0x4d, 0x11, 0x60, 0xf7,
			0xc9, 0xf1, 0x2c, 0xd5, 0x91, 0x2a, 0x06, 0xeb,
			0xff, 0x6a, 0xdc, 0xae, 0x89, 0x9d, 0x92, 0x19,
			0x1f, 0xe4, 0x30, 0x56, 0x73, 0xba, 0x2f, 0xfe,
			0x8f, 0xa3, 0xf1, 0xa4, 0xe5, 0xad, 0x79, 0xf3,
			0xf3, 0x34, 0xb3, 0xb2, 0x02, 0xb2, 0x17, 0x3c,
			0x48, 0x6e, 0xa3, 0x7c, 0xe3, 0xd3, 0x97, 0xed,
			0x03, 0x4c, 0x7f, 0x9d, 0xfe, 0xb1, 0x5c, 0x5e,
			0x92, 0x73, 0x36, 0xd0, 0x44, 0x1f, 0x4c, 0x43,
			0x00, 0xe2, 0xcf, 0xf0, 0xd0, 0x90, 0x0b, 0x52,
			0xd3, 0xb4,
		},
	},
	{
		sha1.New,
		[]byte{
			0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
			0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
			0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
		},
		[]byte{},
		[]byte{},
		[]byte{
			0x0a, 0xc1, 0xaf, 0x70, 0x02, 0xb3, 0xd7, 0x61,
			0xd1, 0xe5, 0x52, 0x98, 0xda, 0x9d, 0x05, 0x06,
			0xb9, 0xae, 0x52, 0x05, 0x72, 0x20, 0xa3, 0x06,
			0xe0, 0x7b, 0x6b, 0x87, 0xe8, 0xdf, 0x21, 0xd0,
			0xea, 0x00, 0x03, 0x3d, 0xe0, 0x39, 0x84, 0xd3,
			0x49, 0x18,
		},
	},
	{
		sha1.New,
		[]byte{
			0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c,
			0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c,
			0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c,
		},
		nil,
		[]byte{},
		[]byte{
			0x2c, 0x91, 0x11, 0x72, 0x04, 0xd7, 0x45, 0xf3,
			0x50, 0x0d, 0x63, 0x6a, 0x62, 0xf6, 0x4f, 0x0a,
			0xb3, 0xba, 0xe5, 0x48, 0xaa, 0x53, 0xd4, 0x23,
			0xb0, 0xd1, 0xf2, 0x7e, 0xbb, 0xa6, 0xf5, 0xe5,
			0x67, 0x3a, 0x08, 0x1d, 0x70, 0xcc, 0xe7, 0xac,
			0xfc, 0x48,
		},
	},
}

func TestHKDF(t *testing.T) {
	for i, tt := range hkdfTests {
		hkdf := New(tt.hash, tt.master, tt.salt, tt.info)
		out := make([]byte, len(tt.out))

		n, err := io.ReadFull(hkdf, out)
		if n != len(tt.out) || err != nil {
			t.Errorf("test %d: not enough output bytes: %d.", i, n)
		}

		if !bytes.Equal(out, tt.out) {
			t.Errorf("test %d: incorrect output: have %v, need %v.", i, out, tt.out)
		}
	}
}

func TestHKDFMultiRead(t *testing.T) {
	for i, tt := range hkdfTests {
		hkdf := New(tt.hash, tt.master, tt.salt, tt.info)
		out := make([]byte, len(tt.out))

		for b := 0; b < len(tt.out); b++ {
			n, err := io.ReadFull(hkdf, out[b:b+1])
			if n != 1 || err != nil {
				t.Errorf("test %d.%d: not enough output bytes: have %d, need %d .", i, b, n, len(tt.out))
			}
		}

		if !bytes.Equal(out, tt.out) {
			t.Errorf("test %d: incorrect output: have %v, need %v.", i, out, tt.out)
		}
	}
}

func TestHKDFLimit(t *testing.T) {
	hash := sha1.New
	master := []byte{0x00, 0x01, 0x02, 0x03}
	info := []byte{}

	hkdf := New(hash, master, nil, info)
	limit := hash().Size() * 255
	out := make([]byte, limit)

	// The maximum output bytes should be extractable
	n, err := io.ReadFull(hkdf, out)
	if n != limit || err != nil {
		t.Errorf("not enough output bytes: %d, %v.", n, err)
	}

	// Reading one more should fail
	n, err = io.ReadFull(hkdf, make([]byte, 1))
	if n > 0 || err == nil {
		t.Errorf("key expansion overflowed: n = %d, err = %v", n, err)
	}
}

func Benchmark16ByteMd5Single(b *testing.B) {
	benchmarkHKDFSingle(md5.New, 16, b)
}

func Benchmark20ByteSha1Single(b *testing.B) {
	benchmarkHKDFSingle(sha1.New, 20, b)
}

func Benchmark32ByteSha256Single(b *testing.B) {
	benchmarkHKDFSingle(sha256.New, 32, b)
}

func Benchmark64ByteSha512Single(b *testing.B) {
	benchmarkHKDFSingle(sha512.New, 64, b)
}

func Benchmark8ByteMd5Stream(b *testing.B) {
	benchmarkHKDFStream(md5.New, 8, b)
}

func Benchmark16ByteMd5Stream(b *testing.B) {
	benchmarkHKDFStream(md5.New, 16, b)
}

func Benchmark8ByteSha1Stream(b *testing.B) {
	benchmarkHKDFStream(sha1.New, 8, b)
}

func Benchmark20ByteSha1Stream(b *testing.B) {
	benchmarkHKDFStream(sha1.New, 20, b)
}

func Benchmark8ByteSha256Stream(b *testing.B) {
	benchmarkHKDFStream(sha256.New, 8, b)
}

func Benchmark32ByteSha256Stream(b *testing.B) {
	benchmarkHKDFStream(sha256.New, 32, b)
}

func Benchmark8ByteSha512Stream(b *testing.B) {
	benchmarkHKDFStream(sha512.New, 8, b)
}

func Benchmark64ByteSha512Stream(b *testing.B) {
	benchmarkHKDFStream(sha512.New, 64, b)
}

func benchmarkHKDFSingle(hasher func() hash.Hash, block int, b *testing.B) {
	master := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	salt := []byte{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17}
	info := []byte{0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27}
	out := make([]byte, block)

	b.SetBytes(int64(block))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hkdf := New(hasher, master, salt, info)
		io.ReadFull(hkdf, out)
	}
}

func benchmarkHKDFStream(hasher func() hash.Hash, block int, b *testing.B) {
	master := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}
	salt := []byte{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17}
	info := []byte{0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27}
	out := make([]byte, block)

	b.SetBytes(int64(block))
	b.ResetTimer()

	hkdf := New(hasher, master, salt, info)
	for i := 0; i < b.N; i++ {
		_, err := io.ReadFull(hkdf, out)
		if err != nil {
			hkdf = New(hasher, master, salt, info)
			i--
		}
	}
}
