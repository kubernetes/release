#!/usr/bin/env bash
#
# gitlib.sh unit tests
#
source "$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/common.sh"
source $TOOL_LIB_PATH/gitlib.sh
source $TOOL_LIB_PATH/releaselib.sh

##############################################################################
# TESTING VER_REGEX regex's
##############################################################################
echo "Testing VER_REGEX:"
echo
# Ex. v1.5.0-alpha.2.435+8c67d08e3a535d
DOTZERO="v1.4.0"
SHORT="v1.4.5"
LONG="v1.4.5-alpha.0"
SHA="v1.4.5-alpha.0.435+8c67d08e3a535d"
RC="v1.4.5-rc.1"

printf "%-40s" "$DOTZERO : "
if [[ $DOTZERO =~ ${VER_REGEX[dotzero]} ]]; then
  echo "$PASSED Value: ${BASH_REMATCH[0]}"
else
  echo "$FAILED"
fi

printf "%-40s" "$SHORT : "
if [[ $SHORT =~ ${VER_REGEX[release]} ]]; then
  echo "$PASSED Value: ${BASH_REMATCH[0]}"
else
  echo "$FAILED"
fi

printf "%-40s" "$LONG : "
if [[ $LONG =~ ${VER_REGEX[release]} ]]; then
  echo "$PASSED Value: ${BASH_REMATCH[0]}"
else
  echo "$FAILED"
fi

printf "%-40s" "$SHA : "
if [[ $SHA =~ ${VER_REGEX[release]}\.${VER_REGEX[build]} ]]; then
  echo "$PASSED Value: ${BASH_REMATCH[0]}"
else
  echo "$FAILED"
fi

printf "%-40s" "$RC : "
if [[ $RC =~ ${VER_REGEX[release]} ]]; then
  echo "$PASSED Value: ${BASH_REMATCH[0]}"
else
  echo "$FAILED"
fi

##############################################################################
# END TESTING VER_REGEX regex's
##############################################################################
