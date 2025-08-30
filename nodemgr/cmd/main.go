package main

import (
	"log"
	"time"

	"nodemgr/internal/adapter"
	"nodemgr/internal/adapter/provision"
	"nodemgr/internal/core/domain"
	"nodemgr/internal/core/port"
	"nodemgr/internal/core/service"
)

func main() {
	templateRepo := adapter.NewRepository[domain.TemplateID, domain.NodeTemplate]()
	templateRepo.Create(domain.NodeTemplate{
		TemplateID: "ubuntu-worker-small",
		Name:       "ubuntu-worker-small",
		CPUs:       2,
		MemoryMB:   256,
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

	mappingRepo := adapter.NewRepository[domain.MappingID, domain.TemplateMapping]()
	mappingRepo.Create(domain.TemplateMapping{
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

	templateService := service.NewTemplateService(templateRepo, mappingRepo)
	spec, err := templateService.RenderTemplate("ubuntu-worker-small", "docker")
	if err != nil {
		log.Fatalf("failed to render template: %v", err)
	}

	log.Printf("Rendered Spec: %+v", spec)

	var provider port.NodeProvider = provision.NewDockerProvider()

	node, err := provider.Provision(spec)
	if err != nil {
		log.Fatalf("failed to provision node: %v", err)
	}

	log.Println("Provisioned Node:")
	log.Printf("  ID: %s\n", node.ID())
	log.Printf("  ProviderID (container ID): %s\n", node.ProviderID)
	log.Printf("  Status: %s\n", node.Status)
	log.Printf("  Meta: %+v\n", node.Meta)

	// // Try to get a lifecycle controller
	ctrl, err := provider.Controller(&node)
	if err != nil {
		log.Println("Controller not available:", err)
	} else {
		// Calls will error with "lifecycle.* not supported"
		if err := ctrl.Start(); err != nil {
			log.Println("Start error:", err)
		}
	}

	time.Sleep(5 * time.Second)

	// // Destroy the container (removes the Pulumi stack and the container)
	if err := provider.Destroy(node.ID()); err != nil {
		log.Fatalf("failed to destroy node: %v", err)
	}
	log.Println("Node destroyed successfully")
}
