%global debug_package %{nil}

Name: kubeadm
Version: {{ .Version }}
Release: {{ .Revision }}
Summary: Command-line utility for administering a Kubernetes cluster

%if "%{_vendor}" == "debbuild"
Group: admin
%endif

Packager: Kubernetes Authors <dev@kubernetes.io>
License: Apache-2.0
URL: https://kubernetes.io
Source0: %{name}_%{version}.orig.tar.gz

Requires: kubelet >= {{ index .Dependencies "kubelet" }}
Requires: kubectl >= {{ index .Dependencies "kubectl" }}
Requires: kubernetes-cni >= {{ index .Dependencies "kubernetes-cni" }}
Requires: cri-tools >= {{ index .Dependencies "cri-tools" }}

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
mkdir -p %{buildroot}%{_bindir}
mkdir -p %{buildroot}%{_unitdir}/kubelet.service.d/

install -p -m 755 ${KUBE_ARCH}/kubeadm %{buildroot}%{_bindir}/kubeadm
install -p -m 644 10-kubeadm.conf %{buildroot}%{_unitdir}/kubelet.service.d/10-kubeadm.conf

%files
%{_bindir}/kubeadm
%dir %{_unitdir}/kubelet.service.d
%{_unitdir}/kubelet.service.d/10-kubeadm.conf
%license LICENSE
%doc README.md

%changelog
