# godist
A simple generator for creating easily distributable Go packages

This automation is providing an easy way to generate variables that can be configured,
globally across multiple go packages, things like trademarks, product names, websites,
etc., this generator should be defined as a `go:generate` comment and run at build time
using the `go generate` command.

### Example-1
Simple use within a single main package, create a file called dist_gen.go with:

```go
package main
//go:generate go run github.com/afiune/godist
```

The automation will deploy a file called `dist.go` with all the variables defined inside
the JSON file `glob_dist.json` inside this repository. (find a real example at
[example-main/](example-main))

### Example-2
Multi-package, create a go package called dist/ with a file called `gen.go` with:

```go
package dist
//go:generate go run github.com/afiune/godist global.go dist
```

This usage is for go projects that has multiple packages, by creating a single `dist` package
inside your repository, you can import the generated package in any other packages. (find a
real example at [example-multi-pkg/](example-multi-pkg))

### Example-3
Using a custom JSON file containing global variables

```go
package dist
//go:generate go run github.com/afiune/godist global.go dist http://shorturl.at/arxQX
```

To fully customize this automation, a user can also provide a custom URL poiting to a JSON file
that contains the the global variables to define.
