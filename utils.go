package sapi

import (
	"errors"
)

var (
	ErrorOutOfRange = errors.New("Bit index out of range")
)

type BitMap struct {
	Min, Max uint32
	Bits     []uint8
}

func NewBitmap(min, max uint32) *BitMap {
	return &BitMap{
		Min:  min,
		Max:  max,
		Bits: make([]uint8, (max+8)/8),
	}
}

func (b *BitMap) GetUnusedBit(value *uint32) bool {
	for i := b.Min; i <= b.Max; i++ {
		if !b.bitOn(i) {
			b.Setbit(i)
			*value = i
			return true
		}
	}
	return false
}

func (b *BitMap) UnsetBit(bit uint32) error {
	if bit > b.Max || bit < b.Min {
		return ErrorOutOfRange
	}

	aIndex := b.getArrayIndex(bit)
	bIndex := b.getBitIndex(bit)
	b.Bits[aIndex] &= (^(1 << (8 - bIndex)))

	return nil
}

func (b *BitMap) getArrayIndex(bit uint32) uint32 {
	return (bit - 1) / 8
}

func (b *BitMap) getBitIndex(bit uint32) uint32 {
	index := bit % 8
	if index == 0 {
		index += 8
	}
	return index
}

func (b *BitMap) Setbit(bit uint32) error {
	if bit > b.Max || bit < b.Min {
		return ErrorOutOfRange
	}

	aIndex := b.getArrayIndex(bit)
	bIndex := b.getBitIndex(bit)
	b.Bits[aIndex] |= (1 << (8 - bIndex))

	return nil
}

func (b *BitMap) bitOn(bit uint32) bool {
	aIndex := b.getArrayIndex(bit)
	bIndex := b.getBitIndex(bit)
	v := uint8((1 << (8 - bIndex)))
	v1 := b.Bits[aIndex]

	if (v1 & v) == v {
		return true
	}
	return false
}
