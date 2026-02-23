pkg_name=chef-analyze
pkg_origin=chef
pkg_description="A CLI to analyze artifacts from a Chef Infra Server"
pkg_maintainer="Chef Software Inc. <support@chef.io>"
pkg_bin_dirs=(bin)
pkg_deps=(core/glibc core/go25 core/gcc)


pkg_version() {
  cat "$SRC_PATH/VERSION"
}

do_before() {
  do_default_before
  update_pkg_version
}

do_prepare() {
  build_line "Setting GOFLAGS=\"-mod=vendor\""
  export GOFLAGS="-mod=vendor"

  build_line "Running all 'go generate' statements before building"
  ( cd "$SRC_PATH" || exit_with "unable to cd into source directory" 1
    go generate ./...
  )
}
do_build() {
  echo "Building $pkg_name"
  go build -o "$pkg_name"
}

do_install() {
  echo "installing $pkg_name"
  cp "$pkg_name" "$pkg_prefix/bin/$pkg_name"
}

do_check() {
  export CHEF_FEAT_ANALYZE="true"
  "$pkg_prefix/bin/$pkg_name" version | grep "^Analyze your Chef Infra Server artifacts"
}