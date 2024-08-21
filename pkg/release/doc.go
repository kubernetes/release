/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package release

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt releasefakes/fake_version_client.go > releasefakes/_fake_version_client.go && mv releasefakes/_fake_version_client.go releasefakes/fake_version_client.go"
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt releasefakes/fake_repository.go > releasefakes/_fake_repository.go && mv releasefakes/_fake_repository.go releasefakes/fake_repository.go"
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt releasefakes/fake_publisher_client.go > releasefakes/_fake_publisher_client.go && mv releasefakes/_fake_publisher_client.go releasefakes/fake_publisher_client.go"
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt releasefakes/fake_prerequisites_checker_impl.go > releasefakes/_fake_prerequisites_checker_impl.go && mv releasefakes/_fake_prerequisites_checker_impl.go releasefakes/fake_prerequisites_checker_impl.go"
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt releasefakes/fake_image_impl.go > releasefakes/_fake_image_impl.go && mv releasefakes/_fake_image_impl.go releasefakes/fake_image_impl.go"
//go:generate /usr/bin/env bash -c "cat ../../hack/boilerplate/boilerplate.generatego.txt releasefakes/fake_branch_checker_impl.go > releasefakes/_fake_branch_checker_impl.go && mv releasefakes/_fake_branch_checker_impl.go releasefakes/fake_branch_checker_impl.go"
