package main

import (
	"flag"
	"log"

	"github.com/autonomy/talos/src/initramfs/cmd/init/pkg/constants"
	"github.com/autonomy/talos/src/initramfs/cmd/trustd/pkg/reg"
	"github.com/autonomy/talos/src/initramfs/pkg/grpc/factory"
	"github.com/autonomy/talos/src/initramfs/pkg/grpc/gen"
	"github.com/autonomy/talos/src/initramfs/pkg/grpc/tls"
	"github.com/autonomy/talos/src/initramfs/pkg/userdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	dataPath *string
	generate *bool
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds | log.Ltime)
	dataPath = flag.String("userdata", "", "the path to the user data")
	generate = flag.Bool("generate", false, "generate the TLS certificate using one of the Root of Trusts")
	flag.Parse()
}

func main() {
	data, err := userdata.Open(*dataPath)
	if err != nil {
		log.Fatalf("open user data: %v", err)
	}

	if *generate {
		var generator *gen.Generator
		generator, err = gen.NewGenerator(data, constants.TrustdPort)
		if err != nil {
			log.Fatal(err)
		}
		if err = generator.Identity(data.Security); err != nil {
			log.Fatalf("generate identity: %v", err)
		}
	}

	config, err := tls.NewConfig(tls.Mutual, data.Security.OS)
	if err != nil {
		log.Fatalf("credentials: %v", err)
	}

	log.Println("Starting blockd")
	err = factory.Listen(
		&reg.Registrator{Data: data.Security.OS},
		factory.Network("unix"),
		factory.ServerOptions(
			grpc.Creds(
				credentials.NewTLS(config),
			),
		),
	)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
}
