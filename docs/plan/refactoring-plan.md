# ppopcode 리팩토링 계획

> 작성일: 2026-01-20

## 1. 프로젝트 개요

ppopcode를 TUI 기반 멀티 에이전트 시스템으로 리팩토링한다.

### 핵심 목표
- OpenCode 스타일 TUI 구현
- Claude Code 기반 오케스트레이션
- 멀티 에이전트 (작업별 모델 배정)
- Cursor를 통한 코드 수정

## 2. 아키텍처

```
┌─────────────────────────────────────────────────────────────┐
│                    TUI (통합 진입점)                         │
│              OpenCode 구조 참고하여 새로 구현                 │
├─────────────────────────────────────────────────────────────┤
│  [💬 대화로 시작]              [📋 워크플로우 선택]           │
│                              (cc-wf-studio)                 │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              오케스트레이터 (Claude Code)                    │
│                    Opus 4.5 - 작업 조율                      │
├─────────────────────────────────────────────────────────────┤
│                      모델 배정                               │
│  ┌──────────────┬──────────────┬──────────────┐             │
│  │ Gemini 3 Pro │  GPT 5.2     │ Sonnet 4.5   │             │
│  │   UX/UI      │ 설계/디버깅  │  기본 작업    │             │
│  └──────────────┴──────────────┴──────────────┘             │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                 Cursor 에이전트                              │
│                  코드 수정 실행                              │
│              (Cursor 토큰 소모)                              │
└─────────────────────────────────────────────────────────────┘
```

## 3. 컴포넌트 역할

| 컴포넌트 | 역할 | 기술 |
|----------|------|------|
| **TUI** | 통합 진입점, 사용자 인터페이스 | Go (Bubble Tea) 또는 Python (Textual) |
| **오케스트레이터** | 작업 분석, 모델 배정, 전체 조율 | Claude Code SDK |
| **멀티 에이전트** | 작업별 전문 모델 실행 | 각 AI Provider API |
| **Cursor** | 코드 수정 실행 | cursor-agent CLI |

## 4. 멀티 에이전트 모델 배정

| 에이전트 | 모델 | 역할 |
|----------|------|------|
| **오케스트레이터** | Opus 4.5 | 메인 조율, 작업 분석, 모델 배정 결정 |
| **프론트엔드** | Gemini 3 Pro | UX/UI 개발 |
| **설계/디버깅** | GPT 5.2 | 아키텍처 설계, 디버깅, 고수준 전략 |
| **기본 작업** | Sonnet 4.5 | 일반적인 코딩 작업 |

## 5. 실행 흐름

### 5.1 대화로 시작
```
사용자 → TUI → 대화 입력
              → 오케스트레이터 (Opus 4.5)
              → 작업 분석 후 적절한 모델 배정
              → 해당 모델이 계획/분석 수행
              → Cursor가 코드 수정
              → 결과 반환
```

### 5.2 워크플로우 선택
```
사용자 → TUI → 워크플로우 선택 (cc-wf-studio)
              → 워크플로우 JSON 로드
              → 오케스트레이터가 노드 순서대로 실행
              → 각 노드에서 모델 배정
              → Cursor가 코드 수정
              → 결과 반환
```

## 6. 기술 스택

### 6.1 TUI 옵션

| 옵션 | 언어 | 프레임워크 | 장점 |
|------|------|-----------|------|
| **A** | Go | Bubble Tea | OpenCode와 동일, 빠름, 단일 바이너리 |
| **B** | Python | Textual | 배우기 쉬움, 풍부한 위젯 |
| **C** | TypeScript | Ink | React 문법, npm 생태계 |

### 6.2 AI 연동

- **Claude Code SDK**: 오케스트레이션
- **Anthropic API**: Claude 모델 (Opus, Sonnet)
- **OpenAI API**: GPT 모델
- **Google AI API**: Gemini 모델
- **Cursor CLI**: 코드 수정 위임

## 7. 폴더 구조 (예상)

```
ppopcode/
├── cmd/                    # CLI 진입점
│   └── ppopcode/
│       └── main.go
├── internal/
│   ├── tui/               # TUI 컴포넌트
│   │   ├── app.go
│   │   ├── chat.go
│   │   ├── workflow.go
│   │   └── status.go
│   ├── orchestrator/      # 오케스트레이터
│   │   ├── orchestrator.go
│   │   └── router.go
│   ├── agents/            # 멀티 에이전트
│   │   ├── agent.go
│   │   ├── gemini.go
│   │   ├── gpt.go
│   │   └── sonnet.go
│   ├── cursor/            # Cursor 연동
│   │   └── bridge.go
│   └── workflow/          # cc-wf-studio 연동
│       └── loader.go
├── config/
│   ├── ppopcode.yaml      # 메인 설정
│   ├── agents.yaml        # 에이전트 설정
│   └── models.yaml        # 모델 설정
├── .vscode/
│   └── workflows/         # cc-wf-studio 워크플로우
├── docs/
│   ├── plan/
│   │   └── refactoring-plan.md
│   └── research/
│       └── oh-my-opencode-analysis.md
└── README.md
```

## 8. 구현 단계

### Phase 1: TUI 기본 구조
- [ ] TUI 프레임워크 선택 및 셋업
- [ ] 메인 메뉴 구현 (대화/워크플로우 선택)
- [ ] 채팅 인터페이스 구현
- [ ] 상태 표시 구현

### Phase 2: Claude Code 연동
- [ ] Claude Code SDK 통합
- [ ] 오케스트레이터 구현
- [ ] 기본 대화 흐름 연결

### Phase 3: 멀티 에이전트
- [ ] 에이전트 인터페이스 정의
- [ ] 모델별 에이전트 구현 (Gemini, GPT, Sonnet)
- [ ] 라우팅 로직 구현

### Phase 4: Cursor 연동
- [ ] cursor-agent CLI 통합
- [ ] 코드 수정 요청/응답 처리
- [ ] 에러 핸들링

### Phase 5: 워크플로우 통합
- [ ] cc-wf-studio JSON 파서
- [ ] 워크플로우 실행 엔진
- [ ] 노드별 처리 로직

### Phase 6: 고급 기능
- [ ] 세션 관리
- [ ] 히스토리 저장
- [ ] 설정 UI

## 9. 현재 vs 새 구조 비교

| 항목 | 현재 ppopcode | 새 ppopcode |
|------|--------------|-------------|
| **UI** | 텍스트 CLI | TUI (비주얼) |
| **진입점** | Claude Code CLI | TUI → Claude Code |
| **에이전트** | 단일 (Claude) | 멀티 (Opus, GPT, Gemini, Sonnet) |
| **코드 수정** | Cursor 위임 | Cursor 위임 (동일) |
| **워크플로우** | cc-wf-studio | cc-wf-studio (동일) |

## 10. 참고 자료

- [oh-my-opencode](https://github.com/code-yeongyu/oh-my-opencode)
- [OpenCode TUI](https://github.com/sst/opencode)
- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [Textual](https://github.com/Textualize/textual)
