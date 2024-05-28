%global debug_package %{nil}

Name: kubelet
Version: {{ .RPMVersion }}
Release: {{ .Revision }}
Summary: Node agent for Kubernetes clusters

%if "%{_vendor}" == "debbuild"
Group: net
%endif

Packager: Kubernetes Authors <dev@kubernetes.io>
License: Apache-2.0
URL: https://kubernetes.io
Source0: %{name}_%{version}.orig.tar.gz
Source1: %{name}.rpmlintrc

BuildRequires: systemd
Requires: iptables >= 1.4.21
{{ range $dep := .Metadata.Dependencies }}
Requires: {{ $dep.Name }} {{ $dep.VersionConstraint }}
{{ end }}
%if "%{_vendor}" == "debbuild"
Requires: iproute2
Requires: mount
%else
Requires: iproute
%endif
Requires: socat
Requires: util-linux
Requires: ethtool
Requires: ebtables
Requires: conntrack

%if "%{_vendor}" == "debbuild"
BuildRequires: systemd-deb-macros
%else
BuildRequires: systemd-rpm-macros
%endif

%description
%{summary}.

%prep
%setup -q -c

%build
# Nothing to build

%install
# Detect host arch
KUBE_ARCH="$(uname -m)"

# Install files
mkdir -p %{buildroot}%{_unitdir}/
mkdir -p %{buildroot}%{_bindir}/
mkdir -p %{buildroot}%{_sharedstatedir}/kubelet/
mkdir -p %{buildroot}%{_sysconfdir}/kubernetes/manifests/

install -p -m 755 ${KUBE_ARCH}/kubelet %{buildroot}%{_bindir}/kubelet
install -p -m 644 kubelet.service %{buildroot}%{_unitdir}/kubelet.service

# Required because dpkg-deb doesn't keep empty directories
%if "%{_vendor}" == "debbuild"
touch %{buildroot}%{_sharedstatedir}/kubelet/.kubelet-keep
touch %{buildroot}%{_sysconfdir}/kubernetes/manifests/.kubelet-keep
%endif

%if "%{_vendor}" == "debbuild"
mkdir -p %{buildroot}%{_sysconfdir}/default/
install -p -m 644 -T kubelet.env %{buildroot}%{_sysconfdir}/default/kubelet
%else
mkdir -p %{buildroot}%{_sysconfdir}/sysconfig/
install -p -m 644 -T kubelet.env %{buildroot}%{_sysconfdir}/sysconfig/kubelet
%endif

%files
%{_bindir}/kubelet
%{_unitdir}/kubelet.service
%dir %{_sharedstatedir}/kubelet
%dir %{_sysconfdir}/kubernetes
%dir %{_sysconfdir}/kubernetes/manifests
%if "%{_vendor}" == "debbuild"
%{_sharedstatedir}/kubelet/.kubelet-keep
%{_sysconfdir}/kubernetes/manifests/.kubelet-keep
%config(noreplace) %{_sysconfdir}/default/kubelet
%else
%config(noreplace) %{_sysconfdir}/sysconfig/kubelet
%endif
%license LICENSE
%doc README.md

%preun
%systemd_preun kubelet.service

%post
%systemd_post kubelet.service

%postun
%systemd_postun kubelet.service

%changelog
