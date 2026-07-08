package main

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestRunRejectsMissingBaseURLBadFlagAndWriteFailure(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing base", args: nil, want: "-base-url is required"},
		{name: "bad flag", args: []string{"-nope"}, want: "flag provided but not defined"},
		{name: "bad out", args: []string{"-base-url", testServer(t, false).URL, "-out", t.TempDir() + "/missing/out.json"}, want: "no such file"},
	}
	for _, tc := range cases {
		err := run(tc.args, errWriter{}, time.Now(), http.DefaultClient)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%s error = %v, want %q", tc.name, err, tc.want)
		}
	}
}

func TestCollectRejectsFetchDecodeAndValidationFailures(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		html   string
		status string
		pages  string
		badge  string
		want   string
		code   int
	}{
		{name: "status fetch", code: http.StatusInternalServerError, want: "status 500"},
		{name: "status decode", status: `{`, pages: validPagesJSON(), badge: validBadgeJSON(), want: "unexpected end"},
		{name: "pages decode", status: validStatusJSON(), pages: `{`, badge: validBadgeJSON(), want: "unexpected end"},
		{name: "badge decode", status: validStatusJSON(), pages: validPagesJSON(), badge: `{`, want: "unexpected end"},
		{name: "validation", status: `{"overall":"degraded","visibility":"private"}`, pages: validPagesJSON(), badge: validBadgeJSON(), want: "failed validation"},
	}
	for _, tc := range cases {
		server := liveServer(liveCase{
			html: tc.html, status: tc.status,
			pages: tc.pages, badge: tc.badge, code: tc.code,
		})
		_, err := collect(server.URL, time.Now(), server.Client())
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%s error = %v, want %q", tc.name, err, tc.want)
		}
		server.Close()
	}
}

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
