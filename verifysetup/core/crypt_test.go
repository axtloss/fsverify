package core

import "testing"

func TestCalculateBlockHash(t *testing.T) {
	got, err := CalculateBlockHash([]byte("hello"))
	want := "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"

	if got != want || err != nil {
		t.Errorf("got %s, wanted %s", got, want)
	}
}
