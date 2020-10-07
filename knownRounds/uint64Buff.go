package knownRounds

import (
	jww "github.com/spf13/jwalterweatherman"
	"math"
)

type uint64Buff []uint64

// Get returns the value of the bit at the given position.
func (u64b uint64Buff) get(pos int) bool {
	bin, offset := u64b.convertLoc(pos)

	return u64b[bin]>>(63-offset)&1 == 1
}

// set modifies the bit at the specified position to be 1.
func (u64b uint64Buff) set(pos int) {
	bin, offset := u64b.convertLoc(pos)
	u64b[bin] |= 1 << (63 - offset)
}

// set modifies the bit at the specified position to be 1.
func (u64b uint64Buff) clear(pos int) {
	bin, offset := u64b.convertLoc(pos)
	u64b[bin] &= ^(1 << (63 - offset))
}

// clearRange clears all the bits in the buffer between the given range
// (including the start and end bits).
//
// If start is greater than end, then the selection is inverted.
func (u64b uint64Buff) clearRange(start, end int) {

	// Determine the starting positions the buffer
	numBlocks := u64b.delta(start, end)
	firstBlock, firstBit := u64b.convertLoc(start)

	// Loop over every the blocks in u64b that are in the range
	for blockIndex := 0; blockIndex < numBlocks; blockIndex++ {
		// Get index where the block appears in the buffer
		buffBlock := u64b.getBin(firstBlock + blockIndex)

		// Get the position of the last bit in the current block
		lastBit := 64
		if blockIndex == numBlocks-1 {
			_, lastBit = u64b.convertEnd(end)
		}

		// Generate bit mask for the range and apply it
		bm := bitMaskRange(firstBit, lastBit)
		u64b[buffBlock] &= bm

		// Set position to the first bit in the next block
		firstBit = 0
	}
}

// copy returns a copy of the bits from start to end (inclusive) from u64b.
func (u64b uint64Buff) copy(start, end int) uint64Buff {
	startBlock, startPos := u64b.convertLoc(start)

	numBlocks := u64b.delta(start, end)
	copied := make(uint64Buff, numBlocks)

	// Copy all blocks in range
	for i := 0; i < numBlocks; i++ {
		realBlock := u64b.getBin(startBlock + i)
		copied[i] = u64b[realBlock]
	}

	// Set all bits before the start
	copied[0] |= ^bitMaskRange(0, startPos)

	// Clear all bits after end
	_, endPos := u64b.convertEnd(end)
	copied[numBlocks-1] &= ^bitMaskRange(0, endPos)

	return copied
}

// implies applies the material implication of mask and u64b in the given range
// (including the start and end bits) and places the result in masked starting
// at position maskedStart. An error is returned if the range is larger than the
// length of masked.
//
// If u64bStart is greater than u64bEnd, then the selection is inverted.
//
// More info on material implication:
//   https://en.wikipedia.org/wiki/Material_conditional
func (u64b uint64Buff) implies(mask uint64Buff) uint64Buff {
	if len(u64b) != len(mask) {
		jww.FATAL.Panicf("Cannot imply two buffers of different lengths "+
			"(%v and %v).", len(u64b), len(mask))
	}
	result := make(uint64Buff, len(u64b))

	for i := 0; i < len(u64b); i++ {
		result[i] = ^mask[i] | u64b[i]
	}
	return result
}

// extend increases the length of the buffer to the given size and fills in the
// values with zeros.
func (u64b uint64Buff) extend(numBlocks int) uint64Buff {
	// The created buffer is all zeroes per go spec
	ext := make(uint64Buff, numBlocks)
	copy(ext[:len(u64b)], u64b)
	return ext
}

// convertLoc returns the block index and the position of the bit in that block
// for the given position in the buffer.
func (u64b uint64Buff) convertLoc(pos int) (int, int) {
	// Block index in buffer (position / 64)
	bin := pos >> 6 % len(u64b)

	// Position of bit in block
	offset := pos % 64

	return bin, offset
}

func (u64b uint64Buff) convertEnd(pos int) (int, int) {
	bin := (pos - 1) / 64

	offset := (pos-1)%64 + 1

	return bin, offset
}

// getBin returns the block index in the buffer for the given absolute index.
func (u64b uint64Buff) getBin(block int) int {
	return block % len(u64b)
}

// delta calculates the number of blocks or parts of blocks contained within the
// range between start and end. If the start and end appear in the same block,
// then delta returns 1.
func (u64b uint64Buff) delta(start, end int) int {
	if end == start {
		return 1
	}
	end--
	if end < start {
		return len(u64b) - start/64 + end/64 + 1
	} else {
		return end/64 - start/64 + 1
	}
}

// bitMaskRange generates a bit mask that targets the bits in the provided
// range. The resulting value has 0s in that range and 1s everywhere else.
func bitMaskRange(start, end int) uint64 {
	s := uint64(math.MaxUint64 << uint(64-start))
	e := uint64(math.MaxUint64 >> uint(end))
	return (s | e) & (getInvert(end < start) ^ (s ^ e))
}

func getInvert(b bool) uint64 {
	switch b {
	case true:
		return math.MaxUint64
	default:
		return 0
	}
}
