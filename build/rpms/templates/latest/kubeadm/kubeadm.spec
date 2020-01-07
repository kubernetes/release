Name: kubeadm
Version: {{ .Version }}
Release: {{ .Revision }}
Summary: Command-line utility for administering a Kubernetes cluster.

License: ASL 2.0
URL: https://kubernetes.io
Source0: {{ .DownloadLinkBase }}/bin/linux/{{ .GoArch }}/kubeadm
Source1: 10-kubeadm.conf

# TODO: Need to templatize dependencies
BuildRequires: systemd
BuildRequires: curl
Requires: kubelet >= {{ index .Dependencies "kubelet" }}
Requires: kubectl >= {{ index .Dependencies "kubectl" }}
Requires: kubernetes-cni >= {{ index .Dependencies "kubernetes-cni" }}
Requires: cri-tools >= {{ index .Dependencies "cri-tools" }}

%description
Command-line utility for administering a Kubernetes cluster.

%prep
cp -p %SOURCE0 %{_builddir}/
cp -p %SOURCE1 %{_builddir}/

# TODO: Do we need these?
#%autosetup
#%build
#%configure
#%make_build

%install
# TODO: Do we need this?
#rm -rf $RPM_BUILD_ROOT

cd %{_builddir}
install -m 755 -d %{buildroot}%{_sysconfdir}/kubernetes/manifests/
install -p -m 755 -t %{buildroot}%{_bindir}/ kubeadm
install -p -m 644 -t %{buildroot}%{_unitdir}/kubelet.service.d/ 10-kubeadm.conf

# TODO: Do we need this?
#%make_install

%files
%{_bindir}/kubeadm
%{_unitdir}/kubelet.service.d/10-kubeadm.conf

# TODO: Do we need these?
#%license add-license-file-here
#%doc add-docs-here


%changelog
* Sat Jan  4 2020 Stephen Augustus <saugustus@vmware.com> - 1.18.0
- Create separate spec file for kubeadm
