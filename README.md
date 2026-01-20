# ppopcode

> **[한국어 설치 가이드 보기](README.ko.md)**

**Multi-agent AI coding assistant with TUI interface.**

---

### Why ppopcode?

**Tired of switching between ChatGPT, Claude, and Cursor?**

ppopcode brings all your AI tools into one terminal. Just say what you want, and it picks the right AI for the job.

---

### Key Features

| Feature | Description |
|---------|-------------|
| **One Interface** | Multiple AIs in a single terminal |
| **Smart Routing** | Auto-selects the best AI for each task |
| **Cursor Integration** | Code edits handled automatically |
| **Workflows** | Automate repetitive tasks |

---

### Who is this for?

- Developers tired of copy-pasting between AI tools
- Cursor subscribers who want to maximize their plan
- Anyone who prefers a clean, terminal-based workflow

---

### Before vs After

**Before (The Old Way)**
```
1. Open ChatGPT in browser
2. Ask question, copy answer
3. Open Cursor
4. Paste code
5. Stuck? Go back to step 1...
→ Repeat, repeat, repeat 😩
```

**After (ppopcode)**
```
1. Run ppopcode in terminal
2. Type what you want
3. Done! 😊
```

---

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

### cc-wf-studio Extension (Optional)

For workflow features, install the cc-wf-studio VSCode extension:

1. Open VSCode/Cursor
2. Go to Extensions (`Ctrl+Shift+X`)
3. Search `cc-wf-studio`
4. Click Install

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
