package server_test

import (
	"net"
	"strconv"

	"github.com/MaanasSathaye/swiss/server"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetFreePort", func() {
	It("should return a free port", func() {
		port, err := server.GetFreePort()
		Expect(err).To(BeNil())
		Expect(port).Should(BeNumerically(">", 0))

		// Try to listen on the port
		l, err := net.Listen("tcp", "localhost"+":"+strconv.Itoa(port))
		Expect(err).To(BeNil())
		defer l.Close()
	})
})
