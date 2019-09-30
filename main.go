package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
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
	if clientName == "" || clientKey == "" || chefServerUrl == "" {
		fmt.Println("One or more parameters missing.\n\nRequired:\n\t-user USER\n\t-key KEY\n\t-chef_server_url URL")
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

	query := make(map[string]interface{})
	query["name"] = []string{"name"}
	query["chef_version"] = []string{"chef_packages", "chef", "version"}
	query["os"] = []string{"platform"}
	query["os_version"] = []string{"platform_version"}
	pres, err := client.Search.PartialExec("node", "*:*", query)
	if err != nil {
		log.Fatal("Error running Search.PartialExec()", err)
	}
	// using 'v' not 's' because not all fields will have values.
	formatString := "%30s   %-12v   %-15v   %-10v\n"
	fmt.Printf(formatString, "Node Name", "Chef Version", "OS", "OS Version")
	for _, element := range pres.Rows {
		v := element.(map[string]interface{})["data"].(map[string]interface{})
		fmt.Printf(formatString, v["name"], v["chef_version"], v["os"], v["os_version"])
	}
}
