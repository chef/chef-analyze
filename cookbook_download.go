package main

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/go-chef/chef"
)

// DownloadCookbook downloads the latest version of a cookbook
// TODO @afiune contribute back to go-chef/chef
func DownloadCookbook(client *chef.Client, name string) error {
	cookbook, err := client.Cookbooks.GetVersion(name, "_latest")
	if err != nil {
		return err
	}

	fmt.Printf("Downloading %s cookbook version %s\n", cookbook.CookbookName, cookbook.Version)
	// TODO @afiune maybe this could be configurable for processes that need to download
	// a single cookbook to verify XYZ or execute verifications on it but not actually
	// persist the cookbook on disk. (temporal files)
	//
	// => IDEA: DownloadCookbookAt(name, path)
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// TODO @afiune Verify if the cookbook directory already exists
	cookbookPath := path.Join(cwd, cookbook.Name)

	downloadsErr := []error{
		downloadItems(client, cookbook.RootFiles, "root_files", cookbookPath),
		downloadItems(client, cookbook.Files, "files", path.Join(cookbookPath, "files")),
		downloadItems(client, cookbook.Templates, "templates", path.Join(cookbookPath, "templates")),
		downloadItems(client, cookbook.Attributes, "attributes", path.Join(cookbookPath, "attributes")),
		downloadItems(client, cookbook.Recipes, "recipes", path.Join(cookbookPath, "recipes")),
		downloadItems(client, cookbook.Definitions, "definitions", path.Join(cookbookPath, "definitions")),
		downloadItems(client, cookbook.Libraries, "libraries", path.Join(cookbookPath, "libraries")),
		downloadItems(client, cookbook.Providers, "providers", path.Join(cookbookPath, "providers")),
		downloadItems(client, cookbook.Resources, "resources", path.Join(cookbookPath, "resources")),
	}

	for _, err := range downloadsErr {
		if err != nil {
			return err
		}
	}

	// TODO @afiun how do we download spec/ and test/
	fmt.Printf("Cookbook downloaded to %s\n", cookbookPath)
	return nil
}

func downloadItems(client *chef.Client, items []chef.CookbookItem, itemType, localPath string) error {
	if len(items) == 0 {
		return nil
	}

	fmt.Printf("Downloading %s\n", itemType)
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return err
	}

	for _, item := range items {
		itemPath := path.Join(localPath, item.Name)
		if err := downloadFile(client, item.Url, itemPath); err != nil {
			return err
		}
	}

	return nil
}

func downloadFile(client *chef.Client, url, file string) error {
	request, err := client.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request, nil)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}

	io.Copy(f, response.Body)
	return nil
}
