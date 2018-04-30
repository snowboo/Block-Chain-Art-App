package main

import (
	"encoding/gob"
	"net"
	"crypto/elliptic"
	"os"
	"crypto/ecdsa"
	"fmt"
	"testing"
	"os/exec"
	"bytes"
	"path/filepath"
)


var localHost = "localhost"
var serverPort = "12345"

func TestMinerMiner(t *testing.T) {
	gob.Register(&net.TCPAddr{})
	gob.Register(&elliptic.CurveParams{})
	// usage: ./single_client [server-address]
	if len(os.Args) != 2 {
	}

	// this creates a directory (to be used as localPath) for each client.
	// The directories will have the format "./client{A,B}NNNNNNNNN", where
	// N is an arbitrary number. Feel free to change these local paths
	// to best fit your environment
	r, _ := os.Open("/dev/urandom")
	defer r.Close()
	cAPrivateKey, _ := ecdsa.GenerateKey(elliptic.P384(), r)
	fmt.Println(cAPrivateKey)

	//if err := clientA(serverAddr, localIpPort, cAPrivateKey); err != nil {
	//	reportError(err)
	//}

	//   go run ink-miner.go 127.0.0.1:12345 3076301006072a8648ce3d020106052b810400220362000461521b69e8fc90c3a87d194db94b61a1a09594e54b4602edb2a10f03b4d08d02016234b37ae3cc136dcef0e890786ff926acc74ad376eaeab9bf5fff92ba150685ba1a4918d2ba369b34c9b247f424c561d82f63ce43fd7e116f4871a9cdf9e5 3081a40201010430dd09bbc48d497df5fa20be98e42cc57b11705d324a1ecac4c04572897fa71accf45d69b90073bbc4f58fb67f235742c9a00706052b81040022a1640362000461521b69e8fc90c3a87d194db94b61a1a09594e54b4602edb2a10f03b4d08d02016234b37ae3cc136dcef0e890786ff926acc74ad376eaeab9bf5fff92ba150685ba1a4918d2ba369b34c9b247f424c561d82f63ce43fd7e116f4871a9cdf9e5
	minerPort1 := "12345"
	minerAddr1 := localHost + ":" + minerPort1
	pubKey1 := "3076301006072a8648ce3d020106052b810400220362000461521b69e8fc90c3a87d194db94b61a1a09594e54b4602edb2a10f03b4d08d02016234b37ae3cc136dcef0e890786ff926acc74ad376eaeab9bf5fff92ba150685ba1a4918d2ba369b34c9b247f424c561d82f63ce43fd7e116f4871a9cdf9e5"
	privKey1 := "3081a40201010430dd09bbc48d497df5fa20be98e42cc57b11705d324a1ecac4c04572897fa71accf45d69b90073bbc4f58fb67f235742c9a00706052b81040022a1640362000461521b69e8fc90c3a87d194db94b61a1a09594e54b4602edb2a10f03b4d08d02016234b37ae3cc136dcef0e890786ff926acc74ad376eaeab9bf5fff92ba150685ba1a4918d2ba369b34c9b247f424c561d82f63ce43fd7e116f4871a9cdf9e5"


	//minerPort2 := "5002"
	//minerAddr2 := localHost + ":" + minerPort2
	pubKey2 := "3076301006072a8648ce3d020106052b8104002203620004d15dd793e07cde2b3d892cfec2c3ea46e0a4a8da30f0ff4b21731f50743269f239a3b3c1ddeeb5920092fe9ef65fd9c46d7ddec3befdfcc7f732bd5c3f9dbe9f70aa8204e2ba21a62182576111ca66d1d575f1cafda47cb52d680629c0e9983d"
	privKey2 := "3081a402010104301f3db045d12d94a49a113e18e2f808ba62e3947acf9963e2e8f81b8b3ca73296e324029a4a9c5b160f4450dd920f46eda00706052b81040022a16403620004d15dd793e07cde2b3d892cfec2c3ea46e0a4a8da30f0ff4b21731f50743269f239a3b3c1ddeeb5920092fe9ef65fd9c46d7ddec3befdfcc7f732bd5c3f9dbe9f70aa8204e2ba21a62182576111ca66d1d575f1cafda47cb52d680629c0e9983d"


	//
	//go SetupTestInkMiner(minerAddr1, pubKey1, privKey1)
	//time.Sleep(1 * time.Second)
	//go SetupTestInkMiner(minerAddr2, pubKey2, privKey2)

	path := "/Users/trentyou/Documents/UBC/Winter 2018 2/CPSC416/Proj1/proj1_c6w8_d3n0b_o1a0b_t3b0b_v4c9/"
	//serverPath := "/Users/trentyou/Documents/UBC/Winter 2018 2/CPSC416/Proj1/proj1_c6w8_d3n0b_o1a0b_t3b0b_v4c9/proj1-server/"


	//cmd := exec.Command("go", "run",  serverPath + "server.go", "-c", serverPath + "config.json")
	cmd := exec.Command("go", "run",  path + "ink-miner.go", minerAddr1, pubKey1, privKey1)

	//time.Sleep(2 * time.Second)

	cmd2 := exec.Command("go", "run",  path + "ink-miner.go", minerAddr1, pubKey2, privKey2)


	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	fmt.Println("PATH: ", exPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	fmt.Println(cmd.Output())
	fmt.Println(stderr.String())

	if err != nil {
		fmt.Print(err)
	}

	var stderr2 bytes.Buffer
	cmd2.Stderr = &stderr2

	fmt.Println(cmd2.Output())
	fmt.Println(stderr2.String())
}

func StartMinerInstance() {
	minerPort1 := "12345"
	minerAddr1 := localHost + ":" + minerPort1
	pubKey1 := "3076301006072a8648ce3d020106052b810400220362000461521b69e8fc90c3a87d194db94b61a1a09594e54b4602edb2a10f03b4d08d02016234b37ae3cc136dcef0e890786ff926acc74ad376eaeab9bf5fff92ba150685ba1a4918d2ba369b34c9b247f424c561d82f63ce43fd7e116f4871a9cdf9e5"
	privKey1 := "3081a40201010430dd09bbc48d497df5fa20be98e42cc57b11705d324a1ecac4c04572897fa71accf45d69b90073bbc4f58fb67f235742c9a00706052b81040022a1640362000461521b69e8fc90c3a87d194db94b61a1a09594e54b4602edb2a10f03b4d08d02016234b37ae3cc136dcef0e890786ff926acc74ad376eaeab9bf5fff92ba150685ba1a4918d2ba369b34c9b247f424c561d82f63ce43fd7e116f4871a9cdf9e5"


	path := "/Users/trentyou/Documents/UBC/Winter 2018 2/CPSC416/Proj1/proj1_c6w8_d3n0b_o1a0b_t3b0b_v4c9/"

}
