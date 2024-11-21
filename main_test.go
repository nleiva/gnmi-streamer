package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	pb "github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/protobuf/proto"

	"github.com/openconfig/gnmi/client"
	gnmiclient "github.com/openconfig/gnmi/client/gnmi"
)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := createCache("testdata/updates.json")
	if err != nil {
		log.Fatalf("ERROR creating cache: %v\n", err)
	}
	// This is a blocking call, so we run it in the background.
	go setup(ctx, stream)

	now := time.Now()
	sendUpdates(stream.cache, stream.updates, &now)

	// Run test cases
	code := m.Run()

	// Teardown
	cancel()
	time.Sleep(300 * time.Millisecond)

	os.Exit(code)
}

func setup(ctx context.Context, s Stream) {
	err := run(ctx, s)
	if err != nil {
		log.Fatalf("ERROR setting up server: %v\n", err)
	}
}

func TestSubscribeOnce(t *testing.T) {
	testCases := []struct {
		dev   string
		query client.Path
		count int
		err   bool
	}{
		// These cases will be found.
		{"dev1", client.Path{"state"}, 2, false},
		{"dev2", client.Path{"a"}, 1, false},
		{"dev1", client.Path{"terminal-device"}, 1, false},
		{"dev1", client.Path{"*", "router"}, 2, false},
		// This case is not found.
		{"dev1", client.Path{"b"}, 0, false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("target: %q query: %q", tc.dev, tc.query), func(t *testing.T) {
			count := 0
			sync := 0
			q := client.Query{
				Addrs:   []string{net.JoinHostPort(HOST, PORT)},
				Target:  tc.dev,
				Queries: []client.Path{tc.query},
				Type:    client.Once,

				ProtoHandler: func(msg proto.Message) error {
					resp, ok := msg.(*pb.SubscribeResponse)
					if !ok {
						t.Errorf("failed to type assert message %#v", msg)
					}
					switch v := resp.Response.(type) {
					case *pb.SubscribeResponse_Update:
						count++
					case *pb.SubscribeResponse_Error:
						t.Errorf("error in response: %s", v)
					case *pb.SubscribeResponse_SyncResponse:
						sync++
					default:
						t.Errorf("unknown response %T: %s", v, v)
					}

					return nil
				},
				TLS: &tls.Config{InsecureSkipVerify: true},
			}

			c := client.BaseClient{}
			err := c.Subscribe(context.Background(), q, gnmiclient.Type)
			defer c.Close()

			if err != nil && !tc.err {
				t.Fatalf("unexpected error: %v", err)
			}
			if err == nil && tc.err {
				t.Fatal("didn't get expected error")
			}

			if tc.err {
				return
			}
			if sync != 1 {
				t.Errorf("got %d sync messages, want 1", sync)
			}
			if count != tc.count {
				t.Errorf("got %d updates, want %d", count, tc.count)
			}
		})

	}
}

func TestConfig(t *testing.T) {
	tt := []struct {
		name string
		file string
		err  string
	}{
		{
			name: "valid file",
			file: "testdata/updates.json",
		},
		{
			name: "invalid file",
			file: "testdata/not-updates.json",
			err:  "can't read Updates file: open testdata/not-updates.json: no such file or directory",
		},
		{
			name: "invalid file format",
			file: "testdata/wrong-updates.json",
			err:  "can't parse Updates info: can't decode Targets file: can't decode object: json: cannot unmarshal array into Go value of type main.Updates",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, err := createCache(tc.file)
			if err != nil && err.Error() != tc.err {
				t.Errorf("got %v, want %v", err, tc.err)
			}
		})
	}
}
