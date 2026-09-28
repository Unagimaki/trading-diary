#!/bin/sh
set -eu

chown app:app /data/uploads
exec su-exec app:app /app/api
