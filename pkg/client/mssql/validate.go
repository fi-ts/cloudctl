package mssql

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
		return fmt.Errorf("invalid database name %q: name must be between %d and %d characters", name, nameMinLength, nameMaxLength)
	}
	return nil
}

func validateVersion(version string) error {
	if !IsAllowedVersion(version) {
		return fmt.Errorf("invalid version %q: accepted values are %s", version, strings.Join(AllowedVersions, ", "))
	}
	return nil
}

func validateStorageGB(storageGB int) error {
	if storageGB < 1 {
		return fmt.Errorf("invalid storage_gb %d: must be at least 1", storageGB)
	}
	if storageGB > 4096 {
		return fmt.Errorf("invalid storage_gb %d: must not exceed 4096", storageGB)
	}
	return nil
}
