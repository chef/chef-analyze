package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	chef "github.com/afiune/go-chef"
	log "github.com/sirupsen/logrus"
)

func main() {
	// TODO @afiune migrate these flags to use viper since it is our standard @Chef
	var (
		clientName           string
		clientKey            string
		chefServerUrl        string
		downloadCookbookName string
	)

	flag.StringVar(&clientName, "user", "", "Chef Infra Server API client username.")
	flag.StringVar(&clientKey, "key", "", "Chef Infra Server API client key.")
	flag.StringVar(&chefServerUrl, "chef_server_url", "", "Chef Infra Server URL.")
	flag.StringVar(&downloadCookbookName, "download_cookbook", "", "Tests a cookbook download")
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

	query := map[string]interface{}{
		"name":         []string{"name"},
		"chef_version": []string{"chef_packages", "chef", "version"},
		"os":           []string{"platform"},
		"os_version":   []string{"platform_version"}}

	pres, err := client.Search.PartialExec("node", "*:*", query)
	if err != nil {
		log.Fatal("Unable to collect nodes information", err)
	}

	// using 'v' not 's' because not all fields will have values.
	formatString := "%30s   %-12v   %-15v   %-10v\n"
	fmt.Printf(formatString, "Node Name", "Chef Version", "OS", "OS Version")
	for _, element := range pres.Rows {
		v := element.(map[string]interface{})["data"].(map[string]interface{})
		fmt.Printf(formatString, v["name"], v["chef_version"], v["os"], v["os_version"])
	}

	if downloadCookbookName != "" {
		fmt.Println("\n * Flag -download_cookbook provided.")
		err := DownloadCookbook(client, downloadCookbookName)
		if err != nil {
			log.Fatal("Unable to download cookbook", err)
		}
	}
}
