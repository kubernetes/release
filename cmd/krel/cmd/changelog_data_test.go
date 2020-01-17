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

package cmd

const patchReleaseExpectedTOC = `<!-- BEGIN MUNGE: GENERATED_TOC -->

- [v1.16.3](#v1163)
  - [Downloads for v1.16.3](#downloads-for-v1163)
    - [Client Binaries](#client-binaries)
    - [Server Binaries](#server-binaries)
    - [Node Binaries](#node-binaries)
  - [Changelog since v1.16.2](#changelog-since-v1162)
  - [Changes by Kind](#changes-by-kind)
    - [Failing Test](#failing-test)
    - [Other (Bug, Cleanup or Flake)](#other-bug-cleanup-or-flake)`

const patchReleaseExpectedContent = `<!-- END MUNGE: GENERATED_TOC -->

# v1.16.3

[Documentation](https://docs.k8s.io)

## Downloads for v1.16.3

filename | sha512 hash
-------- | -----------
[kubernetes.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes.tar.gz) | ` + "`e467bea81ea10461af9600e54a0cc4fcdc8dd3557a23f387882aad6c17a10c4a34cf22ab831c4c4ba88891f7c37343212704182e6dbf63dfa7c3ccb89613fad0`" + `
[kubernetes-src.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-src.tar.gz) | ` + "`a92479445c2daebf82b0398abc656629c3627e879eaf05e2b05b92e9c0451149b97564147fc5172503ea27f2bb0d714be09bcf1b3bbf34b0c06795312513dba6`" + `

### Client Binaries

filename | sha512 hash
-------- | -----------
[kubernetes-client-darwin-386.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-client-darwin-386.tar.gz) | ` + "`8c09bfe56d6a0cb7fdfc4673fd7f8e957864b614cec30c6fa93bd2eec4067e2e97d7dfec269793639443b6130fa6a62fd61410825b948ccd71382495c19a57b9`" + `
[kubernetes-client-darwin-amd64.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-client-darwin-amd64.tar.gz) | ` + "`9891da997f09e0692158dad4a01f46f3d14430152f5f9fd0dc225816e81c6be73e56fecc7ce74a10b166482724e12e963a929254ffee2d6145f4e99817fd2a5a`" + `
[kubernetes-client-linux-386.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-client-linux-386.tar.gz) | ` + "`7347d4f90ea00d42428896f71e83c7c7c2fd079c2afd471d06e0772a6b18304bd271a7c907fd8188e0ba2aac4609d7a7f87d6ebdbd4ef7110c1aff2fe391de51`" + `
[kubernetes-client-linux-amd64.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-client-linux-amd64.tar.gz) | ` + "`e6d50e009dbe5d3c549731880a1fa55cdc7f57a2b50e0f6d65afc0fe5d06736387de4208535b0b1eab13d9fa331840e804266d12a542a3437b0768282ac10a46`" + `
[kubernetes-client-linux-arm.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-client-linux-arm.tar.gz) | ` + "`7c286e088d296ea5351545dd91189979157e57d508b6f44ee1c4d18018224cc20a6a2a88bb47b26497a53e9f623d2360faae7fdfca5f7c3e3cfd71e63eef1cbc`" + `
[kubernetes-client-linux-arm64.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-client-linux-arm64.tar.gz) | ` + "`ef018e1285ed5e984f6c0c84963454f5c849ed99584cba7965cbd9e7ee44259210b79f2bd8abe9d00324d2414f749cf1e111b754930ba30eee26577607511d37`" + `
[kubernetes-client-linux-ppc64le.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-client-linux-ppc64le.tar.gz) | ` + "`1d834d32cd2c3b328695b9fe81c912d624ba3a5a9ae927becff906dfc407917ef0055f20a9eb53d26fd4b831ee027c2cedb31fb9155e70d8fdcf04ed2596acde`" + `
[kubernetes-client-linux-s390x.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-client-linux-s390x.tar.gz) | ` + "`88a95b311635729b71e12e5e5de5a849b5630d1f292040d3b913df8fbdf4f63b50822341a66109fe6700bd3c9376a4538c589855127a77390c85691c057738d8`" + `
[kubernetes-client-windows-386.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-client-windows-386.tar.gz) | ` + "`433d2d7908d42d2438041ced20ac43d454055cff2eea4ec434a8f8e169c2510b6b68faa94a2c7d1b7223415a7e780681df60a6d98d11c13f7d51bcce647f7cf6`" + `
[kubernetes-client-windows-amd64.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-client-windows-amd64.tar.gz) | ` + "`2d56da8aa2d14ed8585aa2375a325d5fcc216173e69c5e8fd54795ebfa154eb13494c2754dff6d856ef9b86c6be95b0f4767f04b9d68916d70a292acacac81b0`" + `

### Server Binaries

filename | sha512 hash
-------- | -----------
[kubernetes-server-linux-amd64.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-server-linux-amd64.tar.gz) | ` + "`360fd871f93101263e7727f295e7513fab5951d7fd5da209a586606c9c763429e867ae70ae5c8c427064cb084eb3adc69c465ef3193897acf1ec80a235bbaaa6`" + `
[kubernetes-server-linux-arm.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-server-linux-arm.tar.gz) | ` + "`6f84dc231f0b7ee7e24c989ab03353f910ff3c71d544b75d780cb3e56deff8f4fe77d06c5ee3ba98cef3619a71a1cc1af37b81b239c5a776f34e4a3e99a3269d`" + `
[kubernetes-server-linux-arm64.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-server-linux-arm64.tar.gz) | ` + "`1619c4e277b2cba182c08355ae733ee7694e0e57ec5c6661b650bcf94d7d30223aa3f84163c19f373de58306ec355e9ff766088094ddbb03708c585f35389177`" + `
[kubernetes-server-linux-ppc64le.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-server-linux-ppc64le.tar.gz) | ` + "`ca050289318b1f7db49a81009075f468363da9c4fc8561ef3aaa89ba6cbd8fb3f0e78045a57c01ef76b42193eaba675220daa7a5badcd6b45195ff8047e1a63d`" + `
[kubernetes-server-linux-s390x.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-server-linux-s390x.tar.gz) | ` + "`544b4144f728a233866b8b71edd2e0b90c1da6b10790a4ef3578e9ea4cc5dd48bc86aaf0de212614b980f052919e7bfc8d708dec785b14e0b731c298b45bcc9c`" + `

### Node Binaries

filename | sha512 hash
-------- | -----------
[kubernetes-node-linux-amd64.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-node-linux-amd64.tar.gz) | ` + "`b51c3e58c1b004b50ef40f8351c2e6c6822a2f4af100c7872da0d9c862fed58e0a5e8d4a33eaaa070782fdc09a8b75f62c77aeb1e80049c20d7eaf3d0c4d73f6`" + `
[kubernetes-node-linux-arm.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-node-linux-arm.tar.gz) | ` + "`594eb789d58a7903a432b122d810025cf643a2a2c378abdca4a98e9196699b9543c9f7ca782fd02ccb90363ff17d505b3b37fc181bda4ba5587e3d25f75a9114`" + `
[kubernetes-node-linux-arm64.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-node-linux-arm64.tar.gz) | ` + "`fe5e49b76960d8ad774371a93200c195cde1bfe286d07aab1935abff211833a052084b4920ca28ca2edeee1a13dc4cecdd7040b028ecfb63041a7b3579a16394`" + `
[kubernetes-node-linux-ppc64le.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-node-linux-ppc64le.tar.gz) | ` + "`cc63cfc5c46117619b0418b7b5eb37ffd4c628bd38a816071f0e15d73c65e874ef8428ce6df948f00dec28b8a99ba75cdc3997baedd1ceb66f395a007c4e9a5d`" + `
[kubernetes-node-linux-s390x.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-node-linux-s390x.tar.gz) | ` + "`8b7a11fe2a1bd5759e1e89d965568f0880e17dacbcd5d9ce686478695d70f7896f1d0b3f7ab763b27af8f5da9d874138fdca2921ad0083a41ad771aa24607e62`" + `
[kubernetes-node-windows-amd64.tar.gz](https://dl.k8s.io/v1.16.3/kubernetes-node-windows-amd64.tar.gz) | ` + "`578c1c22fd517f089caafe6981f73e451fc06318e4fb91e545f56a1ef0d3766c4acfc5e728879a41378a76e97df02b9e067b45f4179be3f9600a9bc1d81f5bba`" + `

## Changelog since v1.16.2

## Changes by Kind

### Failing Test`
