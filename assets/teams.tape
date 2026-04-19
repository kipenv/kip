# envshare — Teams demo GIF
# Records: create team → join → push --all → pull
# Install VHS: https://github.com/charmbracelet/vhs

Output assets/teams.gif

Set Shell "bash"
Set FontSize 14
Set Width 900
Set Height 500
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 40ms

# Create team
Type "envshare team create my-startup"
Enter
Sleep 2s

# Simulated output:
#   ✓ Team created: my-startup
#   Invite code: my-startup-a8f3k2
#   Share this code with your teammates

Sleep 1.5s

# Init in project
Type "envshare init"
Enter
Sleep 1s

# Push to team
Type "envshare push .env --all --pin"
Enter
Sleep 2s

# Simulated output:
#   ✓ Encrypted with AES-256-GCM
#   ✓ Shared with all members of my-startup
#   ✓ Pinned as official .env

Sleep 1.5s

# Teammate's side
Type "# --- Teammate joins and pulls ---"
Enter
Sleep 1s

Type "envshare team join my-startup-a8f3k2"
Enter
Sleep 1.5s

Type "envshare pull"
Enter
Sleep 2s

# Simulated output:
#   ✓ Decrypted .env from my-startup (pinned)
#   Saved to .env (12 variables)

Sleep 2s
