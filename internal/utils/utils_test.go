package utils

import (
	"testing"
)

func TestItemInSlice (t *testing.T) {
	yah := "kaka"
	yahSlice := []string{"kaka", "tui", "kea"}
	inSlice, err := ItemInSlice(yah, yahSlice)
	if err != nil {
		t.Errorf("Unexepcted error checking if item in slice: %v", err)
	}
	if inSlice == false {
		t.Errorf("Expected true. Item in slice. item: %v slice: %v", yah, yahSlice)
	}

	yahInt := 3
	yahIntSlice := []int{1,2,3,4}
	inSlice, err = ItemInSlice(yahInt, yahIntSlice)
	if err != nil {
		t.Errorf("Unexepcted error checking if item in slice: %v", err)
	}
	if inSlice == false {
		t.Errorf("Expected true. Item in slice. item: %v slice: %v", yah, yahSlice)
	}

	nah := "gecko"
	nahSlice := []string{"tui,", "kea", "kakapo"}
	inSlice, err = ItemInSlice(nah, nahSlice)
	if err != nil {
		t.Errorf("Unexepcted error checking if item in slice: %v", err)
	}
	if inSlice == true {
		t.Errorf("Expected False. Item not in slice. item: %v slice: %v", nah, nahSlice)
	}
}
