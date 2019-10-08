pkg_name=chef-analyze
pkg_origin=chef
pkg_description="Chef Inventory analyzer command line interface (CLI)"
pkg_maintainer="Chef Software Inc. <support@chef.io>"
pkg_bin_dirs=(bin)
pkg_deps=(core/glibc)
pkg_scaffolding=core/scaffolding-go
scaffolding_go_module=on

pkg_version() {
  cat "$SRC_PATH/VERSION"
}

do_before() {
  do_default_before
  update_pkg_version
}
