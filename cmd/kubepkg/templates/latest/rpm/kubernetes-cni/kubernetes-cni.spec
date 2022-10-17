Name: kubernetes-cni
Version: {{ .Version }}
Release: {{ .Revision }}
Summary: Binaries required to provision kubernetes container networking

License: ASL 2.0
URL: https://kubernetes.io
Source0: {{ .CNIDownloadLink }}

BuildRequires: systemd
BuildRequires: curl
Requires: kubelet

%description
Binaries required to provision container networking.

%prep
%setup -c -D -T -a 0 -n cni-plugins

%install
cd %{_builddir}
mkdir -p %{buildroot}%{_sysconfdir}/cni/net.d/
mkdir -p %{buildroot}/opt/cni/bin
install -m 755 -d %{buildroot}%{_sysconfdir}/cni/net.d/
install -m 755 -d %{buildroot}/opt/cni/bin
mv cni-plugins/* %{buildroot}/opt/cni/bin/

%files
/opt/cni

# TODO: Do we need these?
#%license add-license-file-here
#%doc add-docs-here


%changelog
* Sat Jan  4 2020 Stephen Augustus <saugustus@vmware.com> - 1.18.0
- Create separate spec file for kubernetes-cni
