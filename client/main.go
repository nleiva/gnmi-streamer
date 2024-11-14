package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/protobuf/proto"

	"github.com/openconfig/gnmi/client"
	gnmiclient "github.com/openconfig/gnmi/client/gnmi"
)

const (
	gNMIHOST = ""
	gNMIPORT = "9339"
)

func main() {
	device := "dev1"
	query := client.Path{"a"}
	count := 0
	sync := 0

	q := client.Query{
		Addrs:   []string{net.JoinHostPort(gNMIHOST, gNMIPORT)},
		Target:  device,
		Queries: []client.Path{query},
		Type:    client.Stream,

		ProtoHandler: func(msg proto.Message) error {
			resp, ok := msg.(*pb.SubscribeResponse)
			if !ok {
				return fmt.Errorf("failed to type assert message %#v", msg)
			}
			switch v := resp.Response.(type) {
			case *pb.SubscribeResponse_Update:
				{
					count++
					fmt.Printf("RESPONSE:\n  PATH: %v\n  VALUE: %v\n", v.Update.Update[0].Path, v.Update.Update[0].Val)
				}
			case *pb.SubscribeResponse_Error:
				return fmt.Errorf("error in response: %s", v)
			case *pb.SubscribeResponse_SyncResponse:
				sync++
			default:
				return fmt.Errorf("unknown response %T: %s", v, v)
			}

			return nil
		},
		TLS: &tls.Config{InsecureSkipVerify: true},
	}

	log.Printf("Subscribing")
	c := client.BaseClient{}
	err := c.Subscribe(context.Background(), q, gnmiclient.Type)
	if err != nil {
		log.Fatalf("can't subscribe to gNMI server: %v", err)
	}

	time.Sleep(60 * time.Second)
}
