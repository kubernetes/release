#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

docker_exec() {
  docker exec --tty --interactive "${@}"
}

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
git_toplevel="$(git rev-parse --show-toplevel)"
packages_sub="debian/bin/unstable/xenial"
packages_dir="${git_toplevel}/${packages_sub}"

docker build --tag=xenial-verify "${script_dir}"

c="$(docker run --detach --privileged --volume="${packages_dir}:/pkg" --env=TERM xenial-verify)"

until docker_exec "${c}" systemctl is-system-running | grep -q running ; do sleep 1 ; done

docker_exec "${c}" systemctl start docker.service


docker_exec "${c}" bash -c "find /pkg -iname '*.deb' | xargs dpkg --install"

docker_exec "${c}" kubeadm init --token=389977.b2e7847655c5c561
