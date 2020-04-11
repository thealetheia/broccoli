// Package z85 implements ZeroMQ Base-85 encoding as specified by http://rfc.zeromq.org/spec:32/Z85
package z85

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// ErrLength results from encoding or decoding wrongly aligned input data.
var ErrLength = errors.New("z85: wrongly aligned input data")

// InvalidByteError values describe errors resulting from an invalid byte in a z85 encoded data.
type InvalidByteError byte

func (e InvalidByteError) Error() string {
	return fmt.Sprintf("z85: invalid input byte: %#U", rune(e))
}

const (
	minDigit = '!'
	maxDigit = '}'
)

var (
	digits = []byte(
		"0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.-:+=^!/*?&<>()[]{}@%$#")
	decodeLookup = [maxDigit - minDigit + 1]byte{}
)

func init() {
	for i, d := range digits {
		decodeLookup[d-minDigit] = byte(i + 1) // +1 to use 0 as an invalid byte marker
	}
}

// Decode decodes z85 encoded src into DecodedLen(len(src)) bytes of dst, returning the
// number of bytes written to dst, always DecodedLen(len(src)).
// The len(src) must be divisible by 5, otherwise an ErrLength is returned.
// If Decode encounters invalid input bytes, it returns an InvalidByteError.
func Decode(dst, src []byte) (n int, err error) {
	if len(src)%5 != 0 {
		return 0, ErrLength
	}
	n = DecodedLen(len(src))
	for len(src) > 0 {
		var v uint32
		for i := 0; i < 5; i++ {
			digit := src[i]
			if digit < minDigit || digit > maxDigit {
				return 0, InvalidByteError(digit)
			}
			m := uint32(decodeLookup[digit-minDigit])
			if m == 0 {
				return 0, InvalidByteError(digit)
			}
			v = v*85 + (m - 1) // -1 readjust due to invalid byte marker
		}
		binary.BigEndian.PutUint32(dst, v)
		src = src[5:]
		dst = dst[4:]
	}
	return
}

// DecodedLen returns the length in bytes of the decoded data corresponding to n bytes of
// z85-encoded data.
func DecodedLen(n int) int {
	return n * 4 / 5
}

// Encode encodes src into EncodedLen(len(src)) bytes of dst using z85 encoding, returning the
// number of bytes written to dst, always EncodedLen(len(src)).
// The len(src) must be divisible by 4, otherwise an ErrLength is returned.
func Encode(dst, src []byte) (n int, err error) {
	if len(src)%4 != 0 {
		return 0, ErrLength
	}
	n = EncodedLen(len(src))
	for len(src) > 0 {
		v := binary.BigEndian.Uint32(src)
		for i := 4; i >= 0; i-- {
			dst[i] = digits[v%85]
			v /= 85
		}
		src = src[4:]
		dst = dst[5:]
	}
	return
}

// EncodedLen returns the length in bytes of the z85 encoding of an input buffer of length n.
func EncodedLen(n int) int {
	return n * 5 / 4
}
