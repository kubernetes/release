%global KUBE_MAJOR 1
%global KUBE_MINOR 15
%global KUBE_PATCH 0
%global KUBE_VERSION %{KUBE_MAJOR}.%{KUBE_MINOR}.%{KUBE_PATCH}
%global RPM_RELEASE 0
%global ARCH amd64

# This expands a (major, minor, patch) tuple into a single number so that it
# can be compared against other versions. It has the current implementation
# assumption that none of these numbers will exceed 255.
%define semver() (%1 * 256 * 256 + %2 * 256 + %3)
%global KUBE_SEMVER %{semver %{KUBE_MAJOR} %{KUBE_MINOR} %{KUBE_PATCH}}

%global CNI_VERSION 0.7.5

License: ASL 2.0
URL: https://kubernetes.io
Source0: https://dl.k8s.io/v%{KUBE_VERSION}/bin/linux/%{ARCH}/kubelet
Source1: https://dl.k8s.io/v%{KUBE_VERSION}/bin/linux/%{ARCH}/kubectl
Source2: https://dl.k8s.io/v%{KUBE_VERSION}/bin/linux/%{ARCH}/kubeadm
Source3: kubelet.service
Source4: 10-kubeadm.conf
Source5: kubelet.env

Name: kubelet
Version: %{KUBE_VERSION}
Release: %{RPM_RELEASE}
Summary: Kubernetes cluster node agent
BuildRequires: systemd
BuildRequires: curl
Requires: conntrack
Requires: ebtables
Requires: ethtool
Requires: iproute
Requires: iptables >= 1.4.21
Requires: kubernetes-cni >= %{CNI_VERSION}
Requires: socat
Requires: util-linux

%description
The node agent of Kubernetes, the container cluster manager.

%package -n kubectl
Version: %{KUBE_VERSION}
Release: %{RPM_RELEASE}
Summary: Kubernetes kubectl client utility

%description -n kubectl
Command-line for interacting with a Kubernetes cluster.

%package -n kubeadm
Version: %{KUBE_VERSION}
Release: %{RPM_RELEASE}
Summary: Kubernetes kubeadm cluster management tool
Requires: kubelet >= 1.14.0
Requires: kubectl >= 1.14.0
Requires: kubernetes-cni >= %{CNI_VERSION}
Requires: cri-tools >= 1.11.0

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

ln -s 10-kubeadm-post-1.11.conf %SOURCE4

cp -p %SOURCE0 %{_builddir}/
cp -p %SOURCE1 %{_builddir}/
cp -p %SOURCE2 %{_builddir}/
cp -p %SOURCE3 %{_builddir}/
cp -p %SOURCE4 %{_builddir}/
cp -p %SOURCE5 %{_builddir}/

%install

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

install -m 755 -d %{buildroot}%{_sysconfdir}/sysconfig/
install -p -m 644 -T kubelet.env %{buildroot}%{_sysconfdir}/sysconfig/kubelet

%files
%{_bindir}/kubelet
%{_unitdir}/kubelet.service
%{_sysconfdir}/kubernetes/manifests/

%config(noreplace) %{_sysconfdir}/sysconfig/kubelet

%files -n kubectl
%{_bindir}/kubectl

%files -n kubeadm
%{_bindir}/kubeadm
%{_unitdir}/kubelet.service.d/10-kubeadm.conf

%doc


%changelog
* Thu Jun 13 2019 Tim Pepper <tpepper@vmware.com> - 1.15.0
- Split KUBE_SEMVER's into separate spec files.
- Split kubernetes-cni into its own spec file.
- Split cri-tools into its own spec file.
- Bump CNI version requirement from "=" to ">=" for easier maintenance.
