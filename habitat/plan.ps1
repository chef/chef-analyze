$pkg_name="chef-analyze"
$pkg_origin="chef"
$pkg_description="Chef Inventory analyzer command line interface (CLI)"
$pkg_maintainer="Chef Software Inc. <support@chef.io>"
$pkg_bin_dirs=@("bin")
$pkg_deps=@("core/go")

function pkg_version {
  Get-Content "$SRC_PATH/VERSION"
}

function Invoke-Before {
  Set-PkgVersion
}

function Invoke-Prepare {
  $Env:GOFLAGS = "-mod=vendor"
  Write-BuildLine "Setting GOFLAGS=$env:GOFLAGS"
}

function Invoke-Install {
  Push-Location "$SRC_PATH"
  go build -o $pkg_prefix/bin/chef-analyze.exe
}

function Invoke-Check{
  $Env:CHEF_FEAT_ANALYZE = "true"
  (& "$pkg_prefix/bin/chef-analyze.exe" version).StartsWith("Analyze your Chef Infra Server artifacts")
}
