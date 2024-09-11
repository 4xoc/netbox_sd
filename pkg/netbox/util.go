package netbox

import (
	"fmt"

	"github.com/Masterminds/semver"
)

// netboxIsCompatible returns true when a version string of Netbox is supported.
func netboxIsCompatible(version string) bool {
	var (
		compatibleVersion *semver.Constraints
		givenVersion      *semver.Version
		err               error
	)

	compatibleVersion, err = semver.NewConstraint(compatibleNetboxVersion)
	if err != nil {
		panic(fmt.Sprintf("could not parse Netbox version constraint '%s'", compatibleNetboxVersion))
	}

	givenVersion, err = semver.NewVersion(version)
	if err != nil {
		panic(fmt.Sprintf("could not parse given Netbox version '%s'", version))
	}

	return compatibleVersion.Check(givenVersion)
}
