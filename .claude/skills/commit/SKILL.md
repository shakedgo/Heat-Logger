---
description: Analyzes staged git changes and generates a standardized, formatted commit message.
---

# Git Commit Generator

Your task is to analyze the currently staged git changes and create a high-quality commit message.

## Step 1: Gather Context
Use your bash tool to silently run the following commands and read the output:
1. `git status`
2. `git diff --staged`

## Step 2: Analyze and Draft
Analyze the staged changes. Draft a commit message using the present tense and explain "why" something has changed, not just "what" has changed.

Only use the following emojis and types: 
- ✨ `feat:` - New feature
- 🐛 `fix:` - Bug fix
- 🔨 `refactor:` - Refactoring code
- 📝 `docs:` - Documentation
- 🎨 `style:` - Styling/formatting
- ✅ `test:` - Tests
- ⚡ `perf:` - Performance

Use this exact format:
<emoji> <type>: <concise_description>
<optional_body_explaining_why>

## Step 3: Output and Confirm
1. Show a brief summary of the changes currently staged.
2. Propose the formatted commit message.
3. **CRITICAL:** Do NOT auto-commit. Ask me for confirmation. If I approve, use your bash tool to execute the commit.