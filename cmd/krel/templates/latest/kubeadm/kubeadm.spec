%global debug_package %{nil}

Name: kubeadm
Version: {{ .RPMVersion }}
Release: {{ .Revision }}
Summary: Command-line utility for administering a Kubernetes cluster

%if "%{_vendor}" == "debbuild"
Group: admin
%endif

Packager: Kubernetes Authors <dev@kubernetes.io>
License: Apache-2.0
URL: https://kubernetes.io
Source0: %{name}_%{version}.orig.tar.gz

{{ range $dep := .Metadata.Dependencies }}
Requires: {{ $dep.Name }} {{ $dep.VersionConstraint }}
{{ end }}

%if "%{_vendor}" == "debbuild"
BuildRequires: systemd-deb-macros
BuildRequires: sed
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
mkdir -p %{buildroot}%{_bindir}
mkdir -p %{buildroot}%{_unitdir}/kubelet.service.d/

%if "%{_vendor}" == "debbuild"
sed -i 's;/etc/sysconfig/kubelet;/etc/default/kubelet;g' 10-kubeadm.conf
%endif

install -p -m 755 ${KUBE_ARCH}/kubeadm %{buildroot}%{_bindir}/kubeadm
install -p -m 644 10-kubeadm.conf %{buildroot}%{_unitdir}/kubelet.service.d/10-kubeadm.conf

%files
%{_bindir}/kubeadm
%dir %{_unitdir}/kubelet.service.d
%{_unitdir}/kubelet.service.d/10-kubeadm.conf
%license LICENSE
%doc README.md

%changelog
