package dockerfile

import "testing"

func TestInstructionHelpers(t *testing.T) {
	instruction, rest, ok := splitInstruction("env NAME=value")
	if !ok || instruction != "ENV" || rest != "NAME=value" {
		t.Fatalf("instruction/rest/ok = %q/%q/%v", instruction, rest, ok)
	}
	if name, value := splitKeyValue(`NAME="value"`); name != "NAME" || value != "value" {
		t.Fatalf("key value = %q/%q", name, value)
	}
	if base, alias := parseFrom("alpine AS final"); base != "alpine" || alias != "final" {
		t.Fatalf("from = %q/%q", base, alias)
	}
	copyInstruction := parseCopy("--chmod=755 --from=build /out/app /app")
	if copyInstruction.From != "build" || copyInstruction.Src != "/out/app" || copyInstruction.Dst != "/app" {
		t.Fatalf("copy = %#v", copyInstruction)
	}
}

func TestSortedIntsDoesNotMutateInput(t *testing.T) {
	in := []int{3, 1, 2}
	got := SortedInts(in)
	if got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("sorted = %#v", got)
	}
	if in[0] != 3 {
		t.Fatalf("input mutated: %#v", in)
	}
}
