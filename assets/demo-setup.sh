#!/bin/bash
ASSETS_DIR="$(cd "$(dirname "$0")" && pwd)"

rm -rf /tmp/envdemo /tmp/envbin
mkdir -p /tmp/envdemo/teammate /tmp/envbin

cp "$ASSETS_DIR/.env.demo" /tmp/envdemo/env.staging
cp "$ASSETS_DIR/.env.demo" /tmp/envdemo/teammate/.env

cp "$ASSETS_DIR/fake-envshare" /tmp/envbin/kip
chmod +x /tmp/envbin/kip

# Init scripts sourced by VHS — \001/\002 are the \[/\] equivalents inside $'...'
cat > /tmp/kip-sender-init.sh << 'SETUP'
export PATH=/tmp/envbin:$PATH
cd /tmp/envdemo
PS1=$'\001\033[38;5;111m\002~/project\001\033[0m\002 \001\033[38;5;240m\002\u203a\001\033[0m\002 '
SETUP

cat > /tmp/kip-receiver-init.sh << 'SETUP'
export PATH=/tmp/envbin:$PATH
cd /tmp/envdemo/teammate
PS1=$'\001\033[38;5;210m\002~/teammate\001\033[0m\002 \001\033[38;5;240m\002\u203a\001\033[0m\002 '
SETUP

# Keep old init for test-visual.tape compatibility
cp /tmp/kip-sender-init.sh /tmp/kip-demo-init.sh

echo "READY"
