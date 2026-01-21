# ppopcode

> **[한국어 설치 가이드 보기](README.ko.md)**

**AI coding assistant with TUI interface, powered by Claude.**

---

### Why ppopcode?

**Tired of switching between your IDE and AI chat windows?**

ppopcode brings Claude directly into your terminal. Just say what you want, and it handles everything - from answering questions to editing code via Cursor.

---

### Key Features

| Feature | Description |
|---------|-------------|
| **Terminal UI** | Clean, efficient TUI interface |
| **Claude Powered** | Uses Claude Code for intelligent responses |
| **Cursor Integration** | Code edits handled automatically |
| **Workflows** | Automate repetitive tasks |
| **Session Context** | Maintains conversation history |

---

### Who is this for?

- Developers who prefer terminal-based workflows
- Cursor subscribers who want to maximize their plan
- Anyone who wants AI coding assistance without leaving the terminal

---

### Before vs After

**Before (The Old Way)**
```
1. Open browser for AI chat
2. Ask question, copy answer
3. Open Cursor
4. Paste code
5. Stuck? Go back to step 1...
→ Repeat, repeat, repeat
```

**After (ppopcode)**
```
1. Run ppopcode in terminal
2. Type what you want
3. Done!
```

---

## Quick Start

### Prerequisites
- Go 1.21+
- **Claude Code subscription** - Run `claude login` to authenticate
- **Cursor subscription** - For code editing features (optional)

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

**Cursor (Optional, for code editing)**
- Open Cursor IDE and sign in with your subscription

## How It Works

```
You: "Create a login form"
     ↓
ppopcode: Sends request to Claude
     ↓
Claude analyzes and responds
     ↓
Code edits are applied via Cursor
```

## Controls

- `↑/↓` or `j/k`: Navigate
- `Enter`: Select
- `Esc`: Back
- `q`: Quit
- `/clean`: Clear chat history

## Documentation

- [한국어 가이드](README.ko.md)

## Version

**v1.0.0** - Initial stable release

## License

MIT
