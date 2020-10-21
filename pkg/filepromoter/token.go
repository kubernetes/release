/*
Copyright 2019 The Kubernetes Authors.

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

package filepromoter

import (
	"sync"

	"golang.org/x/oauth2"

	// TODO: Use k/release/pkg/log instead
	"k8s.io/klog/v2"
	"sigs.k8s.io/k8s-container-image-promoter/pkg/gcloud"
)

// gcloudTokenSource implements oauth2.TokenSource.
type gcloudTokenSource struct {
	mutex          sync.Mutex
	ServiceAccount string
}

// Token implements TokenSource.Token.
func (s *gcloudTokenSource) Token() (*oauth2.Token, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	klog.Infof("getting service-account-token for %q", s.ServiceAccount)

	token, err := gcloud.GetServiceAccountToken(s.ServiceAccount, true)
	if err != nil {
		klog.Warningf("failed to get service-account-token for %q: %v",
			s.ServiceAccount, err)
		return nil, err
	}
	return &oauth2.Token{
		AccessToken: string(token),
	}, nil
}
