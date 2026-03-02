Output docs/demo-showcase.gif
Set Shell "bash"
Set FontSize 16
Set Width 1200
Set Height 700
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms

Hide
Type "cd /tmp/grove-demo/rails-app"
Enter
Sleep 500ms
Show

# Launch dashboard — hero shot
Type "./bin/grove"
Enter
Sleep 2.5s

# Navigate through worktrees — detail panel updates
Down
Sleep 1s
Down
Sleep 1s
Down
Sleep 1s

# Toggle to compact view
Type "v"
Sleep 1.5s

# Toggle back to rich view
Type "v"
Sleep 1s

# Filter by name
Type "/"
Sleep 500ms
Type "feat"
Sleep 1.2s

# Clear filter
Escape
Sleep 800ms

# Sort cycle — dirty first
Type "o"
Sleep 1s

# Sort cycle — back to default
Type "o"
Sleep 1s

# Quit
Type "q"
