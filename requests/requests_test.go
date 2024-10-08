package requests_test

import (
	"github.com/MaanasSathaye/swiss/requests"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Requests", func() {
	Describe("NewConstantRequest", func() {
		It("should generate a constant size request", func() {
			url := "https://localhost:4001"
			req := requests.NewConstantRequest(url)
			Expect(req).NotTo(BeNil()) // Check if request is not nil

		})
	})

	Describe("NewVariableRequest", func() {
		It("should generate requests with varying sizes", func() {
			var (
				sizes []int64
			)
			url := "https://localhost:4001"

			for i := 0; i < 100; i++ {
				req := requests.NewVariableRequest(url)
				Expect(req).NotTo(BeNil())
				sizes = append(sizes, req.ContentLength)
			}

			Expect(sizes).To(HaveLen(100))
		})
	})
})
