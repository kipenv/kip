# envshare — AI Scan demo GIF
# Records: scan detecting dangerous keys before sharing
# Install VHS: https://github.com/charmbracelet/vhs

Output assets/scan.gif

Set Shell "bash"
Set FontSize 14
Set Width 900
Set Height 500
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 40ms

# Scan
Type "envshare scan .env.production"
Enter
Sleep 2s

# Simulated output:
#   Scanning .env.production (14 variables)...
#
#   WARN  AWS_SECRET_ACCESS_KEY    live AWS credential detected
#   WARN  STRIPE_SECRET_KEY        live Stripe key (sk_live_*)
#   WARN  JWT_SECRET               weak secret (< 32 chars)
#   WARN  DATABASE_URL             points to production host
#   OK    REDIS_URL                no issues
#   OK    LOG_LEVEL                no issues
#   OK    PORT                     no issues
#
#   4 warnings found. Review before sharing.

Sleep 3s

# Push with confirmation
Type "envshare push .env.production --expires 10m"
Enter
Sleep 2s

# Simulated output:
#   ⚠ 4 warnings detected. Continue? [y/N] y
#   ✓ Encrypted with AES-256-GCM
#   Link: https://envshare.dev/s/m4kd92#7Hj2pQ...
#   Expires: 10 minutes | Reads: 1

Sleep 2s
