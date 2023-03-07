Name: kubelet
Version: {{ .Version }}
Release: {{ .Revision }}
Summary: Container cluster management

%if "%{_vendor}" == "debbuild"
Group: net
%endif

Packager: Kubernetes Authors <dev@kubernetes.io>
License: Apache-2.0
URL: https://kubernetes.io
Source0: %{name}_%{version}.orig.tar.gz

BuildRequires: systemd
Requires: iptables >= 1.4.21
Requires: kubernetes-cni >= {{ index .Dependencies "kubernetes-cni" }}
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

%if 0%{?suse_version}
Requires(post,postun): %fillup_prereq
%endif

%description
The node agent of Kubernetes, the container cluster manager.

%prep
%setup -c -D -T -a 0 -n kubelet

%build

%install

KUBE_ARCH="$(uname -m)"

cd %{_builddir}/kubelet/
mkdir -p %{buildroot}%{_unitdir}/kubelet.service.d/
mkdir -p %{buildroot}%{_bindir}/
mkdir -p %{buildroot}/var/lib/kubelet/

install -m 755 -d %{buildroot}%{_unitdir}
install -m 755 -d %{buildroot}%{_unitdir}/kubelet.service.d/
install -m 755 -d %{buildroot}%{_bindir}
install -m 755 -d %{buildroot}/var/lib/kubelet/
install -p -m 755 -t %{buildroot}%{_bindir}/ ${KUBE_ARCH}/kubelet
install -p -m 644 -t %{buildroot}%{_unitdir}/ kubelet.service

%if 0%{?suse_version}
mkdir -p %{buildroot}%{_fillupdir}/
install -m 644 -T kubelet.env %{buildroot}%{_fillupdir}/sysconfig.kubelet
%else
mkdir -p %{buildroot}%{_sysconfdir}/sysconfig/
install -m 755 -d %{buildroot}%{_sysconfdir}/sysconfig/
install -p -m 644 -T kubelet.env %{buildroot}%{_sysconfdir}/sysconfig/kubelet
%endif

%files
%{_bindir}/kubelet
%{_unitdir}/kubelet.service
%if 0%{?suse_version}
%{_fillupdir}/sysconfig.kubelet
%else
%config(noreplace) %{_sysconfdir}/sysconfig/kubelet
%endif

%if "%{_vendor}" == "debbuild"
%license %{_builddir}/kubelet/LICENSE
%doc %{_builddir}/kubelet/README.md
%else
%license LICENSE
%doc README.md
%endif

%preun
%systemd_preun kubelet.service

%post
%if 0%{?suse_version}
%fillup_only kubelet
%endif
%systemd_post kubelet.service

%postun
%systemd_postun kubelet.service

%changelog
