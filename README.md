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
- **Claude Code subscription** - Run `claude login` to authenticate
- **Cursor subscription** - For code editing features
- (Optional) OpenAI/Google API keys for GPT/Gemini agents

### Installation

**Global Install (Recommended)**

Install once, use anywhere!

```bash
# Clone and install
git clone https://github.com/ppopcode/ppopcode.git
cd ppopcode
make install
```

That's it! Now you can run `ppopcode` from any directory.

**Windows (PowerShell)**
```powershell
git clone https://github.com/ppopcode/ppopcode.git
cd ppopcode
make install
# Or manually: .\scripts\install.ps1
```

### Running ppopcode

After global installation:
```bash
# Run from anywhere!
ppopcode
```

For local development:
```bash
# Build and run in one command
make run

# Or run the binary directly
./ppopcode        # Linux/Mac
.\ppopcode.exe    # Windows
```

### Uninstall

```bash
# Linux/Mac
make uninstall
# Or: ./scripts/install.sh uninstall

# Windows (PowerShell)
.\scripts\install.ps1 -Uninstall
```

### cc-wf-studio Extension (Optional)

For workflow features, install the cc-wf-studio VSCode extension:

1. Open VSCode/Cursor
2. Go to Extensions (`Ctrl+Shift+X`)
3. Search `cc-wf-studio`
4. Click Install

### Authentication Setup

**Use the "Link Accounts" menu in ppopcode to check and setup authentication.**

**Claude (Required)**
```bash
# Login with your Claude Code subscription
claude login
```

**Cursor (Required for code editing)**
- Open Cursor IDE and sign in with your subscription

**Optional API Keys (for GPT/Gemini agents)**
```bash
export OPENAI_API_KEY="your-key"    # For GPT agent
export GOOGLE_API_KEY="your-key"    # For Gemini agent
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

## License

MIT
