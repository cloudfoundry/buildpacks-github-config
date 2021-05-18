#!/usr/bin/env bash

set -eu
set -o pipefail

readonly ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# shellcheck source=SCRIPTDIR/.util/print.sh
source "${ROOT_DIR}/scripts/.util/print.sh"

function main() {
  local target repo_type

  while [[ "${#}" != 0 ]]; do
    case "${1}" in
      --target)
        target="${2}"
        shift 2
        ;;

      --help|-h)
        shift 1
        usage
        exit 0
        ;;

      "")
        # skip if the argument is empty
        shift 1
        ;;

      *)
        util::print::error "unknown argument \"${1}\""
    esac
  done

  if [[ -z "${target:-}" ]]; then
    usage
    echo
    util::print::error "--target is a required flag"
  fi

  bootstrap "${target}"
}

function usage() {
  cat <<-USAGE
bootstrap.sh --target <target> [OPTIONS]

Bootstraps a repository with github configuration and scripts.

OPTIONS
  --help  -h         prints the command usage
  --target <target>  path to a buildpack repository
USAGE
}

function bootstrap() {
  local target
  target="${1}"

  if [[ ! -d "${target}" ]]; then
    util::print::error "cannot bootstrap: \"${target}\" does not exist"
  fi

  cp -pR "${ROOT_DIR}/buildpack/." "${target}"
}

main "${@:-}"
