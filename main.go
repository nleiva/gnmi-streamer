/*
Based on: https://github.com/openconfig/gnmi/tree/master/subscribe
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	log "github.com/golang/glog"
	"github.com/openconfig/gnmi/cache"
	"github.com/openconfig/gnmi/client"
	"github.com/openconfig/gnmi/subscribe"
	"github.com/openconfig/gnmi/testing/fake/testing/grpc/config"
	"github.com/openconfig/gnmi/value"
	"github.com/openconfig/gnmic/pkg/api/path"
	"google.golang.org/grpc"

	pb "github.com/openconfig/gnmi/proto/gnmi"
	"go.uber.org/automaxprocs/maxprocs"
)

const (
	HOST    = ""
	PORT    = "9339"
	CADENCE = 5
	UPDATES = "updates.json"
)

func startServer(ctx context.Context, c *cache.Cache, opts ...subscribe.Option) (string, *subscribe.Server, func(), error) {
	p, err := subscribe.NewServer(c, opts...)
	if err != nil {
		return "", nil, nil, fmt.Errorf("can't instantiate Server: %w", err)
	}

	lc := net.ListenConfig{}
	lis, err := lc.Listen(ctx, "tcp", net.JoinHostPort(HOST, PORT))
	if err != nil {
		return "", nil, nil, fmt.Errorf("can't set listener: %w", err)
	}
	opt, err := config.WithSelfTLSCert()
	if err != nil {
		return "", nil, nil, fmt.Errorf("can't create self-signed certificate: %w", err)
	}

	srv := grpc.NewServer(opt)
	pb.RegisterGNMIServer(srv, p)
	go srv.Serve(lis)

	return lis.Addr().String(), p, func() {
		lis.Close()
	}, nil
}

// sendUpdatesNew generates an update for each supplied path incrementing the
// timestamp and value for each using Elem instead of Elements
func sendUpdates(c *cache.Cache, updates map[string][]string, timestamp *time.Time) {
	for device, paths := range updates {
		*timestamp = timestamp.Add(time.Nanosecond)
		stream := make([]*pb.Update, 0, len(paths))

		for _, p := range paths {
			u, err := path.ParsePath(p)
			if err != nil {
				log.Errorf("error parsing path %s: %v", p, err)
				continue
			}

			val, err := value.FromScalar(rand.IntN(1000))
			if err != nil {
				log.Errorf("error creating scalar value for %s: %v", p, err)
				continue
			}
			update := &pb.Update{
				Path: u,
				Val:  val,
			}
			stream = append(stream, update)
		}

		noti := &pb.Notification{
			Prefix:    &pb.Path{Target: device},
			Timestamp: timestamp.UnixNano(),
			Update:    stream,
		}
		if err := c.GnmiUpdate(noti); err != nil {
			log.Errorf("error streaming update to %v: %v", device, err)
		}
	}

}

func periodic(period time.Duration, fn func()) {
	if period == 0 {
		return
	}
	t := time.NewTicker(period)
	defer t.Stop()
	for range t.C {
		fn()
	}
}

func createCache(file string) (Stream, error) {
	// Read Targets config file.
	u, err := os.Open(file)
	if err != nil {
		return Stream{}, fmt.Errorf("can't read Updates file: %w", err)
	}
	updates, err := GetUpdates(u)
	if err != nil {
		return Stream{}, fmt.Errorf("can't parse Updates info: %w", err)
	}
	// Get Target devices
	t := make([]string, 0, len(updates))
	for k := range updates {
		t = append(t, k)
	}

	// Setups a Cache for the list of targets
	targets := client.Path(t)

	return Stream{
		cache:    cache.New(targets),
		updates:  updates,
		interval: CADENCE * time.Second,
	}, nil
}

type Stream struct {
	cache    *cache.Cache
	updates  Updates
	interval time.Duration
}

func run(ctx context.Context, stream Stream) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// Start functions to periodically update metadata stored in the cache for each target.
	go periodic(stream.interval, stream.cache.UpdateMetadata)
	go periodic(stream.interval, stream.cache.UpdateSize)

	addr, server, teardown, err := startServer(ctx, stream.cache)
	if err != nil {
		return fmt.Errorf("can't start server: %w", err)
	}
	defer teardown()

	// Registers a callback function to receive calls for each update accepted by the cache
	stream.cache.SetClient(server.Update)

	log.Infof("listening on %v", addr)

	ticker := time.NewTicker(stream.interval)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				timestamp := time.Now()
				sendUpdates(stream.cache, stream.updates, &timestamp)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		close(quit)
	}()
	wg.Wait()
	return nil
}

func main() {
	flag.Parse()
	_, err := maxprocs.Set()
	if err != nil {
		log.Exitf("error setting GOMAXPROCS: %s\n", err)
	}

	stream, err := createCache(UPDATES)
	if err != nil {
		log.Exitf("error creating cache: %s\n", err)
	}

	if err := run(context.Background(), stream); err != nil {
		log.Exitf("error starting the gNMI server: %s\n", err)
	}
}
