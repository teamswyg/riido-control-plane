package riidoaiserver

import "testing"

func BenchmarkRenderProgressMessage(b *testing.B) {
	args := map[string]string{
		"label":       "테스트 실행",
		"description": "진행 메시지 렌더링 hot path를 검증합니다.",
	}
	b.ReportAllocs()
	for b.Loop() {
		rendered, key, ok := renderProgressMessage(1103, args)
		if !ok || rendered == "" || key == "" {
			b.Fatalf("renderProgressMessage failed: rendered=%q key=%q ok=%v", rendered, key, ok)
		}
	}
}
