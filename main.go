/*
Based on: https://github.com/openconfig/gnmi/tree/master/subscribe
*/

package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/openconfig/gnmi/cache"
	"github.com/openconfig/gnmi/client"
	"github.com/openconfig/gnmi/testing/fake/testing/grpc/config"
	"github.com/openconfig/gnmi/value"
	"google.golang.org/grpc"

	pb "github.com/openconfig/gnmi/proto/gnmi"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/nleiva/gnmi-streamer/path"
	"github.com/nleiva/gnmi-streamer/subscribe"
)

const (
	gNMIHOST    = ""
	gNMIPORT    = "9339"
	gNMICadence = 5
)

func startServer(ctx context.Context, targets []string, opts ...subscribe.Option) (string, *subscribe.Server, *cache.Cache, func(), error) {
	c := cache.New(targets)
	p, err := subscribe.NewServer(c, opts...)
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("can't instantiate Server: %w", err)
	}

	c.SetClient(p.Update)

	lc := net.ListenConfig{}
	lis, err := lc.Listen(ctx, "tcp", net.JoinHostPort(gNMIHOST, gNMIPORT))
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("can't set listener: %w", err)
	}
	opt, err := config.WithSelfTLSCert()
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("can't create self-signed certificate: %w", err)
	}

	srv := grpc.NewServer(opt)
	pb.RegisterGNMIServer(srv, p)
	go srv.Serve(lis)

	return lis.Addr().String(), p, p.C, func() {
		lis.Close()
	}, nil
}

// sendUpdatesNew generates an update for each supplied path incrementing the
// timestamp and value for each using Elem instead of Elements
func sendUpdatesNew(c *cache.Cache, updates map[string][]string, timestamp *time.Time) {
	*timestamp = timestamp.Add(time.Nanosecond)

	for device, paths := range updates {
		stream := make([]*pb.Update, 0, len(paths))

		for _, p := range paths {
			u, err := path.Parse(p)
			if err != nil {
				log.Printf("error parsing path %s: %v", p, err)
				continue
			}

			val, err := value.FromScalar(rand.IntN(1000))
			if err != nil {
				log.Printf("error creating scalar value for %s: %v", p, err)
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
			log.Printf("error streaming update to %v: %v", device, err)
		}
	}

}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	addr, _, cache, teardown, err := startServer(ctx, client.Path{"dev1", "dev2"})
	if err != nil {
		return fmt.Errorf("can't start server: %w", err)
	}
	defer teardown()

	log.Printf("listening on %v", addr)

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
				sendUpdatesNew(cache, updates, &timestamp)
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
