package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/teamswyg/riido-control-plane/internal/riidoaiserver"
)

func verifyHTTPCase(tc caseSpec) (caseEvidence, error) {
	handler, err := reviewHTTPHandler()
	if err != nil {
		return caseEvidence{}, err
	}
	catalogStatus, err := getCatalogStatus(handler)
	if err != nil {
		return caseEvidence{}, err
	}
	providerStatus, err := getProviderStatus(handler)
	if err != nil {
		return caseEvidence{}, err
	}
	pollStatus := postPollStatus(handler)
	result := caseEvidence{
		Name: tc.Name, Kind: tc.Kind,
		CatalogStatus: catalogStatus, ProviderStatus: providerStatus, PollStatus: pollStatus,
	}
	if catalogStatus != tc.WantCatalogStatus || providerStatus != tc.WantProviderStatus || pollStatus != tc.WantPollStatus {
		return result, fmt.Errorf("%s result=%+v", tc.Name, result)
	}
	return result, nil
}

func reviewHTTPHandler() (http.Handler, error) {
	provisioning, err := reviewProvisioning()
	if err != nil {
		return nil, err
	}
	store := riidoaiserver.NewStore()
	if err := store.ApplyReviewAccountProvisioning(context.Background(), provisioning); err != nil {
		store.Close()
		return nil, err
	}
	authorizer, err := riidoaiserver.NewStaticTokenAuthorizer([]riidoaiserver.StaticTokenCredential{provisioning.Credential})
	if err != nil {
		store.Close()
		return nil, err
	}
	return riidoaiserver.NewServer(riidoaiserver.ServerConfig{Assignment: store, Authorizer: authorizer}).Handler(), nil
}
