# Single-pass response delivery

This change retains one provider turn on the normal path. There is no classifier,
reviewer, summarizer or automatic quality-retry model call. Existing explicitly
configured verification continuations and provider recovery are unchanged.

## Contract

- Answer-only requests need no repository, file mutation or shell check.
- Answer-only intent recognition requires directive syntax, not a substring in
  prose. Nominal requests (summary, translation, explanation, comparison) are
  recognized in the task title or explicit user instruction, not document headings.
  This remains a deterministic heuristic, not a semantic language classifier.
- The latest user request controls scope; historical task instructions are context.
- Requested execution must be reported with observed evidence, not proposed steps.
- A provider success with no non-whitespace answer and no verified host result is
  a failed empty response, never a generic successful completion.
- Terminal acknowledgements use a deterministic key scoped to execution, lease
  and result, including retries beyond the HTTP retry budget.
- Empty history entries do not consume the recent-message budget.
- Buffer streamed text before network state updates. Failed delivery must not erase
  the only answer; final delivery can recover the accumulated text.
- Measure daemon claim-to-first-text, first delivery and terminal acknowledgement
  with its own clock. Measure server event handling and assignment age with the
  server clock. Do not subtract clocks across machines or call these UI paint time.
- Logs contain identifiers, status and timing only, never prompts, tokens or answers.

## Verification and boundaries

Focused regressions cover normal single-turn answers, text-only completion,
whitespace/empty completion, failed first delivery and terminal fallback, scoped
history and the latest directive. Local authenticated Codex runs measure first SSE
answer and terminal state before/after with the same prompt and runtime-default model configuration. They are
small observations, not a statistical speedup claim. Test timing overhead separately
with local benchmarks; no model calls occur in ordinary unit tests.

No frontend, dependency, deployment, container isolation or new worktree lifecycle
is introduced. Existing session recovery and opt-in host checks remain authoritative.

## Local measurements — 2026-10-16

Answer-only task: summarize the supplied note in two Korean sentences, with no
repository, tools or file creation. macOS arm64 / Apple M3 Pro, Go 1.26.5,
Codex CLI 0.161.0, same runtime-default model configuration (resolved model ID
not pinned). Baselines: daemon `96304799a998791e2485d974a4a774c9166743cb`,
control-plane `82787a78f678300ce7dc9faafde4bca190ba56b9`.
Each run used disposable local server state, a separate local test account/device,
and the existing user's authenticated Codex CLI. No production state was used.

| Run | Before first SSE (ms) | After first SSE (ms) | Before terminal observed (ms) | After terminal observed (ms) |
| --- | ---: | ---: | ---: | ---: |
| 1 | 6324 | 7139 | 7051 | 8281 |
| 2 | 6751 | 12265 | 7255 | 12767 |
| 3 | 6824 | 6857 | 7300 | 7459 |
| 4 | 7494 | 7347 | 8138 | 8075 |
| 5 | 6275 | 6253 | 7064 | 6923 |
| Median | 6751 | 7139 | 7255 | 8075 |

Observed medians increased by 388 ms (5.7%) and 820 ms (11.3%). This is **not a
speedup or no-regression claim**: five unrandomized runs per revision, variable
provider latency and workstation load, one 12.265 s first-answer outlier, and no
pinned model ID. Runs 1–4 exercised the same answer path before the final terminal
retry-key/history-empty-row hardening; run 5 used the final functional code.
Provider/internal API request counts are not measured by session counts.

After runs recorded first-text-to-server-ack delays of 22, 9, 10, 16 and 11 ms.
For the outlier, claim-to-first-text was 11295 ms, then first delivery took 9 ms;
the delay occurred before answer text reached the daemon, not in its subsequent
HTTP delivery. This does not identify a particular upstream cause.
Timers start immediately before assignment POST after server/daemon readiness
and SSE connection. They include assignment scheduling; they exclude initial
service startup. Terminal observation polls every 100 ms. Baseline runs 1–3
used 10 ms SSE timestamp resolution. SSE receipt is not desktop UI paint time.
All ten runs produced streamed and final Korean answers; no manual desktop UI
validation is claimed here.

The opt-in native integration probe counted **one `turn/start` and one process
start**, with provider elapsed time 5272 ms. It checks no host verification was
required. This verifies the ordinary daemon turn path, not provider-internal
inference requests. Existing explicitly enabled continuations/recovery can still
start additional turns; this change adds none.

Prompt-construction microbenchmark (three samples, same 4800-character context):
72901/82149/93555 ns/op before, 20101/19714/17469 ns/op after; approximately
47176 B/op and 20 allocs/op in both. Both are below 0.1 ms in these samples;
these are prompt-building costs only, not model or end-to-end latency.

## Reproduce locally

Clone both PR branches into sibling `riido-daemon` and `riido-control-plane`
directories. From each repository respectively:

```sh
# riido-daemon
 go build -o /tmp/riido-response-daemon ./cmd/riido
# riido-control-plane
 go build -o /tmp/riido-response-server ./cmd/riido_ai_server
```

Then, from riido-daemon:

```sh
python3 scripts/measure-single-pass-response.py \
  --daemon /tmp/riido-response-daemon \
  --server /tmp/riido-response-server
AGENTBRIDGE_INTEGRATION=1 go test ./internal/agentbridge/supervisor \
  -run '^TestIntegrationAnswerOnlyUsesOneProviderTurn$' -v -count=1
```

Requires a locally installed and authenticated Codex CLI; these two commands
consume actual provider usage. To pin a supported model, pass `--model MODEL_ID`
and set `AGENTBRIDGE_CODEX_MODEL=MODEL_ID` for the native test. The script prints
its measurements and private temporary evidence path and stops its own processes.
Do not commit the retained state: it includes disposable device credentials and
local run content. Existing desktop login/profile is not needed for this probe.

Ordinary validation (does not opt into real provider calls):

```sh
go test ./...
golangci-lint run ./... --timeout=5m
# riido-daemon
 go test -race ./internal/agentbridge/controlplane/saasplane ./internal/agentbridge/supervisor
# riido-control-plane
 go test -race ./internal/riidoaiserver
 go test ./internal/riidoaiserver -run '^$' -bench '^BenchmarkSinglePassAssignmentPrompt$' -benchmem -count=3
```

Both full Go suites and the listed race checks passed. Lint is recorded in the PR.
Raw local logs are intentionally excluded; sanitized measurement values are kept
in the companion executable manifest.

## PR review follow-up

PR #1094 review regressions reproduce prose keyword false positives and missed
nominal requests on the pre-fix head. The matcher now retains UserInstruction as a
separate field, requires anchored directive syntax and recognizes nominal requests
in titles/user instructions without treating document headings as commands.
The earlier live SSE measurements predate this matcher correction. No new provider
run was needed for this deterministic classification fix. The same local prompt
benchmark after the correction measured 28382/21454/35903 ns/op (20 allocs/op),
which measures prompt construction only and does not establish model latency.

Empty editor HTML messages are normalized with the same renderer as prompt history and excluded before the recent-message budget. `TestAcceptanceRenderedEmptyHistoryPreservesEarlierRequirement` preserves earlier user constraints across empty editor status rows.

## Direct assignment execution · RIID-5827

Assigning an agent requests execution of the task's intended outcome. Keyword
classification is advisory metadata only: neither missing command verbs nor a
marketing/planning task type may prevent publication to the existing provider
turn. All intent classes have `intent_gate_required: false`. The same turn
interprets the title, current document and latest user instruction, proceeds
with supported work and asks only a concrete question about information that
actually blocks progress. Generic requests to repeat the desired outcome are
not the default first response. Tool authorization and genuine provider
`needs_input` results retain their existing behavior; persisted legacy intent
threads remain readable and can accept follow-ups.

HTTP regressions cover legacy and additive assignment APIs, requests without
command keywords, task-body and explicit-instruction preservation, and duplicate
request IDs resolving to the same assignment. Verified comment delivery must
retain the comment body and task snapshot as well. No classifier/reviewer
request, automatic follow-up or second provider run is introduced. UI assignee
selection sends an agent ID; the authorized server reader loads the saved task
document. Unsaved editor text is not part of that snapshot. Existing document
size limits still explicitly mark truncation; no test claims unlimited context
or proves the contents of an unobserved customer request.

## Delivery latency attribution

Successful partial/terminal timing logs separate durable assignment commit
(`assignment_commit_ms`) from client projection (`client_projection_ms`). Their
sum is `event_processing_ms`; authentication and request decoding occur before
this interval. These durations are measured on the server clock, not client paint
latency. No answer text, credentials or additional model calls are introduced.

The 2026-10-17 paired local experiment injected 200 ms per daemon event HTTP
acknowledgement. Daemon grouping reduced event requests from 87 to 25 and
provider-completion-to-daemon-completion from 11.740 s to 0.577 s. Request-to-first
SSE was 10.388 s / 8.098 s; request-to-terminal was 19.924 s / 9.729 s. The model
itself took 5.918 s / 6.959 s. Server processing medians were 12.5 / 14 ms
(commit 5.5 / 6 ms, client projection 6 / 7 ms). Each revision had one sample
and one native Codex session, with runtime-default model configuration. These
are controlled network-backpressure observations, not production or UI paint
latency guarantees. This control-plane change adds attribution only; the
network request reduction is implemented in the daemon.
