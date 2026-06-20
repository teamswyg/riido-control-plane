package main

import (
	"net/http"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func metricsServer(store *riidoaiserver.Store, metrics *riidoaiserver.HTTPTransactionMetrics, auth riidoaiserver.RequestAuthorizer) (http.Handler, error) {
	return riidoaiserver.NewServer(riidoaiserver.ServerConfig{
		Assignment:       store,
		Authorizer:       auth,
		HTTPTransactions: metrics,
	}).Handler(), nil
}

func scopedAuthorizer(token, scope string) riidoaiserver.RequestAuthorizer {
	auth, err := riidoaiserver.NewStaticTokenAuthorizer([]riidoaiserver.StaticTokenCredential{{
		PrincipalID: "metrics-evidence",
		Token:       token,
		Scopes:      []string{scope},
	}})
	if err != nil {
		panic(err)
	}
	return auth
}
