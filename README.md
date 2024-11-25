# gNMI Streamer

[![GoDoc](https://godoc.org/github.com/nleiva/gnmi-streamer?status.svg)](https://godoc.org/github.com/nleiva/gnmi-streamer) 
[![Test](https://github.com/nleiva/gnmi-streamer/actions/workflows/test.yml/badge.svg)](https://github.com/nleiva/gnmi-streamer/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/nleiva/gnmi-streamer/branch/main/graph/badge.svg)](https://codecov.io/gh/nleiva/gnmi-streamer) 
[![Go Report Card](https://goreportcard.com/badge/github.com/nleiva/gnmi-streamer)](https://goreportcard.com/report/github.com/nleiva/gnmi-streamer)

gNMI Server to stream arbitrary data. It produces a random metric for the targets and data paths listed in a JSON file.

## Server

If you have Go installed in your system, run the server on a tab with `make server`.

```bash
$ make server
I1119 09:02:54.822880   29014 main.go:149] listening on [::]:9339
I1119 09:03:02.164796   29014 subscribe.go:283] peer: 127.0.0.1:51101 target: "dev2" subscription: subscribe:{prefix:{target:"dev2"} subscription:{path:{element:"a" elem:{name:"a"}}}}
I1119 09:03:07.014033   29014 subscribe.go:323] peer: 127.0.0.1:51101 target "dev2" subscription: end: "subscribe:{prefix:{target:\"dev2\"} subscription:{path:{element:\"a\" elem:{name:\"a\"}}}}"
```

You can configure the following environmental variables:

- `GNMI_HOST`: Server IP address. Default `""`.
- `GNMI_PORT`: Server port. Default `"9339"`.
- `GNMI_FILE`: JSON file with list of devices and paths to stream. Default `"updates.json"`.
- `GNMI_CADENCE`: How often to generate a metric for the device path. Default `5`.

If you don't have Go installed, you can run one of the executable files available in the [releases](https://github.com/nleiva/gnmi-streamer/releases). For example for a RHEL system:

```bash
wget https://github.com/nleiva/gnmi-streamer/releases/download/v0.1.0/gnmi-streamer_0.1.0_Linux_x86_64.rpm ## Download RPM
rpm -i gnmi-streamer_0.1.0_Linux_x86_64.rpm # Install RPM
gnmi-streamer -logtostderr # Run the server and log to stdout
```

## Client

Run the server on a different tab with `make client`.

```bash
$ make client
2024/11/19 09:03:02 Subscribing
RESPONSE:
  PATH: elem:{name:"a"}  elem:{name:"b"  key:{key:"n"  value:"c"}}  elem:{name:"d"}
  VALUE: int_val:929
RESPONSE:
  PATH: elem:{name:"a"}  elem:{name:"b"  key:{key:"n"  value:"c"}}  elem:{name:"d"}
  VALUE: int_val:18
...
```

### gNMIc

You can alternatively subscribe to the server with [gNMIc](https://gnmic.openconfig.net).


```bash
bash -c "$(curl -sL https://get-gnmic.openconfig.net)"
```

```bash
$ gnmic -a [::]:9339 --skip-verify --target dev1 subscribe --path "/state/router[router-name=*]/interface[interface-name=*]/statistics/ip/in-octets"
{
  "sync-response": true
}
{
  "source": ":9339",
  "subscription-name": "default-1732025210",
  "timestamp": 1732025204640357001,
  "time": "2024-11-19T09:06:44.640357001-05:00",
  "target": "dev1",
  "updates": [
    {
      "Path": "state/router[router-name=dev1]/interface[interface-name=*]/statistics/ip/in-octets",
      "values": {
        "state/router/interface/statistics/ip/in-octets": 213
      }
    }
  ]
}
```


```bash
$ gnmic -a [::]:9339 --skip-verify --target dev2 subscribe --path "a" --format prototext
sync_response: true

update: {
  timestamp: 1732025259640538002
  prefix: {
    target: "dev2"
  }
  update: {
    path: {
      elem: {
        name: "a"
      }
      elem: {
        name: "b"
        key: {
          key: "n"
          value: "c"
        }
      }
      elem: {
        name: "d"
      }
    }
    val: {
      int_val: 424
    }
  }
}
...
```

## Credits

This is copy & paste from the following packages.

- [SUBSCRIBE SERVER](https://github.com/openconfig/gnmi/tree/master/subscribe)
- [COLLECTOR](https://github.com/openconfig/gnmi/tree/master/collector)
- [gNMI PATH](https://github.com/openconfig/gnmic/tree/main/pkg/api/path)

## References

- [SNMP is dead](https://pc.nanog.org/static/published/meetings/NANOG73/1677/20180625_Shakir_Snmp_Is_Dead_v1.pdf): Describes the a gNMI caching collector from multiple gNMI sources (targets) to multiple gNMI clients. [recording](https://youtu.be/McNm_WfQTHw?si=lPy5a7qIdIKMW7ne)
