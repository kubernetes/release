# gazelle:repository_macro repos.bzl%go_repositories
workspace(name = "io_k8s_release")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

################################################################################
# Go Build Definitions
################################################################################

git_repository(
    name = "io_k8s_repo_infra",
    commit = "db6ceb5f992254db76af7c25db2edc5469b5ea82",
    remote = "https://github.com/kubernetes/repo-infra.git",
    shallow_since = "1570128715 -0700",
)

load("@io_k8s_repo_infra//:load.bzl", _repo_infra_repos = "repositories")

_repo_infra_repos()

load("@io_k8s_repo_infra//:repos.bzl", "configure")

configure(go_modules = None)

load("//:repos.bzl", "go_repositories")

go_repositories()

load("@io_k8s_repo_infra//:repos.bzl", _repo_infra_go_repos = "go_repositories")

_repo_infra_go_repos()

http_file(
    name = "jq",
    downloaded_file_path = "jq",
    executable = True,
    sha256 = "af986793a515d500ab2d35f8d2aecd656e764504b789b66d7e1a0b727a124c44",
    urls = ["https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64"],
)
