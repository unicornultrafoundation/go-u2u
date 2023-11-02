package launcher

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/unicornultrafoundation/go-u2u/evmcore/txtracer"
)

func TestTxTracing(t *testing.T) {

	// Start test node on random ports and keep it running for another requests
	port := strconv.Itoa(trulyRandInt(10000, 65536))
	wsport := strconv.Itoa(trulyRandInt(10000, 65536))
	cliNode := exec(t,
		"--fakenet", "1/1", "--enabletxtracer", "--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"--ws", "--ws.port", wsport, "--http", "--http.api", "eth,web3,net,txpool,trace", "--http.port", port, "--allow-insecure-unlock",
		"--cache", "7923", "--datadir.minfreedisk", "1")

	// Wait for node to start
	endpoint := "ws://127.0.0.1:" + wsport
	waitForEndpoint(t, endpoint, 60*time.Second)

	// Deploy a smart contract from the testdata javascript file
	cliConsoleDeploy := exec(t, "attach", "--datadir", cliNode.Datadir, "--exec", "loadScript('testdata/txtracer_test.js')")
	cliConsoleOutput := strings.Split(string(*cliConsoleDeploy.GetOutPipeData()), "\n")
	txHashCall := cliConsoleOutput[1]
	txHashDeploy := cliConsoleOutput[2]
	cliConsoleDeploy.WaitExit()

	traceResult1, err := getTrace(txHashCall, port)
	if err != nil {
		log.Fatalln(err)
	}

	traceResult2, err := getTrace(txHashDeploy, port)
	if err != nil {
		log.Fatalln(err)
	}

	// Stop test node
	cliNode.Kill()
	cliNode.WaitExit()

	// Compare results
	// Test first transaction result trace, which should be
	// just a simple call to a contract function
	require.Equal(t, traceResult1.Result[0].TraceType, "call")

	// Test second transaction result trace, which should be
	// call to a contract, which will create a new contract and
	// call a two other functions on new contract
	require.Equal(t, traceResult2.Result[0].TraceType, "call")
	require.Equal(t, traceResult2.Result[1].TraceType, "create")
	require.Equal(t, traceResult2.Result[2].TraceType, "call")
	require.Equal(t, traceResult2.Result[2].TraceType, "call")

	// Check the addresses of inner traces
	require.Equal(t, len(traceResult2.Result[0].TraceAddress), 0)
	require.Equal(t, int(traceResult2.Result[1].TraceAddress[0]), 0)
	require.Equal(t, int(traceResult2.Result[2].TraceAddress[0]), 1)
	require.Equal(t, int(traceResult2.Result[3].TraceAddress[0]), 2)
}

func getTrace(txHash string, nodePort string) (response, error) {

	jsonStr := []byte(`{"method":"trace_transaction","params":["` + txHash + `"],"id":1,"jsonrpc":"2.0"}`)
	resp, err := http.Post("http://127.0.0.1:"+nodePort, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Fatalln(err)
	}

	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	//Convert the body to type string
	var res response
	err = json.Unmarshal(body, &res)
	return res, err
}

type response struct {
	Jsonrpc string
	Id      int64
	Result  []txtracer.ActionTrace
}
