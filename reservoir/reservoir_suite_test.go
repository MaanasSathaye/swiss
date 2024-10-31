package reservoir_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestReservoir(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Reservoir Suite")
}
