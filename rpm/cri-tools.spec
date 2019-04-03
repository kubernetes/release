%global CRI_TOOLS_VERSION 1.12.0
%global RPM_RELEASE 0
%global ARCH amd64

Name: cri-tools
Version: %{CRI_TOOLS_VERSION}
Release: %{RPM_RELEASE}
License: ASL 2.0
Summary: Container Runtime Interface tools
URL: https://kubernetes.io

Source0: https://github.com/kubernetes-incubator/cri-tools/releases/download/v%{CRI_TOOLS_VERSION}/crictl-v%{CRI_TOOLS_VERSION}-linux-%{ARCH}.tar.gz

%description
Command-line utility for interacting with a container runtime.

%prep
%setup -c

%install
pwd
ls -l
install -m 755 -d %{buildroot}%{_bindir}
install -p -m 755 -t %{buildroot}%{_bindir}/ crictl

%files
%{_bindir}/crictl

%changelog
* Thu Jun 13 2019 Tim Pepper <tpepper@vmware.com> - 1.12.0
- Create cri-tools as its own spec file.
