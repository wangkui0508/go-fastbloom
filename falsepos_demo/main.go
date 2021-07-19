package main

import (
	"fmt"
	"os"

	"github.com/zeebo/xxh3"
	"github.com/coinexchain/randsrc"

	"github.com/wangkui0508/fastbloom"
)

type BloomFilter struct {
	data  []bool
	probePerEntry int
}

func NewBloomFilter(size, probePerEntry int) *BloomFilter {
	return &BloomFilter{
		data: make([]bool, size),
		probePerEntry: probePerEntry,
	}
}

func (bf *BloomFilter) Reset() {
	for i := range bf.data {
		bf.data[i] = false
	}
}

func (bf *BloomFilter) op(in []byte, isAdd bool) bool {
	for i := 0; i < bf.probePerEntry; i++ {
		hash := xxh3.Hash(append([]byte{byte(i)}, in...))
		offset := int(hash>>1)%len(bf.data)
		if isAdd {
			bf.data[offset] = true
		} else if !bf.data[offset] {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) Add(in []byte) {
	bf.op(in, true)
}

func (bf *BloomFilter) Has(in []byte) bool {
	return bf.op(in, false)
}


func CheckParam(rs randsrc.RandSrc, bitsPerEntry, probePerEntry int) {
	const entryCount = 2500000;
	bfRef := NewBloomFilter(bitsPerEntry*entryCount, probePerEntry)
	slotCount := (bitsPerEntry*entryCount + fastbloom.SlotBitCount - 1) / fastbloom.SlotBitCount;
	var seed [8]byte
	bf := fastbloom.NewFastBloom(slotCount, probePerEntry, seed)
	for i := 0; i < entryCount; i++ {
		bz := rs.GetBytes(32)
		bfRef.Add(bz)
		bf.Add(bz)
	}
	var bfRefPosCount, bfPosCount int
	for i := 0; i < entryCount; i++ {
		bz := rs.GetBytes(32)
		if bfRef.Has(bz) {
			bfRefPosCount++
		}
		if bf.Has(bz) {
			bfPosCount++
		}
	}
	fmt.Printf("bitsPerEntry=%d probePerEntry=%d traditional-bloom: %f fastbloom: %f\n", bitsPerEntry, probePerEntry,
		float64(bfRefPosCount)/float64(entryCount), float64(bfPosCount)/float64(entryCount))
}

func main() {
	randFilename := os.Getenv("RANDFILE")
	if len(randFilename) == 0 {
		fmt.Printf("No RANDFILE specified. Exiting...")
		return
	}
	rs := randsrc.NewRandSrcFromFileWithSeed(randFilename, []byte{0})
	for bitsPerEntry := 8; bitsPerEntry <= 24; bitsPerEntry++  {
		for probePerEntry := 3; probePerEntry <= bitsPerEntry; probePerEntry++ {
			CheckParam(rs, bitsPerEntry, probePerEntry)
		}
	}
}

