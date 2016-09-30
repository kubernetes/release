#!/bin/bash
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
  local tempjson=/tmp/$PROG-rgjc.$$
  local logroot="gs://kubernetes-jenkins/logs"
  local version
  local lastversion
  local buildnumber

  logecho "Getting $job build results from Jenkins..."
  mkdir -p ${job_path%/*}

  $GSUTIL -qm cp $logroot/$job/jobResultsCache.json $tempjson 2>&-
  # If there's no file up on $logroot, job doesn't exist.  Skip it.
  [[ -s $tempjson ]] || return

  # Additional select on .version is b/c we have so many empty versions for now
  while read version buildnumber; do
    ((dedup)) && [[ $version == $lastversion ]] && continue
    echo "JOB[$buildnumber]=$version"
    lastversion=$version
  done < <(jq -r '.[] | select(.result == "SUCCESS") | select(.version != "") | [.version,.buildnumber] | "\(.[0]) \(.[1])"' $tempjson |sort -r -k2,2) \
   > $job_path

  rm -f $tempjson
}

##############################################################################
# Sets the JENKINS_BUILD_VERSION global by cross checking it against a set of
# critical jenkins build jobs
# @param branch - branch name
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
  local build_version
  local build_number
  local cache_build
  local last_cache_build
  local build_sha1
  local run
  local n
  local giveup_build_number=999999
  local job_count=0
  local max_job_length
  local other_job
  local good_job
  local branch_head
  local branch_suffix
  [[ $branch =~ release- ]] && branch_suffix="-$branch"
  local main_job="kubernetes-e2e-gce$branch_suffix"
  local job_path=/tmp/buildresults-cache.$$
  local -a JOB

  # Would be nice to pull/generate these jobs dynamically filtered through
  # a pattern like gce,gke so new/changed testing jobs don't require
  # updating here, but it turns out way too much hardcoding would need
  # to be done to make this 'dynamic', so just list these explicitly.
  #all_jobs=$($JCURL -s "http://kubekins.dls.corp.google.com/view/Critical%20Builds/api/json" |jq -r '.jobs[] | .name' | egrep -v '-soak-|-build')
  local -a gce_jobs=("kubernetes-e2e-gce-serial$branch_suffix"
                     "kubernetes-e2e-gce-slow$branch_suffix"
                     "kubernetes-kubemark-5-gce$branch_suffix"
                     "kubernetes-e2e-gce-reboot$branch_suffix"
                     "kubernetes-e2e-gce-scalability$branch_suffix"
                     "kubernetes-test-go$branch_suffix"
                    )

  # kubernetes-e2e-gke-subnet - Uses a branch version?
  # kubernetes-e2e-gke-test - Uses a branch version?
  local -a gke_jobs=(
                     "kubernetes-e2e-gke-serial$branch_suffix"
                     "kubernetes-e2e-gke$branch_suffix"
                     "kubernetes-e2e-gke-slow$branch_suffix"
                    )

  # Combined list for cross-checking against $main_job
  local -a secondary_jobs=(${gce_jobs[@]} ${gke_jobs[@]})

  # Update main cache
  # We dedup the $main_job's list of successful runs and just run through that
  # unique list. We then leave the full state of secondaries below so we have
  # finer granularity at the Jenkin's job level to determine if a build is ok.
  release::get_job_cache -d $job_path/$main_job

  # Update secondary caches limited by main cache last build number
  for other_job in ${secondary_jobs[@]}; do
    release::get_job_cache $job_path/$other_job
  done

  if ((FLAGS_verbose)); then
    # Get the longest line for formatting
    max_job_length=$(echo ${secondary_jobs[*]} |\
     awk '{for (i=1;i<=NF;++i) {l=length($i);if(l>x) x=l}}END{print x}')
    # Pad it a bit
    ((max_job_length+2))

    logecho
    logecho "(*) Primary job (-) Secondary jobs"
    logecho
    logecho "  $(printf '%-'$max_job_length's' "Jenkins Job")" \
            "Run #   Build # Time/Status"
    logecho "= $(common::print_n_char = $max_job_length)" \
            "======  ======= ==========="
  fi

  while read good_job; do
    if [[ $good_job =~ \
          JOB\[([0-9]+)\]=(${VER_REGEX[release]})\.${VER_REGEX[build]} ]]; then
      main_run=${BASH_REMATCH[1]}
      build_number=${BASH_REMATCH[8]}
      build_sha1=${BASH_REMATCH[9]}
      build_version=${BASH_REMATCH[2]}.$build_number+$build_sha1
      build_sha1_date=$($GHCURL $K8S_GITHUB_API/commits?sha=$build_sha1 |\
                        jq -r '.[0] | .commit .author .date')
      build_sha1_date=$(date +"%R %m/%d" -d "$build_sha1_date")
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

    # For release branch, the HEAD revision has to be the one we use because
    # we make code changes on the HEAD of the branch (version.go).
    # Verify if discovered build_version's SHA1 hash == HEAD if branch
    if [[ "$branch" =~ release- ]]; then
      branch_head=$($GHCURL $K8S_GITHUB_API/commits/$branch |jq -r '.sha')

      if [[ $build_sha1 != ${branch_head:0:14} ]]; then
        # TODO: Figure out how to curl a list of last N commits
        #       So we can return a message about how far ahead the top of the
        #       release branch is from the last good commit.
        #commit_count=$(git rev-list $build_sha1..${branch_head:0:14} |wc -l)
        commit_count=some
        logecho
        logecho "$ATTENTION: The $branch branch HEAD is ahead of the last" \
                "good Jenkins run by $commit_count commits." \
                "Wait for Jenkins to catch up."
        return 1
      fi
    fi

    # Deal with far-behind secondary builds and just skip forward
    ((build_number>giveup_build_number)) && continue

    if ((FLAGS_verbose)); then
      logecho "* $(printf \
                   '%-'$max_job_length's %-7s %-7s' \
                   $main_job \#$main_run \#$build_number) [$build_sha1_date]"
      logecho "* (--buildversion=$build_version)"
    fi

    # Check secondaries to ensure that build number is green across "all"
    for other_job in ${secondary_jobs[@]}; do
      ((FLAGS_verbose)) \
       && logecho -n "- $(printf '%-'$max_job_length's ' $other_job)"

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
        # if last_run-run proves consecutive Jenkins jobs AND
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
                    "${TPUT[RED]}GIVE UP${TPUT[OFF]}"
        giveup_build_number=$build_number
        job_count=0
        break
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

  return 0
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
  build_version[label]=${BASH_REMATCH[4]}
  build_version[labelid]=${BASH_REMATCH[5]}

  # RELEASE_VERSION_PRIME is the default release version for this session/type
  # Other labels such as alpha and beta are set as needed
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
    RELEASE_VERSION[beta0]="v${build_version[major]}.${build_version[minor]}"
    RELEASE_VERSION[beta0]+=".${build_version[patch]}-beta.1"
    RELEASE_VERSION[beta1]="v${build_version[major]}.${build_version[minor]}"
    RELEASE_VERSION[beta1]+=".$((${build_version[patch]}+1))-beta.0"
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
# Create a unique bucket name for releasing Kube and make sure it exists.
# @param bucket - The gs release bucket name
# @return 1 if bucket can't be made
release::gcs::ensure_release_bucket() {
  local bucket=$1

  if ! $GSUTIL ls "gs://$bucket" >/dev/null 2>&1 ; then
    logecho -n "Creating Google Cloud Storage bucket $bucket: "
    logrun -s $GSUTIL mb -p "$GCLOUD_PROJECT" "gs://$bucket" || return 1
  fi
}

###############################################################################
# Create a unique bucket name for releasing Kube and make sure it exists.
# TODO: There is a version of this in kubernetes/build/common.sh. Refactor.
# @param gcs_stage - the staging directory
# @param source and destination arguments
# @return 1 if tar fails
release::gcs::stage_and_hash() {
  local gcs_stage=$1
  shift

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
# Copy the release artifacts to staging and push them up to GS
# @param build_type - One of 'release' or 'ci'
# @param version - The version
# @param build_output - build output directory
# @param bucket - GS bucket
# @optparam bucket_mirror - (optional) mirror GS bucket
# @return 1 on failure
release::gcs::copy_release_artifacts() {
  local build_type=$1
  local version=$2
  local build_output=$3
  local bucket=$4
  local bucket_mirror=$5
  local platform
  local platforms
  local release_stage=$build_output/release-stage
  local release_tars=$build_output/release-tars
  local gcs_stage=$build_output/gcs-stage
  local src
  local dst
  local gcs_destination=gs://$bucket/$build_type/$version
  local gcs_mirror=gs://$bucket_mirror/$build_type/$version
  local gce_path=$release_stage/full/kubernetes/cluster/gce
  local gci_path

  # GCI path changed in 1.2->1.3 time period
  if [[ -d $gce_path/gci ]]; then
    gci_path=$gce_path/gci
  else
    gci_path=$gce_path/trusty
  fi

  logrun rm -rf $gcs_stage || return 1
  logrun mkdir -p $gcs_stage || return 1

  logecho "Publish release artifacts to gs://$bucket..."

  # Stage everything in release directory
  logecho "- Staging locally to ${gcs_stage##*/}..."
  release::gcs::stage_and_hash $gcs_stage $release_tars/* . || return 1

  # Having the configure-vm.sh script and and trusty code from the GCE cluster
  # deploy hosted with the release is useful for GKE.
  release::gcs::stage_and_hash $gcs_stage $gce_path/configure-vm.sh extra/gce \
   || return 1
  release::gcs::stage_and_hash $gcs_stage $gci_path/node.yaml extra/gce \
   || return 1
  release::gcs::stage_and_hash $gcs_stage $gci_path/master.yaml extra/gce \
   || return 1
  release::gcs::stage_and_hash $gcs_stage $gci_path/configure.sh extra/gce \
   || return 1

  # Upload the "naked" binaries to GCS.  This is useful for install scripts that
  # download the binaries directly and don't need tars.
  platforms=($(cd "$release_stage/client"; echo *))
  for platform in "${platforms[@]}"; do
    src="$release_stage/client/$platform/kubernetes/client/bin/*"
    dst="bin/${platform/-//}/"
    # We assume here the "server package" is a superset of the "client package"
    if [[ -d "$release_stage/server/$platform" ]]; then
      src="$release_stage/server/$platform/kubernetes/server/bin/*"
    fi
    release::gcs::stage_and_hash $gcs_stage "$src" "$dst" || return 1
  done

  logecho "- Hashing files in ${gcs_stage##*/}..."
  find $gcs_stage -type f | while read path; do
    common::md5 $path > "$path.md5" || return 1
    common::sha $path 1 > "$path.sha1" || return 1
  done

  # Copy the main set from staging to destination
  logecho -n "- Copying release artifacts to $gcs_destination: "
  logrun -s $GSUTIL -qm cp -r $gcs_stage/* $gcs_destination/ || return 1

  logecho -n "- Marking all uploaded objects public: "
  logrun -s $LOGRUN_MOCK $GSUTIL -q -m acl ch -R -g all:R \
                                 "$gcs_destination" || return 1

  logecho -n "- Listing final contents to log file: "
  logrun -s $GSUTIL ls -lhr "$gcs_destination" || return 1

  # Push to mirror if set
  if [[ -n "$bucket_mirror" && "$bucket" != "$bucket_mirror" ]]; then
    logecho -n "- Mirroring build to $gcs_mirror: "
    logrun -s $GSUTIL -q -m rsync -d -r "$gcs_destination" "$gcs_mirror" \
     || return 1
    logecho -n "- Marking all uploaded mirror objects public: "
    logrun -s $LOGRUN_MOCK $GSUTIL -q -m acl ch -R -g all:R \
                                   "$gcs_mirror" || return 1
  fi
}


###############################################################################
# Publish a new version, (latest or stable,) but only if the
# files actually exist on GCS and the artifacts we're dealing with are newer
# than the contents in GCS.
# @param build_type - One of 'release' or 'ci'
# @param version - The version
# @param build_output - build output directory
# @param bucket - GS bucket
# @optparam bucket_mirror - GS mirror bucket
# @return 1 on failure
release::gcs::publish_version () {
  local build_type=$1
  local version=$2
  local build_output=$3
  local bucket=$4
  local bucket_mirror=$5
  local release_dir="gs://$bucket/$build_type/$version"
  local version_major
  local version_minor
  local publish_file
  local -a publish_files
  local -a publish_buckets=($bucket)
  local type="latest"

  # Ensure no duplicate
  [[ "$bucket" != "$bucket_mirror" ]] && publish_buckets+=($bucket_mirror)

  # For release/ targets, type could be 'stable'
  if [[ "$build_type" == release ]]; then
    [[ "$version" =~ alpha|beta ]] || type="stable"
  fi

  if ! $GSUTIL ls $release_dir >/dev/null 2>&1 ; then
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
                )

  for b in ${publish_buckets[*]}; do
    logecho
    logecho "Publish official pointer text files to $b..."

    for publish_file in ${publish_files[@]}; do
      # If there's a version that's above the one we're trying to release, don't
      # do anything, and just try the next one.
      release::gcs::verify_latest_update $build_type/$publish_file.txt \
                                      $b $version || continue
      release::gcs::publish $build_type/$publish_file.txt $build_output \
                            $b $version || return 1
    done
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

  if gcs_version="$($GSUTIL cat $publish_file_dst 2>/dev/null)"; then
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
    # version_prerelease is a string, ("alpha" or "beta",) but first check if
    # either is an official release (i.e. empty prerelease string).
    #
    # We have to do this because lexicographically "beta" > "alpha" > "", but
    # we want official > beta > alpha.
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

  logrun mkdir -p "$release_stage/upload" || return 1
  echo "$version" > "$release_stage/upload/latest" || return 1

  logrun $GSUTIL -m -h "Content-Type:text/plain" cp \
    "$release_stage/upload/latest" \
    "$publish_file_dst" || return 1

  if ((FLAGS_nomock)); then
    logecho -n "Making uploaded version file public and non-cacheable: "
    logrun -s $GSUTIL acl ch -R -g all:R $publish_file_dst || return 1
    $GSUTIL setmeta -h "Cache-Control:private, max-age=0" \
     "$publish_file_dst" >/dev/null 2>&1 || return 1

    # If public, validate public link
    logecho -n "* Validating uploaded version file at $public_link: "
    contents="$(curl -s $public_link)"
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
# Releases all docker images to a docker registry.
#
# @param registry - docker registry
# @param version - version tag
# @return 1 on failure
release::docker::release () {
  local registry=$1
  local version=$2
  local docker_push_cmd=(docker)
  local docker_target
  local legacy_docker_target
  local arch
  local binary
  local binaries=(
    "kube-apiserver"
    "kube-controller-manager"
    "kube-scheduler"
    "kube-proxy"
    "hyperkube"
  )

  [[ "$registry" =~ gcr.io/ ]] && docker_push_cmd=("$GCLOUD" "docker")

  # Activate credentials for the k8s.production.user@gmail.com
  [[ "$registry" == "gcr.io/google_containers" ]] \
   && logrun $GCLOUD config set account k8s.production.user@gmail.com
 
  logecho
  logecho "Send docker containers to $registry..."

  # 'gcloud docker' gives lots of internal_failure's so add retries to 
  # all of the invocations
  for arch in "${KUBE_SERVER_PLATFORMS[@]##*/}"; do
    for binary in "${binaries[@]}"; do
      docker_target="$binary-$arch:$version"
      if ! logrun -r 5 ${docker_push_cmd[@]} \
                       history "$registry/$docker_target"; then
        logecho "$WARNING - Skipping non-existent $docker_target..."
        continue
      fi

      logecho "Release $docker_target:"
      logecho -n "- Pushing: "
      logrun -r 5 -s ${docker_push_cmd[@]} push "$registry/$docker_target"

      # If we have a amd64 docker image. Tag it without -amd64 also
      # and push it for compatibility with earlier versions
      if [[ $arch == "amd64" ]]; then
        legacy_docker_target="$binary:$version"
        logecho "Release legacy $legacy_docker_target:"

        logecho -n "- Tagging: "
        logrun -r 5 -s docker tag -f "$registry/$docker_target" \
                              "$registry/$legacy_docker_target" 2>/dev/null

        logecho -n "- Pushing: "
        logrun -r 5 -s ${docker_push_cmd[@]} \
                       push "$registry/$legacy_docker_target"
      fi
    done
  done

  # Activate default account
  if [[ "$registry" == "gcr.io/google_containers" ]]; then
    logrun $GCLOUD config set account $USER@google.com
  fi
}
