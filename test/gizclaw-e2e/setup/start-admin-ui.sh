#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../../.." && pwd)"
testdata_dir="$repo_root/test/gizclaw-e2e/testdata"
bin_path="$testdata_dir/bin/gizclaw"
config_home="$testdata_dir/admin-config-home"
pid_file="$testdata_dir/admin-ui.pid"
log_file="$testdata_dir/admin-ui.log"
listen_addr="127.0.0.1:8080"

if [[ ! -x "$bin_path" ]]; then
  "$script_dir/build.sh" >/dev/null
fi

if [[ -f "$pid_file" ]]; then
  pid="$(cat "$pid_file")"
  if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
    echo "gizclaw e2e admin UI already running: pid=$pid url=http://$listen_addr"
    exit 0
  fi
  rm -f "$pid_file"
fi

(
  cd "$repo_root"
  export XDG_CONFIG_HOME="$config_home"
  exec "$bin_path" admin --context e2e-admin --listen "$listen_addr"
) >"$log_file" 2>&1 &
pid="$!"
echo "$pid" >"$pid_file"
echo "gizclaw e2e admin UI pid=$pid url=http://$listen_addr log=$log_file"
