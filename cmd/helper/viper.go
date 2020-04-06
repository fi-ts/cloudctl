package helper

import "github.com/spf13/viper"

// AtLeastOneViperStringFlagGiven ensure at least one string flag is given
func AtLeastOneViperStringFlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if viper.GetString(flag) != "" {
			return true
		}
	}
	return false
}

// AtLeastOneViperStringSliceFlagGiven ensure at least one string slice flag is given
func AtLeastOneViperStringSliceFlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if len(viper.GetStringSlice(flag)) > 0 {
			return true
		}
	}
	return false
}

// AtLeastOneViperBoolFlagGiven ensure at least one bool flag is given
func AtLeastOneViperBoolFlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if viper.GetBool(flag) {
			return true
		}
	}
	return false
}

// AtLeastOneViperInt64FlagGiven ensure at least one int64 flag is given
func AtLeastOneViperInt64FlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if viper.GetInt64(flag) != 0 {
			return true
		}
	}
	return false
}

// ViperString returns the string pointer for the given flag
func ViperString(flag string) *string {
	if viper.GetString(flag) == "" {
		return nil
	}
	value := viper.GetString(flag)
	return &value
}

// ViperStringSlice returns the string slice for the given flag
func ViperStringSlice(flag string) []string {
	value := viper.GetStringSlice(flag)
	if len(value) == 0 {
		return nil
	}
	return value
}

// ViperBool returns the bool pointer for the given flag
func ViperBool(flag string) *bool {
	if !viper.GetBool(flag) {
		return nil
	}
	value := viper.GetBool(flag)
	return &value
}

// ViperInt64 returns the int64 pointer for the given flag
func ViperInt64(flag string) *int64 {
	if viper.GetInt64(flag) == 0 {
		return nil
	}
	value := viper.GetInt64(flag)
	return &value
}
