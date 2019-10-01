package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-chef/chef"
)

func main() {
	// TODO @afiune migrate these flags to use viper since it is our standard @Chef
	var (
		clientName    string
		clientKey     string
		chefServerUrl string
	)

	flag.StringVar(&clientName, "user", "", "Chef Infra Server API client username.")
	flag.StringVar(&clientKey, "key", "", "Chef Infra Server API client key.")
	flag.StringVar(&chefServerUrl, "chef_server_url", "", "Chef Infra Server URL.")
	flag.Parse()

	if len(os.Args) > 1 && os.Args[1] == "help" {
		flag.Usage()
		os.Exit(0)
	}

	if clientName == "" || clientKey == "" || chefServerUrl == "" {
		fmt.Println("One or more required parameters missing:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// read a client key
	key, err := ioutil.ReadFile(clientKey)
	if err != nil {
		fmt.Println("Couldn't read key pem:", err)
		os.Exit(1)
	}

	// build a client
	client, err := chef.NewClient(&chef.Config{
		Name:    clientName,
		Key:     string(key),
		BaseURL: chefServerUrl,
	})
	if err != nil {
		fmt.Println("Unable setup Chef client:", err)
		os.Exit(1)
	}

	// List Nodes
	nodesList, err := client.Nodes.List()
	if err != nil {
		fmt.Println("Unable to list nodes:", err)
		os.Exit(1)
	}

	// Print out the list
	fmt.Println("Nodes:")
	for nodeName, _ := range nodesList {
		fmt.Println(nodeName)
	}

	// List Cookbooks
	cookbookList, err := client.Cookbooks.List()
	if err != nil {
		fmt.Println("Unable to list cookbooks:", err)
		os.Exit(1)
	}

	// Print out the list
	fmt.Println("\nCookbooks:")
	for cookbookName, cookbookVersions := range cookbookList {
		versionsArray := cookbookVersionsAsArrayStrings(&cookbookVersions)
		fmt.Printf("%s %v\n", cookbookName, versionsArray)
	}
}

// TODO @afiune contribute back to the community repo, propose to add a function to the struct
// that automatically extracts these values for convenience at:
//
// https://github.com/go-chef/chef/blob/master/cookbook.go#L28
func cookbookVersionsAsArrayStrings(cookbookVersions *chef.CookbookVersions) []string {
	rawVersions := make([]string, len(cookbookVersions.Versions))

	for x, cookbookVersion := range cookbookVersions.Versions {
		rawVersions[x] = cookbookVersion.Version
	}

	return rawVersions
}
