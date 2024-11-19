package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	pb "github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/protobuf/proto"

	"github.com/openconfig/gnmi/client"
	gnmiclient "github.com/openconfig/gnmi/client/gnmi"
)

const (
	HOST = ""
	PORT = "9339"
)

func main() {
	device := "dev2"

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	q := client.Query{
		Addrs:   []string{net.JoinHostPort(HOST, PORT)},
		Target:  device,
		Queries: []client.Path{{"a"}},
		Type:    client.Stream,

		ProtoHandler: func(msg proto.Message) error {
			resp, ok := msg.(*pb.SubscribeResponse)
			if !ok {
				return fmt.Errorf("failed to type assert message %#v", msg)
			}
			switch v := resp.Response.(type) {
			case *pb.SubscribeResponse_Update:
				{
					fmt.Printf("RESPONSE:\n  PATH: %v\n  VALUE: %v\n", v.Update.Update[0].Path, v.Update.Update[0].Val)
				}
			case *pb.SubscribeResponse_Error:
				return fmt.Errorf("error in response: %s", v)
			case *pb.SubscribeResponse_SyncResponse:
			default:
				return fmt.Errorf("unknown response %T: %s", v, v)
			}

			return nil
		},
		TLS: &tls.Config{InsecureSkipVerify: true},
	}

	log.Printf("Subscribing")
	c := client.BaseClient{}

	err := c.Subscribe(ctx, q, gnmiclient.Type)
	defer c.Close()
	if err != nil {
		log.Fatalf("can't subscribe to gNMI server: %v", err)
	}
}
