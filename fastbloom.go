package fastbloom

import (
	"unsafe"

	"github.com/zeebo/xxh3"
)

type FastBloom struct {
	slotData []uint64
	slotCount int
	probePerEntry int
	seed0 [8]byte
	seed1 [8]byte
}

const (
	SlotByteCount = 64 //Intel CPUs have 64-byte cache lines
	SlotBitCount = SlotByteCount * 8
	SlotWordCount = SlotByteCount / 8
)

func NewFastBloom(slotCount, probePerEntry int, seed [8]byte) *FastBloom {
	bf := &FastBloom{
		slotData: make([]uint64, slotCount*SlotWordCount),
		slotCount: slotCount,
		probePerEntry: probePerEntry,
		seed0: seed,
		seed1: seed,
	}
	bf.seed1[0] = ^bf.seed1[0] // make it different from seed0
	ptr := uintptr(unsafe.Pointer(&bf.slotData[0]))
	if ptr%8 != 0 {
		panic("Not 8-byte aligned")
	}
	unalignedWords := (ptr/8)%SlotWordCount
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
	bitOffsetInSlot := int(hash.Lo % SlotBitCount) //Lo's lowest 9 bits as first bitOffsetInSlot
	slotIdx := int(hash.Lo/SlotBitCount) % bf.slotCount //consume Lo's high bits
	currSlotOffset := SlotWordCount*slotIdx
	probeCount := 1
	for {
		wordIdx, bitOffsetInWord := bitOffsetInSlot/64, bitOffsetInSlot%64
		mask := uint64(1) << bitOffsetInWord;
		if isAdd {
			bf.slotData[currSlotOffset+wordIdx] |= mask
		} else if (bf.slotData[currSlotOffset+wordIdx]&mask) == 0 {
			return false // do not has it
		}

		probeCount++
		if probeCount == bf.probePerEntry {
			break
		}
		bitOffsetInSlot = int(hash.Hi%SlotBitCount) //consume bits in high64 as bit_offset
		if probeCount == 8 { //generate more random bits
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

func GetOptimalParams(targetFalsePositiveRatio float64, entryCount int) (slotCount, probePerEntry int) {
	bitsPerEntry := 25
	probePerEntry = 14
	if targetFalsePositiveRatio > 0.023220 {
		bitsPerEntry=8;  probePerEntry=6;
	} else if targetFalsePositiveRatio > 0.014895 {
		bitsPerEntry=9;  probePerEntry=7;
	} else if targetFalsePositiveRatio > 0.009646 {
		bitsPerEntry=10; probePerEntry=8;
	} else if targetFalsePositiveRatio > 0.006303 {
		bitsPerEntry=11; probePerEntry=8;
	} else if targetFalsePositiveRatio > 0.004118 {
		bitsPerEntry=12; probePerEntry=9;
	} else if targetFalsePositiveRatio > 0.002753 {
		bitsPerEntry=13; probePerEntry=9;
	} else if targetFalsePositiveRatio > 0.001856 {
		bitsPerEntry=14; probePerEntry=10;
	} else if targetFalsePositiveRatio > 0.001236 {
		bitsPerEntry=15; probePerEntry=10;
	} else if targetFalsePositiveRatio > 0.000841 {
		bitsPerEntry=16; probePerEntry=10;
	} else if targetFalsePositiveRatio > 0.000575 {
		bitsPerEntry=17; probePerEntry=11;
	} else if targetFalsePositiveRatio > 0.000377 {
		bitsPerEntry=18; probePerEntry=11;
	} else if targetFalsePositiveRatio > 0.000271 {
		bitsPerEntry=19; probePerEntry=12;
	} else if targetFalsePositiveRatio > 0.000206 {
		bitsPerEntry=20; probePerEntry=13;
	} else if targetFalsePositiveRatio > 0.000134 {
		bitsPerEntry=21; probePerEntry=13;
	} else if targetFalsePositiveRatio > 0.000108 {
		bitsPerEntry=22; probePerEntry=13;
	} else if targetFalsePositiveRatio > 0.000068 {
		bitsPerEntry=23; probePerEntry=13;
	} else if targetFalsePositiveRatio > 0.000050 {
		bitsPerEntry=24; probePerEntry=13;
	}
	slotCount = (bitsPerEntry * entryCount + SlotBitCount - 1) / SlotBitCount;
	return
}
