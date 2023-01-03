package memgraph

import (
	"github.com/prahaladd/gograph/core"
	"github.com/prahaladd/gograph/neo"
)

func init() {
	core.RegisterConnectorFactory("memgraph", neo.NewConnection)
}
