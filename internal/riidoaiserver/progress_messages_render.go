package riidoaiserver

import (
	"regexp"
	"strings"
	"sync"

	"github.com/teamswyg/riido-contracts/progressmessage"
)

type progressMessageTemplate struct {
	Key      string
	Template string
}

var (
	progressTemplatePlaceholder  = regexp.MustCompile(`\{\{([a-zA-Z0-9_]+)\}\}`)
	progressMessageTemplateCache = sync.OnceValue(loadProgressMessageTemplateMap)
)

func renderProgressMessage(code int, args map[string]string) (string, string, bool) {
	if code <= 0 {
		return "", "", false
	}
	normalizedArgs := progressmessage.NormalizeArgsForCode(code, args)
	template, ok := progressMessageTemplateCache()[code]
	if !ok {
		return "", "", false
	}
	return renderProgressTemplate(template.Template, normalizedArgs), template.Key, true
}

func loadProgressMessageTemplateMap() map[int]progressMessageTemplate {
	catalog, err := progressmessage.Catalog()
	if err != nil {
		return map[int]progressMessageTemplate{}
	}
	out := make(map[int]progressMessageTemplate, len(catalog.Messages))
	for _, item := range catalog.Messages {
		template := item.Locales[progressmessage.DefaultLocale]
		if template == "" {
			template = item.Locales["ko"]
		}
		if item.Code > 0 && template != "" {
			out[item.Code] = progressMessageTemplate{
				Key:      strings.TrimSpace(item.Key),
				Template: template,
			}
		}
	}
	return out
}

func renderProgressTemplate(template string, args map[string]string) string {
	return progressTemplatePlaceholder.ReplaceAllStringFunc(template, func(match string) string {
		parts := progressTemplatePlaceholder.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}
		value := strings.TrimSpace(args[parts[1]])
		if value == "" {
			return "not provided"
		}
		return value
	})
}
