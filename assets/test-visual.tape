# Quick visual test — kip demo
# Run: bash assets/demo-setup.sh && vhs assets/test-visual.tape

Output assets/test-visual.gif

Set Shell "bash"
Set FontSize 16
Set FontFamily "JetBrains Mono"
Set Width 1200
Set Height 650
Set Theme "TokyoNight"
Set Padding 36
Set Margin 80
Set MarginFill "#0f172a"
Set WindowBar Colorful
Set WindowBarSize 40
Set BorderRadius 10
Set TypingSpeed 0

Hide
Type "source /tmp/kip-demo-init.sh"
Enter
Sleep 400ms
Type "clear"
Enter
Sleep 300ms
Show

Set TypingSpeed 65ms

Type "cat env.staging"
Enter
Sleep 3500ms

Type "kip push env.staging"
Enter
Sleep 3500ms

Sleep 1000ms
Type "clear"
Enter
Sleep 400ms

Type "# -- teammate's machine --"
Enter
Sleep 1200ms

Type "kip pull https://kipenv.dev/s/x8f9k2#3Kj8mNvP7qR2wX5yZ9aB4cD6eF8gH0jK"
Enter
Sleep 3000ms

Type "cat .env"
Enter
Sleep 4000ms
