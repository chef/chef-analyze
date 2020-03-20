package reporting

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	chef "github.com/chef/go-chef"
	"github.com/pkg/errors"
)

type NodeCapture struct {
	Progress      chan int
	Error         error
	name          string
	repositoryDir string
	nodes         NodesInterface
	roles         RolesInterface
	env           EnvironmentInterface
	cookbooks     CookbookInterface
}

// Events that we publish on the Progress channel when
// performing a node capture
const (
	FetchingNode = iota
	FetchingEnvironment
	FetchingRoles
	FetchingCookbooks
	FetchingComplete
)

// Initialize a NodeCapture. Creates a Progress channel that callers can monitor for updates
func NewNodeCapture(nodeName string, repositoryDir string, nodes NodesInterface,
	roles RolesInterface, env EnvironmentInterface, cookbooks CookbookInterface) *NodeCapture {
	return &NodeCapture{name: nodeName, nodes: nodes, roles: roles,
		env: env, cookbooks: cookbooks, repositoryDir: repositoryDir,
		// 5 possible events in a Run - let's not block our activity in case the caller
		// doesn't pick them up
		Progress: make(chan int, 5)}
}

func (nc *NodeCapture) Run() {
	defer func() { close(nc.Progress) }()

	nc.Progress <- FetchingNode
	node, err := nc.captureNodeObject()
	if err != nil {
		nc.Error = errors.Wrapf(err, "unable to capture node '%s'", nc.name)
		return
	}

	nc.Progress <- FetchingCookbooks
	err = nc.captureCookbookObjects(node.AutomaticAttributes["cookbooks"].(map[string]interface{}))
	if err != nil {
		nc.Error = errors.Wrapf(err, "unable to capture node cookbooks for '%s'", nc.name)
		return
	}

	nc.Progress <- FetchingEnvironment
	err = nc.captureEnvObject(node)
	if err != nil {
		nc.Error = errors.Wrapf(err, "unable to capture enviroment")
		return
	}

	nc.Progress <- FetchingRoles
	err = nc.captureRoleObjects(node)
	err = nc.captureEnvObject(node)
	if err != nil {
		nc.Error = errors.Wrapf(err, "unable to capture role(s)")
		return
	}

	nc.Progress <- FetchingComplete
}
func (nc *NodeCapture) captureNodeObject() (*chef.Node, error) {
	node, err := nc.nodes.Get(nc.name)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to retrieve node: %s", nc.name)
	}
	err = saveObjectAsJSON(nc.repositoryDir, "nodes", node.Name, node)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save node. This is a bug, please report it.")
	}

	// In this pass we are not supporting policyfiles. For now, that
	// means we'll raise an errro when the node has policyfile enabled,
	// to avoid the risk of reporting back inaccurate results based on non-PF assumptions.
	if node.PolicyName != "" || node.PolicyGroup != "" {
		return nil, errors.New(fmt.Sprintf("Node %s is managed by Policyfile. Unsupported at this time.", nc.name))
	}
	return &node, nil
}

func (nc *NodeCapture) captureCookbookObjects(cookbooks map[string]interface{}) error {
	// Cookbooks from Node object
	cookbookDir := fmt.Sprintf("%s/cookbooks", nc.repositoryDir)
	for name, version_data := range cookbooks {
		version := safeStringFromMap(version_data.(map[string]interface{}), "version")
		// TODO - this will come back with version-named cookbook dirs. we'll want to make that optional
		// in the underlying interface, and/or clean up after the fact.
		err := nc.cookbooks.DownloadTo(name, version, cookbookDir)
		if err != nil {
			return errors.Wrapf(err, "Failed to download cookbook %s v%s", name, version)
		}
		// The cookbook directory will have been created as 'name-version'.  We need it to be just 'name'
		// so that it can be picked up by chef-zero when requested from within kitchen.

		err = os.Rename(fmt.Sprintf("%s/%s-%s", cookbookDir, name, version), fmt.Sprintf("%s/%s", cookbookDir, name))
		if err != nil {
			return errors.Wrapf(err, "failed to rename cookbook %s-%s to non-versioned name", name, version)
		}
	}
	return nil
}

func (nc *NodeCapture) captureEnvObject(node *chef.Node) error {
	env, err := nc.env.Get(node.Environment)
	if err != nil {
		return errors.Wrapf(err, "unable to retrieve node environment: %s", node.Environment)
	}

	err = saveObjectAsJSON(nc.repositoryDir, "environments", node.Environment, env)
	if err != nil {
		return errors.Wrap(err, "failed to save environment. This is a bug, please report it.")
	}
	return nil
}

func (nc *NodeCapture) captureRoleObjects(node *chef.Node) error {

	roleNames := filterRoles(node.RunList)
	for _, roleName := range roleNames {
		role, err := nc.roles.Get(roleName)
		if err != nil {
			return errors.Wrapf(err, "unable to retrieve role: %s", roleName)
		}
		err = saveObjectAsJSON(nc.repositoryDir, "roles", roleName, role)
		if err != nil {
			return errors.Wrapf(err, "failed to save role %s. This is a bug, please report it:", roleName)
		}
	}
	return nil
}

func saveObjectAsJSON(rootDir string, subDir string, objName string, obj interface{}) error {
	var err error
	dirName := fmt.Sprintf("%s/%s", rootDir, subDir)
	path := fmt.Sprintf("%s/%s/%s.json", rootDir, subDir, objName)
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err = os.MkdirAll(dirName, 0700)
	}
	if err != nil {
		return errors.Wrapf(err, "Failed to create directory %s", dirName)
	}
	body, err := chef.JSONReader(obj)
	if err != nil {
		return errors.Wrapf(err, "Failed to convert %s to json. If you see this, it's a bug.  Please report it.", objName)
	}

	content, err := ioutil.ReadAll(body)
	if err != nil {
		return errors.Wrapf(err, "Failed to read %s %s", subDir, objName)
	}

	err = ioutil.WriteFile(path, content, 0600)
	if err != nil {
		return errors.Wrapf(err, "Failed to save file %s", path)
	}
	return nil
}
func filterRoles(possibleRoles []string) []string {
	roles := make([]string, 0)
	for _, role := range possibleRoles {
		pos := strings.Index(role, "role[")
		if pos == -1 {
			continue
		}
		// can't remember if chef server allows unicode in run lists,
		// so we'll play it safe and convert to runes.
		runes := []rune(role)
		stop := len(role) - 1
		roles = append(roles, string(runes[5:stop]))
	}

	return roles
}

// Node: comes from ndoe endpoint in chef server.
