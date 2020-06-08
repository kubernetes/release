# gazelle:repository_macro repos.bzl%go_repositories
workspace(name = "io_k8s_release")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive", "http_file")

################################################################################
# Go Build Definitions
################################################################################

http_archive(
    name = "io_k8s_repo_infra",
    sha256 = "b1e51445409a02b1b9a5a0bff31560c7096c07656d1f1546489f32879f4bf7ac",
    strip_prefix = "repo-infra-2e8f2d47a547eac83d5c17a4cd417f178ebede82",
    urls = [
        "https://github.com/kubernetes/repo-infra/archive/2e8f2d4.tar.gz",
    ],
)

load("@io_k8s_repo_infra//:load.bzl", repo_infra_repositories = "repositories")

repo_infra_repositories()

load(
    "@io_k8s_repo_infra//:repos.bzl",
    repo_infra_configure = "configure",
    repo_infra_go_repositories = "go_repositories",
)

repo_infra_configure(
    go_version = "1.13.9",
    minimum_bazel_version = "2.2.0",
)

repo_infra_go_repositories()

load("//:repos.bzl", "go_repositories")

go_repositories()

http_file(
    name = "jq",
    downloaded_file_path = "jq",
    executable = True,
    sha256 = "af986793a515d500ab2d35f8d2aecd656e764504b789b66d7e1a0b727a124c44",
    urls = ["https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64"],
)
