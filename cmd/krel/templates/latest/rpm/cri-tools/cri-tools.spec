%undefine _missing_build_ids_terminate_build

Name: cri-tools
Version: {{ .Version }}
Release: {{ .Revision }}
Summary: Command-line utility for interacting with a container runtime

%if "%{_vendor}" == "debbuild"
Group: admin
%endif

Packager: Kubernetes Authors <dev@kubernetes.io>
License: Apache-2.0
URL: https://kubernetes.io
Source0: %{name}_%{version}.orig.tar.gz

%description
Command-line utility for interacting with a container runtime.

%prep
%setup -c -D -T -a 0 -n cri-tools

%build

%install

KUBE_ARCH="$(uname -m)"

cd %{_builddir}/cri-tools/${KUBE_ARCH}/
mkdir -p %{buildroot}%{_bindir}

install -p -m 755 -t %{buildroot}%{_bindir}/ crictl

%files
%{_bindir}/crictl
%if "%{_vendor}" == "debbuild"
%license %{_builddir}/cri-tools/LICENSE
%doc %{_builddir}/cri-tools/README.md
%else
%license LICENSE
%doc README.md
%endif

%changelog
