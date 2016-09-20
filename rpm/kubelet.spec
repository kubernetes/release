%global KUBE_VERSION 1.4.0
%global CNI_RELEASE 07a8a28637e97b22eb8dfe710eeae1344f69d16e
%global RPM_RELEASE 0

Name: kubelet
Version: %{KUBE_VERSION}
Release: %{RPM_RELEASE}.beta.11
Summary: Container cluster management
License: ASL 2.0

URL: https://kubernetes.io
Source0: https://storage.googleapis.com/kubernetes-release/release/v%{KUBE_VERSION}/bin/linux/amd64/kubelet
Source1: kubelet.service

BuildRequires: curl
Requires: iptables >= 1.4.21
Requires: socat
Requires: util-linux
Requires: ethtool

%description
The node agent of Kubernetes, the container cluster manager.

%package -n kubernetes-cni

Version: 0.3.0.1
Release: %{RPM_RELEASE}.07a8a2
Summary: Binaries required to provision kubernetes container networking
Requires: kubelet

%description -n kubernetes-cni
Binaries required to provision container networking.

%package -n kubectl

Summary: Command-line utility for interacting with a Kubernetes cluster.

%description -n kubectl
Command-line utility for interacting with a Kubernetes cluster.

%package -n kubeadm

Version: 1.5.0
Release: %{RPM_RELEASE}.alpha.0.1403.gc19e08e
Summary: Command-line utility for administering a Kubernetes cluster. (ALPHA)
Requires: kubelet >= 1.4.0
Requires: kubectl >= 1.4.0
Requires: kubernetes-cni

%description -n kubeadm
Command-line utility for administering a Kubernetes cluster.

%prep
# Assumes the builder has overridden sourcedir to point to directory
# with this spec file. (where these files are stored) Copy them into
# the builddir so they can be installed.
#
# Example: rpmbuild --define "_sourcedir $PWD" -bb kubelet.spec
#
cp -p %{_sourcedir}/kubelet.service %{_builddir}/
cp -p %{_sourcedir}/10-kubeadm.conf %{_builddir}/

# NOTE: Uncomment if you have these binaries in the directory you're building from.
# This is a useful temporary hack for faster Docker builds when working on the spec.
# Implies you also comment out the curl commands below.
#cp -p %{_sourcedir}/kubelet %{_builddir}/
#cp -p %{_sourcedir}/kubectl %{_builddir}/
#cp -p %{_sourcedir}/kubeadm %{_builddir}/


%install

# For now we have to add -beta.8 to the version here, we can't include that in a spec file Version:
curl -L --fail "https://storage.googleapis.com/kubernetes-release/release/v%{KUBE_VERSION}-beta.11/bin/linux/amd64/kubelet" -o kubelet
curl -L --fail "https://storage.googleapis.com/kubernetes-release/release/v%{KUBE_VERSION}-beta.11/bin/linux/amd64/kubectl" -o kubectl
curl -L --fail "https://storage.googleapis.com/kubeadm/v1.5.0-alpha.0-1403-gc19e08e/bin/kubeadm" -o kubeadm

install -m 755 -d %{buildroot}%{_bindir}
install -m 755 -d %{buildroot}%{_sysconfdir}/systemd/system/
install -m 755 -d %{buildroot}%{_sysconfdir}/systemd/system/kubelet.service.d/
install -m 755 -d %{buildroot}%{_sysconfdir}/cni/net.d/
install -m 755 -d %{buildroot}%{_sysconfdir}/kubernetes/manifests/
install -m 755 -d %{buildroot}/var/lib/kubelet/
install -p -m 755 -t %{buildroot}%{_bindir}/ kubelet
install -p -m 755 -t %{buildroot}%{_bindir}/ kubectl
install -p -m 755 -t %{buildroot}%{_bindir}/ kubeadm
install -p -m 755 -t %{buildroot}%{_sysconfdir}/systemd/system/ kubelet.service
install -p -m 755 -t %{buildroot}%{_sysconfdir}/systemd/system/kubelet.service.d/ 10-kubeadm.conf


install -m 755 -d %{buildroot}/opt/cni
curl -sSL --fail --retry 5 https://storage.googleapis.com/kubernetes-release/network-plugins/cni-amd64-%{CNI_RELEASE}.tar.gz | tar xz
mv bin/ %{buildroot}/opt/cni/


%files
%{_bindir}/kubelet
%{_sysconfdir}/systemd/system/kubelet.service
%{_sysconfdir}/kubernetes/manifests/

%files -n kubernetes-cni
/opt/cni

%files -n kubectl
%{_bindir}/kubectl

%files -n kubeadm
%{_bindir}/kubeadm
%{_sysconfdir}/systemd/system/kubelet.service.d/10-kubeadm.conf

%doc


%changelog
* Tue Sep 20 2016 dgoodwin <dgoodwin@redhat.com> - 1.4.0-0
- Add kubectl and kubeadm sub-packages.
- Rename to kubernetes-cni.
- Update versions of CNI.

* Wed Jul 20 2016 dgoodwin <dgoodwin@redhat.com> - 1.3.4-1
- Initial packaging.
