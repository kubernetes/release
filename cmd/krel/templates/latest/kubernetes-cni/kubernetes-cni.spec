%global debug_package %{nil}
%undefine _missing_build_ids_terminate_build

Name: kubernetes-cni
Version: {{ .RPMVersion }}
Release: {{ .Revision }}
Summary: Binaries required to provision kubernetes container networking

%if "%{_vendor}" == "debbuild"
Group: net
%endif

Packager: Kubernetes Authors <dev@kubernetes.io>
License: Apache-2.0
URL: https://kubernetes.io
Source0: %{name}_%{version}.orig.tar.gz

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
mkdir -p %{buildroot}/opt/cni/bin
mkdir -p %{buildroot}%{_sysconfdir}/cni/net.d/

cp -a ${KUBE_ARCH}/* %{buildroot}/opt/cni/bin/

%if "%{_vendor}" == "debbuild"
touch %{buildroot}%{_sysconfdir}/cni/net.d/.kubernetes-cni-keep
%endif

%files
/opt/cni/
%dir %{_sysconfdir}/cni
%dir %{_sysconfdir}/cni/net.d
%if "%{_vendor}" == "debbuild"
%{_sysconfdir}/cni/net.d/.kubernetes-cni-keep
%endif
%license LICENSE
%doc README.md

%changelog
