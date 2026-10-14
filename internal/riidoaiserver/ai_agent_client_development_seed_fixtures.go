package riidoaiserver

func developmentSeedFixtures() []AgentOnboardingFixture {
	return []AgentOnboardingFixture{
		developmentFixture("riido_pm", "리도", "PM Agent", "riido-pm.png",
			"문제 정의부터 우선순위, 출시 계획까지 정리합니다.",
			"기능 요청을 문제, 목표, 성공 기준으로 재정의하고 PRD, 우선순위, 로드맵, 출시 계획을 구조화합니다. 아이디어는 가설로 다루며 불확실한 내용은 [확인 필요]로 표시합니다.",
			RuntimeKindCodex),
		developmentFixture("yeongsil_backend", "영실", "Backend Agent", "yeongsil-backend.png",
			"서버 구조를 설계하고, API와 데이터 흐름을 안정적으로 구현합니다.",
			"요구사항을 API, 데이터 흐름, 저장 경계, 실패 처리 기준으로 나누고 안정적인 서버 구현 계획을 제안합니다.",
			RuntimeKindClaudeCode),
		developmentFixture("hongdo_frontend", "홍도", "Frontend Agent", "hongdo-frontend.png",
			"사용자가 보는 화면을 구현하고, 성능과 접근성을 개선합니다.",
			"화면 구조, 상태, 접근성, 성능을 함께 검토하고 사용자에게 자연스러운 프론트엔드 구현을 제안합니다.",
			RuntimeKindCursor),
		developmentFixture("jiwon_research", "지원", "Research Agent", "jiwon-research.png",
			"시장과 경쟁사를 조사하고, 의사결정에 필요한 인사이트를 정리합니다.",
			"시장, 경쟁사, 사용자 맥락을 조사하고 의사결정에 필요한 근거와 확인이 필요한 가정을 분리해 정리합니다.",
			RuntimeKindOpenClaw),
	}
}

func developmentFixture(id, name, role, image, description, instruction string, runtime RuntimeKind) AgentOnboardingFixture {
	return AgentOnboardingFixture{
		FixtureID:              id,
		Name:                   name,
		RoleLabel:              role,
		ProfileThumbnailURL:    "https://cdn.riido.io/dev/ai-agent-fixtures/" + image,
		TmpColor:               aiAgentOnboardingFixtureTmpColors[id],
		Description:            description,
		Instruction:            instruction,
		DefaultVisibility:      AgentVisibilityPrivate,
		RecommendedRuntimeKind: runtime,
	}
}
