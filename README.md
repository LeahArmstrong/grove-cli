# Grove has moved

> **This repository has been archived.** Grove is now maintained at **[lost-in-the/grove](https://github.com/lost-in-the/grove)**.

---

## Migrating from `leaharmstrong/grove-cli` to `lost-in-the/grove`

### 1. Uninstall the old version

```bash
brew uninstall grove
brew untap leaharmstrong/tap
```

### 2. Clean up shell integration

Remove the old eval line from your shell config (`~/.zshrc` or `~/.bashrc`). Look for either of these:

```bash
eval "$(grove install zsh)"
eval "$(grove init zsh)"     # older versions used this
```

Delete whichever line you find.

### 3. Install the new version

```bash
brew install lost-in-the/tap/grove
```

### 4. Set up shell integration

```bash
grove setup
```

This auto-detects your shell and adds the eval line to your config. Then reload:

```bash
source ~/.zshrc   # or ~/.bashrc
```

### 5. Verify

```bash
grove version
# Should show: grove 0.6.0 darwin/arm64 (or similar)

type grove
# Should show: grove is a shell function
```

### 6. Re-initialize your projects

Your existing `.grove/` directories are still valid. `grove ls` should work immediately. If you see a shell version warning, `grove setup` already handled it.

---

**If you installed via `go install` instead of Homebrew**, replace steps 1 and 3 with:

```bash
# Uninstall old
rm "$(which grove)"

# Install new
go install github.com/lost-in-the/grove/cmd/grove@latest
```
