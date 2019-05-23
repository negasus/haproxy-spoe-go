# Haproxy SPOE Golang Agent Library

Terms from [Haproxy SPOE specification](https://www.haproxy.org/download/1.9/doc/SPOE.txt) 

```
* SPOE : Stream Processing Offload Engine.

    A SPOE is a filter talking to servers managed ba a SPOA to offload the
    stream processing. An engine is attached to a proxy. A proxy can have
    several engines. Each engine is linked to an agent and only one.

* SPOA : Stream Processing Offload Agent.

    A SPOA is a service that will receive info from a SPOE to offload the
    stream processing. An agent manages several servers. It uses a backend to
    reference all of them. By extension, these servers can also be called
    agents.

* SPOP : Stream Processing Offload Protocol, used by SPOEs to talk to SPOA
         servers.

    This protocol is used by engines to talk to agents. It is an in-house
    binary protocol described in this documentation.
```


This library implements SPOA for Golang applications

## Example

Example from Section 2.5 [SPOE specification](https://www.haproxy.org/download/1.9/doc/SPOE.txt) describe simple IP reputation service.

> Here is a simple but complete example that sends client-ip address to a ip
  reputation service. This service can set the variable "ip_score" which is an
  integer between 0 and 100, indicating its reputation (100 means totally safe
  and 0 a blacklisted IP with no doubt).



Golang backend application for this example
 
```go
package main

import (
	"github.com/negasus/haproxy-spoe-go/action"
	"github.com/negasus/haproxy-spoe-go/agent"
	"github.com/negasus/haproxy-spoe-go/request"
	"log"
	"math/rand"
	"net"
	"os"
)

func main() {

	log.Print("listen 3000")

	listener, err := net.Listen("tcp4", "127.0.0.1:3000")
	if err != nil {
		log.Printf("error create listener, %v", err)
		os.Exit(1)
	}
	defer listener.Close()

	a := agent.New(handler)

	if err := a.Serve(listener); err != nil {
		log.Printf("error agent serve: %+v\n", err)
	}
}

func handler(req *request.Request) {

	log.Printf("handle request EngineID: '%s', StreamID: '%d', FrameID: '%d' with %d messages\n", req.EngineID, req.StreamID, req.FrameID, req.Messages.Len())

	messageName := "get-ip-reputation"

	mes, err := req.Messages.GetByName(messageName)
	if err != nil {
		log.Printf("message %s not found: %v", messageName, err)
		return
	}

	ipValue, ok := mes.KV.Get("ip")
	if !ok {
		log.Printf("var 'ip' not found in message")
		return
	}

	ip, ok := ipValue.(net.IP)
	if !ok {
		log.Printf("var 'ip' has wrong type. expect IP addr")
		return
	}

	ipScore := rand.Intn(100)

	log.Printf("IP: %s, send score '%d'", ip.String(), ipScore)

	req.Actions.SetVar(action.ScopeSession, "ip_score", ipScore)
}
```

