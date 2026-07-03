package llm

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	nameMinLength = 1
	nameMaxLength = 63
)

func validateName(name string) error {
	length := utf8.RuneCountInString(name)
	if length < nameMinLength || length > nameMaxLength {
		return fmt.Errorf("invalid endpoint name %q: name must be between %d and %d characters", name, nameMinLength, nameMaxLength)
	}
	return nil
}

func validateModel(model string) error {
	if !IsAllowedModel(model) {
		return fmt.Errorf("invalid model %q: accepted values are %s", model, strings.Join(AllowedModelList, ", "))
	}
	return nil
}
