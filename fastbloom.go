package fastbloom

import (
	"unsafe"

	"github.com/zeebo/xxh3"
)

type FastBloom struct {
	slotData      []uint64
	slotCount     int
	probePerEntry int
	seed0         [8]byte
	seed1         [8]byte
}

const (
	SlotByteCount = 64 //Intel CPUs have 64-byte cache lines
	SlotBitCount  = SlotByteCount * 8
	SlotWordCount = SlotByteCount / 8
)

func NewFastBloom(slotCount, probePerEntry int, seed [8]byte) *FastBloom {
	bf := &FastBloom{
		slotData:      make([]uint64, slotCount*SlotWordCount),
		slotCount:     slotCount,
		probePerEntry: probePerEntry,
		seed0:         seed,
		seed1:         seed,
	}
	bf.seed1[0] = ^bf.seed1[0] // make it different from seed0
	ptr := uintptr(unsafe.Pointer(&bf.slotData[0]))
	if ptr%8 != 0 {
		panic("Not 8-byte aligned")
	}
	unalignedWords := (ptr / 8) % SlotWordCount
	if unalignedWords != 0 { // enforce cache-line alignment
		bf.slotData = bf.slotData[SlotWordCount-unalignedWords:]
		bf.slotCount--
	}
	return bf
}

func (bf *FastBloom) Reset() {
	for i := range bf.slotData {
		bf.slotData[i] = 0
	}
}

func (bf *FastBloom) op(data []byte, isAdd bool) bool {
	hash := xxh3.Hash128(append(bf.seed0[:], data...))
	bitOffsetInSlot := int(hash.Lo % SlotBitCount)      //Lo's lowest 9 bits as first bitOffsetInSlot
	slotIdx := int(hash.Lo/SlotBitCount) % bf.slotCount //consume Lo's high bits
	currSlotOffset := SlotWordCount * slotIdx
	probeCount := 1
	for {
		wordIdx, bitOffsetInWord := bitOffsetInSlot/64, bitOffsetInSlot%64
		mask := uint64(1) << bitOffsetInWord
		if isAdd {
			bf.slotData[currSlotOffset+wordIdx] |= mask
		} else if (bf.slotData[currSlotOffset+wordIdx] & mask) == 0 {
			return false // do not has it
		}

		probeCount++
		if probeCount == bf.probePerEntry {
			break
		}
		bitOffsetInSlot = int(hash.Hi % SlotBitCount) //consume bits in high64 as bit_offset
		if probeCount == 8 {                          //generate more random bits
			hash = xxh3.Hash128(append(bf.seed1[:], data...))
		} else if probeCount == 15 { //copy random bits from low64 to high64 for future use
			hash.Hi = hash.Lo
		} else { //shift out the consumed bits
			hash.Hi = hash.Hi / SlotBitCount
		}
	}
	return true
}

func (bf *FastBloom) Add(data []byte) {
	_ = bf.op(data, true)
}

func (bf *FastBloom) Has(data []byte) bool {
	return bf.op(data, false)
}

func GetOptimalParams(entryCount int, targetFalsePositiveRatio float64) (slotCount, probePerEntry int) {
	bitsPerEntry := 25
	probePerEntry = 14
	if targetFalsePositiveRatio > 0.023220 {
		bitsPerEntry, probePerEntry = 8, 6
	} else if targetFalsePositiveRatio > 0.014895 {
		bitsPerEntry, probePerEntry = 9, 7
	} else if targetFalsePositiveRatio > 0.009646 {
		bitsPerEntry, probePerEntry = 10, 8
	} else if targetFalsePositiveRatio > 0.006303 {
		bitsPerEntry, probePerEntry = 11, 8
	} else if targetFalsePositiveRatio > 0.004118 {
		bitsPerEntry, probePerEntry = 12, 9
	} else if targetFalsePositiveRatio > 0.002753 {
		bitsPerEntry, probePerEntry = 13, 9
	} else if targetFalsePositiveRatio > 0.001856 {
		bitsPerEntry, probePerEntry = 14, 10
	} else if targetFalsePositiveRatio > 0.001236 {
		bitsPerEntry, probePerEntry = 15, 10
	} else if targetFalsePositiveRatio > 0.000841 {
		bitsPerEntry, probePerEntry = 16, 10
	} else if targetFalsePositiveRatio > 0.000575 {
		bitsPerEntry, probePerEntry = 17, 11
	} else if targetFalsePositiveRatio > 0.000377 {
		bitsPerEntry, probePerEntry = 18, 11
	} else if targetFalsePositiveRatio > 0.000271 {
		bitsPerEntry, probePerEntry = 19, 12
	} else if targetFalsePositiveRatio > 0.000206 {
		bitsPerEntry, probePerEntry = 20, 13
	} else if targetFalsePositiveRatio > 0.000134 {
		bitsPerEntry, probePerEntry = 21, 13
	} else if targetFalsePositiveRatio > 0.000108 {
		bitsPerEntry, probePerEntry = 22, 13
	} else if targetFalsePositiveRatio > 0.000068 {
		bitsPerEntry, probePerEntry = 23, 13
	} else if targetFalsePositiveRatio > 0.000050 {
		bitsPerEntry, probePerEntry = 24, 13
	} else if targetFalsePositiveRatio > 0.000037 {
		bitsPerEntry, probePerEntry = 25, 13
	} else if targetFalsePositiveRatio > 0.000024 {
		bitsPerEntry, probePerEntry = 26, 13
	} else if targetFalsePositiveRatio > 0.000020 {
		bitsPerEntry, probePerEntry = 27, 14
	} else if targetFalsePositiveRatio > 0.000014 {
		bitsPerEntry, probePerEntry = 28, 15
	} else if targetFalsePositiveRatio > 0.000011 {
		bitsPerEntry, probePerEntry = 29, 15
	} else if targetFalsePositiveRatio > 0.000009 {
		bitsPerEntry, probePerEntry = 30, 15
	} else if targetFalsePositiveRatio > 0.000007 {
		bitsPerEntry, probePerEntry = 31, 15
	} else if targetFalsePositiveRatio > 0.000005 {
		bitsPerEntry, probePerEntry = 32, 15
	} else if targetFalsePositiveRatio > 0.000004 {
		bitsPerEntry, probePerEntry = 33, 15
	} else if targetFalsePositiveRatio > 0.000003 {
		bitsPerEntry, probePerEntry = 34, 15
	} else if targetFalsePositiveRatio > 0.000002 {
		bitsPerEntry, probePerEntry = 35, 15
	} else if targetFalsePositiveRatio > 0.000001 {
		bitsPerEntry, probePerEntry = 36, 16
	}
	slotCount = (bitsPerEntry*entryCount + SlotBitCount - 1) / SlotBitCount
	return
}
