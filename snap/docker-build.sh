#!/bin/bash

docker run --rm -v $PWD:/root/snap -w /root/snap -e SNAPCRAFT_SETUP_CORE=1\
    snapcraft/xenial-amd64 make "$@"
