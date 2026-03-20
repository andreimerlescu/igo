# ~/.igosh  (or ~/go/.igosh)

if [[ -n "${IGO_SOURCED:-}" ]]; then
    return 0
fi
IGO_SOURCED=1

export IGO_ROOT="${HOME}/go"
export IGO_SHIMS="${IGO_ROOT}/shims"
export IGO_BIN="${IGO_ROOT}/bin"

case ":${PATH}:" in
  *":${IGO_SHIMS}:"*)  ;;    # already there
  *) export PATH="${IGO_SHIMS}:${IGO_BIN}:${PATH}" ;;
esac

if [[ -f "${IGO_ROOT}/version" ]]; then
    ver=$(<"${IGO_ROOT}/version")
    export GOROOT="${IGO_ROOT}/versions/${ver}/go"
    export GOBIN="${GOROOT}/bin"
    export GOPATH="${IGO_ROOT}/versions/${ver}"
    export GOMODCACHE="${GOROOT}/pkg/mod"
fi

igo-rehash() {
    # can regenerate shims if needed in the future
    :
}