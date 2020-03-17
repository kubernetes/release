Name: kubectl
Version: {{ .Version }}
Release: {{ .Revision }}
Summary: Command-line utility for interacting with a Kubernetes cluster.

License: ASL 2.0
URL: https://kubernetes.io
Source0: {{ .DownloadLinkBase }}/bin/linux/{{ .GoArch }}/kubectl

BuildRequires: systemd
BuildRequires: curl

%description
Command-line utility for interacting with a Kubernetes cluster.

%prep
cp -p %SOURCE0 %{_builddir}/

# TODO: Do we need these?
#%autosetup
#%build
#%%configure
#%make_build

%install
# TODO: Do we need this?
#rm -rf $RPM_BUILD_ROOT

cd %{_builddir}
install -p -m 755 -t %{buildroot}%{_bindir}/ kubectl

# TODO: Do we need this?
#%make_install

%files
%{_bindir}/kubectl

# TODO: Do we need these?
#%license add-license-file-here
#%doc add-docs-here


%changelog
* Sat Jan  4 2020 Stephen Augustus <saugustus@vmware.com> - 1.18.0
- Create separate spec file for kubectl
