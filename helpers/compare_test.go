package helpers

import (
	"testing"
)

func TestConstantTimeBytesEqual(t *testing.T) {
	if ConstantTimeBytesEqual([]byte("a"), []byte("b")) {
		t.Fatal("a == b")
	}

	if ConstantTimeBytesEqual([]byte("a"), []byte("aa")) {
		t.Fatal("a == aa")
	}

	if !ConstantTimeBytesEqual([]byte("a"), []byte("a")) {
		t.Fatal("a != a")
	}
}
