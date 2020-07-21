package output

import (
	"fmt"

	"github.com/fi-ts/cloud-go/api/models"
)

// HTTPError prints an HTTP error
func HTTPError(err *models.HttperrorsHTTPErrorResponse) error {
	return fmt.Errorf("An http error has occurred (status code: %d): %s\n", *err.Statuscode, *err.Message)
}

// UnconventionalError prints an unconventional error
func UnconventionalError(err error) error {
	return fmt.Errorf("An unexpected error has occurred: %v\n", err)
}
