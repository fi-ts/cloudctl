package cmd

import (
	"fmt"
	"github.com/metal-stack/metal-lib/auth"
)

const cloudContext = "cloudctl"

// getAuthContext reads AuthContext from given kubeconfig
func getAuthContext(kubeconfig string) (*auth.AuthContext, error) {
	cs, err := getContexts()
	if err != nil {
		return nil, err
	}
	authContext, err := auth.GetAuthContext(kubeconfig, formatContextName(cloudContext, cs.CurrentContext))
	if err != nil {
		return nil, err
	}

	if !authContext.AuthProviderOidc {
		return nil, fmt.Errorf("active user %s has no oidc authProvider, check config", authContext.User)
	}

	return &authContext, nil
}

// formatContextName returns the contextName for the given suffix. suffix can be empty.
func formatContextName(prefix string, suffix string) string {
	contextName := prefix
	if suffix != "" {
		contextName = fmt.Sprintf("%s-%s", cloudContext, suffix)
	}
	return contextName
}
