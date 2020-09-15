package knownRounds

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

const max = math.MaxUint64

// Happy path of get().
func Test_uint64Buff_get(t *testing.T) {
	// Generate test positions and expected value
	testData := []struct {
		pos   int
		value bool
	}{
		{0, false},
		{64, true},
		{319, false},
		{320, false},
	}
	u64b := uint64Buff{0, max, 0, max, 0}

	for i, data := range testData {
		value := u64b.get(data.pos)
		if value != data.value {
			t.Errorf("get() returned incorrect value for bit at position %d (round %d)."+
				"\n\texpected: %v\n\treceived: %v", data.pos, i, data.value, value)
		}
	}
}

// Happy path of set().
func Test_uint64Buff_set(t *testing.T) {
	// Generate test positions and expected buffers
	testData := []struct {
		pos  int
		buff uint64Buff
	}{
		{0, uint64Buff{0x8000000000000000, max, 0, max, 0}},
		{64, uint64Buff{0, max, 0, max, 0}},
		{320, uint64Buff{0x8000000000000000, max, 0, max, 0}},
		{15, uint64Buff{0x1000000000000, max, 0, max, 0}},
	}

	for i, data := range testData {
		u64b := uint64Buff{0, max, 0, max, 0}
		u64b.set(data.pos)
		if !reflect.DeepEqual(u64b, data.buff) {
			t.Errorf("Resulting buffer after setting bit at position %d (round %d)."+
				"\n\texpected: %064b\n\treceived: %064b"+
				"\n\033[38;5;59m               0123456789012345678901234567890123456789012345678901234567890123 4567890123456789012345678901234567890123456789012345678901234567 8901234567890123456789012345678901234567890123456789012345678901 2345678901234567890123456789012345678901234567890123456789012345 6789012345678901234567890123456789012345678901234567890123456789 0123456789012345678901234567890123456789012345678901234567890123"+
				"\n\u001B[38;5;59m               0         1         2         3         4         5         6          7         8         9         0         1         2          3         4         5         6         7         8         9          0         1         2         3         4         5          6         7         8         9         0         1          2         3         4         5         6         7         8"+
				"\n\u001B[38;5;59m               0         0         0         0         0         0         0          0         0         0         1         1         1          1         1         1         1         1         1         1          2         2         2         2         2         2          2         2         2         2         3         3          3         3         3         3         3         3         3",
				data.pos, i, data.buff, u64b)
		}
	}
}

// Tests that clearRange() clears the correct bits.
func Test_uint64Buff_clearRange(t *testing.T) {
	// Generate test ranges and expected buffer
	testData := []struct {
		start, end int
		buff       uint64Buff
	}{
		{0, 63, uint64Buff{1, max, max, max, max}},
		{0, 64, uint64Buff{0, max, max, max, max}},
		{0, 65, uint64Buff{0, 0x7FFFFFFFFFFFFFFF, max, max, max}},
		{0, 319, uint64Buff{0, 0, 0, 0, 1}},
		{0, 320, uint64Buff{0, 0, 0, 0, 0}},
		{1, 318, uint64Buff{0x8000000000000000, 0, 0, 0, 3}},
		{1, 330, uint64Buff{0, 0, 0, 0, 0}},
		{0, 1200, uint64Buff{0, 0, 0, 0, 0}},
		{0, 400, uint64Buff{0, 0, 0, 0, 0}},
		{36, 354, uint64Buff{0x30000000, 0, 0, 0, 0}},
		{0, 0, uint64Buff{0, 0, 0, 0, 0}},
		{0, 1, uint64Buff{0x7FFFFFFFFFFFFFFF, max, max, max, max}},
		{5, 27, uint64Buff{0xF800001FFFFFFFFF, max, max, max, max}},
		{5, 110, uint64Buff{0xF800000000000000, 0x3FFFF, max, max, max}},
		{310, 5, uint64Buff{0x7FFFFFFFFFFFFFF, max, max, max, 0xFFFFFFFFFFFFFC00}},
	}

	for i, data := range testData {
		u64b := uint64Buff{max, max, max, max, max}
		u64b.clearRange(data.start, data.end)
		if !reflect.DeepEqual(u64b, data.buff) {
			t.Errorf("Resulting buffer after clearing range %d to %d is incorrect (round %d)."+
				"\n\texpected: %064b\n\treceived: %064b"+
				"\n\033[38;5;59m               0123456789012345678901234567890123456789012345678901234567890123 4567890123456789012345678901234567890123456789012345678901234567 8901234567890123456789012345678901234567890123456789012345678901 2345678901234567890123456789012345678901234567890123456789012345 6789012345678901234567890123456789012345678901234567890123456789 0123456789012345678901234567890123456789012345678901234567890123"+
				"\n\u001B[38;5;59m               0         1         2         3         4         5         6          7         8         9         0         1         2          3         4         5         6         7         8         9          0         1         2         3         4         5          6         7         8         9         0         1          2         3         4         5         6         7         8"+
				"\n\u001B[38;5;59m               0         0         0         0         0         0         0          0         0         0         1         1         1          1         1         1         1         1         1         1          2         2         2         2         2         2          2         2         2         2         3         3          3         3         3         3         3         3         3",
				data.start, data.end, i, data.buff, u64b)
		}
	}
}

// Tests that copy() copies the correct bits.
func Test_uint64Buff_copy(t *testing.T) {
	// Generate test ranges and expected copied value

	const numTests = 100
	const maxBuffSize = 100

	prng := rand.New(rand.NewSource(42))

	for i := 0; i < numTests; i++ {
		lenBuf := 0
		for lenBuf == 0 {
			lenBuf = int(prng.Uint32() % maxBuffSize)
		}

		buf := make(uint64Buff, lenBuf)
		for j := 0; j < lenBuf; j++ {
			buf[j] = prng.Uint64()
		}
		subsampleStart, subsampleEnd := 0, 0
		for subsampleEnd-subsampleStart == 0 {
			subsampleStart = int(prng.Uint64() % uint64(lenBuf*64))
			subsampleDelta := int(prng.Uint64() % (uint64(lenBuf*64 - subsampleStart)))
			subsampleEnd = subsampleStart + subsampleDelta
		}

		//delta := subsampleEnd-subsampleStart

		copied := buf.copy(subsampleStart, subsampleEnd)

		//check edge regions
		for j := 0; j < subsampleStart%64; j++ {
			if !copied.get(j) {
				t.Errorf("Round %v position %v < substampeStart %v(%v) is "+
					"false when should be true", i, j, subsampleStart, subsampleStart%64)
			}
		}
		// dont test the edge case where the last element is the last in the
		// last block because nothing will have been filled in to test
		if (subsampleEnd/64 - subsampleStart/64) != len(copied) {
			for j := subsampleEnd % 64; j < 64; j++ {
				if copied.get(((len(copied) - 1) * 64) + j) {
					t.Errorf("Round %v position %v (%v) > substampeEnd %v(%v) is "+
						"true when should be false", i, ((len(copied)-1)*64)+j, j,
						subsampleEnd, subsampleEnd%64)
				}
			}
		}
		//check all in between bits are correct
		for j := subsampleStart % 64; j < subsampleEnd-subsampleStart; j++ {
			if copied.get(j) != buf.get(j+(subsampleStart/64)*64) {
				t.Errorf("Round %v copy position %v not the same as original"+
					" position %v (%v + %v)", i, j%64, (j+subsampleStart)%64,
					subsampleStart, j)
			}
		}
	}
}

// Happy path of convertLoc().
func Test_uint64Buff_convertLoc(t *testing.T) {
	// Generate test position and expected block index and offset
	testData := []struct {
		pos         int
		bin, offset uint
	}{
		{0, 0, 0},
		{5, 0, 5},
		{63, 0, 63},
		{64, 1, 0},
		{127, 1, 63},
		{128, 2, 0},
		{319, 4, 63},
		{320, 0, 0},
	}

	u64b := uint64Buff{0, 0, 0, 0, 0}

	for i, data := range testData {
		bin, offset := u64b.convertLoc(data.pos)
		if bin != data.bin || offset != data.offset {
			t.Errorf("convert() returned incorrect values for position %d "+
				"(round %d).\n\texpected: bin: %3d  offset: %3d"+
				"\n\treceived: bin: %3d  offset: %3d",
				data.pos, i, data.bin, data.offset, bin, offset)
		}
	}
}

// Happy path of convertEnd().
func Test_uint64Buff_convertEnd(t *testing.T) {
	// Generate test position and expected block index and offset
	testData := []struct {
		pos         int
		bin, offset uint
	}{
		{0, 0, 0},
		{5, 0, 5},
		{63, 0, 63},
		{64, 0, 64},
		{65, 1, 1},
		{127, 1, 63},
		{128, 1, 64},
		{319, 4, 63},
		{320, 4, 64},
	}

	u64b := uint64Buff{0, 0, 0, 0, 0}

	for i, data := range testData {
		bin, offset := u64b.convertEnd(data.pos)
		if bin != data.bin || offset != data.offset {
			t.Errorf("convert() returned incorrect values for position %d "+
				"(round %d).\n\texpected: bin: %3d  offset: %3d"+
				"\n\treceived: bin: %3d  offset: %3d",
				data.pos, i, data.bin, data.offset, bin, offset)
		}
	}
}

// Tests happy path of getBin().
func Test_uint64Buff_getBin(t *testing.T) {
	// Generate test block indexes and the expected index in the buffer
	testData := []struct {
		block       uint
		expectedBin uint
	}{
		{0, 0},
		{4, 4},
		{5, 0},
		{15, 0},
		{82, 2},
	}

	u64b := uint64Buff{0, 0, 0, 0, 0}
	for i, data := range testData {
		bin := u64b.getBin(data.block)
		if bin != data.expectedBin {
			t.Errorf("getBin() returned incorrect block index for index %d "+
				"(round %d).\n\texpected: %d\n\treceived: %d",
				data.block, i, data.expectedBin, bin)
		}
	}
}

// Tests that delta() returns the correct delta for the given range.
func Test_uint64Buff_delta(t *testing.T) {
	// Generate test ranges and the expected delta
	testData := []struct {
		start, end    int
		expectedDelta uint
	}{
		{0, 0, 6},
		{5, 5, 6},
		{170, 170, 6},
		{670, 670, 6},
		{63, 64, 1},
		{0, 63, 1},
		{0, 64, 1},
		{0, 65, 2},
		{5, 35, 1},
		{5, 75, 2},
		{0, 75, 2},
		{0, 319, 5},
		{0, 400, 7},
		{35, 354, 6},
		{63, 354, 6},
		{45, 5, 6},
		{130, 5, 4},
		{230, 65, 4},
		{310, 5, 2},
		{310, 64, 2},
		{310, 65, 3},
	}

	u64b := uint64Buff{max, max, max, max, max}

	for i, data := range testData {
		delta := u64b.delta(data.start, data.end)
		if delta != data.expectedDelta {
			t.Errorf("delta() returned incorrect value for range %d to %d (round %d)."+
				"\n\texpected: %d\n\treceived: %d",
				data.start, data.end, i, data.expectedDelta, delta)
		}
	}
}

// Tests that bitMaskRange() produces the correct bit mask for the range.
func Test_bitMaskRange(t *testing.T) {
	// Generate test ranges and the expected mask
	testData := []struct {
		start, end   uint
		expectedMask uint64
	}{
		{0, 0, 0b1111111111111111111111111111111111111111111111111111111111111111},
		{63, 63, 0b1111111111111111111111111111111111111111111111111111111111111111},
		{0, 65, 0b0000000000000000000000000000000000000000000000000000000000000000},
		{5, 25, 0b1111100000000000000000000111111111111111111111111111111111111111},
		{15, 15, 0b1111111111111111111111111111111111111111111111111111111111111111},
		{32, 62, 0b1111111111111111111111111111111100000000000000000000000000000011},
		{62, 32, 0b0000000000000000000000000000000011111111111111111111111111111100},
		{62, 32, 0b0000000000000000000000000000000011111111111111111111111111111100},
		{5, 65, 0b1111100000000000000000000000000000000000000000000000000000000000},
		{75, 85, 0b0000000000000000000000000000000000000000000000000000000000000000},
		{65, 65, 0b0000000000000000000000000000000000000000000000000000000000000000},
	}

	for _, data := range testData {
		testMask := bitMaskRange(data.start, data.end)
		if testMask != data.expectedMask {
			t.Errorf("Generated mask for range %d to %d is incorrect."+
				"\n\texpected: %064b\n\treceived: %064b"+
				"\n              0123456789012345678901234567890123456789012345678901234567890123"+
				"\n              0         1         2         3         4         5         6",
				data.start, data.end, data.expectedMask, testMask)
		}
	}
}

// printBuff prints the buffer and mask in binary with their start and end point
// labeled.
func printBuff(buff, mask uint64Buff, buffStart, buffEnd, maskStart, maskEnd int) {
	fmt.Printf("\n\u001B[38;5;59m         0         0         0         0         0         0         0          0         0         0         1         1         1          1         1         1         1         1         1         1          2         2         2         2         2         2          2         2         2         2         3         3          3         3         3         3         3         3         3" +
		"\n\u001B[38;5;59m         0         1         2         3         4         5         6          7         8         9         0         1         2          3         4         5         6         7         8         9          0         1         2         3         4         5          6         7         8         9         0         1          2         3         4         5         6         7         8" +
		"\n\033[38;5;59m         0123456789012345678901234567890123456789012345678901234567890123 4567890123456789012345678901234567890123456789012345678901234567 8901234567890123456789012345678901234567890123456789012345678901 2345678901234567890123456789012345678901234567890123456789012345 6789012345678901234567890123456789012345678901234567890123456789 01234567890123456789012345678901234567890123456789012345678901\n")
	fmt.Printf("        %*s%*s\n", buffStart+2+buffStart/64, "S", (buffEnd+2+buffEnd/64)-(buffStart+2+buffStart/64), "E")
	fmt.Printf("buff:   %064b\n", buff)
	fmt.Printf("mask:   %064b\n", mask)
	fmt.Printf("        %*s%*s\n", maskStart+2+maskStart/64, "S", (maskEnd+2+maskEnd/64)-(maskStart+2+maskStart/64), "E")
}