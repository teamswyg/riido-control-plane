package riidoaiserver

import "maps"

func cloneMetadata(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	maps.Copy(out, in)
	return out
}

func cloneProviderStatusRecords(in []ProviderStatusRecord) []ProviderStatusRecord {
	if len(in) == 0 {
		return nil
	}
	return append([]ProviderStatusRecord(nil), in...)
}

func cloneProviderStatusResponse(in ProviderStatusSyncResponse) ProviderStatusSyncResponse {
	in.Providers = cloneProviderStatusRecords(in.Providers)
	return in
}
