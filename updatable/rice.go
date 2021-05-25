package updatable

import (
	"errors"
	"io"
	"math"
	"sort"

	. "checklist/safebrowsing"
)

func RiceEncodedHashes(hashInts []uint32) (*RiceDeltaEncoding, error) {
	sort.Slice(hashInts, func(i, j int) bool { return hashInts[i] < hashInts[j] })
	return EncodeRiceIntegers(hashInts)
}

// Copied from Google's code
func DecodeRiceIntegers(rice *RiceDeltaEncoding) ([]uint32, error) {
	if rice == nil {
		return nil, errors.New("safebrowsing: missing rice encoded data")
	}
	if rice.RiceParameter < 0 || rice.RiceParameter > 32 {
		return nil, errors.New("safebrowsing: invalid k parameter")
	}

	values := []uint32{uint32(rice.FirstValue)}
	br := newBitReader(rice.EncodedData)
	rd := NewRiceDecoder(br, uint32(rice.RiceParameter))
	for i := 0; i < int(rice.NumEntries); i++ {
		delta, err := rd.ReadValue()
		if err != nil {
			return nil, err
		}
		values = append(values, values[i]+delta)
	}

	if br.BitsRemaining() >= 8 {
		return nil, errors.New("safebrowsing: unconsumed rice encoded data")
	}
	return values, nil
}

// riceDecoder implements Golomb-Rice decoding for the Safe Browsing API.
//
// In a Rice decoder every number n is encoded as q and r where n = (q<<k) + r.
// k is a constant and a parameter of the Rice decoder and can have values in
// 0..32 inclusive. The values for q and r are encoded in the bit stream using
// different encoding schemes. The quotient comes before the remainder.
//
// The quotient q is encoded in unary coding followed by a 0. E.g., 3 would be
// encoded as 1110, 4 as 11110, and 7 as 11111110.
//
// The remainder r is encoded using k bits as an unsigned integer with the
// least-significant bits coming first in the bit stream.
//
// For more information, see the following:
//	https://en.wikipedia.org/wiki/Golomb_coding
type riceDecoder struct {
	br *bitReader
	k  uint32 // Golomb-Rice parameter
}

func NewRiceDecoder(br *bitReader, k uint32) *riceDecoder {
	return &riceDecoder{br, k}
}

func (rd *riceDecoder) ReadValue() (uint32, error) {
	var q uint32
	for {
		bit, err := rd.br.ReadBits(1)
		if err != nil {
			return 0, err
		}
		q += bit
		if bit == 0 {
			break
		}
	}

	r, err := rd.br.ReadBits(int(rd.k))
	if err != nil {
		return 0, err
	}

	return q<<rd.k + r, nil
}

// The bitReader provides functionality to read bits from a slice of bytes.
//
// Logically, the bit stream is constructed such that the first byte of buf
// represent the first bits in the stream. Within a byte, the least-significant
// bits come before the most-significant bits in the bit stream.
//
// This is the same bit stream format as DEFLATE (RFC 1951).
type bitReader struct {
	buf  []byte
	mask byte
}

func newBitReader(buf []byte) *bitReader {
	return &bitReader{buf, 0x01}
}

func (br *bitReader) ReadBits(n int) (uint32, error) {
	if n < 0 || n > 32 {
		panic("invalid number of bits")
	}

	var v uint32
	for i := 0; i < n; i++ {
		if len(br.buf) == 0 {
			return v, io.ErrUnexpectedEOF
		}
		if br.buf[0]&br.mask > 0 {
			v |= 1 << uint(i)
		}
		br.mask <<= 1
		if br.mask == 0 {
			br.buf, br.mask = br.buf[1:], 0x01
		}
	}
	return v, nil
}

// BitsRemaining reports the number of bits left to read.
func (br *bitReader) BitsRemaining() int {
	n := 8 * len(br.buf)
	for m := br.mask | 1; m != 1; m >>= 1 {
		n--
	}
	return n
}

// Added code for encoding

type bitWriter struct {
	buf  []byte
	pos  int
	mask byte
}

func newBitWriter() *bitWriter {
	return &bitWriter{make([]byte, 0), -1, 0x00}
}

func (bw *bitWriter) Write(v uint32, n int) error {
	if n < 0 {
		panic("invalid number of bits")
	}

	for i := 0; i < n; i++ {
		if bw.mask == 0 {
			bw.buf, bw.pos, bw.mask = append(bw.buf, 0), bw.pos+1, 0x01
		}

		if v&1 > 0 {
			bw.buf[bw.pos] |= bw.mask

		}
		v >>= 1
		bw.mask <<= 1
	}
	return nil
}

// BitsRemaining reports the number of bits left to read.
func (bw *bitWriter) Bytes() []byte {
	return bw.buf
}

func EncodeRiceIntegersWithParam(values []uint32, riceParam int32) (*RiceDeltaEncoding, error) {
	rice := new(RiceDeltaEncoding)
	rice.FirstValue = int64(values[0])
	rice.NumEntries = int32(len(values)) - 1
	if rice.NumEntries == 0 {
		return rice, nil
	}
	rice.RiceParameter = riceParam
	bw := newBitWriter()
	re := NewRiceEncoder(bw, uint32(rice.RiceParameter))
	for i := range values[1:] {
		err := re.WriteValue(values[i+1] - values[i])
		if err != nil {
			return nil, err
		}
	}
	rice.EncodedData = bw.Bytes()
	return rice, nil

}

func EncodeRiceIntegers(values []uint32) (*RiceDeltaEncoding, error) {
	if len(values) == 0 {
		return nil, nil
	}
	var riceParam int32
	valRange := values[len(values)-1] - values[0]
	if valRange <= 0 {
		riceParam = 1
	} else {
		riceParam = int32(math.Log2(float64(valRange) / float64(len(values))))
	}
	return EncodeRiceIntegersWithParam(values, riceParam)
}

type riceEncoder struct {
	bw *bitWriter
	k  uint32 // Golomb-Rice parameter
}

func NewRiceEncoder(bw *bitWriter, k uint32) *riceEncoder {
	return &riceEncoder{bw, k}
}

func (re *riceEncoder) WriteValue(v uint32) error {
	var q uint32 = v >> re.k
	var r uint32 = v & (1<<re.k - 1)

	var qUnary uint32 = 1<<q - 1
	err := re.bw.Write(qUnary, int(q+1))
	if err != nil {
		return err
	}

	err = re.bw.Write(r, int(re.k))
	if err != nil {
		return err
	}
	return nil
}
