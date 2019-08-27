#!/usr/bin/env bash
set -e

DOCKER_OPTS=${DOCKER_OPTS:-""}
IFS=" " read -r -a DOCKER <<< "docker ${DOCKER_OPTS}"
detach=false

# If we have stdin we can run interactive.  This allows things like 'shell.sh'
# to work.  However, if we run this way and don't have stdin, then it ends up
# running in a daemon-ish mode.  So if we don't have a stdin, we explicitly
# attach stderr/stdout but don't bother asking for a tty.
if [[ -t 0 ]]; then
  docker_run_opts+=(--interactive --tty)
elif [[ "${detach}" == false ]]; then
  docker_run_opts+=("--attach=stdout" "--attach=stderr")
fi

docker_run_cmd=("${DOCKER[@]}" run "${docker_run_opts[@]}")

docker build -t kubelet-rpm-builder .
echo "Cleaning output directory..."
sudo rm -rf output/*
mkdir -p output
"${docker_run_cmd[@]}" --rm -v $PWD/output/:/root/rpmbuild/RPMS/ kubelet-rpm-builder $1
sudo chown -R $USER $PWD/output

echo
echo "----------------------------------------"
echo
echo "RPMs written to: "
ls $PWD/output/*/
echo
echo "Yum repodata written to: "
ls $PWD/output/*/repodata/
