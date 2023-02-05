#!/bin/sh

main() {
  if [ "${TRACE-}" ]; then
    set -x
  fi

  CACHE_DIR=$(echo_cache_dir)
  STANDALONE_INSTALL_PREFIX=${STANDALONE_INSTALL_PREFIX:-$HOME/.local}
  VERSION=${VERSION:-$(echo_latest_version)}
    # These can be overridden for testing but shouldn't normally be used as it can
    # result in a broken code-server.
  OS=${OS:-$(os)}
  ARCH=${ARCH:-$(arch)}

  DISTRO=$(distro_name)

  echo "$OS $ARCH $DISTRO $CACHE_DIR $STANDALONE_INSTALL_PREFIX $VERSION"

  case $DISTRO in
    macos) install_standalone ;;
    debian) install_standalone ;;
    fedora | opensuse) install_standalone ;;
    *) echoh "Unsupported package manager." ;;
  esac
}

install_standalone() {
  echoh "Installing v$VERSION of the $ARCH release from GitHub."
  echoh

#  fetch "https://github.com/coder/code-server/releases/download/v$VERSION/code-server-$VERSION-$OS-$ARCH.tar.gz" \
#    "$CACHE_DIR/code-server-$VERSION-$OS-$ARCH.tar.gz"

 # sh_c mkdir -p "$STANDALONE_INSTALL_PREFIX" 2> /dev/null || true

#  sh_c="sh_c"
#  if [ ! -w "$STANDALONE_INSTALL_PREFIX" ]; then
#    sh_c="sudo_sh_c"
#  fi

#  "$sh_c" mkdir -p "$STANDALONE_INSTALL_PREFIX/lib" "$STANDALONE_INSTALL_PREFIX/bin"
#  "$sh_c" tar -C "$STANDALONE_INSTALL_PREFIX/lib" -xzf "$CACHE_DIR/code-server-$VERSION-$OS-$ARCH.tar.gz"
#  "$sh_c" mv -f "$STANDALONE_INSTALL_PREFIX/lib/code-server-$VERSION-$OS-$ARCH" "$STANDALONE_INSTALL_PREFIX/lib/code-server-$VERSION"
#  "$sh_c" ln -fs "$STANDALONE_INSTALL_PREFIX/lib/code-server-$VERSION/bin/code-server" "$STANDALONE_INSTALL_PREFIX/bin/code-server"
}

has_standalone() {
  case $ARCH in
    amd64) return 0 ;;
    # We only have amd64 for macOS.
    arm64)
      [ "$(distro)" != macos ]
      return
      ;;
    *) return 1 ;;
  esac
}

os() {
  uname="$(uname)"
  case $uname in
    Linux) echo linux ;;
    Darwin) echo macos ;;
    FreeBSD) echo freebsd ;;
    *) echo "$uname" ;;
  esac
}

distro() {
  if [ "$OS" = "macos" ] || [ "$OS" = "freebsd" ]; then
    echo "$OS"
    return
  fi

  if [ -f /etc/os-release ]; then
    (
      . /etc/os-release
      if [ "${ID_LIKE-}" ]; then
        for id_like in $ID_LIKE; do
          case "$id_like" in debian | fedora | opensuse | arch)
            echo "$id_like"
            return
            ;;
          esac
        done
      fi

      echo "$ID"
    )
    return
  fi
}

distro_name() {
  if [ "$(uname)" = "Darwin" ]; then
    echo "macOS v$(sw_vers -productVersion)"
    return
  fi

  if [ -f /etc/os-release ]; then
    (
      . /etc/os-release
      echo "$PRETTY_NAME"
    )
    return
  fi

  # Prints something like: Linux 4.19.0-9-amd64
  uname -sr
}

arch() {
  uname_m=$(uname -m)
  case $uname_m in
    aarch64) echo arm64 ;;
    x86_64) echo amd64 ;;
    *) echo "$uname_m" ;;
  esac
}

command_exists() {
  if [ ! "$1" ]; then return 1; fi
  command -v "$@" > /dev/null
}

sudo_sh_c() {
  if [ "$(id -u)" = 0 ]; then
    sh_c "$@"
  elif command_exists doas; then
    sh_c "doas $*"
  elif command_exists sudo; then
    sh_c "sudo $*"
  elif command_exists su; then
    sh_c "su root -c '$*'"
  else
    echoh
    echoerr "This script needs to run the following command as root."
    echoerr "  $*"
    echoerr "Please install doas, sudo, or su."
    exit 1
  fi
}

echoh() {
  echo "$@" | humanpath
}

humanpath() {
  sed "s# $HOME# ~#g; s#\"$HOME#\"\$HOME#g"
}

echo_cache_dir() {
  if [ "${XDG_CACHE_HOME-}" ]; then
    echo "$XDG_CACHE_HOME/dview"
  elif [ "${HOME-}" ]; then
    echo "$HOME/.cache/dview"
  else
    echo "/tmp/dview-cache"
  fi
}

echo_latest_version() {
  version="$(curl -fsSLI -o /dev/null -w "%{url_effective}" https://github.com/dimcz/dview/releases/latest)"
  version="${version#https://github.com/dimcz/dview/releases/tag/}"
  version="${version#v}"
  echo "$version"
}

main "$@"