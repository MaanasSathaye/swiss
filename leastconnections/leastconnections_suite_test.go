package leastconnections_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLeastconnections(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Leastconnections Suite")
}
