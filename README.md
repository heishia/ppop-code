# ppopcode

> **[한국어 설치 가이드 보기](README.ko.md)**

Multi-agent AI coding assistant with TUI interface.

## What is this?

ppopcode combines multiple AI models (Claude, GPT, Gemini) into a single terminal interface. You say what you want, and it automatically picks the right AI for the job.

## Quick Start

### Prerequisites
- Go 1.21+
- (Optional) Cursor for code editing

### Installation

```bash
git clone https://github.com/ppopcode/ppopcode.git
cd ppopcode
go mod tidy
go build -o ppopcode ./cmd/ppopcode
./ppopcode
```

### Environment Variables (Optional)

```bash
export ANTHROPIC_API_KEY="your-key"
export OPENAI_API_KEY="your-key"
export GOOGLE_API_KEY="your-key"
```

## How It Works

```
You: "Create a login form"
     ↓
ppopcode: "UI task? Let me ask Gemini!"
     ↓
Gemini plans the work
     ↓
Cursor edits the code
```

| Task Type | AI Agent |
|-----------|----------|
| UI/UX | Gemini |
| Debug/Design | GPT |
| General coding | Sonnet |

## Controls

- `↑/↓` or `j/k`: Navigate
- `Enter`: Select
- `Esc`: Back
- `q`: Quit

## Documentation

- [한국어 가이드](README.ko.md)
- [Technical Docs](docs/README.md)
- [Architecture Plan](docs/plan/refactoring-plan.md)

## License

MIT
