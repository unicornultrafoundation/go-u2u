package launcher

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/stretchr/testify/require"
	"github.com/unicornultrafoundation/go-hashgraph/abft"
	"github.com/unicornultrafoundation/go-hashgraph/utils/cachescale"

	"github.com/unicornultrafoundation/go-u2u/evmcore"
	"github.com/unicornultrafoundation/go-u2u/gossip"
	"github.com/unicornultrafoundation/go-u2u/gossip/emitter"
	"github.com/unicornultrafoundation/go-u2u/vecmt"
)

func TestConfigFile(t *testing.T) {
	cacheRatio := cachescale.Ratio{
		Base:   uint64(DefaultCacheSize*1 - ConstantCacheSize),
		Target: uint64(DefaultCacheSize*2 - ConstantCacheSize),
	}

	src := config{
		Node:           defaultNodeConfig(),
		U2u:            gossip.DefaultConfig(cacheRatio),
		Emitter:        emitter.DefaultConfig(),
		TxPool:         evmcore.DefaultTxPoolConfig,
		U2uStore:       gossip.DefaultStoreConfig(cacheRatio),
		Hashgraph:      abft.DefaultConfig(),
		HashgraphStore: abft.DefaultStoreConfig(cacheRatio),
		VectorClock:    vecmt.DefaultConfig(cacheRatio),
	}

	canonical := func(nn []*enode.Node) []*enode.Node {
		if len(nn) == 0 {
			return []*enode.Node{}
		}
		return nn
	}

	for name, val := range map[string][]*enode.Node{
		"Nil":     nil,
		"Empty":   {},
		"Default": asDefault,
		"UserDefined": {enode.MustParse(
			"enr:-HW4QIEFxJwyZzPQJPE2DbQpEu7FM1Gv99VqJ3CbLb22fm9_V9cfdZdSBpZCyrEb5UfMeC6k9WT0iaaeAjRcuzCfr4yAgmlkgnY0iXNlY3AyNTZrMaECps0D9hhmXEN5BMgVVe0xT5mpYU9zv4YxCdTApmfP-l0",
		)},
	} {
		t.Run(name+"BootstrapNodes", func(t *testing.T) {
			require := require.New(t)

			src.Node.P2P.BootstrapNodes = val
			src.Node.P2P.BootstrapNodesV5 = val

			stream, err := tomlSettings.Marshal(&src)
			require.NoError(err)

			var got config
			err = tomlSettings.NewDecoder(bytes.NewReader(stream)).Decode(&got)
			require.NoError(err)

			{ // toml workaround
				src.Node.P2P.BootstrapNodes = canonical(src.Node.P2P.BootstrapNodes)
				got.Node.P2P.BootstrapNodes = canonical(got.Node.P2P.BootstrapNodes)
				src.Node.P2P.BootstrapNodesV5 = canonical(src.Node.P2P.BootstrapNodesV5)
				got.Node.P2P.BootstrapNodesV5 = canonical(got.Node.P2P.BootstrapNodesV5)
			}

			require.Equal(src.Node.P2P.BootstrapNodes, got.Node.P2P.BootstrapNodes)
			require.Equal(src.Node.P2P.BootstrapNodesV5, got.Node.P2P.BootstrapNodesV5)
		})
	}
}
