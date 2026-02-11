package main

import (
	"context"
	"log"

	"github.com/cloudfluent/terraform-provider-label/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/cloudfluent/label",
	})
	if err != nil {
		log.Fatal(err)
	}
}
