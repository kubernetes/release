module k8s.io/release

go 1.15

require (
	cloud.google.com/go v0.66.0 // indirect
	cloud.google.com/go/storage v1.11.0
	github.com/GoogleCloudPlatform/testgrid v0.0.22
	github.com/bazelbuild/rules_go v0.24.2
	github.com/blang/semver v3.5.1+incompatible
	github.com/containers/image v3.0.2+incompatible
	github.com/containers/image/v5 v5.6.0
	github.com/go-git/go-git/v5 v5.1.0
	github.com/golang/protobuf v1.4.2
	github.com/google/go-github/v29 v29.0.3
	github.com/google/uuid v1.1.2
	github.com/maxbrunsfeld/counterfeiter/v6 v6.2.3
	github.com/nozzle/throttler v0.0.0-20180817012639-2ea982251481
	github.com/olekukonko/tablewriter v0.0.4
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2-0.20190823105129-775207bd45b6
	github.com/pkg/errors v0.9.1
	github.com/psampaz/go-mod-outdated v0.6.0
	github.com/saschagrunert/go-modiff v1.2.0
	github.com/sendgrid/rest v2.6.1+incompatible
	github.com/sendgrid/sendgrid-go v3.6.3+incompatible
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/yuin/goldmark v1.2.1
	golang.org/x/net v0.0.0-20200904194848-62affa334b73
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	golang.org/x/sys v0.0.0-20200917073148-efd3b9a0ff20 // indirect
	golang.org/x/tools v0.0.0-20200917221617-d56e4e40bc9d // indirect
	google.golang.org/api v0.32.0
	google.golang.org/genproto v0.0.0-20200917134801-bb4cff56e0d0 // indirect
	google.golang.org/grpc v1.32.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/apimachinery v0.19.2
	k8s.io/kubectl v0.19.2
	k8s.io/utils v0.0.0-20200912215256-4140de9c8800
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/bazelbuild/rules_go => github.com/bazelbuild/rules_go v0.23.10 // pinned to match kubernetes/kubernetes@1.19.2
