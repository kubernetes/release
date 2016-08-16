CNI_RELEASE = 8a936732094c0941e1543ef5d292a1f4fffa1ac5

bin/%:
	curl -sSL --fail --retry 5 https://storage.googleapis.com/kubernetes-release/network-plugins/cni-$(CNI_RELEASE).tar.gz | tar xz

deb: bin/%
	dpkg-buildpackage
