package core

import "testing"

func TestGethash(t *testing.T) {
	node := Node{BlockStart: 0, BlockEnd: 0, BlockSum: "AA", PrevNodeSum: "hello"}
	got, err := node.GetHash()
	want := "87cdc950224b3850667c0f6a907a5c0dcf047425"

	if got != want || err != nil {
		t.Errorf("got %s, wanted %s", got, want)
	}
}

func TestParseUnitSpec(t *testing.T) {
	got := parseUnitSpec([]byte{0x0})
	want := 1

	if got != want {
		t.Errorf("got %d, wanted %d", got, want)
	}
}
