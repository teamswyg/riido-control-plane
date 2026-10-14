package riidoaiserver

import "net/http"

func normalizeAssignmentOperationStoreHarnessConfig(
	cfg dynamoDBAssignmentOperationStoreHarnessConfig,
) dynamoDBAssignmentOperationStoreHarnessConfig {
	if cfg.TableName == "" {
		cfg.TableName = "assignments"
	}
	if cfg.RequestBuffer <= 0 {
		cfg.RequestBuffer = 1
	}
	if cfg.AccessKeyID == "" {
		cfg.AccessKeyID = "AKID"
	}
	if cfg.Handler == nil {
		cfg.Handler = func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{}`))
		}
	}
	return cfg
}
