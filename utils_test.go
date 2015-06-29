package sapi

import (
	"testing"
)

func TestOutOfRange(t *testing.T) {
	bitmap := NewBitmap(2, 4094)
	var value, i uint32

	i = 2
	for bitmap.GetUnusedBit(&value) {
		if value != i {
			t.Errorf("Expected value %d, but got %d", i, value)
		}
		i++
	}
	for i = 2; i <= 4094; i++ {
		if err := bitmap.UnsetBit(i); err != nil {
			t.Errorf("Unexpected error %s", err)
		}
	}
	err := bitmap.UnsetBit(50000)
	if err != ErrorOutOfRange {
		t.Errorf("Expected error %s, got %s", ErrorOutOfRange, err)
	}
}

func TestGetSet(t *testing.T) {
	bitmap := NewBitmap(452, 4000)
	var value uint32

	if !bitmap.GetUnusedBit(&value) {
		t.Errorf("Cant got value out of bitmap.")
	}

	if value != 452 {
		t.Errorf("Expected value %d, but got %d", 452, value)
	}

	bitmap.UnsetBit(452)
	bitmap.GetUnusedBit(&value)
	bitmap.GetUnusedBit(&value)

	if value != 453 {
		t.Errorf("Expected value %d, but got %d", 453, value)
	}
}
