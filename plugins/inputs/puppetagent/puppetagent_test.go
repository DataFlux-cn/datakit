package puppetagent

import (
	"testing"
)

const location = "/opt/puppetlabs/puppet/cache/state/last_run_summary.yaml"

func TestMain(t *testing.T) {
	testAssert = true

	var puppet = PuppetAgent{
		Location: location,
		Tags: map[string]string{
			"tags1": "value1",
		},
	}

	puppet.Run()
}
