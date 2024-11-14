/*
Based on: https://github.com/openconfig/gnmi/tree/master/subscribe
*/

package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"time"

	"github.com/openconfig/gnmi/cache"
	"github.com/openconfig/gnmi/client"
	"github.com/openconfig/gnmi/testing/fake/testing/grpc/config"
	"github.com/openconfig/gnmi/value"
	"google.golang.org/grpc"

	pb "github.com/openconfig/gnmi/proto/gnmi"

	"github.com/nleiva/gnmi-streamer/subscribe"
)

const (
	gNMIHOST = ""
	gNMIPORT = "9339"
)

func startServer(targets []string, opts ...subscribe.Option) (string, *subscribe.Server, *cache.Cache, func(), error) {
	c := cache.New(targets)
	p, err := subscribe.NewServer(c, opts...)
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("can't instantiate Server: %w", err)
	}

	c.SetClient(p.Update)

	lis, err := net.Listen("tcp", net.JoinHostPort(gNMIHOST, gNMIPORT))
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("can't set listener: %w", err)
	}
	opt, err := config.WithSelfTLSCert()
	if err != nil {
		return "", nil, nil, nil, fmt.Errorf("config.WithSelfCert: %w", err)
	}

	srv := grpc.NewServer(opt)
	pb.RegisterGNMIServer(srv, p)
	go srv.Serve(lis)

	return lis.Addr().String(), p, p.C, func() {
		lis.Close()
	}, nil
}

// sendUpdates generates an update for each supplied path incrementing the
// timestamp and value for each.
func sendUpdates(c *cache.Cache, paths []client.Path, timestamp *time.Time) {
	for _, path := range paths {
		*timestamp = timestamp.Add(time.Nanosecond)
		sv, err := value.FromScalar(rand.IntN(1000))
		if err != nil {
			log.Printf("error with scalar value: %v", err)
			continue
		}
		noti := &pb.Notification{
			Prefix:    &pb.Path{Target: path[0]},
			Timestamp: timestamp.UnixNano(),
			Update: []*pb.Update{
				{
					Path: &pb.Path{Element: path[1:]},
					Val:  sv,
				},
			},
		}
		if err := c.GnmiUpdate(noti); err != nil {
			log.Printf("error streaming update: %v", err)
		}
	}
}

func main() {
	addr, _, cache, teardown, err := startServer(client.Path{"dev1", "dev2"})
	if err != nil {
		log.Fatal(err)
	}
	defer teardown()

	log.Printf("listening on %v", addr)

	paths := []client.Path{
		{"dev1", "a", "b", "c", "d"},
		{"dev1", "a", "b", "d", "e"},
		{"dev1", "a", "c", "d", "e"},
		{"dev2", "x", "y", "z"},
	}

	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			var timestamp time.Time
			select {
			case <-ticker.C:
				sendUpdates(cache, paths, &timestamp)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	time.Sleep(60 * time.Second)
	close(quit)
}
