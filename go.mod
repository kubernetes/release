module k8s.io/release

go 1.13

require (
	cloud.google.com/go v0.44.3
	github.com/GoogleCloudPlatform/testgrid v0.0.10
	github.com/bazelbuild/rules_go v0.23.1
	github.com/blang/semver v3.5.1+incompatible
	github.com/go-git/go-git/v5 v5.0.0
	github.com/golang/protobuf v1.3.2
	github.com/golangci/golangci-lint v1.25.0
	github.com/google/go-github/v29 v29.0.3
	github.com/google/uuid v1.1.1
	github.com/maxbrunsfeld/counterfeiter/v6 v6.2.3
	github.com/nozzle/throttler v0.0.0-20180817012639-2ea982251481
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/psampaz/go-mod-outdated v0.6.0
	github.com/saschagrunert/go-modiff v1.2.0
	github.com/sendgrid/rest v2.4.1+incompatible
	github.com/sendgrid/sendgrid-go v3.5.0+incompatible
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.5.1
	github.com/yuin/goldmark v1.1.30
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/api v0.21.0
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/golangci/golangci-lint => github.com/golangci/golangci-lint v1.23.3
