# ClusterOperator Custom Resource

The ClusterOperator is a custom resource object which holds the current state of an operator. This object is used by operators to convey their state to the rest of the cluster.

Ref: [godoc](https://godoc.org/github.com/openshift/api/config/v1#ClusterOperator) for more info on the ClusterOperator type.

## Why I want ClusterOperator Custom Resource in /manifests

ClusterVersionOperator sweeps the release payload and applies it to the cluster. So if your operator manages a critical component for the cluster and you want ClusterVersionOperator to wait for your operator to **complete** before it continues to apply other operators, you must include the ClusterOperator Custom Resource in [`/manifests`](operators.md#what-do-i-put-in-manifests).

## How should I include ClusterOperator Custom Resource in /manifests

### How ClusterVersionOperator handles ClusterOperator in release payload

When ClusterVersionOperator encounters a ClusterOperator Custom Resource,

- It uses the `.metadata.name` and `.metadata.namespace` to find the corresponding ClusterOperator instance in the cluster
- It then waits for the instance in the cluster until
  - `.status.version` in the live instance matches the `.status.version` from the release payload and
  - the live instance `.status.conditions` report available, not progressing and not failed
- It then continues to the next task.

ClusterVersionOperator will only deploy files with `.yaml`, `.yml`, or `.json` extensions, like `kubectl create -f DIR`.

**NOTE**: ClusterVersionOperator sweeps the manifests in the release payload in alphabetical order, therefore if the ClusterOperator Custom Resource exists before the deployment for the operator that is supposed to report the Custom Resource, ClusterVersionOperator will be stuck waiting and cannot proceed.

### What should be the contents of ClusterOperator Custom Resource in /manifests

There are 3 important things that need to be set in the ClusterOperator Custom Resource in /manifests for CVO to correctly handle it.

- `.metadata.namespace`: namespace for finding the live instance in cluster
- `.metadata.name`: name for finding the live instance in the namespace
- `.status.version`: this is the version that the operator is expected to report. ClusterVersionOperator only respects the `.status.conditions` from instance that reports `.status.version`

Example:

For a cluster operator `my-cluster-operator` applying version `1.0.0`, that is reporting its status using ClusterOperator instance `my-cluster-operator` in namespace `my-cluster-operator-namespace`.

The ClusterOperator Custom Resource in /manifests should look like,

```yaml
apiVersion: operatorstatus.openshift.io
kind: ClusterOperator
metadata:
  namespace: my-cluster-operator-namespace
  name: my-cluster-operator
status:
  version: 1.0.0
```

## What should an operator report with ClusterOperator Custom Resource

### Status

The operator should ensure that all the fields of `.status` in ClusterOperator are atomic changes. This means that all the fields in the `.status` are only valid together and do not partially represent the status of the operator.

### Version

The operator should report a version which indicates the components that it is applying to the cluster.

### Conditions

Refer [the godocs](https://godoc.org/github.com/openshift/api/config/v1#ClusterStatusConditionType) for conditions.

In general, ClusterOperators should contain at least three core conditions:

* `Progressing` which is true if the controller has detected the state is invalid and is working towards it (or if the controller has never reconciled before)
* `Available` should be true if the controller has successfully reached a working status and no known errors are blocking progress.
* `Failing` should be true if the controller is blocked from reaching the desired state.

The message reported for each of these conditions is important.  All messages should start with a capital letter (like a sentence) and be written for an end user / admin to debug the problem.  `Failing` should describe in detail (a few sentences at most) why the current controller is blocked. The detail should be sufficient for an engineer or support person to triage the problem. `Available` should convey useful information about what is available, and be a single sentence without punctuation.  `Progressing` is the most important message because it is shown by default in the CLI as a column and should be a terse, human-readable message describing the current state of the object in 5-10 words (the more succinct the better).

For instance, if the CVO is working towards 4.0.1 and has already successfully deployed 4.0.0, the conditions might be reporting:

* `Failing` is false with no message
* `Available` is true with message `Cluster has deployed 4.0.0`
* `Progressing` is true with message `Working towards 4.0.1`

If the controller reaches 4.0.1, the conditions might be:

* `Failing` is false with no message
* `Available` is true with message `Cluster has deployed 4.0.1`
* `Progressing` is false with message `Cluster version is 4.0.1`

If an error blocks reaching 4.0.1, the conditions might be:

* `Failing` is true with a detailed message `Unable to apply 4.0.1: could not update 0000_70_network_deployment.yaml because the resource type NetworkConfig has not been installed on the server.`
* `Available` is false with message `Cluster is unable to reach 4.0.1`
* `Progressing` is true with message `Unable to apply 4.0.1: a required object is missing`

The progressing message is the first message a human will see when debugging an issue, so it should be terse, succinct, and summarize the problem well.  The failing message can be more verbose. Start with simple, easy to understand messages and grow them over time to capture more detail.