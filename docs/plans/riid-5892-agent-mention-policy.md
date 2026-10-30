# RIID-5892: Agent 조회와 실행 정책 연결

목표: 원댓글에서 시작한 Agent가 CLI로 현재 후보를 비교하고, 허용된 Agent ID를 명시적으로 멘션하면 같은 댓글 대화에서 실제 후속 실행과 맥락 상속이 일어난다.

범위: 기존 깊이 3·추가 실행 예산 8·조상 순환 규칙을 조회/실행이 공유한다. 세션 충돌은 기존 대화 재개 검사로 사전 확인한다. 조회는 예산을 소비하거나 실행을 만들지 않으며 실행 직전 권한·연결·정책 재검증은 유지한다. 명시적 외부 실행 지침과 역할 근거는 기존 공개 description 및 callable을 사용한다.

현재 staging 확인: control-plane 2fc74c02077c의 readCommentDeliveryEvidence는 X-Riido-Read-Scope-Version 협상과 prepareCommentReadScope를 보존한다. monolith 22da10b9b789의 header-on/off 모두 원댓글·트리거·대화 식별자를 반환한다. 두 계약은 재구현하지 않고 관련 집중 테스트로 보존을 확인한다.

검증: 기본 Server 구성이 policy_checked=true를 반환하며 제한된 후보는 callable=false와 이유를 제공한다. 깊이/예산/순환·동일 대화의 활성 세션·다른 task/workspace·누락된 원댓글 맥락을 검사한다. 기존 HTTP 목록/상세→멘션 완료→B 배정 및 확장 자료 상속·조회 기록 테스트를 재실행한다. 실제 staging 계정에서 A→목록 조회→B 호출을 확인한다.

제외: production 배포, 스키마 변경, 임의 이름 기반 호출, 실시간 대화 동기화, 일반 실행에 새 원댓글 자동 생성. 기존 선택 기준과 명시 멘션 방식은 유지한다. policy_checked는 조회 시점의 판단이며 LIVE_STATUS에 따라 실행 시 다시 검증한다.

릴리즈: 현재 staging 통합 커밋 위에 최소 변경만 배포한다. 새 저장 형식을 도입하지 않는다. 롤백은 현재 staging 2fc74c02077c로 복구하며 기존 대화·반응·확장 조회 자료를 보존한다. daemon 변경이 필요하면 staging 전용 경로만 사용한다.
