package roundrobin_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRoundrobin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Roundrobin Suite")
}
