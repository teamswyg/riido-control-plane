package main

import (
	"html/template"
	"strings"
)

type publicStatusHTMLView struct {
	publicStatus
	BadgeClass string
}

var publicStatusHTMLTemplate = template.Must(template.New("public-status").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Riido Public QA Status</title>
<style>
body{font-family:system-ui,-apple-system,sans-serif;margin:0;background:#f7f8fb;color:#172033}
main{max-width:880px;margin:0 auto;padding:40px 20px}
.badge{display:inline-block;padding:6px 10px;border-radius:6px;font-weight:700}
.status-degraded{background:#fff2cc;color:#6f4d00}
.status-operational{background:#dff7ea;color:#075c2d}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:12px;margin-top:24px}
.card{background:white;border:1px solid #dce3ef;border-radius:8px;padding:16px}
.label{color:#637083;font-size:13px}.value{font-size:24px;font-weight:700;margin-top:6px}
</style>
</head>
<body>
<main>
<h1>Riido Public QA Status</h1>
<p><span class="badge {{.BadgeClass}}">{{.Overall}}</span></p>
<div class="grid">
<div class="card"><div class="label">Visibility</div><div class="value">{{.Visibility}}</div></div>
<div class="card"><div class="label">P0 Cycles</div><div class="value">{{.P0CycleCount}}</div></div>
<div class="card"><div class="label">P0 Partial</div><div class="value">{{.P0PartialCount}}</div></div>
<div class="card"><div class="label">Partial Checks</div><div class="value">{{.PartialCount}}</div></div>
<div class="card"><div class="label">Stale Partials</div><div class="value">{{.StalePartialCount}}</div></div>
<div class="card"><div class="label">Candidates</div><div class="value">{{.ClosedLoopCandidates}}</div></div>
</div>
<p>Raw logs included: <strong>{{.RawLogsIncluded}}</strong></p>
<p>Secrets included: <strong>{{.SecretsIncluded}}</strong></p>
<p>Endpoint details: <strong>{{.EndpointDetails}}</strong></p>
<p>Next artifact: <code>{{.NextArtifact}}</code></p>
</main>
</body>
</html>
`))

func renderPublicStatusHTML(status publicStatus) (string, error) {
	var b strings.Builder
	view := publicStatusHTMLView{publicStatus: status, BadgeClass: "status-" + status.Overall}
	if err := publicStatusHTMLTemplate.Execute(&b, view); err != nil {
		return "", err
	}
	return b.String(), nil
}
