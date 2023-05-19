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
Command-line utility for administering a Kubernetes cluster.

%prep
%setup -c -D -T -a 0 -n kubeadm

%build

%install

KUBE_ARCH="$(uname -m)"

cd %{_builddir}/kubeadm/
mkdir -p %{buildroot}%{_bindir}
mkdir -p %{buildroot}%{_sysconfdir}/kubernetes/manifests/
mkdir -p %{buildroot}%{_unitdir}/kubelet.service.d/

install -p -m 755 -t %{buildroot}%{_bindir}/ ${KUBE_ARCH}/kubeadm
install -m 755 -d %{buildroot}%{_sysconfdir}/kubernetes/manifests/
install -p -m 644 -t %{buildroot}%{_unitdir}/kubelet.service.d/ 10-kubeadm.conf

%files
%{_bindir}/kubeadm
%dir %{_unitdir}/kubelet.service.d
%{_unitdir}/kubelet.service.d/10-kubeadm.conf

%if "%{_vendor}" == "debbuild"
%license %{_builddir}/kubeadm/LICENSE
%doc %{_builddir}/kubeadm/README.md
%else
%license LICENSE
%doc README.md
%endif

%changelog
