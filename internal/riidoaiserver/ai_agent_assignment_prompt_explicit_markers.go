package riidoaiserver

func explicitInstructionMarkers() []string {
	return []string{
		"해줘",
		"해주세요",
		"하라",
		"작성",
		"구현",
		"수정",
		"삭제",
		"추가",
		"생성",
		"만들",
		"실행",
		"테스트",
		"검증",
		"고쳐",
		"리팩토링",
		"배포",
		"설치",
		"업데이트",
		"확인해",
		"조회",
		"조사해",
		"보고해",
		"짜줘",
		"채우",
		"create",
		"implement",
		"fix",
		"update",
		"delete",
		"add",
		"write",
		"run",
		"test",
		"verify",
		"deploy",
		"refactor",
		"install",
		"generate",
		"build",
	}
}

func hasExplicitInstructionSignal(text string) bool {
	return containsAny(text, explicitInstructionMarkers())
}
