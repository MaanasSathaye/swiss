package requests_test

import (
	"context"
	"time"

	"github.com/MaanasSathaye/swiss/requests"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Requests", func() {
	Describe("NewConstantRequest", func() {
		It("should generate a constant size request", func() {
			ctx, done := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer done()
			req, err := requests.NewConstantRequest(ctx, "http://example.com/")
			Expect(err).To(BeNil())
			Expect(req).NotTo(BeNil()) // Check if request is not nil

		})
	})

	Describe("NewVariableRequest", func() {
		It("should generate requests with varying sizes", func() {
			var (
				sizes []int64
			)
			ctx, done := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer done()
			for i := 0; i < 100; i++ {
				req, err := requests.NewConstantRequest(ctx, "http://example.com/")
				Expect(err).To(BeNil())
				Expect(req).NotTo(BeNil())
				sizes = append(sizes, req.ContentLength)
			}

			Expect(sizes).To(HaveLen(100))
		})
	})
})
