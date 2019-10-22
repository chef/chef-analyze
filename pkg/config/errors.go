package config

var (
	ConfigNotFoundErr = `
  configuration file not found. (default: $HOME/.chef/credentials)

  setup your local workstation by following this documentation:
    - https://docs.chef.io/knife_setup.html#knife-profiles 
`
)
