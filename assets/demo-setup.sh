#!/bin/bash
rm -rf /tmp/envdemo /tmp/envbin
mkdir -p /tmp/envdemo/teammate /tmp/envbin

# Copy env file WITHOUT the dot so ls shows it
cp ~/proyectos/kips/kip/assets/.env.demo /tmp/envdemo/env.staging
cp ~/proyectos/kips/kip/assets/.env.demo /tmp/envdemo/teammate/.env

# Fake CLI
cp ~/proyectos/kips/kip/assets/fake-kip /tmp/envbin/kip

echo "READY"
