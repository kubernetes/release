K8S_VERSION ?= 1.3.4
PACKAGE = kubelet
DESTINATION = debian/changelog
DISTRO = xenial

export K8S_VERSION PACKAGE DESTINATION DISTRO

deb: kubelet
	./write_changelog
	dpkg-buildpackage
.PHONY += deb

kubelet:
	curl -L --fail "https://storage.googleapis.com/kubernetes-release/release/v$(K8S_VERSION)/bin/linux/amd64/kubelet" -o kubelet
	chmod +x kubelet
