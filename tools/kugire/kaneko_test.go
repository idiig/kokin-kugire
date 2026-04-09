package kugire

import (
	"strings"
	"testing"
)

func TestLoadKanekoData(t *testing.T) {
	input := strings.NewReader(`$$DATA|

$A|000001
$B|年内に思ひがけず…
$D|年内に思い掛けず春は来たことであるわ、さてはこの同じ一年の内の昨日までを、去年と言おうか、それとも今年と言おうか。
$I|ねんないに…
$Z|2003/09/25

$A|000002
$D|袖が濡れるほど…
$Z|2003/09/25

`)

	data, err := loadKanekoData(input)
	if err != nil {
		t.Fatalf("loadKanekoData error: %v", err)
	}
	if len(data) != 2 {
		t.Errorf("len: got %d, want 2", len(data))
	}
	got, ok := data[1]
	if !ok {
		t.Fatal("poem 1 not found")
	}
	want := "年内に思い掛けず春は来たことであるわ、さてはこの同じ一年の内の昨日までを、去年と言おうか、それとも今年と言おうか。"
	if got != want {
		t.Errorf("poem 1:\n got  %q\n want %q", got, want)
	}
	if _, ok := data[2]; !ok {
		t.Error("poem 2 not found")
	}
}

func TestLoadKanekoData_skipHeader(t *testing.T) {
	// Lines before $$DATA| must be ignored
	input := strings.NewReader(`$$DB_ID|KA
$$DB_NAME|test
$$DATA|

$A|000003
$D|春霞立てるやいづこ。
$Z|2003/09/25

`)
	data, err := loadKanekoData(input)
	if err != nil {
		t.Fatalf("loadKanekoData error: %v", err)
	}
	if len(data) != 1 {
		t.Errorf("len: got %d, want 1", len(data))
	}
	if data[3] != "春霞立てるやいづこ。" {
		t.Errorf("poem 3: %q", data[3])
	}
}

func TestLoadKanekoData_missingD(t *testing.T) {
	// An entry without $D should be silently skipped
	input := strings.NewReader(`$$DATA|

$A|000005
$B|some note
$Z|2003/09/25

`)
	data, err := loadKanekoData(input)
	if err != nil {
		t.Fatalf("loadKanekoData error: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty map, got %d entries", len(data))
	}
}
