# gNMI Streamer

### Server

```bash
$ go run main.go
2024/11/14 20:01:23 listening on [::]:9339
2024/11/14 20:01:31 peer: 127.0.0.1:54710 target: "dev1" subscription: subscribe:{prefix:{target:"dev1"} subscription:{path:{element:"a" elem:{name:"a"}}}}
2024/11/14 20:01:31 start processing Subscription for dev1
2024/11/14 20:01:31 end processSubscription for dev1
```

### Clients

#### Local

```bash
$ cd client
```

```bash
$ go run main.go 
2024/11/14 20:01:31 Subscribing
RESPONSE:
  PATH: element:"a" element:"b" element:"c" element:"d"
  VALUE: int_val:805
RESPONSE:
  PATH: element:"a" element:"b" element:"d" element:"e"
  VALUE: int_val:668
RESPONSE:
  PATH: element:"a" element:"c" element:"d" element:"e"
  VALUE: int_val:935
RESPONSE:
  PATH: element:"a" element:"b" element:"c" element:"d"
  VALUE: int_val:580
RESPONSE:
  PATH: element:"a" element:"b" element:"d" element:"e"
  VALUE: int_val:151
RESPONSE:
  PATH: element:"a" element:"c" element:"d" element:"e"
  VALUE: int_val:494
...
```

#### gNMIc


```bash
bash -c "$(curl -sL https://get-gnmic.openconfig.net)"
```

```bash
unset https_proxy
$ gnmic -a [::]:9339 --skip-verify --target dev1 subscribe --path "a"
{
  "sync-response": true
}
{
  "source": "[::]:9339",
  "subscription-name": "default-1731616427",
  "timestamp": -6795364578871345151,
  "time": "1754-08-30T22:43:41.128654849Z",
  "target": "dev1",
  "updates": [
    {
      "Path": "",
      "values": {
        "": 239
      }
    }
  ]
}
{
  "source": "[::]:9339",
  "subscription-name": "default-1731616427",
  "timestamp": -6795364578871345150,
  "time": "1754-08-30T22:43:41.12865485Z",
  "target": "dev1",
  "updates": [
    {
      "Path": "",
      "values": {
        "": 101
      }
    }
  ]
}
{
  "source": "[::]:9339",
  "subscription-name": "default-1731616427",
  "timestamp": -6795364578871345149,
  "time": "1754-08-30T22:43:41.128654851Z",
  "target": "dev1",
  "updates": [
    {
      "Path": "",
      "values": {
        "": 498
      }
    }
  ]
}
{
  "source": "[::]:9339",
  "subscription-name": "default-1731616427",
  "timestamp": -6795364578871345149,
  "time": "1754-08-30T22:43:41.128654851Z",
  "target": "dev1",
  "updates": [
    {
      "Path": "",
      "values": {
        "": 306
      }
    }
  ]
}
{
  "source": "[::]:9339",
  "subscription-name": "default-1731616427",
  "timestamp": -6795364578871345151,
  "time": "1754-08-30T22:43:41.128654849Z",
  "target": "dev1",
  "updates": [
    {
      "Path": "",
      "values": {
        "": 154
      }
    }
  ]
}
{
  "source": "[::]:9339",
  "subscription-name": "default-1731616427",
  "timestamp": -6795364578871345150,
  "time": "1754-08-30T22:43:41.12865485Z",
  "target": "dev1",
  "updates": [
    {
      "Path": "",
      "values": {
        "": 738
      }
    }
  ]
}
...
```

