# ppopcode

TUI 기반 멀티 에이전트 AI 코딩 어시스턴트

## 개요

ppopcode는 OpenCode 스타일의 TUI를 통해 여러 AI 모델을 오케스트레이션하고, Cursor를 통해 코드를 수정하는 프레임워크입니다.

### 핵심 기능

- **TUI 인터페이스**: Bubble Tea 기반의 터미널 UI
- **멀티 에이전트**: 작업 유형별 최적 모델 자동 배정
- **Cursor 연동**: 코드 수정은 Cursor가 담당
- **워크플로우**: cc-wf-studio 통합

## 아키텍처

```
┌─────────────────────────────────────────────────────────────┐
│                    TUI (통합 진입점)                         │
├─────────────────────────────────────────────────────────────┤
│  [💬 대화로 시작]              [📋 워크플로우 선택]           │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              오케스트레이터 (Claude Code)                    │
│                    Opus 4.5 - 작업 조율                      │
├─────────────────────────────────────────────────────────────┤
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
└─────────────────────────────────────────────────────────────┘
```

## 폴더 구조

```
ppopcode/
├── cmd/ppopcode/main.go          # 진입점
├── internal/
│   ├── tui/                      # TUI 컴포넌트
│   ├── orchestrator/             # 오케스트레이터
│   ├── agents/                   # 멀티 에이전트
│   ├── cursor/                   # Cursor 연동
│   ├── workflow/                 # cc-wf-studio 연동
│   ├── config/                   # 설정
│   └── session/                  # 세션 관리
├── config/
│   └── ppopcode.yaml             # 설정 파일
├── .vscode/
│   └── workflows/                # cc-wf-studio 워크플로우
└── docs/
    └── plan/                     # 계획 문서
```

## 설치

### 요구사항

- Go 1.21+
- Cursor (cursor-agent CLI)

### 빌드

```bash
cd ppopcode
go mod tidy
go build -o ppopcode ./cmd/ppopcode
```

### 환경 변수 (선택)

```bash
export ANTHROPIC_API_KEY="your-key"
export OPENAI_API_KEY="your-key"
export GOOGLE_API_KEY="your-key"
```

## 사용법

### TUI 실행

```bash
./ppopcode
```

### 메뉴

1. **Chat**: AI와 대화 시작
2. **Workflow**: cc-wf-studio 워크플로우 선택 및 실행
3. **Settings**: 에이전트 및 설정 관리

### 키보드 단축키

- `↑/↓` 또는 `j/k`: 이동
- `Enter`: 선택
- `Esc`: 뒤로가기
- `q`: 종료

## 설정

### config/ppopcode.yaml

```yaml
app:
  name: ppopcode
  version: 1.0.0

agents:
  orchestrator:
    type: claude
    model: claude-opus-4-20250514
    role: Main orchestrator

  gemini:
    type: gemini
    model: gemini-2.0-flash
    role: UX/UI specialist

  gpt:
    type: openai
    model: gpt-4o
    role: Design & debugging

  sonnet:
    type: claude
    model: claude-sonnet-4-20250514
    role: General coding

cursor:
  timeout: 300
  max_retry: 2
```

## 에이전트 라우팅

| 작업 유형 | 키워드 | 에이전트 |
|----------|--------|----------|
| UX/UI | ui, ux, frontend, component, css | Gemini 3 Pro |
| 설계/디버깅 | architecture, debug, fix, error | GPT 5.2 |
| 기본 작업 | (기타) | Sonnet 4.5 |

## 라이선스

MIT
