package riidoaiserver

import (
	"regexp"
)

var (
	clientVisibleMarkdownLocalLinkPattern           = regexp.MustCompile(`\[([^\]]+)\]\(\s*<?(?:file://)?(?:/Users|/private/var|/var/folders|/tmp)/[^)>]*>?\s*\)`)
	clientVisibleAngleLocalPathPattern              = regexp.MustCompile(`<(?:file://)?(?:/Users|/private/var|/var/folders|/tmp)/[^>]+>`)
	clientVisibleApplicationSupportLocalPathPattern = regexp.MustCompile("(?:file://)?/Users/[^\\s<>)\\]\"'`]+/Library/Application Support/[^\\s<>)\\]\"'`]+")
	clientVisibleLocalPathPattern                   = regexp.MustCompile("(?:file://)?(?:/Users|/private/var|/var/folders|/tmp)/[^\\s<>)\\]\"'`]+")
)
