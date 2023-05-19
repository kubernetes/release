Name: kubectl
Version: {{ .Version }}
Release: {{ .Revision }}
Summary: Command-line utility for interacting with a Kubernetes cluster

%if "%{_vendor}" == "debbuild"
Group: admin
%endif

Packager: Kubernetes Authors <dev@kubernetes.io>
License: Apache-2.0
URL: https://kubernetes.io
Source0: %{name}_%{version}.orig.tar.gz

%description
Command-line utility for interacting with a Kubernetes cluster.

%prep
%setup -c -D -T -a 0 -n kubectl

%build

%install

KUBE_ARCH="$(uname -m)"

cd %{_builddir}/kubectl/${KUBE_ARCH}/
mkdir -p %{buildroot}%{_bindir}

install -p -m 755 -t %{buildroot}%{_bindir}/ kubectl

%files
%{_bindir}/kubectl
%if "%{_vendor}" == "debbuild"
%license %{_builddir}/kubectl/LICENSE
%doc %{_builddir}/kubectl/README.md
%else
%license LICENSE
%doc README.md
%endif

%changelog
