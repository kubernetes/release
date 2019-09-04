workspace(name = "io_kubernetes_build")

# The native http_archive rule is deprecated. This is a drop-in replacement.
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

################################################################################
# Go Build Definitions
################################################################################

# Download, load, and initialize the Go build rules.
http_archive(
    name = "io_bazel_rules_go",
    strip_prefix = "rules_go-0.16.6",
    urls = ["https://github.com/bazelbuild/rules_go/archive/0.16.6.zip"],
    sha256 = "c0f7e581b17d0b8252e1d0175cb3f6398a579bb91d2e5995f29f6c5985ccd647",
)

load("@io_bazel_rules_go//go:def.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains()

# Download, load, and initialize the Gazelle tool for generating BUILD files for
# Go code.
http_archive(
    name = "bazel_gazelle",
    strip_prefix = "bazel-gazelle-0.16.0",
    urls = ["https://github.com/bazelbuild/bazel-gazelle/archive/0.16.0.zip"],
    sha256 = "a5b329e3d929247279005ba3cfda0c092a220085c0ed0505de1dcdd68dfc53bc",
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

################################################################################
# Go Dependencies
#
# Update this with:
#   bazel run //:gazelle -- update-repos -from_file=Gopkg.lock
################################################################################

go_repository(
    name = "com_github_blang_semver",
    commit = "2ee87856327ba09384cabd113bc6b5d174e9ec0f",
    importpath = "github.com/blang/semver",
)

go_repository(
    name = "com_github_davecgh_go_spew",
    commit = "8991bc29aa16c548c550c7ff78260e27b9ab7c73",
    importpath = "github.com/davecgh/go-spew",
)

go_repository(
    name = "com_github_go_kit_kit",
    commit = "12210fb6ace19e0496167bb3e667dcd91fa9f69b",
    importpath = "github.com/go-kit/kit",
)

go_repository(
    name = "com_github_go_logfmt_logfmt",
    commit = "07c9b44f60d7ffdfb7d8efe1ad539965737836dc",
    importpath = "github.com/go-logfmt/logfmt",
)

go_repository(
    name = "com_github_golang_protobuf",
    commit = "130e6b02ab059e7b717a096f397c5b60111cae74",
    importpath = "github.com/golang/protobuf",
)

go_repository(
    name = "com_github_google_go_github",
    commit = "2406bfd7f32dea4608923b39352fce69e647e1e1",
    importpath = "github.com/google/go-github",
)

go_repository(
    name = "com_github_google_go_querystring",
    commit = "53e6ce116135b80d037921a7fdd5138cf32d7a8a",
    importpath = "github.com/google/go-querystring",
)

go_repository(
    name = "com_github_kolide_kit",
    commit = "c155a91098e3c16721433130c82c3525abe4a450",
    importpath = "github.com/kolide/kit",
)

go_repository(
    name = "com_github_kr_logfmt",
    commit = "b84e30acd515aadc4b783ad4ff83aff3299bdfe0",
    importpath = "github.com/kr/logfmt",
)

go_repository(
    name = "com_github_pkg_errors",
    commit = "ba968bfe8b2f7e042a574c888954fccecfa385b4",
    importpath = "github.com/pkg/errors",
)

go_repository(
    name = "com_github_pmezard_go_difflib",
    commit = "792786c7400a136282c1664665ae0a8db921c6c2",
    importpath = "github.com/pmezard/go-difflib",
)

go_repository(
    name = "com_github_stretchr_testify",
    commit = "ffdc059bfe9ce6a4e144ba849dbedead332c6053",
    importpath = "github.com/stretchr/testify",
)

go_repository(
    name = "gopkg_in_src_d_go_git_v4",
    commit = "f9a30199e7083bdda8adad3a4fa2ec42d25c1fdb",
    importpath = "gopkg.in/src-d/go-git.v4",
)

go_repository(
    name = "org_golang_google_appengine",
    commit = "150dc57a1b433e64154302bdc40b6bb8aefa313a",
    importpath = "google.golang.org/appengine",
)

go_repository(
    name = "org_golang_x_net",
    commit = "a04bdaca5b32abe1c069418fb7088ae607de5bd0",
    importpath = "golang.org/x/net",
)

go_repository(
    name = "org_golang_x_oauth2",
    commit = "bb50c06baba3d0c76f9d125c0719093e315b5b44",
    importpath = "golang.org/x/oauth2",
)
