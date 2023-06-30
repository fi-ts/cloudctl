package cmd

import "testing"

func Test_getGardenerWithLatestVersion(t *testing.T) {
	tests := []struct {
		name             string
		gardenerVersions map[string]int
		wantedVersion    string
		wantedCount      float64
	}{

		{
			name: "one 100%",
			gardenerVersions: map[string]int{
				"1.1.1": 2,
			},
			wantedVersion: "1.1.1",
			wantedCount:   2,
		},
		{
			name: "simple 50%",
			gardenerVersions: map[string]int{
				"1.1.0": 2,
				"1.1.1": 2,
			},
			wantedVersion: "1.1.1",
			wantedCount:   float64(2),
		},

		{
			name: "simple 20%",
			gardenerVersions: map[string]int{
				"1.1.1": 2,
				"0.1.1": 4,
				"1.1.0": 4,
			},
			wantedVersion: "1.1.1",
			wantedCount:   float64(2),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getGardenerWithLatestVersion(tt.gardenerVersions)
			if got != tt.wantedVersion {
				t.Errorf("getGardenerPercentage() got = %v, want %v", got, tt.wantedVersion)
			}
			if got1 != tt.wantedCount {
				t.Errorf("getGardenerPercentage() got1 = %v, want %v", got1, tt.wantedCount)
			}
		})
	}
}
