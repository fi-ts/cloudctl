package main

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"git.f-i-ts.de/cloud-native/cloudctl/cmd"
)

func main() {
	cmd.Execute()
}
