// Package s3 implements a fake s3 server for rclone
package s3

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/JankariTech/gofakes3"
	"github.com/go-chi/chi/v5"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/hash"
	httplib "github.com/rclone/rclone/lib/http"
)

// Options contains options for the http Server
type Options struct {
	//TODO add more options
	pathBucketMode bool
	hashName       string
	hashType       hash.Type
	authPair       []string
	noCleanup      bool
	HTTP           httplib.Config
}

// Server is a s3.FileSystem interface
type Server struct {
	*httplib.Server
	f       fs.Fs
	faker   *gofakes3.GoFakeS3
	handler http.Handler
	ctx     context.Context // for global config
}

// Make a new S3 Server to serve the remote
func newServer(ctx context.Context, f fs.Fs, opt *Options) (s *Server, err error) {
	w := &Server{
		f:   f,
		ctx: ctx,
	}

	var newLogger logger
	w.faker = gofakes3.New(
		newBackend(opt, w),
		gofakes3.WithHostBucket(!opt.pathBucketMode),
		gofakes3.WithLogger(newLogger),
		gofakes3.WithRequestID(rand.Uint64()),
		gofakes3.WithoutVersioning(),
		gofakes3.WithV4Auth(authlistResolver(opt.authPair)),
	)

	w.Server, err = httplib.NewServer(ctx,
		httplib.WithConfig(opt.HTTP),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to init server: %w", err)
	}

	w.handler = w.faker.Server()
	return w, nil
}

// Bind register the handler to http.Router
func (w *Server) Bind(router chi.Router) {
	router.Handle("/*", w.handler)
}
