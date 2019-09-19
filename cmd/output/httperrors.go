package output

import (
	"fmt"
	"git.f-i-ts.de/cloud-native/cloudctl/api/models"
)

// PrintHTTPError prints an HTTP error
func PrintHTTPError(err *models.HttperrorsHTTPErrorResponse) {
	fmt.Printf("An http error has occurred (status code: %d): %s\n", *err.Statuscode, *err.Message)
}

// PrintUnconventionalError prints an unconventional error
func PrintUnconventionalError(err error) {
	fmt.Printf("An unexpected error has occurred: %v\n", err)
}
