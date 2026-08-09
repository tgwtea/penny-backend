package buildinfo

import "os"

// Values are populated with -ldflags by immutable release builds.
var (
	version         = "development"
	commit          = "unknown"
	contractVersion = "1.0.0"
)

// Info is non-secret release provenance exposed by health responses and logs.
type Info struct {
	Version          string
	Commit           string
	ArtifactDigest   string
	ContractVersion  string
	ProductMilestone string
}

// Current returns the process build metadata. The image digest is supplied at
// deployment time because embedding an OCI image's own digest is circular.
func Current() Info {
	return Info{
		Version:          version,
		Commit:           commit,
		ArtifactDigest:   valueOrDefault("PENNY_ARTIFACT_DIGEST", "unknown"),
		ContractVersion:  contractVersion,
		ProductMilestone: valueOrDefault("PENNY_PRODUCT_MILESTONE", "development"),
	}
}

func valueOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
