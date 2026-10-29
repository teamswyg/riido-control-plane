package riidoaiserver

import "unicode/utf8"

func readScopeBodyFragments(page taskReadBodyPage) []taskReadBodyPage {
	if len(page.Content) == 0 {
		return []taskReadBodyPage{page}
	}
	var out []taskReadBodyPage
	for begin := 0; begin < len(page.Content); {
		end := min(begin+8192, len(page.Content))
		for end < len(page.Content) && !utf8.RuneStart(page.Content[end]) {
			end--
		}
		fragment := page
		offset := *page.Offset + begin
		fragment.Offset = &offset
		fragment.Content = page.Content[begin:end]
		if end < len(page.Content) {
			complete := false
			next := *page.Offset + end
			fragment.Complete = &complete
			fragment.NextOffset = &next
			fragment.Status = "partial"
		}
		out = append(out, fragment)
		begin = end
	}
	return out
}
