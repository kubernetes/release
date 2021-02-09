/*
Copyright 2021 The Kubernetes Authors.

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
package licensefakes

import (
	"sync"

	"k8s.io/release/v1/pkg/license"
)

type FakeReaderImplementation struct {
	ClassifyFileStub        func(string) (string, []string, error)
	classifyFileMutex       sync.RWMutex
	classifyFileArgsForCall []struct {
		arg1 string
	}
	classifyFileReturns struct {
		result1 string
		result2 []string
		result3 error
	}
	classifyFileReturnsOnCall map[int]struct {
		result1 string
		result2 []string
		result3 error
	}
	ClassifyLicenseFilesStub        func([]string) ([]license.ClassifyResult, []string, error)
	classifyLicenseFilesMutex       sync.RWMutex
	classifyLicenseFilesArgsForCall []struct {
		arg1 []string
	}
	classifyLicenseFilesReturns struct {
		result1 []license.ClassifyResult
		result2 []string
		result3 error
	}
	classifyLicenseFilesReturnsOnCall map[int]struct {
		result1 []license.ClassifyResult
		result2 []string
		result3 error
	}
	FindLicenseFilesStub        func(string) ([]string, error)
	findLicenseFilesMutex       sync.RWMutex
	findLicenseFilesArgsForCall []struct {
		arg1 string
	}
	findLicenseFilesReturns struct {
		result1 []string
		result2 error
	}
	findLicenseFilesReturnsOnCall map[int]struct {
		result1 []string
		result2 error
	}
	InitializeStub        func(*license.ReaderOptions) error
	initializeMutex       sync.RWMutex
	initializeArgsForCall []struct {
		arg1 *license.ReaderOptions
	}
	initializeReturns struct {
		result1 error
	}
	initializeReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeReaderImplementation) ClassifyFile(arg1 string) (string, []string, error) {
	fake.classifyFileMutex.Lock()
	ret, specificReturn := fake.classifyFileReturnsOnCall[len(fake.classifyFileArgsForCall)]
	fake.classifyFileArgsForCall = append(fake.classifyFileArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.ClassifyFileStub
	fakeReturns := fake.classifyFileReturns
	fake.recordInvocation("ClassifyFile", []interface{}{arg1})
	fake.classifyFileMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeReaderImplementation) ClassifyFileCallCount() int {
	fake.classifyFileMutex.RLock()
	defer fake.classifyFileMutex.RUnlock()
	return len(fake.classifyFileArgsForCall)
}

func (fake *FakeReaderImplementation) ClassifyFileCalls(stub func(string) (string, []string, error)) {
	fake.classifyFileMutex.Lock()
	defer fake.classifyFileMutex.Unlock()
	fake.ClassifyFileStub = stub
}

func (fake *FakeReaderImplementation) ClassifyFileArgsForCall(i int) string {
	fake.classifyFileMutex.RLock()
	defer fake.classifyFileMutex.RUnlock()
	argsForCall := fake.classifyFileArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeReaderImplementation) ClassifyFileReturns(result1 string, result2 []string, result3 error) {
	fake.classifyFileMutex.Lock()
	defer fake.classifyFileMutex.Unlock()
	fake.ClassifyFileStub = nil
	fake.classifyFileReturns = struct {
		result1 string
		result2 []string
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeReaderImplementation) ClassifyFileReturnsOnCall(i int, result1 string, result2 []string, result3 error) {
	fake.classifyFileMutex.Lock()
	defer fake.classifyFileMutex.Unlock()
	fake.ClassifyFileStub = nil
	if fake.classifyFileReturnsOnCall == nil {
		fake.classifyFileReturnsOnCall = make(map[int]struct {
			result1 string
			result2 []string
			result3 error
		})
	}
	fake.classifyFileReturnsOnCall[i] = struct {
		result1 string
		result2 []string
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeReaderImplementation) ClassifyLicenseFiles(arg1 []string) ([]license.ClassifyResult, []string, error) {
	var arg1Copy []string
	if arg1 != nil {
		arg1Copy = make([]string, len(arg1))
		copy(arg1Copy, arg1)
	}
	fake.classifyLicenseFilesMutex.Lock()
	ret, specificReturn := fake.classifyLicenseFilesReturnsOnCall[len(fake.classifyLicenseFilesArgsForCall)]
	fake.classifyLicenseFilesArgsForCall = append(fake.classifyLicenseFilesArgsForCall, struct {
		arg1 []string
	}{arg1Copy})
	stub := fake.ClassifyLicenseFilesStub
	fakeReturns := fake.classifyLicenseFilesReturns
	fake.recordInvocation("ClassifyLicenseFiles", []interface{}{arg1Copy})
	fake.classifyLicenseFilesMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *FakeReaderImplementation) ClassifyLicenseFilesCallCount() int {
	fake.classifyLicenseFilesMutex.RLock()
	defer fake.classifyLicenseFilesMutex.RUnlock()
	return len(fake.classifyLicenseFilesArgsForCall)
}

func (fake *FakeReaderImplementation) ClassifyLicenseFilesCalls(stub func([]string) ([]license.ClassifyResult, []string, error)) {
	fake.classifyLicenseFilesMutex.Lock()
	defer fake.classifyLicenseFilesMutex.Unlock()
	fake.ClassifyLicenseFilesStub = stub
}

func (fake *FakeReaderImplementation) ClassifyLicenseFilesArgsForCall(i int) []string {
	fake.classifyLicenseFilesMutex.RLock()
	defer fake.classifyLicenseFilesMutex.RUnlock()
	argsForCall := fake.classifyLicenseFilesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeReaderImplementation) ClassifyLicenseFilesReturns(result1 []license.ClassifyResult, result2 []string, result3 error) {
	fake.classifyLicenseFilesMutex.Lock()
	defer fake.classifyLicenseFilesMutex.Unlock()
	fake.ClassifyLicenseFilesStub = nil
	fake.classifyLicenseFilesReturns = struct {
		result1 []license.ClassifyResult
		result2 []string
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeReaderImplementation) ClassifyLicenseFilesReturnsOnCall(i int, result1 []license.ClassifyResult, result2 []string, result3 error) {
	fake.classifyLicenseFilesMutex.Lock()
	defer fake.classifyLicenseFilesMutex.Unlock()
	fake.ClassifyLicenseFilesStub = nil
	if fake.classifyLicenseFilesReturnsOnCall == nil {
		fake.classifyLicenseFilesReturnsOnCall = make(map[int]struct {
			result1 []license.ClassifyResult
			result2 []string
			result3 error
		})
	}
	fake.classifyLicenseFilesReturnsOnCall[i] = struct {
		result1 []license.ClassifyResult
		result2 []string
		result3 error
	}{result1, result2, result3}
}

func (fake *FakeReaderImplementation) FindLicenseFiles(arg1 string) ([]string, error) {
	fake.findLicenseFilesMutex.Lock()
	ret, specificReturn := fake.findLicenseFilesReturnsOnCall[len(fake.findLicenseFilesArgsForCall)]
	fake.findLicenseFilesArgsForCall = append(fake.findLicenseFilesArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.FindLicenseFilesStub
	fakeReturns := fake.findLicenseFilesReturns
	fake.recordInvocation("FindLicenseFiles", []interface{}{arg1})
	fake.findLicenseFilesMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeReaderImplementation) FindLicenseFilesCallCount() int {
	fake.findLicenseFilesMutex.RLock()
	defer fake.findLicenseFilesMutex.RUnlock()
	return len(fake.findLicenseFilesArgsForCall)
}

func (fake *FakeReaderImplementation) FindLicenseFilesCalls(stub func(string) ([]string, error)) {
	fake.findLicenseFilesMutex.Lock()
	defer fake.findLicenseFilesMutex.Unlock()
	fake.FindLicenseFilesStub = stub
}

func (fake *FakeReaderImplementation) FindLicenseFilesArgsForCall(i int) string {
	fake.findLicenseFilesMutex.RLock()
	defer fake.findLicenseFilesMutex.RUnlock()
	argsForCall := fake.findLicenseFilesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeReaderImplementation) FindLicenseFilesReturns(result1 []string, result2 error) {
	fake.findLicenseFilesMutex.Lock()
	defer fake.findLicenseFilesMutex.Unlock()
	fake.FindLicenseFilesStub = nil
	fake.findLicenseFilesReturns = struct {
		result1 []string
		result2 error
	}{result1, result2}
}

func (fake *FakeReaderImplementation) FindLicenseFilesReturnsOnCall(i int, result1 []string, result2 error) {
	fake.findLicenseFilesMutex.Lock()
	defer fake.findLicenseFilesMutex.Unlock()
	fake.FindLicenseFilesStub = nil
	if fake.findLicenseFilesReturnsOnCall == nil {
		fake.findLicenseFilesReturnsOnCall = make(map[int]struct {
			result1 []string
			result2 error
		})
	}
	fake.findLicenseFilesReturnsOnCall[i] = struct {
		result1 []string
		result2 error
	}{result1, result2}
}

func (fake *FakeReaderImplementation) Initialize(arg1 *license.ReaderOptions) error {
	fake.initializeMutex.Lock()
	ret, specificReturn := fake.initializeReturnsOnCall[len(fake.initializeArgsForCall)]
	fake.initializeArgsForCall = append(fake.initializeArgsForCall, struct {
		arg1 *license.ReaderOptions
	}{arg1})
	stub := fake.InitializeStub
	fakeReturns := fake.initializeReturns
	fake.recordInvocation("Initialize", []interface{}{arg1})
	fake.initializeMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeReaderImplementation) InitializeCallCount() int {
	fake.initializeMutex.RLock()
	defer fake.initializeMutex.RUnlock()
	return len(fake.initializeArgsForCall)
}

func (fake *FakeReaderImplementation) InitializeCalls(stub func(*license.ReaderOptions) error) {
	fake.initializeMutex.Lock()
	defer fake.initializeMutex.Unlock()
	fake.InitializeStub = stub
}

func (fake *FakeReaderImplementation) InitializeArgsForCall(i int) *license.ReaderOptions {
	fake.initializeMutex.RLock()
	defer fake.initializeMutex.RUnlock()
	argsForCall := fake.initializeArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeReaderImplementation) InitializeReturns(result1 error) {
	fake.initializeMutex.Lock()
	defer fake.initializeMutex.Unlock()
	fake.InitializeStub = nil
	fake.initializeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeReaderImplementation) InitializeReturnsOnCall(i int, result1 error) {
	fake.initializeMutex.Lock()
	defer fake.initializeMutex.Unlock()
	fake.InitializeStub = nil
	if fake.initializeReturnsOnCall == nil {
		fake.initializeReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.initializeReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeReaderImplementation) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.classifyFileMutex.RLock()
	defer fake.classifyFileMutex.RUnlock()
	fake.classifyLicenseFilesMutex.RLock()
	defer fake.classifyLicenseFilesMutex.RUnlock()
	fake.findLicenseFilesMutex.RLock()
	defer fake.findLicenseFilesMutex.RUnlock()
	fake.initializeMutex.RLock()
	defer fake.initializeMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeReaderImplementation) recordInvocation(key string, args []interface{}) {
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

var _ license.ReaderImplementation = new(FakeReaderImplementation)
