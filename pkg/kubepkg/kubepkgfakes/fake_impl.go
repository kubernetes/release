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

// Code generated by counterfeiter. DO NOT EDIT.
package kubepkgfakes

import (
	"os"
	"sync"

	"github.com/google/go-github/v33/github"
	"k8s.io/release/v1/pkg/kubepkg"
	"k8s.io/release/v1/pkg/release"
)

type FakeImpl struct {
	GetKubeVersionStub        func(release.VersionType) (string, error)
	getKubeVersionMutex       sync.RWMutex
	getKubeVersionArgsForCall []struct {
		arg1 release.VersionType
	}
	getKubeVersionReturns struct {
		result1 string
		result2 error
	}
	getKubeVersionReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	ReadFileStub        func(string) ([]byte, error)
	readFileMutex       sync.RWMutex
	readFileArgsForCall []struct {
		arg1 string
	}
	readFileReturns struct {
		result1 []byte
		result2 error
	}
	readFileReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	ReleasesStub        func(string, string, bool) ([]*github.RepositoryRelease, error)
	releasesMutex       sync.RWMutex
	releasesArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 bool
	}
	releasesReturns struct {
		result1 []*github.RepositoryRelease
		result2 error
	}
	releasesReturnsOnCall map[int]struct {
		result1 []*github.RepositoryRelease
		result2 error
	}
	RunSuccessWithWorkDirStub        func(string, string, ...string) error
	runSuccessWithWorkDirMutex       sync.RWMutex
	runSuccessWithWorkDirArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 []string
	}
	runSuccessWithWorkDirReturns struct {
		result1 error
	}
	runSuccessWithWorkDirReturnsOnCall map[int]struct {
		result1 error
	}
	WriteFileStub        func(string, []byte, os.FileMode) error
	writeFileMutex       sync.RWMutex
	writeFileArgsForCall []struct {
		arg1 string
		arg2 []byte
		arg3 os.FileMode
	}
	writeFileReturns struct {
		result1 error
	}
	writeFileReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeImpl) GetKubeVersion(arg1 release.VersionType) (string, error) {
	fake.getKubeVersionMutex.Lock()
	ret, specificReturn := fake.getKubeVersionReturnsOnCall[len(fake.getKubeVersionArgsForCall)]
	fake.getKubeVersionArgsForCall = append(fake.getKubeVersionArgsForCall, struct {
		arg1 release.VersionType
	}{arg1})
	stub := fake.GetKubeVersionStub
	fakeReturns := fake.getKubeVersionReturns
	fake.recordInvocation("GetKubeVersion", []interface{}{arg1})
	fake.getKubeVersionMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeImpl) GetKubeVersionCallCount() int {
	fake.getKubeVersionMutex.RLock()
	defer fake.getKubeVersionMutex.RUnlock()
	return len(fake.getKubeVersionArgsForCall)
}

func (fake *FakeImpl) GetKubeVersionCalls(stub func(release.VersionType) (string, error)) {
	fake.getKubeVersionMutex.Lock()
	defer fake.getKubeVersionMutex.Unlock()
	fake.GetKubeVersionStub = stub
}

func (fake *FakeImpl) GetKubeVersionArgsForCall(i int) release.VersionType {
	fake.getKubeVersionMutex.RLock()
	defer fake.getKubeVersionMutex.RUnlock()
	argsForCall := fake.getKubeVersionArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeImpl) GetKubeVersionReturns(result1 string, result2 error) {
	fake.getKubeVersionMutex.Lock()
	defer fake.getKubeVersionMutex.Unlock()
	fake.GetKubeVersionStub = nil
	fake.getKubeVersionReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeImpl) GetKubeVersionReturnsOnCall(i int, result1 string, result2 error) {
	fake.getKubeVersionMutex.Lock()
	defer fake.getKubeVersionMutex.Unlock()
	fake.GetKubeVersionStub = nil
	if fake.getKubeVersionReturnsOnCall == nil {
		fake.getKubeVersionReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.getKubeVersionReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *FakeImpl) ReadFile(arg1 string) ([]byte, error) {
	fake.readFileMutex.Lock()
	ret, specificReturn := fake.readFileReturnsOnCall[len(fake.readFileArgsForCall)]
	fake.readFileArgsForCall = append(fake.readFileArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.ReadFileStub
	fakeReturns := fake.readFileReturns
	fake.recordInvocation("ReadFile", []interface{}{arg1})
	fake.readFileMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeImpl) ReadFileCallCount() int {
	fake.readFileMutex.RLock()
	defer fake.readFileMutex.RUnlock()
	return len(fake.readFileArgsForCall)
}

func (fake *FakeImpl) ReadFileCalls(stub func(string) ([]byte, error)) {
	fake.readFileMutex.Lock()
	defer fake.readFileMutex.Unlock()
	fake.ReadFileStub = stub
}

func (fake *FakeImpl) ReadFileArgsForCall(i int) string {
	fake.readFileMutex.RLock()
	defer fake.readFileMutex.RUnlock()
	argsForCall := fake.readFileArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeImpl) ReadFileReturns(result1 []byte, result2 error) {
	fake.readFileMutex.Lock()
	defer fake.readFileMutex.Unlock()
	fake.ReadFileStub = nil
	fake.readFileReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *FakeImpl) ReadFileReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.readFileMutex.Lock()
	defer fake.readFileMutex.Unlock()
	fake.ReadFileStub = nil
	if fake.readFileReturnsOnCall == nil {
		fake.readFileReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.readFileReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *FakeImpl) Releases(arg1 string, arg2 string, arg3 bool) ([]*github.RepositoryRelease, error) {
	fake.releasesMutex.Lock()
	ret, specificReturn := fake.releasesReturnsOnCall[len(fake.releasesArgsForCall)]
	fake.releasesArgsForCall = append(fake.releasesArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 bool
	}{arg1, arg2, arg3})
	stub := fake.ReleasesStub
	fakeReturns := fake.releasesReturns
	fake.recordInvocation("Releases", []interface{}{arg1, arg2, arg3})
	fake.releasesMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeImpl) ReleasesCallCount() int {
	fake.releasesMutex.RLock()
	defer fake.releasesMutex.RUnlock()
	return len(fake.releasesArgsForCall)
}

func (fake *FakeImpl) ReleasesCalls(stub func(string, string, bool) ([]*github.RepositoryRelease, error)) {
	fake.releasesMutex.Lock()
	defer fake.releasesMutex.Unlock()
	fake.ReleasesStub = stub
}

func (fake *FakeImpl) ReleasesArgsForCall(i int) (string, string, bool) {
	fake.releasesMutex.RLock()
	defer fake.releasesMutex.RUnlock()
	argsForCall := fake.releasesArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeImpl) ReleasesReturns(result1 []*github.RepositoryRelease, result2 error) {
	fake.releasesMutex.Lock()
	defer fake.releasesMutex.Unlock()
	fake.ReleasesStub = nil
	fake.releasesReturns = struct {
		result1 []*github.RepositoryRelease
		result2 error
	}{result1, result2}
}

func (fake *FakeImpl) ReleasesReturnsOnCall(i int, result1 []*github.RepositoryRelease, result2 error) {
	fake.releasesMutex.Lock()
	defer fake.releasesMutex.Unlock()
	fake.ReleasesStub = nil
	if fake.releasesReturnsOnCall == nil {
		fake.releasesReturnsOnCall = make(map[int]struct {
			result1 []*github.RepositoryRelease
			result2 error
		})
	}
	fake.releasesReturnsOnCall[i] = struct {
		result1 []*github.RepositoryRelease
		result2 error
	}{result1, result2}
}

func (fake *FakeImpl) RunSuccessWithWorkDir(arg1 string, arg2 string, arg3 ...string) error {
	fake.runSuccessWithWorkDirMutex.Lock()
	ret, specificReturn := fake.runSuccessWithWorkDirReturnsOnCall[len(fake.runSuccessWithWorkDirArgsForCall)]
	fake.runSuccessWithWorkDirArgsForCall = append(fake.runSuccessWithWorkDirArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 []string
	}{arg1, arg2, arg3})
	stub := fake.RunSuccessWithWorkDirStub
	fakeReturns := fake.runSuccessWithWorkDirReturns
	fake.recordInvocation("RunSuccessWithWorkDir", []interface{}{arg1, arg2, arg3})
	fake.runSuccessWithWorkDirMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3...)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeImpl) RunSuccessWithWorkDirCallCount() int {
	fake.runSuccessWithWorkDirMutex.RLock()
	defer fake.runSuccessWithWorkDirMutex.RUnlock()
	return len(fake.runSuccessWithWorkDirArgsForCall)
}

func (fake *FakeImpl) RunSuccessWithWorkDirCalls(stub func(string, string, ...string) error) {
	fake.runSuccessWithWorkDirMutex.Lock()
	defer fake.runSuccessWithWorkDirMutex.Unlock()
	fake.RunSuccessWithWorkDirStub = stub
}

func (fake *FakeImpl) RunSuccessWithWorkDirArgsForCall(i int) (string, string, []string) {
	fake.runSuccessWithWorkDirMutex.RLock()
	defer fake.runSuccessWithWorkDirMutex.RUnlock()
	argsForCall := fake.runSuccessWithWorkDirArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeImpl) RunSuccessWithWorkDirReturns(result1 error) {
	fake.runSuccessWithWorkDirMutex.Lock()
	defer fake.runSuccessWithWorkDirMutex.Unlock()
	fake.RunSuccessWithWorkDirStub = nil
	fake.runSuccessWithWorkDirReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeImpl) RunSuccessWithWorkDirReturnsOnCall(i int, result1 error) {
	fake.runSuccessWithWorkDirMutex.Lock()
	defer fake.runSuccessWithWorkDirMutex.Unlock()
	fake.RunSuccessWithWorkDirStub = nil
	if fake.runSuccessWithWorkDirReturnsOnCall == nil {
		fake.runSuccessWithWorkDirReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.runSuccessWithWorkDirReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeImpl) WriteFile(arg1 string, arg2 []byte, arg3 os.FileMode) error {
	var arg2Copy []byte
	if arg2 != nil {
		arg2Copy = make([]byte, len(arg2))
		copy(arg2Copy, arg2)
	}
	fake.writeFileMutex.Lock()
	ret, specificReturn := fake.writeFileReturnsOnCall[len(fake.writeFileArgsForCall)]
	fake.writeFileArgsForCall = append(fake.writeFileArgsForCall, struct {
		arg1 string
		arg2 []byte
		arg3 os.FileMode
	}{arg1, arg2Copy, arg3})
	stub := fake.WriteFileStub
	fakeReturns := fake.writeFileReturns
	fake.recordInvocation("WriteFile", []interface{}{arg1, arg2Copy, arg3})
	fake.writeFileMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeImpl) WriteFileCallCount() int {
	fake.writeFileMutex.RLock()
	defer fake.writeFileMutex.RUnlock()
	return len(fake.writeFileArgsForCall)
}

func (fake *FakeImpl) WriteFileCalls(stub func(string, []byte, os.FileMode) error) {
	fake.writeFileMutex.Lock()
	defer fake.writeFileMutex.Unlock()
	fake.WriteFileStub = stub
}

func (fake *FakeImpl) WriteFileArgsForCall(i int) (string, []byte, os.FileMode) {
	fake.writeFileMutex.RLock()
	defer fake.writeFileMutex.RUnlock()
	argsForCall := fake.writeFileArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeImpl) WriteFileReturns(result1 error) {
	fake.writeFileMutex.Lock()
	defer fake.writeFileMutex.Unlock()
	fake.WriteFileStub = nil
	fake.writeFileReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeImpl) WriteFileReturnsOnCall(i int, result1 error) {
	fake.writeFileMutex.Lock()
	defer fake.writeFileMutex.Unlock()
	fake.WriteFileStub = nil
	if fake.writeFileReturnsOnCall == nil {
		fake.writeFileReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.writeFileReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeImpl) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getKubeVersionMutex.RLock()
	defer fake.getKubeVersionMutex.RUnlock()
	fake.readFileMutex.RLock()
	defer fake.readFileMutex.RUnlock()
	fake.releasesMutex.RLock()
	defer fake.releasesMutex.RUnlock()
	fake.runSuccessWithWorkDirMutex.RLock()
	defer fake.runSuccessWithWorkDirMutex.RUnlock()
	fake.writeFileMutex.RLock()
	defer fake.writeFileMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeImpl) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ kubepkg.Impl = new(FakeImpl)
