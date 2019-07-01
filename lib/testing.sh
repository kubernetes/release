#!/usr/bin/env bash

testMain() {
  local tests=( "$@" )
  local t

  if [ "$#" -lt 1 ]
  then
    # if no functions are given as arguments, find all functions
    # named 'TEST_...' and run those
    mapfile tests <<< "$( declare -F | awk '$3 ~ "^TEST_" { print $3 }' )"
  fi

  for t in "${tests[@]}"
  do
    # run the tests in a subshell, so that they are isolated
    # from each other
    ( $t ; )
    echo
  done
}

assertEqualContent() {
  local actual_file="$1"
  local expected_file="$2"
  local message="${3:-files do not match content}"
  local rc=0

  diff="$( diff -Naur "$expected_file" "$actual_file" )" || rc=$?

  if [ "$rc" -ne 0 ]; then
    echo "${FAILED}: ${message}"
    echo "${diff}"
  else
    echo "${PASSED}: ${message}"
  fi

  return $rc
}
