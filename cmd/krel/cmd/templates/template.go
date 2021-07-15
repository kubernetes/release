/*
Copyright 2020 The Kubernetes Authors.

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

package templates

var MissingPlaceholder = `Hi @{{.Name}}

The docs placeholder deadline is almost here. Please make sure to [create a placeholder PR](https://kubernetes.io/docs/contribute/new-content/new-features/#open-a-placeholder-pr) against the dev-{{.Release}} branch in the k/website before the deadline.
`

var WrongBaseBranch = `Hello @{{.Name}}, {{.Release}} Docs shadow here.

Please make the pull request against the dev-v{{.Release}} branch.`

var UpcomingDeadline = `@{{.Name}}
The Docs deadline is coming up on the 23rd of November. 
It would be great if you make the requested changes plus any additional changes so it can get merged before then. Thank you!`
