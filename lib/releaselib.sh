#!/usr/bin/env bash
#
# Copyright 2016 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

###############################################################################
# CONSTANTS
###############################################################################

readonly DEFAULT_PROJECT="k8s-staging-release-test"
# TODO(prototype): Temporarily setting this to the staging project to test
#                  the staging flow with --nomock set.
readonly PROD_PROJECT="k8s-staging-release-test"
readonly TEST_PROJECT="${TEST_PROJECT:-${PROJECT_ID:-$DEFAULT_PROJECT}}"
readonly OLD_PROJECT="kubernetes-release-test"

readonly DEFAULT_BUCKET="k8s-staging-release-test"
# TODO(prototype): Temporarily setting this to the staging project to test
#                  the staging flow with --nomock set.
readonly PROD_BUCKET="k8s-staging-release-test"
readonly MOCK_BUCKET="k8s-staging-release-test"
readonly OLD_BUCKET="kubernetes-release"

###############################################################################
# FUNCTIONS
###############################################################################

##############################################################################
# Pulls Jenkins server job cache from GS and preprocesses it into a
# dictionary stored in $job_path associating full job and build numbers
# Used by release::set_build_version()
#
# @optparam -d - dedup git's monotonically increasing (describe) build numbers
# @param job_path - Jenkins job name
#
release::get_job_cache () {
  local dedup=0
  if [[ $1 == "-d" ]]; then
    dedup=1
    shift
  fi
  local job_path=$1
  local job=${job_path##*/}
  local tempjson=$TMPDIR/$PROG-$job.$$
  local logroot="gs://kubernetes-jenkins/logs"
  local version
  local lastversion
  local buildnumber

  logecho "Getting $job build results from GCS..."
  mkdir -p ${job_path%/*}

  $GSUTIL -qm cp $logroot/$job/jobResultsCache.json $tempjson 2>&-
  # If there's no file up on $logroot, job doesn't exist.  Skip it.
  [[ -s $tempjson ]] || return

  # Additional select on .version is b/c we have so many empty versions for now
  # 2 passes.  First pass sorts by buildnumber, Second builds the dictionary.
  while read version buildnumber; do
    ((dedup)) && [[ $version == $lastversion ]] && continue
    echo "$version $buildnumber"
    lastversion=$version
  done < <(jq -r '.[] | select(.result == "SUCCESS") | select(.version != null) | [.version,.buildnumber] | "\(.[0]|rtrimstr("\n")) \(.[1])"' $tempjson |\
   LC_ALL=C sort -rn -k2,2) |\
  while read version buildnumber; do
    [[ -n $buildnumber && -n $version ]] && echo "JOB[$buildnumber]=$version"
  done > $job_path

  rm -f $tempjson
}

##############################################################################
# Sets the JENKINS_BUILD_VERSION global by cross checking it against a set of
# blocking CI jobs
# @param branch - branch name
# @optparam job_path - A local directory to store the copied cache entries
# @optparam exclude_suites - A space-separated list of (greedy) patterns to
#                            exclude CI jobs from checking against the primary
#                            job.
# @optparam hard_limit - A hard limit of primary jobs to process so this doesn't
#                        run for hours. Default should be handled by the caller.
#                        A high value default of 1000 is here to maintain
#                        previous functionality.
#
# TODO:
# * Ability to point to a particular primary job hash and validate it
#   - Useful if wanting to go back to last good from a few days ago
# * Make use of builder average time to find a better search algo
#   - find longest running green and use reasonable offsets to search for
#     dependents to reduce search time
# * e2e-gce: :20
# * e2e-gce-scalability: :42
# * e2e-gce-serial: 3:30
# * e2e-gce-slow: 1:30
# * kubemark-5-gce: 1:10
# * e2e-gce-reboot: :47
# * test-go: :45
# * e2e-gke: :37
# * e2e-gke-slow: :50
release::set_build_version () {
  local branch=$1
  local job_path=${2:-"$TMPDIR/buildresults-cache.$$"}
  local -a exclude_suites=($3)
  local hard_limit=${4:-1000}
  local exclude_patterns=$(IFS="|"; echo "${exclude_suites[*]}")
  local build_version
  local build_number
  local cache_build
  local last_cache_build
  local build_sha1
  local run
  local n
  local i
  local re
  local giveup_build_number=999999
  local job_count=0
  local max_job_length
  local other_job
  local good_job
  local retcode=0
  local good_job_count=0
  local first_build_number
  local branch_head=$($GHCURL $K8S_GITHUB_API/commits/$branch |jq -r '.sha')
  # Shorten
  branch_head=${branch_head:0:14}
  # The instructions below for installing yq put it in /usr/local/bin
  local yq=$(which yq || echo "/usr/local/bin/yq")
  local job_prefix="ci-kubernetes-"
  local -a JOB
  local -a secondary_jobs

  # Deal with somewhat inconsistent naming in config.yaml
  case $branch in
   master) branch="release-master"
  esac

  local -a all_jobs=(
    $(common::run_gobin blocking-testgrid-tests "$branch")
  )

  if [[ -z ${all_jobs[*]} ]]; then
    logecho "$FAILED: Curl to testgrid/config/config.yaml"
  fi

  local main_job="${all_jobs[0]}"

  if [[ -z "${all_jobs[*]}" ]]; then
    logecho "No sig-$branch-blocking list found in the testgrid config.yaml!"
    return 1
  fi

  # Loop through the remainder, excluding anything specified by --exclude_suites
  for ((i=1;i<${#all_jobs[*]};i++)); do
    [[ -n $exclude_patterns && ${all_jobs[$i]} =~ $exclude_patterns ]] \
        || secondary_jobs+=(${all_jobs[$i]})
  done

  # Update main cache
  # We dedup the $main_job's list of successful runs and just run through that
  # unique list. We then leave the full state of secondaries below so we have
  # finer granularity at the Jenkin's job level to determine if a build is ok.
  release::get_job_cache -d $job_path/$main_job &

  # If we're forcing a --build-at-head, we only need to capture the $main_job
  # details.
  if ! ((FLAGS_build_at_head)); then
    # Update secondary caches limited by main cache last build number
      for other_job in ${secondary_jobs[@]}; do
    release::get_job_cache $job_path/$other_job &
    done
  fi

  # Wait for background fetches.
  wait

  if ((FLAGS_verbose)); then
    # Get the longest line for formatting
    max_job_length=$(echo ${secondary_jobs[*]/$job_prefix/} |\
     awk '{for (i=1;i<=NF;++i) {l=length($i);if(l>x) x=l}}END{print x}')
    # Pad it a bit
    ((max_job_length+2))

    logecho
    logecho "(*) Primary job (-) Secondary jobs"
    logecho
    logecho "  $(printf '%-'$max_job_length's' "Job #")" \
            "Run #   Build # Time/Status"
    logecho "= $(common::print_n_char = $max_job_length)" \
            "=====   ======= ==========="
  fi

  while read good_job; do
    ((good_job_count++))

    if ((good_job_count>hard_limit)); then
      logecho
      logecho "Hard Limit of $hard_limit exceeded.  Halting test analysis..."
      logecho
      retcode=1
      break
    fi

    if [[ $good_job =~ \
          JOB\[([0-9]+)\]=(${VER_REGEX[release]})\.${VER_REGEX[build]} ]]; then
      main_run=${BASH_REMATCH[1]}
      build_number=${BASH_REMATCH[8]}
      # Save first build_number
      ((good_job_count==1)) && first_build_number=$build_number
      build_sha1=${BASH_REMATCH[9]}
      build_version=${BASH_REMATCH[2]}.$build_number+$build_sha1
      build_sha1_date=$($GHCURL $K8S_GITHUB_API/commits?sha=$build_sha1 |\
                        jq -r '.[0] | .commit .author .date')
      build_sha1_date=$(date +"%R %m/%d" -d "$build_sha1_date")

      # See anago for --build-at-head
      # This method requires a call to release::get_job_cache() to initially
      # set $build_version, however it is less fragile than ls-remote and
      # sorting
      if ((FLAGS_build_at_head)); then
        build_version=${build_version/$build_sha1/$branch_head}
        # Force job_count so we don't fail
        job_count=1
        logecho
        logecho "Forced --build-at-head specified." \
                "Setting build_version=$build_version"
        logecho
        break
      fi
    elif [[ $good_job =~ JOB\[([0-9]+)\]=(${VER_REGEX[release]}) ]]; then
      logecho
      logecho "$ATTENTION: Newly tagged versions exclude git hash." \
              "Unable to continue.  This is not a valid version." \
              "Add some new post-branch commits to the tree." \
              "(see https://github.com/kubernetes/kubernetes/issues/24535)"
      return 1
    else
      logecho
      logecho "Bad build version!"
      return 1
    fi

    # For release branches, the HEAD revision has to be the one we use because
    # we make code changes on the HEAD of the branch (version.go).
    # Verify if discovered build_version's SHA1 hash == HEAD if branch
    # Due to resetting $branch above based on branch==master to workaround
    # inconsistencies in the testgrid config naming conventions, now look
    # *specifically* for a release branch with a version in it (non-master).
    if [[ "$branch" =~ release-([0-9]{1,})\. ]]; then
      if [[ $build_sha1 != $branch_head ]]; then
        commit_count=$((first_build_number-build_number))
        logecho
        logecho "$ATTENTION: The $branch branch HEAD is ahead of the last" \
                "good run by $commit_count commits."
        logecho
        logecho "If you want to use the head of the branch anyway," \
                "--buildversion=${build_version/$build_sha1/$branch_head}"
        if ((FLAGS_allow_stale_build)); then
          logecho
        else
         return 1
       fi
      fi
    fi

    # Deal with far-behind secondary builds and just skip forward
    ((build_number>giveup_build_number)) && continue

    if ((FLAGS_verbose)); then
      logecho "* $(printf \
                   '%-'$max_job_length's %-7s %-7s' \
                   ${main_job/$job_prefix/} \
                   \#$main_run \#$build_number) [$build_sha1_date]"
      logecho "* (--buildversion=$build_version)"
    fi

    # Check secondaries to ensure that build number is green across "all"
    for other_job in ${secondary_jobs[@]}; do
      ((FLAGS_verbose)) \
       && logecho -n "- $(printf '%-'$max_job_length's ' \
                                 ${other_job/$job_prefix/})"

      # Need to kick out when a secondary doesn't exist (anymore)
      if [[ ! -f $job_path/$other_job ]]; then
        ((FLAGS_verbose)) \
         && logecho -r "Does not exist  SKIPPING"
        ((job_count++)) || true
        continue
      elif [[ $(wc -l <$job_path/$other_job) -lt 1 ]]; then
        ((FLAGS_verbose)) \
         && logecho -r "No Good Runs    ${TPUT[YELLOW]}SKIPPING${TPUT[OFF]}"
        ((job_count++)) || true
        continue
      fi

      unset JOB
      source $job_path/$other_job
      last_cache_build=0
      last_run=0
      # We reverse sort the array here so we can descend it
      for run in $(for n in ${!JOB[*]}; do echo $n; done|sort -rh); do
        if ! [[ ${JOB[$run]} =~ ${VER_REGEX[build]} ]]; then
          run=""
          break
        fi
        cache_build=${BASH_REMATCH[1]}
        # if build_number matches the cache's build number we're good
        # OR
        # if last_run-run proves consecutive jobs AND
        # build_number is within a cache_build range, the build was also good
        if ((build_number==cache_build)) || \
           ((($((last_run-run))==1)) && \
           (((build_number>cache_build)) && \
           ((build_number<last_cache_build)))); then
          break
        fi
        last_cache_build=$cache_build
        last_run=$run
        unset run
      done

      if [[ -n $run ]]; then
        ((FLAGS_verbose)) && \
         logecho "$(printf '%-7s %-7s' \#$run \#$cache_build) $PASSED"
        ((job_count++)) || true
        continue
      else
        ((FLAGS_verbose)) && \
         logecho "$(printf '%-7s %-7s' -- --)" \
                    "${TPUT[RED]}FAILED${TPUT[OFF]}"
        giveup_build_number=$build_number
        job_count=0
        # Note: We used to break here to fail fast.
        # However, it's often useful to see the full
        # list of failing jobs, so now we keep going.
        continue
      fi
    done

    ((FLAGS_verbose)) && logecho
    ((job_count>=${#secondary_jobs[@]})) && break
  done < $job_path/$main_job

  if ((job_count==0)); then
    logecho "Unable to find a green set of test results!"
    return 1
  else
    JENKINS_BUILD_VERSION=$build_version
  fi

  ((FLAGS_verbose)) && logecho JENKINS_BUILD_VERSION=$JENKINS_BUILD_VERSION

  rm -rf $job_path

  return $retcode
}


##############################################################################
# Sets global dictionary RELEASE_VERSION based on passed in build version and
# release branch
# @param version - Jenkins build version
# @param branch - branch to check
# @param parent_branch - the parent of a new branch (if new)
release::set_release_version () {
  local version=$1
  local branch=$2
  local parent_branch=$3
  local label
  declare -A release_branch build_version
  declare -Ag RELEASE_VERSION

  if ! [[ $branch =~ $BRANCH_REGEX ]]; then
    logecho "Invalid branch format! $branch"
    return 1
  fi

  release_branch[major]=${BASH_REMATCH[1]}
  release_branch[minor]=${BASH_REMATCH[2]}

  # if branch == master, version is an alpha
  # if branch == release, version is a beta
  # if branch == release+1, version is an alpha
  if ! [[ $version =~ ${VER_REGEX[release]} ]]; then
    logecho "Invalid version format! $version"
    return 1
  fi

  # Split incoming version up into components
  build_version[major]=${BASH_REMATCH[1]}
  build_version[minor]=${BASH_REMATCH[2]}
  build_version[patch]=${BASH_REMATCH[3]}
  build_version[label]=${BASH_REMATCH[4]}       # -alpha, -beta, -rc
  build_version[labelid]=${BASH_REMATCH[5]}     # rev

  # RELEASE_VERSION_PRIME is the default release version for this session/type
  # Other labels such as alpha, beta, and rc are set as needed
  # Index ordering is important here as it's how they are processed
  if [[ "$parent_branch" == master ]]; then
    # This is a new branch, set new alpha and beta versions
    RELEASE_VERSION[alpha]="v${release_branch[major]}"
    RELEASE_VERSION[alpha]+=".$((${release_branch[minor]}+1)).0-alpha.0"
    RELEASE_VERSION[beta]="v${release_branch[major]}.${release_branch[minor]}"
    RELEASE_VERSION[beta]+=".0-beta.0"
    RELEASE_VERSION_PRIME=${RELEASE_VERSION[beta]}
  elif [[ "$parent_branch" =~ release- ]]; then
    # When we do branched branches we end up with two betas so deal with it
    # by creating a couple of beta indexes.
    # beta0 is the branch-point-minor + 1 + beta.1 because
    # branch-point-minor +1 +beta.0 already exists. This tag lands on the new
    # branch.
    # beta1 is the branch-point-minor + 2 + beta.0 to continue the next version
    # on the parent/source branch.
    RELEASE_VERSION[beta0]="v${build_version[major]}.${build_version[minor]}"
    RELEASE_VERSION[beta0]+=".$((${build_version[patch]}+1))-beta.1"
    RELEASE_VERSION[beta1]="v${build_version[major]}.${build_version[minor]}"
    # Need to increment N+2 here.  N+1-beta.0 exists as an artifact of N.
    RELEASE_VERSION[beta1]+=".$((${build_version[patch]}+2))-beta.0"
    RELEASE_VERSION_PRIME="${RELEASE_VERSION[beta0]}"
  elif [[ $branch =~ release- ]]; then
    # Build out the RELEASE_VERSION dict
    RELEASE_VERSION_PRIME="v${build_version[major]}.${build_version[minor]}"
    # If the incoming version is anything bigger than vX.Y.Z, then it's a
    # Jenkin's build version and it stands as is, otherwise increment the patch
    if [[ -n ${build_version[labelid]} ]]; then
      RELEASE_VERSION_PRIME+=".${build_version[patch]}"
    else
      RELEASE_VERSION_PRIME+=".$((${build_version[patch]}+1))"
    fi

    if ((FLAGS_official)); then
      RELEASE_VERSION[official]="$RELEASE_VERSION_PRIME"
      # Only primary branches get beta releases
      if [[ $branch =~ ^release-([0-9]{1,})\.([0-9]{1,})$ ]]; then
        RELEASE_VERSION[beta]="v${build_version[major]}.${build_version[minor]}"
        RELEASE_VERSION[beta]+=".$((${build_version[patch]}+1))-beta.0"
      fi
    elif ((FLAGS_rc)) || [[ "${build_version[label]}" == "-rc" ]]; then
      # betas not allowed after release candidates
      RELEASE_VERSION[rc]="$RELEASE_VERSION_PRIME"
      if [[ "${build_version[label]}" == "-rc" ]]; then
        RELEASE_VERSION[rc]+="-rc.$((${build_version[labelid]}+1))"
      else
        # Start release candidates at 1 instead of 0
        RELEASE_VERSION[rc]+="-rc.1"
      fi
      RELEASE_VERSION_PRIME="${RELEASE_VERSION[rc]}"
    else
      RELEASE_VERSION[beta]="$RELEASE_VERSION_PRIME${build_version[label]}"
      RELEASE_VERSION[beta]+=".$((${build_version[labelid]}+1))"
      RELEASE_VERSION_PRIME="${RELEASE_VERSION[beta]}"
    fi
  else
    RELEASE_VERSION[alpha]="v${build_version[major]}.${build_version[minor]}"
    RELEASE_VERSION[alpha]+=".${build_version[patch]}${build_version[label]}"
    RELEASE_VERSION[alpha]+=".$((${build_version[labelid]}+1))"
    RELEASE_VERSION_PRIME="${RELEASE_VERSION[alpha]}"
  fi

  if ((FLAGS_verbose)); then
    for label in ${!RELEASE_VERSION[*]}; do
      logecho "RELEASE_VERSION[$label]=${RELEASE_VERSION[$label]}"
    done
    logecho "RELEASE_VERSION_PRIME=$RELEASE_VERSION_PRIME"
  fi

  return 0
}

###############################################################################
# Check that the GCS bucket exists and is writable.
#
# @param bucket - The gs release bucket name
# @return 1 if bucket does not exist or is not writable.
release::gcs::check_release_bucket() {
  local bucket=$1
  local tempfile=$TMPDIR/$PROG-gcs-write.$$

  if ! $GSUTIL ls "gs://$bucket" >/dev/null 2>&1 ; then
    logecho "Google Cloud Storage bucket does not exist: $bucket. Create the bucket with this command:"
    logecho "$GSUTIL mb -p \"$GCLOUD_PROJECT\" \"gs://$bucket\""
    logecho "If the bucket should be publicly readable, make it so with this command:"
    logecho "$GSUTIL defacl ch -u AllUsers:R \"gs://$bucket\""
    logecho "WARNING: This affects all objects uploaded to the bucket!"
    return 1
  fi

  logecho -n "Checking write access to bucket $bucket: "
  if logrun touch $tempfile && \
     logrun $GSUTIL cp $tempfile gs://$bucket && \
     logrun $GSUTIL rm gs://$bucket/${tempfile##*/} && \
     logrun rm -f $tempfile; then
    logecho $OK
  else
    logecho "$FAILED: You do not have access/write permission on $bucket." \
            "Unable to continue."
    return 1
  fi
}

###############################################################################
# Creates a tarball for upload to a GCS staging directory.
# @param gcs_stage - the staging directory
# @param source and destination arguments
# @return 1 if tar directory or tarball creation fails
release::gcs::prepare_tarball() {
  local gcs_stage=$1
  shift
  local src
  local srcdir
  local srcthing
  local args
  local split
  local srcs
  local dst

  # Split the args into srcs... and dst
  local args=("$@")
  local split=$((${#args[@]}-1)) # Split point for src/dst args
  local srcs=("${args[@]::$split}" )
  local dst="${args[$split]}"

  logrun mkdir -p $gcs_stage/$dst || return 1

  for src in ${srcs[@]}; do
    srcdir=$(dirname $src)
    srcthing=$(basename $src)
    tar c -C $srcdir $srcthing | tar x -C $gcs_stage/$dst || return 1
  done
}

###############################################################################
# Ensure the destination bucket path doesn't already exist
# @param gcs_destination - GCS destination directory
# @return 1 on failure or if GCS destination already exists
#           and --allow_dup does not set
release::gcs::destination_empty() {
  local gcs_destination=$1

  logecho -n "Checking whether $gcs_destination already exists: "
  if $GSUTIL ls $gcs_destination >/dev/null 2>&1 ; then
    logecho "- Destination exists. To remove, run:"
    logecho "  gsutil -m rm -r $gcs_destination"

    if ((FLAGS_allow_dup)) ; then
      logecho "flag --allow-dup set, continue with overwriting"
    else
      logecho "$FAILED"
      return 1
    fi
  fi
  logecho "$OK"
}

###############################################################################
# Push the release artifacts to GCS
# @param src - Source path
# @param dest - Destination path
# @return 1 on failure
release::gcs::push_release_artifacts() {
  local src=$1
  local dest=$2

  logecho "Publish public release artifacts..."

  # No need to check this for mock or stage runs
  # Overwriting is ok
  if ((FLAGS_nomock)) && ! ((FLAGS_stage)); then
    release::gcs::destination_empty $dest || return 1
  fi

  # Copy the main set from staging to destination
  # We explicitly don't set an ACL in the cp call, since doing so will override
  # any default bucket ACLs.
  logecho -n "- Copying artifacts to $dest: "
  logrun -s $GSUTIL -qm cp -rc $src/* $dest/ || return 1

  # This small sleep gives the eventually consistent GCS bucket listing a chance
  # to stabilize before the diagnostic listing. There's no way to directly
  # query for consistency, but it's OK if something is dropped from the
  # debugging output.
  sleep 5

  logecho -n "- Listing final contents to log file: "
  logrun -s $GSUTIL ls -lhr "$dest" || return 1
}

###############################################################################
# Locally stage the release artifacts to staging directory
# @param build_type - One of 'release' or 'ci'
# @param version - The version
# @param build_output - build output directory
# @optparam release_kind - defaults to kubernetes
# @return 1 on failure
release::gcs::locally_stage_release_artifacts() {
  local build_type=$1
  local version=$2
  local build_output=$3
  # --release-kind used by push-build.sh
  local release_kind=${4:-"kubernetes"}
  local platform
  local platforms
  local release_stage=$build_output/release-stage
  local release_tars=$build_output/release-tars
  local gcs_stage=$build_output/gcs-stage/$version
  local configure_vm
  local src
  local dst

  logecho "Locally stage release artifacts..."

  logrun rm -rf $gcs_stage || return 1
  logrun mkdir -p $gcs_stage || return 1

  # Stage everything in release directory
  logecho "- Staging locally to ${gcs_stage##$build_output/}..."
  release::gcs::prepare_tarball $gcs_stage $release_tars/* . || return 1

  # Controls whether we publish md5 / sha1; clear to prevent publishing
  # Although we would like to encourage users to stop using md5 / sha1,
  # because all releases are published by the master branch, this broke
  # anyone that was verifying the sha1 hash, including in CI builds.
  # kops caught this in e.g. https://prow.k8s.io/view/gcs/kubernetes-jenkins/pr-logs/pull/kops/7462/pull-kops-e2e-kubernetes-aws/1164899358911500292/
  # Instead we need to stop publishing sha1 and md5 on a release boundary.
  local publish_old_hashes="1"

  if [[ "$release_kind" == "kubernetes" ]]; then
    if ! [[ $version =~ ${VER_REGEX[release]} ]]; then
      logecho -r "$FAILED"
      logecho "* Invalid version format! $version"
      return 1
    fi

    local -r version_major="${BASH_REMATCH[1]}"
    local -r version_minor="${BASH_REMATCH[2]}"
    local -r version_patch="${BASH_REMATCH[3]}"

    # Don't publish md5 & sha1 as of Kubernetes 1.18
    if [[ "$version_minor" -ge "18" ]]; then
      publish_old_hashes=""
    fi

    local gce_path=$release_stage/full/kubernetes/cluster/gce
    local gci_path

    # GCI path changed in 1.2->1.3 time period
    if [[ -d $gce_path/gci ]]; then
      gci_path=$gce_path/gci
    else
      gci_path=$gce_path/trusty
    fi

    # Having the configure-vm.sh script and trusty code from the GCE cluster
    # deploy hosted with the release is useful for GKE.
    # Take it if available (Removed in 1.10+)
    [[ -f $gce_path/configure-vm.sh ]] \
     && configure_vm="$gce_path/configure-vm.sh"

    release::gcs::prepare_tarball $gcs_stage $configure_vm extra/gce \
      || return 1
    release::gcs::prepare_tarball $gcs_stage $gci_path/node.yaml extra/gce \
      || return 1
    release::gcs::prepare_tarball $gcs_stage $gci_path/master.yaml extra/gce \
      || return 1
    release::gcs::prepare_tarball $gcs_stage $gci_path/configure.sh extra/gce \
      || return 1

    # shutdown.sh was introduced starting from v1.11 to make Preemptible COS nodes
    # on GCP not reboot immediately when terminated. Avoid including it in the release
    # bundle if it is not found (for backwards compatibility).
    if [[ -f $gci_path/shutdown.sh ]]; then
      release::gcs::prepare_tarball $gcs_stage $gci_path/shutdown.sh extra/gce \
       || return 1
    fi

    # Having the Windows startup scripts from the GCE cluster deploy hosted with
    # the release is useful for GKE.
    windows_local_path=$gce_path/windows
    windows_gcs_path=extra/gce/windows
    if [[ -d $windows_local_path ]]; then
      release::gcs::prepare_tarball $gcs_stage $windows_local_path/configure.ps1 $windows_gcs_path \
        || return 1
      release::gcs::prepare_tarball $gcs_stage $windows_local_path/common.psm1 $windows_gcs_path \
        || return 1
      release::gcs::prepare_tarball $gcs_stage $windows_local_path/k8s-node-setup.psm1 $windows_gcs_path \
        || return 1
      release::gcs::prepare_tarball $gcs_stage $windows_local_path/testonly/install-ssh.psm1 $windows_gcs_path \
        || return 1
      release::gcs::prepare_tarball $gcs_stage $windows_local_path/testonly/user-profile.psm1 $windows_gcs_path \
        || return 1
    fi

  fi

  # Upload the "naked" binaries to GCS.  This is useful for install scripts that
  # download the binaries directly and don't need tars.
  platforms=($(cd "$release_stage/client"; echo *))
  for platform in "${platforms[@]}"; do
    src="$release_stage/client/$platform/$release_kind/client/bin/*"
    dst="bin/${platform/-//}/"
    # We assume here the "server package" is a superset of the "client package"
    if [[ -d "$release_stage/server/$platform" ]]; then
      src="$release_stage/server/$platform/$release_kind/server/bin/*"
    fi
    release::gcs::prepare_tarball $gcs_stage "$src" "$dst" || return 1

    # Upload node binaries if they exist and this isn't a 'server' platform.
    if [[ ! -d "$release_stage/server/$platform" ]]; then
      if [[ -d "$release_stage/node/$platform" ]]; then
        src="$release_stage/node/$platform/$release_kind/node/bin/*"
        release::gcs::prepare_tarball $gcs_stage "$src" "$dst" || return 1
      fi
    fi
  done

  # Write the release checksum files.
  # We checksum everything except our checksum files, which we do next.
  logecho "- Writing artifact hashes to SHA256SUMS/SHA512SUMS files..."
  find "${gcs_stage}" -type f -printf '%P\0' | xargs -0 -I {} shasum -a 256 "${gcs_stage}/{}" >> SHA256SUMS
  find "${gcs_stage}" -type f -printf '%P\0' | xargs -0 -I {} shasum -a 512 "${gcs_stage}/{}" >> SHA512SUMS
  # After all the checksum files are generated, move them into the bucket staging area
  mv SHA256SUMS SHA512SUMS "$gcs_stage"

  logecho "- Hashing files in ${gcs_stage##$build_output/}..."
  find "$gcs_stage" -type f | while read -r path; do
    if [[ -n "${publish_old_hashes}" ]]; then
      common::md5 "$path" > "$path.md5" || return 1
      common::sha "$path" 1 > "$path.sha1" || return 1
    fi

    common::sha "$path" 256 > "$path.sha256" || return 1
    common::sha "$path" 512 > "$path.sha512" || return 1
  done
}

###############################################################################
# Publish a new version, (latest or stable,) but only if the
# files actually exist on GCS and the artifacts we're dealing with are newer
# than the contents in GCS.
# @param build_type - One of 'release' or 'ci'
# @param version - The version
# @param build_output - build output directory
# @param bucket - GS bucket
# @return 1 on failure
release::gcs::publish_version () {
  local build_type=$1
  local version=$2
  local build_output=$3
  local bucket=$4
  local extra_publish_file=$5
  local release_dir="gs://$bucket/$build_type/$version"
  local version_major
  local version_minor
  local publish_file
  local -a publish_files
  local type="latest"

  # For release/ targets, type could be 'stable'
  if [[ "$build_type" == release ]]; then
    [[ "$version" =~ alpha|beta|rc ]] || type="stable"
  fi

  if ! logrun $GSUTIL ls $release_dir; then
    logecho "Release files don't exist at $release_dir"
    return 1
  fi

  if [[ $version =~ ${VER_REGEX[release]} ]]; then
    version_major=${BASH_REMATCH[1]}
    version_minor=${BASH_REMATCH[2]}
  fi

  publish_files=($type
                 $type-$version_major
                 $type-$version_major.$version_minor
                 $extra_publish_file
                )

  logecho
  logecho "Publish official pointer text files to $bucket..."

  for publish_file in ${publish_files[@]}; do
    # If there's a version that's above the one we're trying to release, don't
    # do anything, and just try the next one.
    release::gcs::verify_latest_update $build_type/$publish_file.txt \
                                    $bucket $version || continue
    release::gcs::publish $build_type/$publish_file.txt $build_output \
                          $bucket $version || return 1
  done
}


###############################################################################
# Check if the new version is greater than the version currently published on
# GCS.
# @param publish_file - the GCS location to look in
# @param bucket - GS bucket
# @param version - release version
# @optparam publish_file_dst - (for testing)
# @return 1 if new version is not greater than the GCS version
release::gcs::verify_latest_update () {
  local -r publish_file=$1
  local -r bucket=$2
  local -r version=$3
  local -r publish_file_dst=${4:-"gs://$bucket/$publish_file"}
  local gcs_version
  local greater=1

  logecho -n "Test $version > $publish_file (published): "
  if ! [[ $version =~ ${VER_REGEX[release]}(.${VER_REGEX[build]})* ]]; then
    logecho -r "$FAILED"
    logecho "* Invalid version format! $version"
    return 1
  fi

  local -r version_major="${BASH_REMATCH[1]}"
  local -r version_minor="${BASH_REMATCH[2]}"
  local -r version_patch="${BASH_REMATCH[3]}"
  local -r version_prerelease="${BASH_REMATCH[4]}"
  local -r version_prerelease_rev="${BASH_REMATCH[5]}"
  local -r version_commits="${BASH_REMATCH[7]}"

  local -a catCmd=( 'cat' "$publish_file_dst" )
  [ -n "${GSUTIL:-}" ] && catCmd=( "$GSUTIL" "${catCmd[@]}" )

  if gcs_version="$( "${catCmd[@]}" )"; then
    if ! [[ $gcs_version =~ ${VER_REGEX[release]}(.${VER_REGEX[build]})* ]]; then
      logecho -r "$FAILED"
      logecho "* file contains invalid release version," \
              "can't compare: '$gcs_version'"
      return 1
    fi

    local -r gcs_version_major="${BASH_REMATCH[1]}"
    local -r gcs_version_minor="${BASH_REMATCH[2]}"
    local -r gcs_version_patch="${BASH_REMATCH[3]}"
    local -r gcs_version_prerelease="${BASH_REMATCH[4]}"
    local -r gcs_version_prerelease_rev="${BASH_REMATCH[5]}"
    local -r gcs_version_commits="${BASH_REMATCH[7]}"

    if [[ "$version_major" -lt "$gcs_version_major" ]]; then
      greater=0
    elif [[ "$version_major" -gt "$gcs_version_major" ]]; then
      : # fall out
    elif [[ "$version_minor" -lt "$gcs_version_minor" ]]; then
      greater=0
    elif [[ "$version_minor" -gt "$gcs_version_minor" ]]; then
      : # fall out
    elif [[ "$version_patch" -lt "$gcs_version_patch" ]]; then
      greater=0
    elif [[ "$version_patch" -gt "$gcs_version_patch" ]]; then
      : # fall out
    # Use lexicographic (instead of integer) comparison because
    # version_prerelease is a string, ("alpha", "beta", or "rc"),
    # but first check if either is an official release
    # (i.e. empty prerelease string).
    #
    # We have to do this because lexicographically "rc" > "beta" > "alpha" > "",
    # but we want official > rc > beta > alpha.
    elif [[ -n "$version_prerelease" && -z "$gcs_version_prerelease" ]]; then
      greater=0
    elif [[ -z "$version_prerelease" && -n "$gcs_version_prerelease" ]]; then
      : # fall out
    elif [[ "$version_prerelease" < "$gcs_version_prerelease" ]]; then
      greater=0
    elif [[ "$version_prerelease" > "$gcs_version_prerelease" ]]; then
      : # fall out
    # The only reason we want to NOT update in the 'equals' case
    # is to maintain publish_file timestamps for the original 'equals' version.
    # for release if the new version is <=, Do Not Publish
    # for ci if the new version is strictly <, Do Not Publish
    elif [[ $publish_file =~ ^release && \
            "$version_prerelease_rev" -le "$gcs_version_prerelease_rev" ]] ||\
         [[ $publish_file =~ ^ci && \
            "$version_prerelease_rev" -lt "$gcs_version_prerelease_rev" ]]; then
      greater=0
    elif [[ "$version_prerelease_rev" -gt "$gcs_version_prerelease_rev" ]]; then
      : # fall out
    # If either version_commits is empty, it will be considered less-than, as
    # expected, (e.g. 1.2.3-beta < 1.2.3-beta.1).
    elif [[ "$version_commits" -lt "$gcs_version_commits" ]]; then
      greater=0
    fi

    if ((greater)); then
      logecho -r "$OK"
      logecho "* $version > $gcs_version (published), updating"
    else
      logecho "$WARNING"
      logecho "* $version <= $gcs_version (published) - not updating."
      return 1
    fi
  else
    # gsutil cat failed; file does not exist
    logecho -r "$OK"
    logecho "* $publish_file_dst does not exist yet. It will be created..."
    return 0
  fi
}


###############################################################################
# Publish a release to GCS: upload a version file, if --nomock,
# make it public, and verify the result.
# @param publish_file - the GCS location to look in
# @param build_output - build output directory
# @param bucket - GS bucket
# @param version - release version
# @optparam bucket_mirror - (optional) mirror GS bucket
# @return 1 on failure
release::gcs::publish () {
  local publish_file=$1
  local build_output=$2
  local bucket=$3
  local version=$4
  local release_stage=$build_output/release-stage
  local publish_file_dst="gs://$bucket/$publish_file"
  local contents
  local public_link="https://storage.googleapis.com/$bucket/$publish_file"

  # TODO(prototype): Uncomment this once dl.k8s.io is cut over to new prod.
  #if [[ "$bucket" == "$PROD_BUCKET" ]]; then
  #  public_link="https://dl.k8s.io/$publish_file"
  #fi

  logrun mkdir -p "$release_stage/upload" || return 1
  echo "$version" > "$release_stage/upload/latest" || return 1

  logrun $GSUTIL -m -h "Content-Type:text/plain" \
    -h "Cache-Control:private, max-age=0, no-transform" cp \
    "$release_stage/upload/latest" \
    "$publish_file_dst" || return 1

  if ((FLAGS_nomock)) && ! ((FLAGS_private_bucket)); then
    # If public, validate public link
    logecho -n "* Validating uploaded version file at $public_link: "
    contents="$(curl --retry 5 -Ls $public_link)"
  else
    # If not public, validate using gsutil
    logecho -n "* Validating uploaded version file at $publish_file_dst: "
    contents="$($GSUTIL cat $publish_file_dst)"
  fi

  if [[ "$contents" == "$version" ]]; then
    logecho "$OK"
  else
    logecho "$FAILED (file contents: $contents)"
    return 1
  fi
}

###############################################################################
# Releases all docker images to a docker registry using the docker tarfiles.
#
# @param registry - docker registry
# @param version - version tag
# @param build_output - build output directory
# @return 1 on failure
release::docker::release () {
  local registry=$1
  local push_registry=$registry
  local version=$2
  local build_output=$3
  local release_images=$build_output/release-images
  local docker_target
  local arch
  local -a arches
  local tarfile
  local orig_tag
  local -a new_tags
  local new_tag
  local binary
  local -A manifest_images

  if [[ "$registry" == "$GCRIO_PATH_PROD" ]]; then
    # Switch to the push alias if using the $GCRIO_PATH_PROD alias
    push_registry="$GCRIO_PATH_PROD_PUSH"
  fi

  logecho "Send docker containers from release-images to $push_registry..."

  arches=($(cd "$release_images"; echo *))
  for arch in ${arches[@]}; do
    for tarfile in $release_images/$arch/*.tar; do
      # There may be multiple tags; just get the first
      orig_tag=$(tar xf $tarfile manifest.json -O  | jq -r '.[0].RepoTags[0]')
      if [[ ! "$orig_tag" =~ ^.+/(.+):.+$ ]]; then
        logecho "$FAILED: malformed tag in $tarfile:"
        logecho $orig_tag
        return 1
      fi
      binary=${BASH_REMATCH[1]}

      new_tag="$push_registry/${binary/-$arch/}"
      new_tag_with_arch=("$new_tag-$arch:$version")
      manifest_images["${new_tag}"]+=" $arch"

      logrun docker load -qi $tarfile
      logrun docker tag $orig_tag ${new_tag_with_arch}
      logecho -n "Pushing ${new_tag_with_arch}: "
      # TODO: Use docker direct when fixed later
      #logrun -r 5 -s docker push "${new_tag_with_arch}" || return 1
      logrun -r 5 -s $GCLOUD docker -- push "${new_tag_with_arch}" || return 1
      if [[ "${PURGE_IMAGES:-yes}" == "yes" ]] ; then
        logrun docker rmi $orig_tag ${new_tag_with_arch} || true
      fi

    done
  done

  # Ensure that the docker command line supports the manifest images
  export DOCKER_CLI_EXPERIMENTAL=enabled

  for image in "${!manifest_images[@]}"; do
    local archs=$(echo "${manifest_images[$image]}" | sed -e 's/^[[:space:]]*//')
    local manifest=$(echo $archs | sed -e "s~[^ ]*~$image\-&:$version~g")
    # This command will push a manifest list: "${registry}/${image}-ARCH:${version}" that points to each architecture depending on which platform you're pulling from
    logecho "Creating manifest image ${image}:${version}..."
    logrun -r 5 -s docker manifest create --amend ${image}:${version} ${manifest} || return 1
    for arch in ${archs}; do
      logecho "Annotating ${image}-${arch}:${version} with --arch ${arch}..."
      logrun -r 5 -s docker manifest annotate --arch ${arch} ${image}:${version} ${image}-${arch}:${version} || return 1
    done
    logecho "Pushing manifest image ${image}:${version}..."
    local purge=""
    if [[ "${PURGE_IMAGES:-yes}" == "yes" ]] ; then
      purge="--purge"
    fi
    logrun -r 5 -s docker manifest push ${image}:${version} ${purge} || return 1
  done

  # Always reset back to $GCP_USER
  # This is set in push-build.sh and anago
  ((FLAGS_gcb)) || logrun $GCLOUD config set account $GCP_USER

  return 0
}


###############################################################################
# Get the kubecross image version for a given release branch.
# @params branches - the release branches we care about
release::kubecross_version() {
  local branches=( "$@" )
  local urlFmt='https://raw.githubusercontent.com/kubernetes/kubernetes/%s/build/build-image/cross/VERSION'
  local ver url

  for branch in "${branches[@]}"
  do
    # shellcheck disable=SC2059
    # ... because we want to store the printf format string in a variable
    url="$( printf "$urlFmt" "$branch" )"

    logecho -n "Trying to get kubecross version for ${branch} ..." >&2

    ver="$( curl -sSL --fail "$url" 2>/dev/null )" && {
      logecho " ${OK}: ${ver}" >&2
      echo "$ver"
      return 0
    }

    logecho " ${FAILED}" >&2
  done

  # If we didn't return yet, we couldn't find the images version and thus need
  # to return an error
  logecho "Unable to find kubecross version in '${branches[*]}'." >&2
  return 1
}

###############################################################################
# Determine whether the release was most recently built with Bazel.
# This is achieved by looking for the most recent kubernetes.tar.gz tarball
# in both the dockerized and Bazel output trees.
# @param kube_root - Root of kubernetes tree
# @param release_kind - Kind of release. kubernetes or federation
# @return 0 if built with Bazel, 1 otherwise
release::was_built_with_bazel() {
  local kube_root=$1
  local release_kind=$2
  local most_recent_release_tar=$( (ls -t \
    $kube_root/{_output,bazel-bin/build}/release-tars/$release_kind.tar.gz \
    2>/dev/null || true) | head -n 1)

  [[ $most_recent_release_tar =~ /bazel-bin/ ]]
}

###############################################################################
# Copy the release artifacts to a local staging dir and push them up to GCS
# using the Bazel "push-build" rule.
# @param build_type - One of 'release' or 'ci'
# @param version - The version
# @param build_output - legacy build output directory; artifacts will be copied
#                       here for use by CI
# @param bucket - GS bucket
# @return 1 on failure
release::gcs::bazel_push_build() {
  local build_type=$1
  local version=$2
  local build_output=$3
  local bucket=$4
  # We don't need to locally stage, but CI uses this.
  # TODO: make it configurable whether we also stage locally
  local gcs_stage=$build_output/gcs-stage/$version
  local gcs_destination=gs://$bucket/$build_type/$version

  logrun rm -rf $gcs_stage || return 1
  logrun mkdir -p $gcs_stage || return 1

  # No need to check this for mock or stage runs
  # Overwriting is ok
  if ((FLAGS_nomock)) && ! ((FLAGS_stage)); then
    release::gcs::destination_empty $gcs_destination || return 1
  fi

  logecho "Publish release artifacts to gs://$bucket using Bazel..."
  logecho "- Hashing and copying public release artifacts to $gcs_destination: "
  bazel run //:push-build $gcs_stage $gcs_destination || return 1

  # This small sleep gives the eventually consistent GCS bucket listing a chance
  # to stabilize before the diagnostic listing. There's no way to directly
  # query for consistency, but it's OK if something is dropped from the
  # debugging output.
  sleep 5

  logecho -n "- Listing final contents to log file: "
  logrun -s $GSUTIL ls -lhr "$gcs_destination" || return 1
}

###############################################################################
# Mail out the announcement
PROGSTEP[release::send_announcement]="SEND ANNOUNCEMENT"
release::send_announcement () {
  local announcement_file="$WORKDIR/announcement.html"
  local subject_file="$WORKDIR/announcement-subject.txt"
  local announcement_text="/tmp/$PROG-rsa.$$"
  local subject
  local mailto="kubernetes-dev@googlegroups.com"
        mailto+=",kubernetes-announce@googlegroups.com"
  # Bucket for the purposes of announcement meta is either in the "GCB" for
  # mock runs or in the standard location defined in set_globals().
  local bucket="$RELEASE_BUCKET_GCB"
  ((FLAGS_nomock)) && bucket="$RELEASE_BUCKET"
  local archive_root="gs://$bucket/archive/anago-$RELEASE_VERSION"

  ((FLAGS_nomock)) || mailto=$GCP_USER
  mailto=${FLAGS_mailto:-$mailto}

  # Announcement file is stored normally in WORKDIR, else check GCS.
  if [[ -f "$announcement_file" ]]; then
    announcement_text="$announcement_file"
  else
    announcement_file="$archive_root/announcement.html"
    if ! $GSUTIL cp $announcement_file $announcement_text >/dev/null 2>&1; then
      logecho "Unable to find an announcement locally or on GCS!"
      return 1
    fi
  fi

  # Subject should be in the same place as announcement_text
  if [[ -f "$subject_file" ]]; then
    subject=$(<"$subject_file")
  else
    subject_file="$archive_root/announcement-subject.txt"
    if ! subject="$($GSUTIL cat $subject_file 2>&-)"; then
      logecho "Unable to find an announcement subject file locally or on GCS!"
      return 1
    fi
  fi

  logecho
  ((FLAGS_yes)) \
   || common::askyorn -e "Pausing here. Confirm announce to $mailto" \
   || common::exit 1 "Exiting..."

  logecho "Announcing \"$subject\" to $mailto..."

  # Always cc invoker
  # Due to announcements landing on public mailing lists requiring membership,
  # post from the invoking user (for now until this is productionized further)
  # and use reply-to to ensure replies go to the right place.
  common::sendmail -h "$mailto" "K8s-Anago<$GCP_USER>" \
                   "K8s-Anago<cloud-kubernetes-release@google.com>" \
                   "$subject" "$GCP_USER" \
                   "$announcement_text" || return 1

  # Cleanup
  logrun rm -f $announcement_text
}

##############################################################################
# Sets major global variables
# Used only in anago and release-notify
# RELEASE_GB - space requirements per build
# GCP_USER - The user that drives the entire release
# RELEASE_BUCKET - mock or standard release bucket location
# BUCKET_TYPE - stage or release
# WRITE_RELEASE_BUCKETS - array of writable buckets
# READ_RELEASE_BUCKETS - array of readable buckets for multiple sourcing of
#                        mock staged builds
# GCRIO_PATH - GCR path based on mock or --nomock
# ALL_CONTAINER_REGISTRIES - when running mock (via GCB) this array also
#                            contains k8s.gcr.io so we can check access in mock
#                            mode before an actual release occurs
release::set_globals () {
  logecho -n "Setting global variables: "

  # Default disk requirements per version - Modified in found_staged_location()
  RELEASE_GB="75"

  if ! ((FLAGS_gcb)); then
    GCP_USER=$($GCLOUD auth list --filter=status:ACTIVE \
                                 --format="value(account)" 2>/dev/null)
    if [[ -z "$GCP_USER" ]]; then
      logecho $FAILED
      logecho "Unable to set a valid GCP credential!"
      return 1
    fi

    # Lowercase GCP user
    GCP_USER="${GCP_USER,,}"
  fi

  if ((FLAGS_stage)); then
    BUCKET_TYPE="stage"
  else
    BUCKET_TYPE="release"
  fi

  # Set GCR values
  # The "production" GCR path is now multi-region alias
  # TODO(prototype): Temporarily setting this to the staging project to test
  #                  the staging flow with --nomock set.
  #GCRIO_PATH_PROD="k8s.gcr.io"
  GCRIO_PATH_PROD="gcr.io/$PROD_PROJECT"
  # TODO(prototype): Temporarily setting this to the staging project to test
  #                  the staging flow with --nomock set.
  # TODO(prototype): Once access has been configured, we should set this to
  #                  k8s-release-test-prod and test image promotion.
  GCRIO_PATH_PROD_PUSH="gcr.io/$PROD_PROJECT"
  # The "test" GCR path
  GCRIO_PATH_TEST="gcr.io/$TEST_PROJECT"

  GCRIO_PATH="${FLAGS_gcrio_path:-$GCRIO_PATH_TEST}"
  ALL_CONTAINER_REGISTRIES=("$GCRIO_PATH")

  if ((FLAGS_nomock)); then
    RELEASE_BUCKET="$PROD_BUCKET"

    GCRIO_PATH="${FLAGS_gcrio_path:-$GCRIO_PATH_PROD}"
  elif ((FLAGS_gcb)); then
    RELEASE_BUCKET="$MOCK_BUCKET"

    # This is passed to logrun() where appropriate when we want to mock
    # specific activities like pushes
    LOGRUN_MOCK="-m"
  else
    # GCS buckets cannot contain @ or "google", so for those users, just use
    # the "$USER" portion of $GCP_USER
    RELEASE_BUCKET_USER="$RELEASE_BUCKET-${GCP_USER%%@google.com}"
    RELEASE_BUCKET_USER="${RELEASE_BUCKET_USER/@/-at-}"

    # GCP also doesn't like anything even remotely looking like a domain name
    # in the bucket name so convert . to -
    RELEASE_BUCKET_USER="${RELEASE_BUCKET_USER/\./-}"
    RELEASE_BUCKET="$RELEASE_BUCKET_USER"

    READ_RELEASE_BUCKETS=("$MOCK_BUCKET")

    # This is passed to logrun() where appropriate when we want to mock
    # specific activities like pushes
    LOGRUN_MOCK="-m"
  fi

  WRITE_RELEASE_BUCKETS=("$RELEASE_BUCKET")
  READ_RELEASE_BUCKETS+=("$RELEASE_BUCKET")

  ALL_CONTAINER_REGISTRIES=("$GCRIO_PATH")

  # TODO:
  # These KUBE_ globals extend beyond the scope of the new release refactored
  # tooling so to pass these through as flags will require fixes across
  # kubernetes/kubernetes and kubernetes/release which we can do at a later time
  export KUBE_DOCKER_REGISTRY="$GCRIO_PATH"
  export KUBE_RELEASE_RUN_TESTS=n
  export KUBE_SKIP_CONFIRMATIONS=y

  logecho $OK
}
