# oh-my-opencode 구현 분석

> 참고: https://github.com/code-yeongyu/oh-my-opencode

## 1. 프로젝트 개요

**oh-my-opencode**는 OpenCode TUI 기반 코딩 에이전트의 플러그인이다. 터미널에서 실행되는 비주얼 인터페이스를 제공하며, 다중 AI 모델을 오케스트레이션하여 코딩 작업을 수행한다.

### 핵심 가치
- "Batteries-Included" - 설치 즉시 모든 기능 사용 가능
- 다중 모델 오케스트레이션
- Claude Code 호환성 레이어

## 2. 아키텍처

```
┌─────────────────────────────────────────────┐
│                 OpenCode TUI                │
│        (터미널 기반 비주얼 인터페이스)         │
├─────────────────────────────────────────────┤
│           oh-my-opencode Plugin             │
├──────────┬──────────┬──────────┬────────────┤
│ Agents   │  Hooks   │  MCPs    │  LSP/AST   │
└──────────┴──────────┴──────────┴────────────┘
```

### 폴더 구조
```
oh-my-opencode/
├── .opencode/           # OpenCode 설정
├── packages/            # 모노레포 패키지
├── src/                 # 메인 소스코드
├── bin/                 # CLI 실행 파일
├── docs/                # 문서
├── script/              # 스크립트
├── signatures/          # 서명 관련
└── assets/              # 리소스
```

## 3. 멀티 에이전트 시스템

### 에이전트 구성

| 에이전트 | 모델 | 역할 |
|---------|------|------|
| **Sisyphus** | Opus 4.5 High | 메인 오케스트레이터, 전체 작업 조율 |
| **Prometheus** | - | 플래너, 작업 계획 수립 |
| **Metis** | - | 계획 컨설턴트 |
| **Oracle** | GPT 5.2 Medium | 설계, 디버깅, 고수준 전략 |
| **Frontend Engineer** | Gemini 3 Pro | UI/UX 개발 전담 |
| **Librarian** | Claude Sonnet 4.5 | 공식 문서, 오픈소스 코드 탐색 |
| **Explore** | Grok Code | 빠른 코드베이스 grep |
| **Document Writer** | - | 문서 작성 |
| **Multimodal Looker** | - | 이미지/멀티모달 처리 |

### 에이전트 동작 방식
1. Sisyphus가 작업을 받으면 직접 파일 탐색하지 않음
2. 백그라운드 태스크로 더 빠르고 저렴한 모델에 병렬 위임
3. UI 작업은 Gemini 3 Pro에게 위임
4. 막히면 GPT 5.2에게 전략적 조언 요청
5. 복잡한 프레임워크 작업 시 서브에이전트가 소스코드/문서 실시간 분석

## 4. 핵심 기능

### 4.1 Claude Code 호환 레이어
- **Commands**: 슬래시 명령어
- **Agents**: 에이전트 정의
- **Skills**: 스킬 시스템
- **MCPs**: Model Context Protocol 통합
- **Hooks**: PreToolUse, PostToolUse, UserPromptSubmit, Stop

### 4.2 백그라운드 에이전트
- 병렬로 여러 에이전트 실행
- Provider/모델별 동시성 제한 설정 가능
- 실제 개발팀처럼 분업 구조

### 4.3 LSP/AST 도구
- 전체 LSP 지원
- AST-Grep 통합
- 리팩토링, 리네임, 진단 기능
- AST 기반 코드 검색

### 4.4 생산성 기능

| 기능 | 설명 |
|-----|------|
| **Todo Continuation Enforcer** | 작업 완료까지 에이전트 강제 실행 |
| **Comment Checker** | 과도한 주석 방지, 인간이 쓴 것처럼 유지 |
| **Ralph Loop** | 반복 실행 패턴 |
| **Think Mode** | 심층 분석 모드 |
| **ultrawork/ulw** | 매직 키워드로 모든 기능 자동 활성화 |

### 4.5 내장 MCP
- **Exa**: 웹 검색
- **Context7**: 공식 문서 검색
- **Grep.app**: GitHub 코드 검색

### 4.6 세션 도구
- 세션 히스토리 목록/읽기/검색/분석

## 5. 설정 시스템

### 설정 파일 위치
```
~/.config/opencode/
├── opencode.json          # OpenCode 기본 설정
└── oh-my-opencode.json    # 플러그인 설정

프로젝트/
└── .opencode/
    └── oh-my-opencode.json    # 프로젝트별 설정
```

### JSONC 지원
- 주석 허용
- 후행 쉼표 허용

### 에이전트 설정 예시
```jsonc
{
  "agents": {
    "oracle": {
      "model": "gpt-5.2-medium",
      "temperature": 0.7,
      "permissions": ["read", "write", "bash"]
    }
  }
}
```

### 백그라운드 태스크 설정
```jsonc
{
  "background_tasks": {
    "concurrency": {
      "openai": 3,
      "anthropic": 2
    }
  }
}
```

## 6. Hook 시스템

25개 이상의 내장 Hook 제공:
- `PreToolUse`: 도구 사용 전
- `PostToolUse`: 도구 사용 후
- `UserPromptSubmit`: 사용자 프롬프트 제출 시
- `Stop`: 에이전트 중지 시

비활성화 설정:
```jsonc
{
  "disabled_hooks": ["comment-checker"]
}
```

## 7. 기술 스택

- **언어**: TypeScript 99.8%, JavaScript 0.2%
- **패키지 매니저**: Bun
- **TUI 기반**: OpenCode
- **라이선스**: 커스텀 (LICENSE.md)

## 8. ppopcode와의 비교

| 항목 | oh-my-opencode | ppopcode |
|------|---------------|----------|
| **기반** | OpenCode TUI | Claude Code CLI |
| **UI** | 터미널 TUI (비주얼) | 텍스트 + WF CC Studio |
| **에이전트** | 다중 모델 오케스트레이션 | 단일 (Claude) |
| **IDE 연동** | 독립 실행 | Cursor 연동 |
| **설정** | JSONC | YAML + Markdown |
| **플랫폼** | 크로스 플랫폼 | Windows 중심 |

## 9. 적용 가능한 아이디어

### 단기
- [ ] Todo Continuation Enforcer 개념 도입
- [ ] Comment Checker Hook 구현
- [ ] 멀티 에이전트 위임 패턴 연구

### 중기
- [ ] 백그라운드 태스크 시스템 설계
- [ ] MCP 서버 통합 (Exa, Context7)
- [ ] 세션 히스토리 관리

### 장기
- [ ] TUI 인터페이스 구축 (Textual/Ink)
- [ ] 다중 모델 오케스트레이션

## 10. 참고 링크

- GitHub: https://github.com/code-yeongyu/oh-my-opencode
- Discord: https://discord.gg/PUwSMR9XNk
- 설치 가이드: https://raw.githubusercontent.com/code-yeongyu/oh-my-opencode/refs/heads/master/docs/guide/installation.md
