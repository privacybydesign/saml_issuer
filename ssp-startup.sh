#!/bin/bash
set -e

/opt/simplesaml/run-on-start.sh

exec apache2-foreground
