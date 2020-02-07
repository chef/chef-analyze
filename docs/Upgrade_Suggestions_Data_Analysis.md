# Upgrade Suggestion (Data Analysis)

Upgrading old Chef infrastructure could be very challenging, in former times,
users would upload cookbooks and register Chef nodes (bootstrap) to a Chef
Infra Server, such nodes would have a defined run-list, that is, the list of
policies (cookbooks) applied to the node. The run-list could also be modified
by roles and environments, giving the user a lot of flexibility, but with a
very high level of complexity.

As an example, let us think about a user with X number of nodes running Chef
11, in order to upgrade these nodes to Chef 12 a series of things needs to
happen:

* Cookbooks need to be updated and tested to run on Chef 12
* The Chef client version on the node(s) needs to be upgraded
* (in some cases) A Chef Infra Server upgrade could be required

This problem gets bigger as the number of nodes and cookbooks increases. There
could be cookbooks that are being used by more than one node (a very common
thing to do), and during an upgrade on a single node, without knowing, the user
could break other nodes accidentally.

Through the analysis of the Chef Infra Server data, we could give our users
a set of suggestions to point them to where they could start taking immediate
actions to upgrade their infrastructure to the latest version of our tools.
By starting with the most simple tasks, users could have fast results in short
cycles, and by leaving the hard tasks at the end, through time, users will
automatically reduce the level of complexity of such tasks until the
completion of the upgrade.

## Goals
1. Help users start upgrading their infrastructure immediately
1. Reduce the risk of executing upgrades
1. Better understanding about current state of user’s infrastructure (level of complexity)
1. Induce users to migrate to new desired pattern, like Policyfiles, or even Effortless

## Motivation

    As a Chef IT Operator,
    I want to be guided through the upgrade process of my infrastructure,
    so that I can start upgrading immediately with short but fast cycles,
    and reduce the amount of work until I am completely up-to-date.

## Specification

Currently, the chef-analyze tool helps users understand the current state of
their infrastructure. It focuses on two main reports, a nodes and a cookbooks
report.

* Nodes report: helps the user understand the list of nodes and their Chef
Client version that the nodes are running, as well as the expanded run-list
applied directly to the node
* Cookbooks report: verifies the compatibility of the cookbooks with the
latest version of Chef Client plus, which nodes are currently using each
cookbook

One outstanding detail for us to verify is that the current reports are
covering Chef Environments and Chef Roles correctly.

The tool currently is not aware of Policyfiles and Policygroups, for the nodes
report we need to add a column that tell the user that a node has a policy
attached, and for the cookbooks report we need to, in addition to gather the
cookbooks from the Chef Infra Server, fetch the list of policies that are being
use by one or more nodes and resolve the cookbooks listed it on each policy.

To implement the above statement, it will be neccesary to change the way we
collect the data. We will start storing the data locally so that we can
leverage it from other parts of the tool. For instance, after the generation
of a nodes report, we would have on disk the information needed for the
cookbooks report that requires the verification of the policies that are being
used by one or more nodes.

A few implications of this approach are:
* When and how do we reload the data stored on disk?
* Do we need to add a process to collect all the data at once?

Once we are generating accurate reports, the next step would be to analyze the
data gathered and give upgrade suggestions to the user. We will provide ranked
suggestions based on measured impact to the infrastructure, preferring least-impactful
changes so that the operator can move forward with some assurance of safety.

The recommended usage of these suggestions is to apply and deploy them iteratively.

A few suggestions we can make to our users are:

1. **List of unused cookbooks that can be deleted**: for IT Operators, cleaning
up their environments should be a very common task. We should be able to give
them a list of cookbooks that aren’t used by any node, so that they can safely
remove them from the Chef Infra Server
1. **The easiest node to start upgrading**: this suggestion should include a
list of steps to follow to complete the upgrade in the order that the user must
execute them. The node must be isolated so that there is no impact with other
nodes in their infrastructure
1. **A grouped list of nodes that look similar or identical**: this would be
beneficial for the user to understand batches of nodes that can be upgraded
together, once the user selects a pattern of nodes, we should be able to give
instructions to upgrade them in the correct order. Depending on the complexity
of this upgrade, users might have to create environments and move nodes to the
upgraded environment as appropriate
1. **A list of actions and impact from a node on demand**: our users might know
better where do they want to start upgrading their infrastructure, and so, we
should provide a way for them to select a node that they would like to upgrade
and we will give a recommendation of the steps to follow to do the upgrade, as
well as any impact that they might have while they execute the recommendations.

## Downstream Impact

We do not consider to have any downstream impact since this is a new tool.
