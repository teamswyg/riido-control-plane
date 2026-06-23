package riidoaiserver

import (
	"testing"

	"github.com/teamswyg/riido-contracts/progressmessage"
)

func TestRenderProgressMessageMatchesContractsCatalog(t *testing.T) {
	t.Parallel()
	catalog, err := progressmessage.Catalog()
	if err != nil {
		t.Fatalf("progressmessage.Catalog: %v", err)
	}
	for _, item := range catalog.Messages {
		args := sampleProgressMessageArgs(item.Args)
		want, ok := progressmessage.Render(item.Code, args, progressmessage.DefaultLocale)
		if !ok {
			t.Fatalf("contracts Render(%d) returned false", item.Code)
		}
		got, key, ok := renderProgressMessage(item.Code, args)
		if !ok || got != want || key != item.Key {
			t.Fatalf("renderProgressMessage(%d) = (%q, %q, %v), want (%q, %q, true)", item.Code, got, key, ok, want, item.Key)
		}
	}
}

func sampleProgressMessageArgs(args []progressmessage.MessageArg) map[string]string {
	out := make(map[string]string, len(args))
	for _, arg := range args {
		out[arg.Name] = "sample"
	}
	return out
}
