%global CNI_VERSION 0.7.5
%global RPM_RELEASE 0
%global ARCH amd64

Name: kubernetes-cni
Version: %{CNI_VERSION}
Release: %{RPM_RELEASE}
License: ASL 2.0
Summary: Kubernetes container networking plugins
URL: https://kubernetes.io

Requires: kubelet

Source: https://dl.k8s.io/network-plugins/cni-plugins-%{ARCH}-v%{CNI_VERSION}.tgz

%description -n kubernetes-cni
Binaries required to provision container networking.

%prep
%setup -c

%install
install -m 755 -d %{buildroot}/opt/cni/bin
install -p -m 755 -t %{buildroot}/opt/cni/bin/ ./*

%files
/opt/cni

%changelog
* Thu Jun 13 2019 Tim Pepper <tpepper@vmware.com> - 0.8.1
- Create kubernetes-cni as its own spec file.
