%global KUBE_VERSION 1.5.1
%global KUBEADM_VERSION 1.6.0-alpha.0.2074+a092d8e0f95f52
%global CNI_RELEASE 07a8a28637e97b22eb8dfe710eeae1344f69d16e
%global RPM_RELEASE 0

Name: kubelet
Version: %{KUBE_VERSION}
Release: %{RPM_RELEASE}
Summary: Container cluster management
License: ASL 2.0

URL: https://kubernetes.io
Source0: https://dl.k8s.io/v%{KUBE_VERSION}/bin/linux/amd64/kubelet
Source1: kubelet.service
Source2: https://dl.k8s.io/v%{KUBE_VERSION}/bin/linux/amd64/kubectl
Source3: https://dl.k8s.io/ci-cross/v%{KUBEADM_VERSION}/bin/linux/amd64/kubeadm
Source4: 10-kubeadm.conf
Source5: https://dl.k8s.io/network-plugins/cni-amd64-%{CNI_RELEASE}.tar.gz


BuildRequires: curl
Requires: iptables >= 1.4.21
Requires: kubernetes-cni >= 0.3.0.1
Requires: socat
Requires: util-linux
Requires: ethtool
Requires: iproute
Requires: ebtables


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

Version: %{KUBE_VERSION}
Release: %{RPM_RELEASE}
Summary: Command-line utility for interacting with a Kubernetes cluster.

%description -n kubectl
Command-line utility for interacting with a Kubernetes cluster.

%package -n kubeadm

Version: 1.6.0
Release: %{RPM_RELEASE}.alpha.0.2074.a092d8e0f95f52
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
# This is a useful hack for faster Docker builds when working on the spec or
# with locally obtained sources.
#
# Example:
#   spectool -gf kubelet.spec
#   rpmbuild --define "_sourcedir $PWD" -bb kubelet.spec
#

cp -p %SOURCE0 %{_builddir}/
cp -p %SOURCE1 %{_builddir}/
cp -p %SOURCE2 %{_builddir}/
cp -p %SOURCE3 %{_builddir}/
cp -p %SOURCE4 %{_builddir}/
%setup -D -T -a 5 -n %{_builddir}/


%install

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
# bin directory from cni-amd64-%{CNI_RELEASE}.tar.gz with a list of cni plugins (among other things)
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
* Tue Dec 13 2016 Lucas Käldström <lucas.kaldstrom@hotmail.co.uk> - 1.5.1
- Bump version of kubelet and kubectl to v1.5.1, plus kubeadm to the third stable version

* Fri Dec 6 2016 Lucas Käldström <lucas.kaldstrom@hotmail.co.uk> - 1.5.0-beta.2
- Bump version of kubelet and kubectl

* Wed Nov 16 2016 Alexander Kanevskiy <alexander.kanevskiy@intel.com>
- fix iproute and mount dependencies (#204)

* Sun Nov 6 2016 Lucas Käldström <lucas.kaldstrom@hotmail.co.uk>
- Sync the debs and rpm files; add some kubelet dependencies to the rpm manifest

* Wed Nov 2 2016 Lucas Käldström <lucas.kaldstrom@hotmail.co.uk>
- Bump version of kubeadm to v1.5.0-alpha.2.380+85fe0f1aadf91e

* Fri Oct 21 2016 Ilya Dmitrichenko <errordeveloper@gmail.com> - 1.4.4-0
- Bump version of kubelet and kubectl

* Mon Oct 17 2016 Lucas Käldström <lucas.kaldstrom@hotmail.co.uk> - 1.4.3-0
- Bump version of kubeadm

* Fri Oct 14 2016 Matthew Mosesohn  <mmosesohn@mirantis.com> - 1.4.0-1
- Allow locally built/previously downloaded binaries

* Tue Sep 20 2016 dgoodwin <dgoodwin@redhat.com> - 1.4.0-0
- Add kubectl and kubeadm sub-packages.
- Rename to kubernetes-cni.
- Update versions of CNI.

* Wed Jul 20 2016 dgoodwin <dgoodwin@redhat.com> - 1.3.4-1
- Initial packaging.
