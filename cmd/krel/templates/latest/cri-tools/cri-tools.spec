%global debug_package %{nil}
%undefine _missing_build_ids_terminate_build

Name: cri-tools
Version: {{ .RPMVersion }}
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
%{summary}.

%prep
%setup -q -c

%build
# Nothing to build

%install
# Detect host arch
KUBE_ARCH="$(uname -m)"

# Install binaries
mkdir -p %{buildroot}%{_bindir}
install -p -m 755 ${KUBE_ARCH}/crictl %{buildroot}%{_bindir}/crictl

%files
%{_bindir}/crictl
%license LICENSE
%doc README.md

%changelog
