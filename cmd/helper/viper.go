package helper

import "github.com/spf13/viper"

func AtLeastOneViperStringFlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if viper.GetString(flag) != "" {
			return true
		}
	}
	return false
}

func AtLeastOneViperStringSliceFlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if len(viper.GetStringSlice(flag)) > 0 {
			return true
		}
	}
	return false
}

func AtLeastOneViperBoolFlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if viper.GetBool(flag) {
			return true
		}
	}
	return false
}

func AtLeastOneViperInt64FlagGiven(flags ...string) bool {
	for _, flag := range flags {
		if viper.GetInt64(flag) != 0 {
			return true
		}
	}
	return false
}

func ViperString(flag string) *string {
	if viper.GetString(flag) == "" {
		return nil
	}
	value := viper.GetString(flag)
	return &value
}

func ViperStringSlice(flag string) []string {
	value := viper.GetStringSlice(flag)
	if len(value) == 0 {
		return nil
	}
	return value
}

func ViperBool(flag string) *bool {
	if !viper.GetBool(flag) {
		return nil
	}
	value := viper.GetBool(flag)
	return &value
}

func ViperInt64(flag string) *int64 {
	if viper.GetInt64(flag) == 0 {
		return nil
	}
	value := viper.GetInt64(flag)
	return &value
}
