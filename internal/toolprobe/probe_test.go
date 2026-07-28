package toolprobe

import "testing"

func TestRunIsDeterministic(t *testing.T) {
	a, err := Run(8, "full", 100, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Run(8, "full", 100, 2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if a.Correct != b.Correct || a.Total != b.Total {
		t.Fatalf("probe changed across identical runs: %+v vs %+v", a, b)
	}
}
