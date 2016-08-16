CNI_RELEASE = 8a936732094c0941e1543ef5d292a1f4fffa1ac5
PACKAGE = k8s-cni
VERSION = 0.0.1-0
DESTINATION = debian/changelog
DISTRO = xenial

export VERSION PACKAGE DESTINATION DISTRO


bin/%:
	curl -sSL --fail --retry 5 https://storage.googleapis.com/kubernetes-release/network-plugins/cni-$(CNI_RELEASE).tar.gz | tar xz

deb: bin/%
	./write_changelog
	dpkg-buildpackage
