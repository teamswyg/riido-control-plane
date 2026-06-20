package riidoaiserver

import "strings"

// agentResponseGuidelines is appended to every assignment's agent instruction.
const agentResponseGuidelines = `[응답 지침]
- 모든 답변은 한국어로 작성하세요.
- 작업 디렉터리에 코드/리포지터리가 없거나 도구 사용이 제한되어 실제 파일 작업을 할 수 없더라도, "작업 공간이 없어 실패했다"로 끝내지 마세요.
  - 질문이거나 설명/조언으로 충분한 요청이면 도구 없이 바로 한국어로 답하세요.
  - 파일·코드 작업이 꼭 필요하면 도구를 무리하게 반복 호출하지 말고, 사용자에게 "어느 경로 또는 리포지터리에서 작업할까요?"처럼 필요한 정보를 물어보거나, 이후 진행해야 할 일을 단계로 제시하며 마무리하세요.`

// augmentAgentInstruction appends shared guidelines before runtime execution.
func augmentAgentInstruction(instruction string) string {
	instruction = strings.TrimSpace(instruction)
	if instruction == "" {
		return agentResponseGuidelines
	}
	return instruction + "\n\n" + agentResponseGuidelines
}
