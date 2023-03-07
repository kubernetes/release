%undefine _missing_build_ids_terminate_build

Name: kubernetes-cni
Version: {{ .Version }}
Release: {{ .Revision }}
Summary: Binaries required to provision kubernetes container networking

%if "%{_vendor}" == "debbuild"
Group: net
%endif

Packager: Kubernetes Authors <dev@kubernetes.io>
License: Apache-2.0
URL: https://kubernetes.io
Source0: %{name}_%{version}.orig.tar.gz

Requires: kubelet

%description
Binaries required to provision container networking.

%prep
%setup -c -D -T -a 0 -n cni-plugins

%build

%install

KUBE_ARCH="$(uname -m)"

cd %{_builddir}/cni-plugins/${KUBE_ARCH}/
mkdir -p %{buildroot}/opt/cni/bin
mkdir -p %{buildroot}%{_sysconfdir}/cni/net.d/

install -m 755 -d %{buildroot}%{_sysconfdir}/cni/net.d/
install -m 755 -d %{buildroot}/opt/cni/bin
mv ./* %{buildroot}/opt/cni/bin/

%files
/opt/cni
%if "%{_vendor}" == "debbuild"
%license %{_builddir}/cni-plugins/LICENSE
%doc %{_builddir}/cni-plugins/README.md
%else
%license LICENSE
%doc README.md
%endif

%changelog
