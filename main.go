package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-chef/chef"
	log "github.com/sirupsen/logrus"
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

	if err := nodesReport(client); err != nil {
		log.Fatal("Unable to collect nodes information", err)
	}

	// Separator
	fmt.Println("\n")

	if err := cookbooksReport(client); err != nil {
		log.Fatal("Unable to collect cookbooks information", err)
	}
}

func nodesReport(client *chef.Client) error {
	query := map[string]interface{}{
		"name":         []string{"name"},
		"chef_version": []string{"chef_packages", "chef", "version"},
		"os":           []string{"platform"},
		"os_version":   []string{"platform_version"}}

	pres, err := client.Search.PartialExec("node", "*:*", query)
	if err != nil {
		return err
	}

	// using 'v' not 's' because not all fields will have values.
	formatString := "%30s   %-12v   %-15v   %-10v\n"
	fmt.Printf(formatString, "Node Name", "Chef Version", "OS", "OS Version")
	for _, element := range pres.Rows {
		v := element.(map[string]interface{})["data"].(map[string]interface{})
		fmt.Printf(formatString, v["name"], v["chef_version"], v["os"], v["os_version"])
	}

	return nil
}

func cookbooksReport(client *chef.Client) error {
	cbooksList, err := client.Cookbooks.ListAvailableVersions("all")
	if err != nil {
		return err
	}

	formatString := "%30s   %-12v\n"
	fmt.Printf(formatString, "Cookbook Name", "Version(s)")
	for cookbook, cbookVersions := range cbooksList {
		versionsArray := make([]string, len(cbookVersions.Versions))
		for i, details := range cbookVersions.Versions {
			versionsArray[i] = details.Version
		}
		fmt.Printf(formatString, cookbook, strings.Join(versionsArray[:], ", "))
	}

	return nil
}
