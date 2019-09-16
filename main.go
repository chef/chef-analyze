package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-chef/chef"
)

// TODO @afiune migrate these flags to use viper since it is our standard @Chef
var (
	clientName    string
	clientKey     string
	chefServerUrl string
)

func main() {
	// TODO @afiune migrate these flags to use viper since it is our standard @Chef
	flag.StringVar(&clientName, "user", "", "Chef Infra Server API client username.")
	flag.StringVar(&clientKey, "key", "", "Chef Infra Server API client key.")
	flag.StringVar(&chefServerUrl, "chef_server_url", "", "Chef Infra Server URL.")
	flag.Parse()

	if clientName == "" || clientKey == "" || chefServerUrl == "" {
		fmt.Println("One or more parameters missing. Required: -user USER -key KEY -chef_server_url URL")
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
	}

	// List Nodes
	nodesList, err := client.Nodes.List()
	if err != nil {
		fmt.Println("Unable to list nodes:", err)
		os.Exit(1)
	}

	// Print out the list
	fmt.Println("Nodes:")
	fmt.Println(nodesList)

	// List Cookbooks
	cookList, err := client.Cookbooks.List()
	if err != nil {
		fmt.Println("Unable to list cookbooks:", err)
		os.Exit(1)
	}

	// Print out the list
	fmt.Println("Cookbooks:")
	fmt.Println(cookList)
}
