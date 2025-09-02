---
title: Node manager
description: Architecture of node manager service
---

Node manager is an independent subsystem of remote make that povers the empheral worker node provisioning. Its main role is to provide dynamic computing assets like virtual machines, docker containers or AWS EC2/spot instances while being independent from any underlying provider or framework. Currently the `nodemgr` project consists of few subsystems:
- template - the role of this service is to provide universal templating schema that allows to describe compute instances across diffrent providers
- mapping - this service handles mapping of provider specific resources like generic image names like `ubuntu` into iso paths, docker images or AWS amis
- schedule - role of this service is to decide which template and provider best suits the policy
- provision - this system provisions and destroys worker nodes while keeping track of already existing deployments
- lifecycle - this is optional subsystem which allows to control the individuall node lifecycle through external means like docker api or AWS EC2 api (useful for dev nodes)
- execute - this service providers capabilities of executing commands, transfering files and changing the lifecycle state directly from worker node. Most instances implement this through SSH
- orchestrate - this service allows to run the whole workflow of automatic deployment by orchestrating the runtime of other services

## Templates
Currently templates are the generic descriptors of coputing hardware. Currently they provide a few common fields that should be respected across all providers:
```go
Name  string
Image string
User  string

CPUs     int // vCPUs
MemoryMB int
DiskMB   int
```
the template also includes the special field reserved for provider ovverides:
```go
ProviderOverrides map[ProviderID]map[string]any
```
this for example may provide additional fields like network interface for the libvirt provider. At last both common fields and ovverides are combined into one common hashmap called `Extra`:
```go
Extra             map[string]any
```
This field contains properties directly consumed by compute providers and is provider dependent. For example passing `ami` or `ami_lookup` fields in extra will only affect the behavior of AWS provider. To allow for easier setup templates can also be loaded from file:
```yml
name: ubuntu-worker-small
cpu: 2
memoryMB: 256
image: ubuntu:24.04
user: ubuntu
overrides:
  docker:
    env: ["FOO=bar"]
  libvirt:
    memoryMB: 1024
    diskGB: 3
    network: default
```

## Mapppers
Because certain resources like images do not share common names across the providers but encompase the same resource. Mappers match the avalible extra fileds using eighter `exact` or `glob` targets and replace them according to provider requirments. For example to support `ubuntu` generic name across diffrent providers this mapper is used:
```yml
mapping:
  - match:
      image: "ubuntu:24.04"
      match_type: "glob" # glob | exact
    overrides:
      all:
        user: "ubuntu"
      aws:
        # TODO: this is bas idea as AWS ami are region and arch dependent
        ami: "ami-0a116fa7c861dd5f9"
        image_type: "ami"
      libvirt:
        # TODO: fix as this does not take the arch of the server into the account
        # TODO: support auto download for known distros
        iso: "ubuntu-24.04.3-live-server-amd64.iso"
        image_type: "iso"
      docker:
        image: "ubuntu:24.04"    # keep same
```

## Schedulers
TODO: Unimplemented feature

## Provisioners
Provisioners are the most basic adaapters that provide infrastructure capabilities. They implement two major functions `Provision()` and `Destroy()` which are used to construct new resources. Most of the providers wrap around the Pulumi library or Terraform cli to make this process easier but this approach has some limitations. IAC does not care about resources between their creation and destruction thus lifecycle API is exposed to partially mitigate this problem. There is also dummy provider for local execution which always returns the same node populated with the data of the host machine. Currently avalible are these providers:
- local (insecure, use only for testing)
- pulumi based:
  - docker
  - libvirt
  - AWS

## Lifecycle
Lifecycle api is modeled after the EC2 lifecycle diagram and optionally extends the capabilities of existing provsioners by hooking into provider specific api:
![EC2 lifecycle](https://docs.aws.amazon.com/images/AWSEC2/latest/UserGuide/images/instance_lifecycle.png)
this api is used to provide finer control over the specific node besides provisioning and destroying methods. This is usefull for targets like docker and libvirt where most often inactive instances do not generate costs thus stopping/hibernating instance instead of compleatly scrapping it every time is beneficial. These controls may also be used to provide simple reemote dev enviorment where instance can be power on or off remotly to save up resources.

## Executors
Executers are simple adapters which allow for both code execution and file transfer. For example local executor can only consume nodes with capability `local:exec` and runs over the golang process api. Most of the nodes will utilize remote access apis such as ssh, sftp or rsync but their are also other interfaces. Currently these executors are supported:
- local
- docker exec
- ssh:
  - sftp
  - rsync

## Orchestrator
TODO: Unimplemented feature
