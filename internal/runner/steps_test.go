package runner

import (
	"testing"
)

func TestItemInSlice(t *testing.T) {
	slice := []string{"kaka","tui","kakapo","kea"}
	yah := "tui"
	nah := "gecko"

	if v := itemInSlice(yah, slice); v != true {
		t.Errorf("Item in slice, but fn returned %v", v)

	}
	if v := itemInSlice(nah, slice); v != false {
		t.Errorf("Item not in slice, but fn returned %v", v)
	}
}


func TestStringsMatch(t *testing.T)  {
	example := "kakapo"
	yah := "kakapo"
	nah := "kea"

	if match := stringsMatch(example, yah); match != true {
		t.Errorf("fn says these strings don't match: %v, %v", example, yah)
	}
	if match := stringsMatch(example, nah); match != false {
		t.Errorf("fn says these strings match: %v, %v", example, yah)
	}
}

func TestResourcesMatch(t *testing.T) {
	example := []string{"kaka","tui","kakapo"}
	yah := []string{"kakapo", "tui", "kaka"}
	// actual response can have more than what's expected and still be valid.
	yah2 := []string{"kakapo", "kaka","kakapo","kea", "tui","takahe"}
	nah := []string{}
	nah2 := []string{"gecko"}
	nah3 := []string{"tui", "gecko", "kakapo"}

	if match := resourcesMatch(example, yah); match != true {
		t.Errorf("This is a valid resource match, but returning false: %v %v", example, yah)
	}
	if match := resourcesMatch(example, yah2); match != true {
		t.Errorf("This is a valid resource match, but returning false: %v %v", example, yah2)
	}
	if match := resourcesMatch(example, nah); match != false {
		t.Errorf("This is not a valid resource match, but returning true: %v %v", example, nah)
	}
	if match := resourcesMatch(example, nah2); match != false {
		t.Errorf("This is not a valid resource match, but returning true: %v %v", example, nah)
	}
	if match := resourcesMatch(example, nah3); match != false {
		t.Errorf("This is not a valid resource match, but returning true: %v %v", example, nah)
	}
}
