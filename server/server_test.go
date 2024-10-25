package server_test

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gofrs/uuid/v5"

	"github.com/MaanasSathaye/swiss/server"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCP Server", func() {
	Context("Starting and stopping a single server", func() {
		It("should start and stop successfully", func() {
			var (
				err error
				srv *server.Server
			)
			ctx, done := context.WithTimeout(context.Background(), 15*time.Second)
			defer done()
			srv, err = server.NewServer(ctx)
			Expect(err).To(BeNil())
			Expect(srv).NotTo(BeNil())

			srv.Start(ctx)
			Expect(srv.Alive).To(BeTrue())

			time.Sleep(1 * time.Second)

			srv.Stop()
			Expect(srv.Alive).To(BeFalse())
		})
	})

	Context("Starting and stopping multiple servers", func() {
		It("should start and stop multiple servers successfully", func() {
			var (
				servers []*server.Server
				err     error
				ctx     context.Context
				srv     *server.Server
			)
			ctx, done := context.WithTimeout(context.Background(), 15*time.Second)
			defer done()
			for i := 0; i < 3; i++ {
				srv, err = server.NewServer(ctx)
				Expect(err).To(BeNil())
				Expect(srv).NotTo(BeNil())
				servers = append(servers, srv)
				srv.Start(ctx)
				Expect(srv.Alive).To(BeTrue())
			}

			time.Sleep(1 * time.Second)

			for _, srv := range servers {
				srv.Stop()
				Expect(srv.Alive).To(BeFalse())
			}
		})
	})

	Context("Starting a server and handling a connection", func() {
		It("should accept a connection and handle it", func() {
			var (
				srv *server.Server
				err error
				id  uuid.UUID
				n   int
			)
			ctx, done := context.WithTimeout(context.Background(), 15*time.Second)
			defer done()
			srv, err = server.NewServer(ctx)
			Expect(err).To(BeNil())
			Expect(srv).NotTo(BeNil())

			srv.Start(ctx)
			Expect(srv.Alive).To(BeTrue())

			conn, err := net.Dial("tcp", srv.Address())
			Expect(err).To(BeNil())
			Expect(conn).NotTo(BeNil())

			id, err = uuid.NewV4()
			Expect(err).To(BeNil())
			resp := fmt.Sprint("hello ", id)
			_, err = conn.Write([]byte(resp))
			Expect(err).To(BeNil())

			buffer := make([]byte, 1024)
			n, err = conn.Read(buffer)
			Expect(err).To(BeNil())
			Expect(string(buffer[:n])).To(ContainSubstring("Acknowledged"))

			conn.Close()
			srv.Stop()
			Expect(srv.Alive).To(BeFalse())
		})
	})

	Context("Starting multiple servers and handling multiple connections", func() {
		It("should accept multiple connections on multiple servers", func() {
			var (
				srv     *server.Server
				servers []*server.Server
				err     error
				id      uuid.UUID
				conn    net.Conn
			)
			ctx, done := context.WithTimeout(context.Background(), 15*time.Second)
			defer done()

			for i := 0; i < 2; i++ {
				srv, err = server.NewServer(ctx)
				Expect(err).To(BeNil())
				Expect(srv).NotTo(BeNil())
				servers = append(servers, srv)
				srv.Start(ctx)
				Expect(srv.Alive).To(BeTrue())
			}

			for _, srv := range servers {
				conn, err = net.Dial("tcp", srv.Address())
				Expect(err).To(BeNil())
				Expect(conn).NotTo(BeNil())

				id, err = uuid.NewV4()
				Expect(err).To(BeNil())
				resp := fmt.Sprint("hello ", id)
				_, err = conn.Write([]byte(resp))
				Expect(err).To(BeNil())

				buffer := make([]byte, 1024)
				n, err := conn.Read(buffer)
				Expect(err).To(BeNil())
				Expect(string(buffer[:n])).To(ContainSubstring("Acknowledged"))

				conn.Close()
			}

			for _, srv := range servers {
				srv.Stop()
				Expect(srv.Alive).To(BeFalse())
			}
		})
	})
})
