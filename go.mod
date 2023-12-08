module github.com/chef/chef-analyze

go 1.13

require (
	github.com/aws/aws-sdk-go v1.44.108
	github.com/chef/go-libs v0.4.1
	github.com/cheggaaa/pb/v3 v3.1.0
	github.com/go-chef/chef v0.27.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pkg/errors v0.9.1
	github.com/r3labs/diff v1.1.0 // indirect
	github.com/smartystreets/assertions v1.1.1 // indirect
	github.com/spf13/cobra v1.5.0
	github.com/spf13/viper v1.18.1
	github.com/stretchr/testify v1.8.4
	golang.org/x/crypto v0.16.0
)

replace github.com/go-chef/chef v0.27.0 => github.com/chef/go-chef v0.4.5
