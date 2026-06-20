package main

import "strings"

func writeCoreRuntime(b *strings.Builder) {
	writeJSDoc(b, "앱에서 사용하는 fetch 구현을 주입하기 위한 타입입니다.")
	b.WriteString("export type RiidoFetcher = typeof fetch;\n\n")
	writeJSDoc(
		b,
		"control-plane 호출에 필요한 기본 설정입니다.",
		"`baseUrl`은 예: `https://<control-plane-host>`처럼 마지막 슬래시 없이 전달해도 됩니다.",
		"`aiAgentToken`은 기존 Riido 앱 로그인 토큰과 구분되는 AI Agent SaaS 전용 토큰입니다.",
		"요청에는 `X-Riido-AI-Agent-Token` 헤더로 전달됩니다.",
		"`fetcher`는 테스트나 앱 공통 transport 래핑이 필요할 때만 주입합니다.",
	)
	b.WriteString("export interface RiidoClientConfig {\n  baseUrl: string;\n  aiAgentToken: string;\n  fetcher?: RiidoFetcher;\n}\n\n")
	writeJSDoc(b, "요청 단위로 전달할 수 있는 옵션입니다. 현재는 AbortSignal만 전달합니다.")
	b.WriteString("export interface RiidoRequestOptions {\n  signal?: AbortSignal;\n}\n\n")
	writeJSDoc(b, "React Query query option에 Riido 요청 옵션을 함께 전달하기 위한 타입입니다.")
	b.WriteString("export type RiidoQueryOptions<TData> = Omit<UseQueryOptions<TData>, 'queryKey' | 'queryFn'> & RiidoRequestOptions;\n\n")
	writeJSDoc(b, "React Query mutation option을 Riido endpoint 변수 타입과 묶은 타입입니다.")
	b.WriteString("export type RiidoMutationOptions<TData, TVariables> = UseMutationOptions<TData, Error, TVariables>;\n\n")
	writeRequestRuntime(b)
	writeRawRequestRuntime(b)
}
