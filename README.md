# gNMI Streamer

## Server

```bash
$ make server
2024/11/15 17:41:09 listening on [::]:9339
2024/11/15 17:41:12 peer: 127.0.0.1:44106 target: "dev1" subscription: subscribe:{prefix:{target:"dev1"} subscription:{path:{elem:{name:"state"} elem:{name:"router" key:{key:"router-name" value:"*"}} elem:{name:"interface" key:{key:"interface-name" value:"*"}} elem:{name:"statistics"} elem:{name:"ip"} elem:{name:"in-octets"}}}}
2024/11/15 17:41:12 start processing Subscription for dev1
2024/11/15 17:41:12 end processSubscription for dev1
2024/11/15 17:41:37 peer: 127.0.0.1:44106 target "dev1" subscription: end: "subscribe:{prefix:{target:\"dev1\"} subscription:{path:{elem:{name:\"state\"} elem:{name:\"router\" key:{key:\"router-name\" value:\"*\"}} elem:{name:\"interface\" key:{key:\"interface-name\" value:\"*\"}} elem:{name:\"statistics\"} elem:{name:\"ip\"} elem:{name:\"in-octets\"}}}}"
```

## Client

### Local

```bash
$ cd client
```

```bash
$ make client
2024/11/15 17:52:55 Subscribing
RESPONSE:
  PATH: elem:{name:"a"} elem:{name:"b" key:{key:"n" value:"c"}} elem:{name:"d"}
  VALUE: int_val:635
RESPONSE:
  PATH: elem:{name:"a"} elem:{name:"b" key:{key:"n" value:"c"}} elem:{name:"d"}
  VALUE: int_val:903
...
```

### gNMIc


```bash
bash -c "$(curl -sL https://get-gnmic.openconfig.net)"
```

```bash
$ gnmic -a [::]:9339 --skip-verify --target dev1 subscribe --path "/state/router[router-name=*]/interface[interface-name=*]/statistics/ip/in-octets"
{
  "sync-response": true
}
{
  "source": "[::]:9339",
  "subscription-name": "default-1731692472",
  "timestamp": -6795364578871345151,
  "time": "1754-08-30T22:43:41.128654849Z",
  "target": "dev1",
  "updates": [
    {
      "Path": "state/router[router-name=dev1]/interface[interface-name=*]/statistics/ip/in-octets",
      "values": {
        "state/router/interface/statistics/ip/in-octets": 5
      }
    }
  ]
}
{
  "source": "[::]:9339",
  "subscription-name": "default-1731692472",
  "timestamp": -6795364578871345151,
  "time": "1754-08-30T22:43:41.128654849Z",
  "target": "dev1",
  "updates": [
    {
      "Path": "state/router[router-name=dev1]/interface[interface-name=*]/statistics/ip/in-octets",
      "values": {
        "state/router/interface/statistics/ip/in-octets": 783
      }
    }
  ]
}
```


```bash
$ unset https_proxy
$ unset HTTPS_PROXY
$ gnmic -a [::]:9339 --skip-verify --target dev2 subscribe --path "a" --format prototext
update: {
  timestamp: -6795364578871345151
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
      int_val: 81
    }
  }
}

sync_response: true

update: {
  timestamp: -6795364578871345151
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
      int_val: 634
    }
  }
}
...
```

## Credits

- [SUBSCRIBE](https://github.com/openconfig/gnmi/tree/master/subscribe)
- [PATH](https://github.com/openconfig/gnmic/tree/main/pkg/api/path)

## References

- [SNMP is dead](https://pc.nanog.org/static/published/meetings/NANOG73/1677/20180625_Shakir_Snmp_Is_Dead_v1.pdf): Describes the a gNMI caching collector from multiple gNMI sources (targets) to multiple gNMI clients. [recording](https://youtu.be/McNm_WfQTHw?si=lPy5a7qIdIKMW7ne)