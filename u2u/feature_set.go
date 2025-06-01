package u2u

// FeatureSet is an enumeration of different releases, each one enabling a
// different set of features. These are an abstraction that allows reason
// about the different releases instead of isolated upgrades.
// Feature sets are exclusive, and a (fake-)net cannot be configured with
// more than one feature set at a time.
type FeatureSet int

const (
	VitriolFeatures FeatureSet = iota // < enables the Vitriol hardfork features
)

// ToUpgrades returns the Upgrades that are enabled by the feature set.
// If called from an unknown feature set, it will return the pre-Vitriol upgrades.
func (fs FeatureSet) ToUpgrades() Upgrades {
	vitriol := fs == VitriolFeatures
	res := Upgrades{
		Berlin:  true,
		London:  true,
		Llr:     true,
		Vitriol: vitriol,
	}
	return res
}

func (fs FeatureSet) String() string {
	switch fs {
	case VitriolFeatures:
		return "vitriol"
	default:
		return "llr"
	}
}
