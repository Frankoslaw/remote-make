package main

import (
	"log"

	"nodemgr/internal/adapter/execute"
	"nodemgr/internal/adapter/provision"
	"nodemgr/internal/core/domain"
	"nodemgr/internal/core/port"
	"nodemgr/internal/core/service"
	"nodemgr/internal/core/util"
)

func main() {
	templateRepo := util.NewRepository[domain.TemplateID, domain.NodeTemplate]()
	templateRepo.Create(domain.NodeTemplate{
		TemplateID: "ubuntu-worker-small",
		Image:      "ubuntu",
		User:       "ubuntu",
		ProviderOverrides: map[domain.ProviderID]map[string]any{
			"libvirt": {
				"memory_mb": 512,
				"disk_gb":   5,
				"network":   "default",
			},
		},
	})

	templateService := service.NewTemplateService(templateRepo)
	spec, err := templateService.RenderTemplate("ubuntu-worker-small", "docker")
	if err != nil {
		log.Fatalf("failed to render template: %v", err)
	}

	mappingRepo := util.NewRepository[domain.MappingID, domain.NodeSpecMapping]()
	mappingRepo.Create(domain.NodeSpecMapping{
		MappingID: "ubuntu",
		Match: map[string]string{
			"image": "ubuntu*",
		},
		MatchType: domain.MatchTypeGlob,
		ProviderOverrides: map[domain.ProviderID]map[string]any{
			"all": {
				"user": "ubuntu",
			},
			"docker": {
				"image":      "ubuntu:24.04",
				"image_type": "docker",
			},
			"libvirt": {
				"image":      "ubuntu-24.04.3-live-server-amd64.iso",
				"image_type": "iso",
			},
		},
	})

	mappingService := service.NewMappingService(mappingRepo)
	spec, err = mappingService.ResolveSpecAliases(spec)
	if err != nil {
		log.Fatalf("failed to resolve spec aliases: %v", err)
	}

	log.Printf("Rendered Spec: %+v", spec)

	var provider port.NodeProvider = provision.NewDockerProvider("unix:///var/run/docker.sock")

	node, err := provider.Provision(spec)
	if err != nil {
		log.Fatalf("failed to provision node: %v", err)
	}

	log.Println("Provisioned Node:")
	log.Printf("  ID: %s\n", node.ID())
	log.Printf("  ProviderID (container ID): %s\n", node.ProviderID)
	log.Printf("  State: %s\n", node.State)
	log.Printf("  Meta: %+v\n", node.Meta)

	execHandleRepo := util.NewRepository[domain.ExecHandleID, port.ExecHandle]()
	var executor port.NodeExecProvider = execute.NewDockerExecProvider(execHandleRepo)

	execHandle, err := executor.OpenExecHandle(node)
	if err != nil {
		log.Fatalf("failed to open exec handle: %v", err)
	}

	execReq := domain.ExecRequest{
		Command: []string{"sh", "-c", "time uname -a"},
	}
	execResp, err := execHandle.Exec(execReq)
	if err != nil {
		log.Fatalf("failed to execute command: %v", err)
	}

	log.Printf("Command executed with exit code %d", execResp.ExitCode)
	log.Printf("STDOUT:\n%s\n", execResp.Stdout)
	log.Printf("STDERR:\n%s\n", execResp.Stderr)

	if err := provider.Destroy(node.ID()); err != nil {
		log.Fatalf("failed to destroy node: %v", err)
	}
	log.Println("Node destroyed successfully")
}
