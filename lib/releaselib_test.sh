#!/bin/bash
#
# releaselib.sh unit tests
#
source $(dirname $(readlink -ne $BASH_SOURCE))/common.sh
source $TOOL_LIB_PATH/gitlib.sh
source $TOOL_LIB_PATH/releaselib.sh


##############################################################################
# TESTING release::gcs::verify_latest_update()
##############################################################################
# stable is always vX.Y.Z
# latest is always vX.Y.Z-(alpha|beta).N

# > vs >= scenarios only relevant to 'latest' .N's.
# type release updates on >
# type ci updates on >=

published_file=/tmp/published.$$

# Fill $data with testing values and expected state
read -r -d '' data <<'EOF' || true
###############################################
# stable test scenarios
# release and ci use same logic here
#  type    incoming       published      State
#  ----    --------       ---------      -----
## release v2.0.0         v1.0.0         0
## release v1.0.0         v1.0.0         1
## release v0.9.9         v1.0.0         1

## release v1.1.0         v1.0.0         0
## release v1.1.0         v1.1.0         1
## release v1.0.0         v1.1.0         1

## release v1.3.4         v1.3.3         0
## release v1.3.3         v1.3.3         1
## release v1.3.2         v1.3.3         1


###############################################
# latest test scenarios
#  type    incoming       published      State
#  ----    --------       ---------      -----
## release v1.4.0-alpha.0 v1.4.0         1
## release v1.4.0-alpha.0 v1.4.0-beta.0  1
## release v1.4.0-beta.0  v1.4.0-alpha.0 0
## ci      v1.4.0-alpha.1 v1.4.0-alpha.0 0
## ci      v1.4.0-alpha.0 v1.4.0-alpha.0 0
## release v1.4.0-alpha.1 v1.4.0-alpha.0 0
## release v1.4.0-alpha.0 v1.4.0-alpha.0 1
EOF


# Test the data
while read comment type version pub_version expected; do
  # Prepare test
  echo $pub_version > $published_file

  # $type value passed in simply to trigger > vs. >= condition
  # arg 2 (bucket) not used with optional arg 4 passed in
  # arg 3 (version) is the incoming version to check
  # arg 4 simply points to a local file to set a 'published' version
  if release::gcs::verify_latest_update $type "" $version $published_file; then
    echo -n "TEST CASE: "
    case $expected in
      0) echo "$PASSED" ;;
      *) echo "$FAILED" ;;
    esac
  else
    echo -n "TEST CASE: "
    case $expected in
      1) echo "$PASSED" ;;
      *) echo "$FAILED" ;;
    esac
  fi
  echo
done < <(echo "$data" |egrep '^## ')

# Garbage collection
rm -f $published_file

##############################################################################
# END TESTING release::gcs::verify_latest_update()
##############################################################################
