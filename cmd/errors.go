package cmd

//
// The intend of this file is to have a single place where we can easily
// visualize the list of all error messages that we present to users.
//

const (
	MissingMinimumParametersErr = `
  there are missing parameters for this tool to work, provide a credentials file:
    --credentials string

  or the following required flags:
    --client_key string
    --client_name string
    --chef_server_url string
`
)
