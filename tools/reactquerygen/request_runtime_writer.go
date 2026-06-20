package main

import "strings"

func writeRequestRuntime(b *strings.Builder) {
	b.WriteString("async function riidoRequest<T>(config: RiidoClientConfig, path: string, init: RequestInit = {}): Promise<T> {\n")
	b.WriteString("  const fetcher = config.fetcher ?? fetch;\n")
	b.WriteString("  const response = await fetcher(`${config.baseUrl.replace(/\\/$/, '')}${path}`, {\n")
	b.WriteString("    ...init,\n    headers: {\n      Accept: 'application/json',\n      'X-Riido-AI-Agent-Token': config.aiAgentToken,\n      ...(init.body ? { 'Content-Type': 'application/json' } : {}),\n      ...init.headers,\n    },\n  });\n")
	b.WriteString("  if (!response.ok) {\n    throw new Error(`Riido API ${response.status}: ${await response.text()}`);\n  }\n")
	b.WriteString("  return response.json() as Promise<T>;\n}\n\n")
}

func writeRawRequestRuntime(b *strings.Builder) {
	b.WriteString("async function riidoRawRequest(config: RiidoClientConfig, path: string, init: RequestInit = {}): Promise<Response> {\n")
	b.WriteString("  const fetcher = config.fetcher ?? fetch;\n")
	b.WriteString("  const response = await fetcher(`${config.baseUrl.replace(/\\/$/, '')}${path}`, {\n")
	b.WriteString("    ...init,\n    headers: {\n      'X-Riido-AI-Agent-Token': config.aiAgentToken,\n      ...init.headers,\n    },\n  });\n")
	b.WriteString("  if (!response.ok) {\n    throw new Error(`Riido API ${response.status}: ${await response.text()}`);\n  }\n")
	b.WriteString("  return response;\n}\n\n")
}
