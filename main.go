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
	gNMIHOST    = ""
	gNMIPORT    = "9339"
	gNMICadence = 5
)

func startServer(ctx context.Context, c *cache.Cache, opts ...subscribe.Option) (string, *subscribe.Server, func(), error) {
	p, err := subscribe.NewServer(c, opts...)
	if err != nil {
		return "", nil, nil, fmt.Errorf("can't instantiate Server: %w", err)
	}

	lc := net.ListenConfig{}
	lis, err := lc.Listen(ctx, "tcp", net.JoinHostPort(gNMIHOST, gNMIPORT))
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
	*timestamp = timestamp.Add(time.Nanosecond)

	for device, paths := range updates {
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

func run(ctx context.Context) error {

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// Setups a Cache for a list of targets
	targets := client.Path{"dev1", "dev2"}
	c := cache.New(targets)

	addr, server, teardown, err := startServer(ctx, c)
	if err != nil {
		return fmt.Errorf("can't start server: %w", err)
	}
	defer teardown()

	// Registers a callback function to receive calls for each update accepted by the cache
	c.SetClient(server.Update)

	log.Infof("listening on %v", addr)

	updates := map[string][]string{
		"dev1": {
			"/state/router[router-name=dev1]/interface[interface-name=*]/statistics/ip/in-octets",
			"/state/router[router-name=dev1]/interface[interface-name=*]/statistics/ip/out-octets",
			"/terminal-device/logical-channels/channel[index=*]/otn/state/esnr/instant",
		},
		"dev2": {
			"/a/b[n=c]/d",
		},
	}

	ticker := time.NewTicker(gNMICadence * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			var timestamp time.Time
			select {
			case <-ticker.C:
				sendUpdates(c, updates, &timestamp)
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
		fmt.Fprintf(os.Stderr, "error setting GOMAXPROCS: %s\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error starting the gNMI server: %s\n", err)
		os.Exit(1)
	}
}
