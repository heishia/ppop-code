# ppopcode

**여러 AI를 하나로! 터미널에서 실행되는 AI 코딩 비서**

---

### 왜 ppopcode인가요?

**"ChatGPT 열고, Claude 열고, Cursor 열고... 너무 번거로워요"**

이런 경험 있으시죠? ppopcode는 이 문제를 해결합니다.

---

### 핵심 기능

| 기능 | 설명 |
|------|------|
| **원스톱 AI** | 여러 AI를 하나의 터미널에서 |
| **자동 라우팅** | 작업에 맞는 AI를 알아서 선택 |
| **Cursor 연동** | 코드 수정은 Cursor가 자동 처리 |
| **워크플로우** | 반복 작업을 자동화 |

---

### 이런 분께 추천해요

- 여러 AI 도구를 왔다 갔다 하느라 지치신 분
- Cursor 구독료를 제대로 활용하고 싶은 분
- 터미널 기반의 깔끔한 작업 환경을 원하시는 분
- 작업별로 최적의 AI를 자동으로 쓰고 싶은 분

---

### Before vs After

**Before (기존 방식)**
```
1. 브라우저에서 ChatGPT 열기
2. 질문하고 답변 복사
3. Cursor 열기
4. 붙여넣기
5. 막히면 다시 1번으로...
→ 반복, 반복, 반복 😩
```

**After (ppopcode)**
```
1. 터미널에서 ppopcode 실행
2. 원하는 것 입력
3. 끝! 😊
```

---

## 이게 뭔가요?

ppopcode는 여러 AI(Claude, GPT, Gemini)를 하나의 터미널 화면에서 사용할 수 있게 해주는 프로그램이에요.

당신이 "로그인 화면 만들어줘"라고 하면:
1. ppopcode가 알아서 적절한 AI를 선택해요 (UI니까 Gemini!)
2. AI가 계획을 세워요
3. 실제 코드 수정은 Cursor가 해줘요

복사 붙여넣기 없이, 창 왔다 갔다 없이, 터미널 하나에서 끝!

## 설치 방법

### 1단계: Go 설치하기

Go가 없으면 먼저 설치해야 해요.

**Windows:**
1. https://go.dev/dl/ 에서 Windows용 설치파일 다운로드
2. 설치파일 실행
3. 터미널(PowerShell) 재시작

**Mac:**
```bash
brew install go
```

**Linux:**
```bash
sudo apt install golang-go
```

설치 확인:
```bash
go version
```
버전이 나오면 성공!

### 2단계: ppopcode 다운로드

```bash
git clone https://github.com/ppopcode/ppopcode.git
cd ppopcode
```

### 3단계: 빌드하기

```bash
go mod tidy
go build -o ppopcode ./cmd/ppopcode
```

Windows에서는:
```powershell
go mod tidy
go build -o ppopcode.exe ./cmd/ppopcode
```

### 4단계: 실행하기

```bash
./ppopcode
```

Windows에서는:
```powershell
.\ppopcode.exe
```

## cc-wf-studio 설치 (선택사항)

워크플로우 기능을 사용하려면 cc-wf-studio VSCode 익스텐션이 필요해요.

### 설치 방법

1. **VSCode 또는 Cursor 열기**

2. **익스텐션 탭 열기**
   - 왼쪽 사이드바에서 네모 4개 아이콘 클릭
   - 또는 `Ctrl+Shift+X` (Windows) / `Cmd+Shift+X` (Mac)

3. **검색창에 입력**
   ```
   cc-wf-studio
   ```

4. **Install 버튼 클릭**

### 사용법

1. VSCode/Cursor에서 `Ctrl+Shift+P` (또는 `Cmd+Shift+P`)
2. `cc-wf-studio` 입력
3. 워크플로우 에디터가 열려요
4. 노드를 드래그해서 작업 흐름을 만들어요
5. 저장하면 `.vscode/workflows/` 폴더에 JSON 파일이 생겨요

### 워크플로우란?

워크플로우는 "작업 순서"를 미리 정해놓는 거예요.

예를 들어 "새 프로젝트 시작" 워크플로우를 만들면:
1. 아이디어 문서 읽기
2. 계획 세우기
3. 코드 작성하기

이 순서를 매번 말할 필요 없이, 워크플로우 하나만 선택하면 자동으로 진행돼요!

## API 키 설정 (선택사항)

AI 기능을 사용하려면 API 키가 필요해요. 없어도 프로그램은 실행돼요!

**Windows (PowerShell):**
```powershell
$env:ANTHROPIC_API_KEY="여기에-클로드-API키"
$env:OPENAI_API_KEY="여기에-OpenAI-API키"
$env:GOOGLE_API_KEY="여기에-구글-API키"
```

**Mac/Linux:**
```bash
export ANTHROPIC_API_KEY="여기에-클로드-API키"
export OPENAI_API_KEY="여기에-OpenAI-API키"
export GOOGLE_API_KEY="여기에-구글-API키"
```

## 사용법

### 메인 메뉴

프로그램을 실행하면 메뉴가 나와요:

- **Chat**: AI와 대화하면서 코딩하기
- **Workflow**: 미리 만든 작업 흐름 실행하기
- **Settings**: 설정 바꾸기

### 키보드 조작

- `↑` `↓` 또는 `j` `k`: 위/아래 이동
- `Enter`: 선택
- `Esc`: 뒤로가기
- `q`: 종료

### 예시

```
Chat 선택 → "로그인 폼 만들어줘" 입력 → Enter

ppopcode: UI 작업이네요! Gemini에게 맡길게요.
[gemini] 로그인 폼을 만들어드릴게요...
         Cursor로 코드를 수정합니다.
```

## 어떤 AI가 어떤 일을 하나요?

| 이런 말을 하면 | 이 AI가 담당해요 |
|--------------|-----------------|
| UI, 디자인, 화면, 컴포넌트 | Gemini |
| 버그, 에러, 수정, 디버깅 | GPT |
| 설계, 구조, 아키텍처 | GPT |
| 그 외 일반 코딩 | Sonnet |

## 자주 묻는 질문

**Q: Cursor가 꼭 있어야 하나요?**
A: 아니요! 코드 수정 기능은 Cursor가 있으면 자동으로 연동되고, 없으면 그냥 답변만 받을 수 있어요.

**Q: API 키 없이도 되나요?**
A: 프로그램 자체는 실행돼요. 다만 AI 응답은 "API 키가 없어요"라는 안내 메시지가 나와요.

**Q: 무료인가요?**
A: ppopcode 자체는 무료예요! 다만 AI API를 사용하면 각 서비스(OpenAI, Anthropic, Google)에 비용이 발생할 수 있어요.

## 문제가 생겼을 때

### "go를 찾을 수 없어요" 에러
→ Go 설치 후 터미널을 껐다 켜세요.

### 빌드할 때 에러가 나요
→ `go mod tidy`를 먼저 실행해보세요.

### 실행은 되는데 화면이 깨져요
→ 터미널 크기를 키워보세요. 최소 80x24 정도 필요해요.

## 기여하기

버그를 발견하거나 아이디어가 있으면 Issue를 열어주세요!

## 라이선스

MIT - 마음대로 쓰세요!
