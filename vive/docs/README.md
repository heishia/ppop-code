# ppopcode

Claude Code + cc-wf-studio + Cursor를 연결하는 AI Agent Orchestration Framework

## 개요

ppopcode는 여러 AI 도구를 연결하는 "접착제" 역할을 합니다:

- **Claude Code**: 진입점 + Skills 실행 + Orchestration
- **cc-wf-studio**: 워크플로우 시각화 및 설계
- **Cursor**: 실제 코드 수정 (Max usage 활용)

## 폴더 구조

```
vive/
├── .claude/
│   ├── skills/           # Claude Code Skills
│   │   ├── cursor-edit/  # Cursor로 코드 수정 위임
│   │   ├── analyze/      # 코드 분석
│   │   └── verify/       # 결과 검증
│   └── hooks/            # 실행 전/후 hook
├── .vscode/
│   └── workflows/        # cc-wf-studio 워크플로우
├── config/
│   ├── ppopcode.yaml     # 프레임워크 설정
│   └── bridges.yaml      # 외부 도구 연동 설정
├── tests/                # 테스트
├── docs/                 # 문서
└── project-rules/        # 대상 프로젝트 규칙
```

## 실행 흐름

```
1. cc-wf-studio에서 워크플로우 설계
   └─> .vscode/workflows/*.json
   └─> Export -> .claude/commands/, .claude/agents/

2. Claude Code CLI로 실행
   $ claude /my-workflow
   └─> Skills 자동 호출 (analyze -> cursor-edit -> verify)

3. cursor-edit Skill이 Cursor 호출
   └─> cursor-agent --prompt "..."
   └─> Cursor Max usage로 실제 코드 수정
```

## 설치 요구사항

1. **Claude Code CLI**
   - https://docs.anthropic.com/claude-code

2. **Cursor CLI**
   - https://docs.cursor.com/cli/installation
   ```bash
   curl https://cursor.com/install -fsS | bash
   ```

3. **cc-wf-studio** (VSCode Extension)
   - VSCode Extensions에서 "cc-wf-studio" 검색

## 사용법

### 1. 워크플로우 설계 (cc-wf-studio)

1. VSCode에서 cc-wf-studio 열기
2. 노드 드래그 & 드롭으로 워크플로우 구성
3. Export -> .claude/ 디렉토리에 자동 생성

### 2. 워크플로우 실행 (Claude Code)

```bash
cd your-project
claude /refactor
```

### 3. 직접 Skill 호출

```bash
claude "analyze this codebase"
claude "edit this file using cursor"
```

## Skills

### cursor-edit

Cursor CLI를 통해 실제 코드 수정을 수행합니다.

```
트리거: 코드 수정, 리팩토링, 기능 추가 등
실행: cursor-agent --prompt "..."
결과: 파일 변경
```

### analyze

코드베이스를 분석하고 구조를 파악합니다.

```
트리거: 분석 요청, 코드 이해 필요
결과: 구조 분석, 문제점 식별, 제안
```

### verify

코드 변경 후 결과를 검증합니다.

```
트리거: 변경 후 검증 필요
결과: diff 확인, 테스트, 린트 검사
```

## 설정

### config/ppopcode.yaml

```yaml
framework:
  name: ppopcode
  version: 1.0.0

error_handling:
  on_cursor_fail: retry
  max_retries: 2
```

### config/bridges.yaml

```yaml
cursor:
  windows:
    command: cursor-agent
    shell: powershell
  
  common:
    timeout: 300
```

## 테스트

```bash
cd vive
python tests/test_skills.py
python tests/test_bridges.py
```

## 라이선스

MIT
