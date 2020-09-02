%global KUBE_MAJOR 1
%global KUBE_MINOR 13
%global KUBE_PATCH 0
%global KUBE_VERSION %{KUBE_MAJOR}.%{KUBE_MINOR}.%{KUBE_PATCH}
%global RPM_RELEASE 0
%global ARCH amd64

# This expands a (major, minor, patch) tuple into a single number so that it
# can be compared against other versions. It has the current implementation
# assumption that none of these numbers will exceed 255.
%define semver() (%1 * 256 * 256 + %2 * 256 + %3)
%global KUBE_SEMVER %{semver %{KUBE_MAJOR} %{KUBE_MINOR} %{KUBE_PATCH}}

%global CNI_VERSION 0.8.7
%global CRI_TOOLS_VERSION 1.13.0

Name: kubelet
Version: %{KUBE_VERSION}
Release: %{RPM_RELEASE}
Summary: Container cluster management
License: ASL 2.0

URL: https://kubernetes.io
Source0: https://dl.k8s.io/v%{KUBE_VERSION}/bin/linux/%{ARCH}/kubelet
Source1: kubelet.service
Source2: https://dl.k8s.io/v%{KUBE_VERSION}/bin/linux/%{ARCH}/kubectl
Source3: https://dl.k8s.io/v%{KUBE_VERSION}/bin/linux/%{ARCH}/kubeadm
Source4: 10-kubeadm.conf
Source5: https://storage.googleapis.com/k8s-artifacts-cni/release/v%{CNI_VERSION}/cni-plugins-linux-%{ARCH}-v%{CNI_VERSION}.tgz
Source6: kubelet.env
Source7: https://storage.googleapis.com/k8s-artifacts-cri-tools/release/v%{CRI_TOOLS_VERSION}/crictl-v%{CRI_TOOLS_VERSION}-linux-%{ARCH}.tar.gz

BuildRequires: systemd
BuildRequires: curl
Requires: iptables >= 1.4.21
Requires: kubernetes-cni >= %{CNI_VERSION}
Requires: socat
Requires: util-linux
Requires: ethtool
Requires: iproute
Requires: ebtables
Requires: conntrack


%description
The node agent of Kubernetes, the container cluster manager.

%package -n kubernetes-cni

Version: %{CNI_VERSION}
Release: %{RPM_RELEASE}
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

Version: %{KUBE_VERSION}
Release: %{RPM_RELEASE}
Summary: Command-line utility for administering a Kubernetes cluster.
Requires: kubelet >= 1.13.0
Requires: kubectl >= 1.13.0
Requires: kubernetes-cni >= 0.8.6
Requires: cri-tools >= 1.13.0

%description -n kubeadm
Command-line utility for administering a Kubernetes cluster.

%package -n cri-tools

Version: %{CRI_TOOLS_VERSION}
Release: %{RPM_RELEASE}
Summary: Command-line utility for interacting with a container runtime.

%description -n cri-tools
Command-line utility for interacting with a container runtime.

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

ln -s 10-kubeadm-post-1.11.conf %SOURCE4

cp -p %SOURCE0 %{_builddir}/
cp -p %SOURCE1 %{_builddir}/
cp -p %SOURCE2 %{_builddir}/
cp -p %SOURCE3 %{_builddir}/
cp -p %SOURCE4 %{_builddir}/
cp -p %SOURCE6 %{_builddir}/
%setup -c -D -T -a 5 -n cni-plugins
%setup -c -a 7 -T -n cri-tools

%install

# The setup macro from prep will make install start in the cni-plugins directory, so cd back to the root.
cd %{_builddir}
install -m 755 -d %{buildroot}%{_unitdir}
install -m 755 -d %{buildroot}%{_unitdir}/kubelet.service.d/
install -m 755 -d %{buildroot}%{_bindir}
install -m 755 -d %{buildroot}%{_sysconfdir}/cni/net.d/
install -m 755 -d %{buildroot}%{_sysconfdir}/kubernetes/manifests/
install -m 755 -d %{buildroot}/var/lib/kubelet/
install -p -m 755 -t %{buildroot}%{_bindir}/ kubelet
install -p -m 755 -t %{buildroot}%{_bindir}/ kubectl
install -p -m 755 -t %{buildroot}%{_bindir}/ kubeadm
install -p -m 644 -t %{buildroot}%{_unitdir}/ kubelet.service
install -p -m 644 -t %{buildroot}%{_unitdir}/kubelet.service.d/ 10-kubeadm.conf
install -p -m 755 -t %{buildroot}%{_bindir}/ cri-tools/crictl

install -m 755 -d %{buildroot}%{_sysconfdir}/sysconfig/
install -p -m 644 -T kubelet.env %{buildroot}%{_sysconfdir}/sysconfig/kubelet

install -m 755 -d %{buildroot}/opt/cni/bin
# bin directory from cni-plugins-linux-%{ARCH}-%{CNI_VERSION}.tgz with a list of cni plugins (among other things)
mv cni-plugins/* %{buildroot}/opt/cni/bin/

%files
%{_bindir}/kubelet
%{_unitdir}/kubelet.service
%{_sysconfdir}/kubernetes/manifests/

%config(noreplace) %{_sysconfdir}/sysconfig/kubelet

%files -n kubernetes-cni
/opt/cni

%files -n kubectl
%{_bindir}/kubectl

%files -n kubeadm
%{_bindir}/kubeadm
%{_unitdir}/kubelet.service.d/10-kubeadm.conf

%files -n cri-tools
%{_bindir}/crictl

%doc


%changelog
* Wed Sep 2 2020 Stephen Augustus <saugustus@vmware.com> - 1.19.1
- Update CNI plugins to v0.8.7

* Mon Jun 22 2020 Stephen Augustus <saugustus@vmware.com> - 1.18.4
- Unbundle CNI plugins (v0.8.6) from kubelet package and release as kubernetes-cni

* Fri May 29 2020 Stephen Augustus <saugustus@vmware.com> - 1.18.4
- Source cri-tools from https://storage.googleapis.com/k8s-artifacts-cri-tools/release
  instead of https://github.com/kubernetes-sigs/cri-tools

* Fri May 22 2020 Stephen Augustus <saugustus@vmware.com> - 1.18.4
- Bundle CNI plugins (v0.8.6) in kubelet package

* Fri May 22 2020 Stephen Augustus <saugustus@vmware.com> - 1.18.4
- Source CNI plugins from https://storage.googleapis.com/k8s-artifacts-cni/release
  instead of https://dl.k8s.io/network-plugins

* Mon Jun 24 2019 Stephen Augustus <saugustus@vmware.com> - 1.15.1
- Bump minimum versions of all kubernetes dependencies
- Remove conditional logic for unsupported versions of Kubernetes

* Sun Jun 23 2019 Stephen Augustus <saugustus@vmware.com> - 1.15.1
- Source cri-tools from https://github.com/kubernetes-sigs/cri-tools
  instead of https://github.com/kubernetes-incubator/cri-tools

* Thu May 30 2019 Tim Pepper <tpepper@vmware.com>
- Change CNI version check to ">="

* Wed Mar 20 2019 Lubomir I. Ivanov <lubomirivanov@vmware.com>
- Bump CNI version to v0.7.5.

* Tue Sep 25 2018 Chuck Ha <chuck@heptio.com> - 1.12.1
- Bump cri-tools to 1.12.0.

* Fri Jul 13 2018 Lantao Liu <lantaol@google.com> - 1.11.0
- Bump cri-tools to 1.11.1.

* Tue Jun 19 2018 Chuck Ha <chuck@heptio.com> - 1.11.0
- Bump cri-tools to GA version.

* Thu Jun 14 2018 Chuck Ha <chuck@heptio.com> - 1.11.0
- Add a crictl sub-package.

* Fri Jun 8 2018 Chuck Ha <chuck@heptio.com> - 1.11.0
- Bump version and update rpm manifest for kubeadm.

* Fri Dec 15 2017 Anthony Yeh <enisoc@google.com> - 1.9.0
- Release of Kubernetes 1.9.0.

* Thu Oct 19 2017 Di Xu <stephenhsu90@gmail.com>
- Bump CNI version to v0.6.0.

* Fri Sep 29 2017 Jacob Beacham <beacham@google.com> - 1.8.0
- Bump version of kubelet and kubectl to v1.8.0.

* Thu Aug 3 2017 Jacob Beacham <beacham@google.com> - 1.7.3
- Bump version of kubelet and kubectl to v1.7.3.

* Wed Jul 26 2017 Jacob Beacham <beacham@google.com> - 1.7.2
- Bump version of kubelet and kubectl to v1.7.2.

* Fri Jul 14 2017 Jacob Beacham <beacham@google.com> - 1.7.1
- Bump version of kubelet and kubectl to v1.7.1.

* Mon Jun 30 2017 Mike Danese <mikedanese@google.com> - 1.7.0
- Bump version of kubelet and kubectl to v1.7.0.

* Fri May 19 2017 Jacob Beacham <beacham@google.com> - 1.6.4
- Bump version of kubelet and kubectl to v1.6.4.

* Wed May 10 2017 Jacob Beacham <beacham@google.com> - 1.6.3
- Bump version of kubelet and kubectl to v1.6.3.

* Wed Apr 26 2017 Jacob Beacham <beacham@google.com> - 1.6.2
- Bump version of kubelet and kubectl to v1.6.2.

* Mon Apr 3 2017 Mike Danese <mikedanese@google.com> - 1.6.1
- Bump version of kubelet and kubectl to v1.6.1.

* Tue Mar 28 2017 Lucas Käldström <lucas.kaldstrom@hotmail.co.uk>
- Bump CNI version to v0.5.1.

* Wed Mar 15 2017 Lucas Käldström <lucas.kaldstrom@hotmail.co.uk> - 1.6.0
- Bump version of kubelet, kubectl and kubeadm to v1.6.0.

* Tue Dec 13 2016 Mike Danese <mikedanese@google.com> - 1.5.4
- Bump version of kubelet and kubectl to v1.5.4.

* Tue Dec 13 2016 Lucas Käldström <lucas.kaldstrom@hotmail.co.uk> - 1.5.1
- Bump version of kubelet and kubectl to v1.5.1, plus kubeadm to the third stable version

* Tue Dec 6 2016 Lucas Käldström <lucas.kaldstrom@hotmail.co.uk> - 1.5.0-beta.2
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
