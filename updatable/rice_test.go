package updatable

import (
	"encoding/binary"
	"encoding/hex"
	"reflect"
	"sort"
	"testing"

	sb "checklist/safebrowsing"

	"github.com/golang/protobuf/proto"
	"gotest.tools/assert"
)

func mustDecodeHex(t *testing.T, s string) []byte {
	buf, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return buf
}

func TestRiceEncoder(t *testing.T) {
	// These vectors were randomly generated using the server-side Rice
	// compression implementation.
	vectors := []struct {
		k      uint32   // Golomb-Rice value
		output string   // The input encoded string in hexadecimal format
		input  []uint32 // The expected output values
	}{{
		2,
		"f702",
		[]uint32{15, 9},
	}, {
		5,
		"00",
		[]uint32{0},
	}, {
		10,
		"",
		[]uint32{},
	}, {
		28,
		"54607be70a5fc1dcee69defe583ca3d6a5f2108c4a595600",
		[]uint32{62763050, 1046523781, 192522171, 1800511020, 4442775, 582142548},
	}, {
		28,
		"06861b2314cb46f2af0708c988541f4104d51a03ebe63a8013917bbf83f3b785f12918b36109",
		[]uint32{26067715, 344823336, 8420095, 399843890, 95029378, 731622412, 35811335, 1047558127, 1117722715, 78698892},
	}, {
		27,
		"8998d875bc4491eb390c3e309a78f36ad4d9b19ffb703e443ea3086742c22b46698e3cebd9105a439a32a52d4e770f877820b6ab7198480c9e9ed7230c13432ca901",
		[]uint32{225846818, 328287420, 166748623, 29117720, 552397365, 350353215, 558267528, 4738273, 567093445, 28563065, 55077698, 73091685, 339246010, 98242620, 38060941, 63917830, 206319759, 137700744},
	}, {
		28,
		"21c50291f982d757b8e93cf0c84fe8648d776204d6853f1c9700041b17c6",
		[]uint32{339784008, 263128563, 63871877, 69723256, 826001074, 797300228, 671166008, 207712688},
	}, {
		28,
		"959c7db08fe8d9bdfe8c7f81530d75dc4e40180c9a453da8dcfa2659409e16084377c34e0401a4e65d00",
		[]uint32{471820069, 196333855, 855579133, 122737976, 203433838, 85354544, 1307949392, 165938578, 195134475, 553930435, 49231136},
	}, {
		27,
		"1a4f692a639af6c62eaf73d06fd731eb771d43e32b93ce678b59f998d4da4f3c6fb0e8a5788d623618fe081e78d814322484611cf33763c4a0887b74cb64c85cba05",
		[]uint32{87336845, 129291033, 30906211, 433549264, 30899891, 53207875, 11959529, 354827862, 82919275, 489637251, 53561020, 336722992, 408117728, 204506246, 188216092, 9047110, 479817359, 230317256},
	}, {
		28,
		"f1940a876c5f9690e3abf7c0cb2de976dbf85963c16f7c99e3875fc704deb9468e54c0ac4a030d6c8f00",
		[]uint32{297968956, 19709657, 259702329, 76998112, 1023176123, 29296013, 1602741145, 393745181, 177326295, 55225536, 75194472},
	}, {
		28,
		"412ce4fe06dc0dbd31a504d56edd9b43b73f11245210804f964bd48067b2dd52c94e02c6d760de0692521edd356471262cfecf8146b27901",
		[]uint32{532220688, 780594691, 436816483, 163436269, 573044456, 1069604, 39629436, 211410997, 227714491, 381562898, 75610008, 196754597, 40310339, 15204118, 99010842},
	}, {
		28,
		"b22c263acd669cdb5f072e6fe6f9211052d594f4822248f99d24f6ff2ffc6d3f21651b363456eac42100",
		[]uint32{219354713, 389598618, 750263679, 554684211, 87381124, 4523497, 287633354, 801308671, 424169435, 372520475, 277287849},
	}}

loop:
	for _, v := range vectors {
		bw := newBitWriter()
		re := NewRiceEncoder(bw, v.k)

		for i := 0; i < len(v.input); i++ {
			err := re.WriteValue(v.input[i])
			if err != nil {
				t.Errorf("test %d, unexpected error: %v", i, err)
				continue loop
			}
		}

		buf := bw.Bytes()
		s := mustDecodeHex(t, v.output)
		assert.DeepEqual(t, s, buf)
	}
}

func TestRiceDecoder(t *testing.T) {
	// These vectors were randomly generated using the server-side Rice
	// compression implementation.
	vectors := []struct {
		k      uint32   // Golomb-Rice value
		input  string   // The input encoded string in hexadecimal format
		output []uint32 // The expected output values
	}{{
		2,
		"f702",
		[]uint32{15, 9},
	}, {
		5,
		"00",
		[]uint32{0},
	}, {
		10,
		"",
		[]uint32{},
	}, {
		28,
		"54607be70a5fc1dcee69defe583ca3d6a5f2108c4a595600",
		[]uint32{62763050, 1046523781, 192522171, 1800511020, 4442775, 582142548},
	}, {
		28,
		"06861b2314cb46f2af0708c988541f4104d51a03ebe63a8013917bbf83f3b785f12918b36109",
		[]uint32{26067715, 344823336, 8420095, 399843890, 95029378, 731622412, 35811335, 1047558127, 1117722715, 78698892},
	}, {
		27,
		"8998d875bc4491eb390c3e309a78f36ad4d9b19ffb703e443ea3086742c22b46698e3cebd9105a439a32a52d4e770f877820b6ab7198480c9e9ed7230c13432ca901",
		[]uint32{225846818, 328287420, 166748623, 29117720, 552397365, 350353215, 558267528, 4738273, 567093445, 28563065, 55077698, 73091685, 339246010, 98242620, 38060941, 63917830, 206319759, 137700744},
	}, {
		28,
		"21c50291f982d757b8e93cf0c84fe8648d776204d6853f1c9700041b17c6",
		[]uint32{339784008, 263128563, 63871877, 69723256, 826001074, 797300228, 671166008, 207712688},
	}, {
		28,
		"959c7db08fe8d9bdfe8c7f81530d75dc4e40180c9a453da8dcfa2659409e16084377c34e0401a4e65d00",
		[]uint32{471820069, 196333855, 855579133, 122737976, 203433838, 85354544, 1307949392, 165938578, 195134475, 553930435, 49231136},
	}, {
		27,
		"1a4f692a639af6c62eaf73d06fd731eb771d43e32b93ce678b59f998d4da4f3c6fb0e8a5788d623618fe081e78d814322484611cf33763c4a0887b74cb64c85cba05",
		[]uint32{87336845, 129291033, 30906211, 433549264, 30899891, 53207875, 11959529, 354827862, 82919275, 489637251, 53561020, 336722992, 408117728, 204506246, 188216092, 9047110, 479817359, 230317256},
	}, {
		28,
		"f1940a876c5f9690e3abf7c0cb2de976dbf85963c16f7c99e3875fc704deb9468e54c0ac4a030d6c8f00",
		[]uint32{297968956, 19709657, 259702329, 76998112, 1023176123, 29296013, 1602741145, 393745181, 177326295, 55225536, 75194472},
	}, {
		28,
		"412ce4fe06dc0dbd31a504d56edd9b43b73f11245210804f964bd48067b2dd52c94e02c6d760de0692521edd356471262cfecf8146b27901",
		[]uint32{532220688, 780594691, 436816483, 163436269, 573044456, 1069604, 39629436, 211410997, 227714491, 381562898, 75610008, 196754597, 40310339, 15204118, 99010842},
	}, {
		28,
		"b22c263acd669cdb5f072e6fe6f9211052d594f4822248f99d24f6ff2ffc6d3f21651b363456eac42100",
		[]uint32{219354713, 389598618, 750263679, 554684211, 87381124, 4523497, 287633354, 801308671, 424169435, 372520475, 277287849},
	}}

loop:
	for i, v := range vectors {
		br := newBitReader(mustDecodeHex(t, v.input))
		rd := NewRiceDecoder(br, v.k)

		vals := []uint32{}
		for i := 0; i < len(v.output); i++ {
			val, err := rd.ReadValue()
			if err != nil {
				t.Errorf("test %d, unexpected error: %v", i, err)
				continue loop
			}
			vals = append(vals, val)
		}
		if !reflect.DeepEqual(vals, v.output) {
			t.Errorf("test %d, output mismatch:\ngot  %v\nwant %v", i, vals, v.output)
		}
	}
}

func TestBitReader(t *testing.T) {
	vectors := []struct {
		cnt int    // Number of bits to read
		val uint32 // Expected output value to read
		rem int    // Number of bits remaining in the bitReader
	}{
		{cnt: 0, val: 0, rem: 56},
		{cnt: 1, val: 1, rem: 55},
		{cnt: 1, val: 0, rem: 54},
		{cnt: 1, val: 1, rem: 53},
		{cnt: 1, val: 1, rem: 52},
		{cnt: 8, val: 0x20, rem: 44},
		{cnt: 32, val: 0x40000000, rem: 12},
		{cnt: 9, val: 0x00000170, rem: 3},
		{cnt: 3, val: 0x00000001, rem: 0},
	}

	// Test bitReader with data.
	br := newBitReader(mustDecodeHex(t, "0d020000000437"))
	for i, v := range vectors {
		val, err := br.ReadBits(v.cnt)
		if err != nil {
			t.Errorf("test %d, unexpected error: %v", i, err)
		}
		if val != v.val {
			t.Errorf("test %d, ReadBits() = 0x%08x, want 0x%08x", i, val, v.val)
		}
		if rem := br.BitsRemaining(); rem != v.rem {
			t.Errorf("test %d, BitsRemaining() = %d, want %d", i, rem, v.rem)
		}
	}

	// Test empty bitReader.
	br = newBitReader(mustDecodeHex(t, ""))
	if rem := br.BitsRemaining(); rem != 0 {
		t.Errorf("BitsRemaining() = %d, want 0", rem)
	}
	if _, err := br.ReadBits(1); err == nil {
		t.Errorf("unexpected ReadBits success")
	}
}

func TestBitWriter(t *testing.T) {
	vectors := []struct {
		cnt int    // Number of bits to read
		val uint32 // Expected output value to read
	}{
		{cnt: 0, val: 0},
		{cnt: 1, val: 1},
		{cnt: 1, val: 0},
		{cnt: 1, val: 1},
		{cnt: 1, val: 1},
		{cnt: 8, val: 0x20},
		{cnt: 32, val: 0x40000000},
		{cnt: 9, val: 0x00000170},
		{cnt: 3, val: 0x00000001},
	}

	// Test bitReader with data.
	br := newBitWriter()
	for i, v := range vectors {
		err := br.Write(v.val, v.cnt)
		if err != nil {
			t.Errorf("test %d, unexpected error: %v", i, err)
		}
	}

	assert.DeepEqual(t, mustDecodeHex(t, "0d020000000437"), br.Bytes())

	// Test empty bitWriter.
	br = newBitWriter()
	assert.Equal(t, len(br.Bytes()), 0)
}

func TestDecodeHashes(t *testing.T) {
	// These vectors were randomly generated using the server-side Rice
	// compression implementation.
	vectors := []struct {
		input  []string // Hex encoded serialized ThreatEntrySet proto
		output []string // Hex encoded expected hash prefixes
	}{{
		[]string{"0802222308a08fcb6d101c18062218dda588628aad88f883e2421a66384d10bce123dd22030202"},
		[]string{"17f15426", "47ba02b7", "573373a2", "a0c7b20d", "a19edd3e", "d2c60aef", "f1fa25a2"},
	}, {
		[]string{"0802221808f8adb660101c1803220d5fe56d19f084ea8fffeacca708"},
		[]string{"2b9f705e", "8d4e735c", "a085c4f2", "f8960d0c"},
	}, {
		[]string{"0802221408dfea9d4e101c1802220991280dd01f2c9d5301"},
		[]string{"33341993", "5f75c709", "83bfca1d"},
	}, {
		[]string{"0802222008c2ee8cc801101c18052214b0e8007275b8ba8319e39f2d4ea7f1df8b1a1202"},
		[]string{"0f64be45", "42370319", "6650275f", "95ba6fd7", "9aab0322", "db7cbd52"},
	}, {
		[]string{"0802222b08ded89aba03101c1808221fce4f1cc9b81c12c9e0142610815a766d8771f0b282397d6779d2cbe98e2f01"},
		[]string{"17ad48ca", "28471d40", "41e3df44", "45d4d43b", "51643a4b", "5eac4637", "7e261bf6", "beebab7b", "cc9d97ff"},
	}, {
		[]string{"0802221508e9f2f7ee01101c18022209ff93bf27d073d3bb37"},
		[]string{"5bf1e2c7", "69f9dd1d", "c96bdafe"},
	}, {
		[]string{"0802221108d793e4ff0a101c180122052f701aef01"},
		[]string{"58dd71ff", "d709f9af"},
	}}

loop:
	for i, v := range vectors {
		var got []string
		for _, in := range v.input {
			set := &sb.ThreatEntrySet{}
			if err := proto.Unmarshal(mustDecodeHex(t, in), set); err != nil {
				t.Errorf("test %d, unexpected proto.Unmarshal error: %v", i, err)
				continue loop
			}

			hashes, err := DecodeRiceIntegers(set.GetRiceHashes())
			if err != nil {
				t.Errorf("test %d, unexpected decodeHashes error: %v", i, err)
				continue loop
			}
			for _, h := range hashes {
				var buf [4]byte
				binary.LittleEndian.PutUint32(buf[:], h)
				got = append(got, string(buf[:]))
			}
		}

		sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
		want := make([]string, 0, len(v.output))
		for _, h := range v.output {
			want = append(want, string(mustDecodeHex(t, h)[:]))
		}
		assert.DeepEqual(t, got, want)
	}
}

func TestDecodeIndices(t *testing.T) {
	// These vectors were randomly generated using the server-side Rice
	// compression implementation.
	vectors := []struct {
		input  []string // Hex encoded serialized ThreatEntrySet proto
		output []int32  // Expected output indices
	}{{
		[]string{"08022a1c08ac01101c18052213720000c0210000100400001a01006017000000"},
		[]int32{172, 229, 364, 494, 776, 963},
	}, {
		[]string{"08022a22084b101c1807221a34010000110000300500000a0000e0100000a80100007a000000"},
		[]int32{75, 229, 297, 463, 473, 608, 714, 958},
	}, {
		[]string{"08022a1e0823101c18062216f80100800f000050050000c50000c0020000b4020000"},
		[]int32{35, 287, 349, 519, 716, 738, 911},
	}, {
		[]string{"08022a0308e607"},
		[]int32{998},
	}, {
		[]string{"08022a1408c101101c1803220b360300c02b0000b8040000"},
		[]int32{193, 604, 779, 930},
	}, {
		[]string{"08022a23088001101c1807221a140000c000000088050000520000c03200004402000024000000"},
		[]int32{128, 138, 141, 318, 400, 806, 951, 1023},
	}, {
		[]string{"08022a1c088f02101c1805221348000000560000c0010000cd00000002000000"},
		[]int32{271, 307, 651, 707, 912, 928},
	}, {
		[]string{"08022a1808f103101c1804220fde0000c00000006805000085000000"},
		[]int32{497, 608, 611, 784, 917},
	}}

loop:
	for i, v := range vectors {
		var got []uint32
		for _, in := range v.input {
			set := &sb.ThreatEntrySet{}
			if err := proto.Unmarshal(mustDecodeHex(t, in), set); err != nil {
				t.Errorf("test %d, unexpected proto.Unmarshal error: %v", i, err)
				continue loop
			}

			hashes, err := DecodeRiceIntegers(set.GetRiceIndices())
			if err != nil {
				t.Errorf("test %d, unexpected decodeHashes error: %v", i, err)
				continue loop
			}
			got = append(got, hashes...)
		}

		indices := make([]int32, 0, len(got))
		for _, v := range got {
			indices = append(indices, int32(v))
		}

		sort.Slice(indices, func(i, j int) bool { return indices[i] < indices[j] })

		assert.DeepEqual(t, indices, v.output)
	}
}

func TestEncodeIndices(t *testing.T) {
	// These vectors were randomly generated using the server-side Rice
	// compression implementation.
	vectors := []struct {
		output string   // Hex encoded serialized ThreatEntrySet proto
		input  []uint32 // Expected output indices
	}{{
		"08022a1c08ac01101c18052213720000c0210000100400001a01006017000000",
		[]uint32{172, 229, 364, 494, 776, 963},
	}, {
		"08022a22084b101c1807221a34010000110000300500000a0000e0100000a80100007a000000",
		[]uint32{75, 229, 297, 463, 473, 608, 714, 958},
	}, {
		"08022a1e0823101c18062216f80100800f000050050000c50000c0020000b4020000",
		[]uint32{35, 287, 349, 519, 716, 738, 911},
	}, {
		"08022a0308e607",
		[]uint32{998},
	}, {
		"08022a1408c101101c1803220b360300c02b0000b8040000",
		[]uint32{193, 604, 779, 930},
	}, {
		"08022a23088001101c1807221a140000c000000088050000520000c03200004402000024000000",
		[]uint32{128, 138, 141, 318, 400, 806, 951, 1023},
	}, {
		"08022a1c088f02101c1805221348000000560000c0010000cd00000002000000",
		[]uint32{271, 307, 651, 707, 912, 928},
	}, {
		"08022a1808f103101c1804220fde0000c00000006805000085000000",
		[]uint32{497, 608, 611, 784, 917},
	}}

loop:
	for i, v := range vectors {
		set := &sb.ThreatEntrySet{}
		if err := proto.Unmarshal(mustDecodeHex(t, v.output), set); err != nil {
			t.Errorf("test %d, unexpected proto.Unmarshal error: %v", i, err)
			continue loop
		}
		rice, err := EncodeRiceIntegersWithParam(v.input, set.RiceIndices.RiceParameter)
		if err != nil {
			t.Errorf("test %d, unexpected encode error: %v", i, err)
			continue loop
		}

		assert.DeepEqual(t, rice, set.RiceIndices)
	}
}
