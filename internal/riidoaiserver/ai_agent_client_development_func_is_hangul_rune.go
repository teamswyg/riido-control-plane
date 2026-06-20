package riidoaiserver

func isHangulRune(r rune) bool {
	return (r >= 0xAC00 && r <= 0xD7A3) || (r >= 0x3130 && r <= 0x318F)
}
