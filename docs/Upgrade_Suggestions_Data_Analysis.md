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

1. Cookbooks need to be updated and tested to run on Chef 12
1. The Chef client version on the node(s) needs to be upgraded
1. (in some cases) A Chef Infra Server upgrade could be required

This problem gets bigger as the number of nodes and cookbooks increases. There
could be cookbooks that are being used by more than one node (a very common
thing to do), and during an upgrade on a single node, without knowing, the user
could break other nodes accidentally.

Through the analysis of the Chef Infra Server data, we could give our users
a set of suggestions to point them to where they could start taking immediate
actions to upgrade their infrastructure to the latest version of our tools.
By starting with the most simple tasks, users could have fast results in short
cycles, and by leaving the hard tasks at the end, though time, users will
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
Client version that the nodes are running, as well as the expanded run-listxi
applied
* Cookbooks report: verifies the compatibility of the cookbooks with the
latest version of Chef Client plus, which nodes are currently using eachxi
cookbook

The next step for us would be to analyze the data gathered from the reports
and give upgrade suggestions to the user. Guiding them with a few actions
they can take to start upgrading their infrastructure. We are aiming to
start with the smallest tasks that the user can take, so that they start
immediately, and to reduce the upcoming tasks ahead.

## Downstream Impact

TBA
