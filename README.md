# registry-servicecomb (*This is a community driven project*)

Use [service-comb](https://github.com/apache/servicecomb-service-center) as service registry for `Kitex`.

## How to use?

### Server
```go
import (
	// ...
	kitexregistry "github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/server"
	"github.com/kitex-contrib/registry-servicecomb/registry"
)

// ...

func main() {
	r, err := registry.NewDefaultSCRegistry()
	if err != nil {
		panic(err)
	}
	svr := hello.NewServer(
		new(HelloImpl),
		server.WithRegistry(r),
		server.WithRegistryInfo(&kitexregistry.Info{ServiceName: "Hello"}),
		server.WithServiceAddr(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}),
	)
	if err := svr.Run(); err != nil {
		log.Println("server stopped with error:", err)
	} else {
		log.Println("server stopped")
	}
	// ...
}

```

### Client
```go
import (
	// ...
	"github.com/cloudwego/kitex/client"
	"github.com/kitex-contrib/registry-servicecomb/resolver"
)

func main() {
	r, err := resolver.NewDefaultSCResolver()
	if err != nil {
		panic(err)
	}
	newClient := hello.MustNewClient("Hello", client.WithResolver(r))
	// ...
}
```

maintained by: [bodhisatan](https://github.com/bodhisatan)