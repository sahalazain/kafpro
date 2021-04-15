package main

import (
	"github.com/sahalazain/kafpro/cmd"
	_ "gocloud.dev/pubsub/kafkapubsub"
	_ "gocloud.dev/pubsub/mempubsub"
)

func main() {
	cmd.Run()
}
