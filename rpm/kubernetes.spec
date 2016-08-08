%global KUBE_VERSION 1.3.4
%global CNI_RELEASE 8a936732094c0941e1543ef5d292a1f4fffa1ac5

Name: kubernetes
Version: %{KUBE_VERSION}
Release: 1
Summary: Container cluster management
License: ASL 2.0

URL: https://kubernetes.io
Source0: https://storage.googleapis.com/kubernetes-release/release/v%{KUBE_VERSION}/bin/linux/amd64/kubelet
Source1: kubelet.service

BuildRequires: curl
Requires: docker-engine >= 1.10
Requires: iptables >= 1.4.21
Requires: socat
Requires: util-linux
Requires: ethtool

%description
The node agent of Kubernetes, the container cluster manager.

%package plugin-cni

Summary: Binaries required to provision kubernetes container networking
Requires: kubernetes

%description plugin-cni
Binaries required to provision container networking.


%prep
# Assumes the builder has overridden sourcedir to point to directory
# with this spec file. (where these files are stored) Copy them into
# the builddir so they can be installed.
#
# Example: rpmbuild --define "_sourcedir $PWD" -bb kubernetes.spec
#
cp -p %{_sourcedir}/kubelet.service %{_builddir}/
cp -p %{_sourcedir}/kubelet-wrapper %{_builddir}/
cp -p %{_sourcedir}/99_bridge.conf %{_builddir}/

#cp -p %{_sourcedir}/kubelet %{_builddir}/


%install

curl -L --fail "https://storage.googleapis.com/kubernetes-release/release/v%{KUBE_VERSION}/bin/linux/amd64/kubelet" -o kubelet

install -m 755 -d %{buildroot}%{_bindir}
install -m 755 -d %{buildroot}%{_sysconfdir}/systemd/system/
install -m 755 -d %{buildroot}%{_sysconfdir}/cni/net.d/
install -m 755 -d %{buildroot}%{_sysconfdir}/kubernetes/manifests/
install -m 755 -d %{buildroot}/var/lib/kubelet/
install -p -m 755 -t %{buildroot}%{_bindir}/ kubelet
install -p -m 755 -t %{buildroot}%{_sysconfdir}/systemd/system/ kubelet.service
install -p -m 755 -t %{buildroot}%{_sysconfdir}/cni/net.d/ 99_bridge.conf
install -p -m 755 -t %{buildroot}/var/lib/kubelet/ kubelet-wrapper

install -m 755 -d %{buildroot}/opt/cni
curl -sSL --fail --retry 5 https://storage.googleapis.com/kubernetes-release/network-plugins/cni-%{CNI_RELEASE}.tar.gz | tar xz
mv bin/ %{buildroot}/opt/cni/


%files
%{_bindir}/kubelet
%{_sysconfdir}/systemd/system/kubelet.service
/var/lib/kubelet/kubelet-wrapper

%files plugin-cni
%{_sysconfdir}/cni/net.d/99_bridge.conf
/opt/cni

%doc


%changelog

* Wed Jul 20 2016 dgoodwin <dgoodwin@redhat.com> - 1.3.0-1
- Initial packaging.
