package riidoaiserver

var (
	_ AssignmentStore      = (*Store)(nil)
	_ ProviderStatusStore  = (*Store)(nil)
	_ ProviderStatusReader = (*Store)(nil)
)
