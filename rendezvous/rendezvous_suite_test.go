package rendezvous_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRendezvous(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rendezvous Suite")
}
