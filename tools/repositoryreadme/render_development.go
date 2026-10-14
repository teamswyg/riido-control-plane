package main

import (
	"fmt"
	"strings"
)

func renderDevelopment(b *strings.Builder, dev development) {
	b.WriteString("## AI Agent Development Testnet\n\n")
	renderParagraphs(b, dev.Intro)
	if len(dev.Env) > 0 {
		b.WriteString("```bash\n")
		for _, env := range dev.Env {
			fmt.Fprintf(b, "%s=%s\n", env.Name, env.Value)
		}
		b.WriteString("```\n\n")
	}
	renderList(b, "Development Workflows", dev.Workflows)
	renderList(b, "검증하는 Endpoint", dev.Endpoints)
}

func renderDocLinks(b *strings.Builder, links []docLink) {
	b.WriteString("## 어떤 문서를 보면 되나\n\n")
	b.WriteString("| 알고 싶은 것 | 문서 |\n| --- | --- |\n")
	for _, link := range links {
		fmt.Fprintf(b, "| %s | [`%s`](%s) |\n", link.Topic, link.Path, link.Path)
	}
	b.WriteByte('\n')
}
