package launcher

import (
	"crypto/rand"
	"math/big"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/unicornultrafoundation/go-u2u/integration/makefakegenesis"
	"github.com/unicornultrafoundation/go-u2u/params"
)

const (
	ipcAPIs  = "abft:1.0 admin:1.0 dag:1.0 debug:1.0 net:1.0 personal:1.0 rpc:1.0 trace:1.0 txpool:1.0 web3:1.0"
	httpAPIs = "abft:1.0 dag:1.0 rpc:1.0 web3:1.0"
)

// Tests that a node embedded within a console can be started up properly and
// then terminated by closing the input stream.
func TestConsoleWelcome(t *testing.T) {
	// Start an u2u console, make sure it's cleaned up and terminate the console
	cli := exec(t,
		"--fakenet", "0/1", "--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none", "--cache", "7923",
		"console")

	// Gather all the infos the welcome message needs to contain
	cli.SetTemplateFunc("goos", func() string { return runtime.GOOS })
	cli.SetTemplateFunc("goarch", func() string { return runtime.GOARCH })
	cli.SetTemplateFunc("gover", runtime.Version)
	cli.SetTemplateFunc("version", func() string { return params.VersionWithCommit("", "") })
	cli.SetTemplateFunc("niltime", genesisStart)
	cli.SetTemplateFunc("apis", func() string { return ipcAPIs })

	// Verify the actual welcome message to the required template
	cli.Expect(`
Welcome to the Consensus JavaScript console!

instance: go-u2u/v{{version}}/{{goos}}-{{goarch}}/{{gover}}
coinbase: {{.Coinbase}}
at block: 1 ({{niltime}})
 datadir: {{.Datadir}}
 modules: {{apis}}

To exit, press ctrl-d
> {{.InputLine "exit"}}
`)
	cli.ExpectExit()
}

// Tests that a console can be attached to a running node via various means.
func TestAttachWelcome(t *testing.T) {
	var (
		ipc      string
		httpPort string
		wsPort   string
	)
	// Configure the instance for IPC attachment
	if runtime.GOOS == "windows" {
		ipc = `\\.\pipe\u2u.ipc` + strconv.Itoa(trulyRandInt(100000, 999999))
	} else {
		dir := tmpdir(t)
		ipc = filepath.Join(dir, "u2u.ipc")
	}
	// And HTTP + WS attachment
	p := trulyRandInt(1024, 65533) // Yeah, sometimes this will fail, sorry :P
	httpPort = strconv.Itoa(p)
	wsPort = strconv.Itoa(p + 1)
	cli := exec(t,
		"--fakenet", "0/1", "--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"--ipcpath", ipc, "--http", "--http.port", httpPort, "--ws", "--ws.port", wsPort, "--cache", "7923",
		"--datadir.minfreedisk", "1")

	t.Run("ipc", func(t *testing.T) {
		waitForEndpoint(t, ipc, 4*time.Second)
		testAttachWelcome(t, cli, "ipc:"+ipc, ipcAPIs)
	})
	t.Run("http", func(t *testing.T) {
		endpoint := "http://127.0.0.1:" + httpPort
		waitForEndpoint(t, endpoint, 4*time.Second)
		testAttachWelcome(t, cli, endpoint, httpAPIs)
	})
	t.Run("ws", func(t *testing.T) {
		endpoint := "ws://127.0.0.1:" + wsPort
		waitForEndpoint(t, endpoint, 4*time.Second)
		testAttachWelcome(t, cli, endpoint, httpAPIs)
	})

	cli.Kill()
	cli.WaitExit()
}

func testAttachWelcome(t *testing.T, cli *testcli, endpoint, apis string) {
	// Attach to a running u2u node and terminate immediately
	attach := exec(t, "attach", endpoint)

	// Gather all the infos the welcome message needs to contain
	attach.SetTemplateFunc("goos", func() string { return runtime.GOOS })
	attach.SetTemplateFunc("goarch", func() string { return runtime.GOARCH })
	attach.SetTemplateFunc("gover", runtime.Version)
	attach.SetTemplateFunc("version", func() string { return params.VersionWithCommit("", "") })
	attach.SetTemplateFunc("niltime", genesisStart)
	attach.SetTemplateFunc("ipc", func() bool { return strings.HasPrefix(endpoint, "ipc") })
	attach.SetTemplateFunc("datadir", func() string { return cli.Datadir })
	attach.SetTemplateFunc("coinbase", func() string { return cli.Coinbase })
	attach.SetTemplateFunc("apis", func() string { return apis })

	// Verify the actual welcome message to the required template
	attach.Expect(`
Welcome to the Consensus JavaScript console!

instance: go-u2u/v{{version}}/{{goos}}-{{goarch}}/{{gover}}
coinbase: {{coinbase}}
at block: 1 ({{niltime}}){{if ipc}}
 datadir: {{datadir}}{{end}}
 modules: {{apis}}

To exit, press ctrl-d
> {{.InputLine "exit" }}
`)
	attach.ExpectExit()
}

// trulyRandInt generates a crypto random integer used by the console tests to
// not clash network ports with other tests running cocurrently.
func trulyRandInt(lo, hi int) int {
	num, _ := rand.Int(rand.Reader, big.NewInt(int64(hi-lo)))
	return int(num.Int64()) + lo
}

func genesisStart() string {
	return time.Unix(int64(makefakegenesis.FakeGenesisTime.Unix()), 0).Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
}
