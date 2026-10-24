package eventlog

import (
	"testing"
	"time"
)

const goodCommit = "9dd7ebf4c2a1e8d3b5f6a7c8d9e0f1a2b3c4d5e6"
const goodLine = `{"version":1,"seq":1,"event":"slice.created","at":"2026-10-24T07:42:00Z","data":{"from":{"branch":"main","commit":"` + goodCommit + `"}}}`

func TestEncodeCreated(t *testing.T) {
	at := time.Date(2026, 10, 24, 7, 42, 0, 0, time.UTC)
	from := From{Branch: "main", Commit: goodCommit}
	got, err := Encode(NewCreated(at, from, ""))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != goodLine+"\n" {
		t.Errorf("got %q", got)
	}
	got, _ = Encode(NewCreated(at, from, "Login with Google"))
	want := goodLine[:len(goodLine)-2] + `,"title":"Login with Google"}}` + "\n"
	if string(got) != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestEncodeLocalTimeIsUTC(t *testing.T) {
	at := time.Date(2026, 10, 24, 9, 42, 0, 0, time.FixedZone("x", 2*3600))
	got, _ := Encode(NewCreated(at, From{Branch: "main", Commit: goodCommit}, ""))
	if string(got) != goodLine+"\n" {
		t.Errorf("got %q", got)
	}
}
