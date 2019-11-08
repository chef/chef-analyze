package main

import (
	"fmt"

	"github.com/chef/chef-analyze/pkg/featflag"
)

func main() {
	fmt.Println("Features:")

	// read global features that comes out-of-the-box from the featflag libary
	fmt.Println(" *", featflag.ChefFeatAll.String(), featflag.ChefFeatAll.Enabled())
	fmt.Println(" *", featflag.ChefFeatAnalyze.String(), featflag.ChefFeatAnalyze.Enabled())

	// defining a new local feature flag
	//
	// example: New Preferences Dialog Feature Flag
	//  * environment variable:
	//      CHEF_FEAT_PREFERENCES_DIALOG=1
	//  * config key:
	//      [features]
	//      preferences_dialog = true
	chefFeatPreferencesDialog := featflag.New("CHEF_FEAT_PREFERENCES_DIALOG", "preferences_dialog")
	fmt.Println(" *", chefFeatPreferencesDialog.String(), chefFeatPreferencesDialog.Enabled())
}
