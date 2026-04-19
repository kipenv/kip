# envshare — Main demo GIF (fully simulated, no server needed)
# BEFORE RECORDING: run `bash assets/demo-setup.sh`
# THEN: cd ~/proyectos/envshares/envshare && vhs assets/demo.tape

Output assets/demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 900
Set Height 520
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 0
Type "export PATH=/tmp/envbin:$PATH && cd /tmp/envdemo"
Enter
Sleep 500ms
Type "clear"
Enter
Sleep 500ms

Set TypingSpeed 100ms

Type "ls"
Enter
Sleep 3s

Type "cat env.staging"
Enter
Sleep 5s

Type "envshare push env.staging"
Enter
Sleep 5s

Type "# On your teammate's machine:"
Enter
Sleep 2s

Type "cd teammate"
Enter
Sleep 1.5s

Type "ls"
Enter
Sleep 3s

Type "envshare pull https://envshare.dev/s/x8f9k2#3Kj8mNvP7qR2wX5yZ9aB4cD6eF8gH0jK"
Enter
Sleep 4s

Type "ls -a"
Enter
Sleep 3s

Type "cat .env"
Enter
Sleep 5s
