
kubelet:
	curl -L --fail "https://storage.googleapis.com/kubernetes-release/release/v1.3.0/bin/linux/amd64/kubelet" -o kubelet
	chmod +x kubelet

deb: kubelet
	dpkg-buildpackage #-S
.PHONY += deb
