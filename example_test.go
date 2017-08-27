package anser

import (
	"testing"

	"github.com/mongodb/grip"
)

func TestExampleApp(t *testing.T) {
	if err := proofOfConcept(); err != nil {
		grip.Error(err)
		t.FailNow()
	}
}
