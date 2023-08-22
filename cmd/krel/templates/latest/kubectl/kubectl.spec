%global debug_package %{nil}

Name: kubectl
Version: {{ .RPMVersion }}
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
%{summary}.

%prep
%setup -q -c

%build
# Nothing to build

%install
# Install binaries
mkdir -p %{buildroot}%{_bindir}
install -p -m 755 %{_arch}/kubectl %{buildroot}%{_bindir}/kubectl

%files
%{_bindir}/kubectl
%license LICENSE
%doc README.md

%changelog
