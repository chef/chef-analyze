package config

//
// The intend of this file is to have a single place where we can easily
// visualize the list of all error messages that we present to users.
//

const (
	ProfileNotFoundErr = `
  profile not found in credentials file.

  verify the format of the credentials file by following this documentation:
    - https://docs.chef.io/knife_setup.html#knife-profiles
`
	CredentialsNotFoundErr = `
  credentials file not found. (default: $HOME/.chef/credentials)

  setup your local credentials config by following this documentation:
    - https://docs.chef.io/knife_setup.html#knife-profiles
`
	MalformedCredentialsFileErr = `
  unable to parse credentials file.

  verify the format of the credentials file by following this documentation:
    - https://docs.chef.io/knife_setup.html#knife-profiles
`
)
