# Sender terminal (left) — kip push flow

Output assets/sender.mp4

Set Shell "bash"
Set FontSize 15
Set FontFamily "JetBrains Mono"
Set Width 620
Set Height 660
Set Theme "TokyoNight"
Set Padding 28
Set WindowBar Colorful
Set WindowBarSize 36
Set BorderRadius 8
Set TypingSpeed 0

Hide
Type "source /tmp/kip-sender-init.sh"
Enter
Sleep 300ms
Type "clear"
Enter
Sleep 200ms
Show

Sleep 800ms

Set TypingSpeed 65ms

Type "cat env.staging"
Enter
Sleep 3500ms

Type "kip push env.staging"
Enter
Sleep 3500ms

# Idle: 3s arrows window + receiver pull(4s) + cat(0.5s) + sleep(4s) = ~14.5s
Sleep 14600ms
