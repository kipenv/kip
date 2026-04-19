# Receiver terminal (right) — kip pull flow

Output assets/receiver.mp4

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
Type "source /tmp/kip-receiver-init.sh"
Enter
Sleep 300ms
Type "clear"
Enter
Sleep 200ms
Show

# sender: 0.8 idle + cat(1.6s type+3.5s sleep) + push(1.2s type+3.5s sleep) + 3s arrows window ≈ 13.6s
Sleep 13600ms

Set TypingSpeed 65ms

Type "kip pull https://kipenv.dev/s/x8f9k2#3Kj8mNvP7qR2wX5yZ9aB4cD6eF8gH0jK"
Enter
Sleep 3000ms

Type "cat .env"
Enter
Sleep 4000ms
