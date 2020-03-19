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
	name          string
	repositoryDir string
	nodes         NodesInterface
	roles         RolesInterface
	env           EnvironmentInterface
	cookbooks     CookbookInterface
	Progress      chan int
	Error         error
	// Maybe workers
}

const (
	FetchingNode = iota
	FetchingEnvironment
	FetchingRoles
	FetchingCookbooks
	FetchingComplete
)

type CapturedNode struct {
	Name      string
	Cookbooks map[string]interface{}

	Node  chef.Node
	Env   *chef.Environment
	Roles []*chef.Role
}

// Initialize a NodeCapture
func NewNodeCapture(nodeName string, repositoryDir string, nodes NodesInterface,
	roles RolesInterface, env EnvironmentInterface, cookbooks CookbookInterface) *NodeCapture {
	return &NodeCapture{name: nodeName, nodes: nodes, roles: roles,
		env: env, cookbooks: cookbooks, repositoryDir: repositoryDir,
		Progress: make(chan int)}
}

func (nc *NodeCapture) Run() {
	defer func() {
		nc.Progress <- FetchingComplete
		close(nc.Progress)
	}()

	nc.Progress <- FetchingNode
	node, err := nc.nodes.Get(nc.name)
	if err != nil {
		nc.Error = errors.Wrapf(err, "unable to retrieve node: %s", nc.name)
		return
	}
	err = saveObjectAsJSON(nc.repositoryDir, "nodes", node.Name, node)
	if err != nil {
		nc.Error = errors.Wrap(err, "failed to save node. This is a bug, please report it.")
		return
	}

	// In this pass we are not supporting policyfiles. For now, that
	// means we'll raise an errro when the node has policyfile enabled,
	// to avoid the risk of reporting back inaccurate results based on non-PF assumptions.
	if node.PolicyName != "" || node.PolicyGroup != "" {
		nc.Error = errors.New(fmt.Sprintf("Node %s is managed by Policyfile. Unsupported at this time.", nc.name))
		return
	}
	// Cookbooks from Node object
	cookbooks := node.AutomaticAttributes["cookbooks"].(map[string]interface{})

	nc.Progress <- FetchingCookbooks
	for name, version_data := range cookbooks {
		version := safeStringFromMap(version_data.(map[string]interface{}), "version")
		// TODO - this will come back with version-named cookbook dirs. we'll want to make that optional
		// in the underlying interface, and/or clean up after the fact.
		err = nc.cookbooks.DownloadTo(name, version, fmt.Sprintf("%s/cookbooks", nc.repositoryDir))
		if err != nil {
			nc.Error = errors.Wrapf(err, "Failed to download cookbook %s v%s", name, version)
			return
		}
	}

	// Environment
	nc.Progress <- FetchingEnvironment
	env, err := nc.env.Get(node.Environment)
	if err != nil {
		nc.Error = errors.Wrapf(err, "unable to retrieve node environment: %s", node.Environment)
		return
	}

	err = saveObjectAsJSON(nc.repositoryDir, "environments", node.Environment, env)
	if err != nil {
		nc.Error = errors.Wrap(err, "failed to save environment. This is a bug, please report it.")
		return
	}

	// Roles
	nc.Progress <- FetchingRoles
	roleNames := filterRoles(node.RunList)
	for _, roleName := range roleNames {
		role, err := nc.roles.Get(roleName)
		if err != nil {
			nc.Error = errors.Wrapf(err, "unable to retrieve role: %s", roleName)
			return
		}
		err = saveObjectAsJSON(nc.repositoryDir, "roles", roleName, role)
		if err != nil {
			nc.Error = errors.Wrapf(err, "failed to save role %s. This is a bug, please report it:", roleName)
			return
		}
	}

	// return &CapturedNode{Name: node.Name, Node: node, Cookbooks: cookbooks, Env: env, Roles: roles}, nil
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
