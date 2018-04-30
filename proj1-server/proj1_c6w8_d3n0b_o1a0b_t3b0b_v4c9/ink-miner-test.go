// Usage:
//
// $ ./single_client [server-address]

// go run ink-miner.go ink-miner-test.go 127.0.0.1:12345
// Need to comment out function in the ink-miner.go file (main function)

package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"

	"./shared"
)

const (
	DEADLINE = "2018-02-16T23:59:59-08:00" // project deadline :-)
)

//////////////////////////////////////////////////////////////////////
// helper functions -- no need to look at these
type testLogger struct {
	prefix string
}

func NewLogger(prefix string) testLogger {
	return testLogger{prefix: prefix}
}

func (l testLogger) log(message string) {
	fmt.Printf("[%s][%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), l.prefix, message)
}

func (l testLogger) TestResult(description string, success bool) {
	var label string
	if success {
		label = "OK"
	} else {
		label = "ERROR"
	}

	l.log(fmt.Sprintf("%-70s%-10s", description, label))
}

func usage() {
	fmt.Fprintf(os.Stderr, "%s [server-address]\n", os.Args[0])
	os.Exit(1)
}

func reportError(err error) {
	timeWarning := []string{}

	deadlineTime, _ := time.Parse(time.RFC3339, DEADLINE)
	timeLeft := deadlineTime.Sub(time.Now())
	totalHours := timeLeft.Hours()
	daysLeft := int(totalHours / 24)
	hoursLeft := int(totalHours) - 24*daysLeft

	if daysLeft > 0 {
		timeWarning = append(timeWarning, fmt.Sprintf("%d days", daysLeft))
	}

	if hoursLeft > 0 {
		timeWarning = append(timeWarning, fmt.Sprintf("%d hours", hoursLeft))
	}

	timeWarning = append(timeWarning, fmt.Sprintf("%d minutes", int(timeLeft.Minutes())-60*int(totalHours)))
	warning := strings.Join(timeWarning, ", ")

	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	fmt.Fprintf(os.Stderr, "\nPlease fix the bug above and run this test again. Time remaining before deadline: %s\n", warning)
	os.Exit(1)
}

//////////////////////////////////////////////////////////////////////

func clientA(serverAddr, localIpPort string, privKey *ecdsa.PrivateKey) (err error) {
	logger := NewLogger("Client A")

	testCase := fmt.Sprintf("Resolving address ('%s')", serverAddr)
	addr1, err := net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		logger.TestResult(testCase, false)
	}

	testCase = fmt.Sprintf("Connection to address ('%s')", serverAddr)
	c, err := rpc.Dial("tcp", serverAddr)
	if err != nil {
		logger.TestResult(testCase, false)
	}
	defer c.Close()

	var settings shared.MinerNetSettings

	// normal registration
	testCase = fmt.Sprintf("Connection to server ('%s')", serverAddr)
	err = c.Call("RServer.Register", shared.MinerInfo{Address: addr1, Key: privKey.PublicKey}, &settings)
	if err != nil {
		logger.TestResult(testCase, false)
	}

	// Testing for hash validation (nonce and seret)
	testCase = fmt.Sprintf("Testing for hash to compute nonce('%s')", localIpPort)
	hash, firstBlock := CreateFirstBlock(settings.GenesisBlockHash, settings.PoWDifficultyNoOpBlock, privKey.PublicKey)
	secret := ConvertBlockToString(firstBlock)
	hash_val := ComputeNonceSecretHash(secret, strconv.Itoa(int(firstBlock.Nonce)))
	if hash != hash_val {
		logger.TestResult(testCase, false)
	}

	logger.TestResult(testCase, true)

	return
}

func main() {
	gob.Register(&net.TCPAddr{})
	gob.Register(&elliptic.CurveParams{})
	// usage: ./single_client [server-address]
	if len(os.Args) != 2 {
		usage()
	}

	serverAddr := os.Args[1]
	localIpPort := "127.0.0.1:9090"

	// this creates a directory (to be used as localPath) for each client.
	// The directories will have the format "./client{A,B}NNNNNNNNN", where
	// N is an arbitrary number. Feel free to change these local paths
	// to best fit your environment
	r, _ := os.Open("/dev/urandom")
	defer r.Close()
	cAPrivateKey, _ := ecdsa.GenerateKey(elliptic.P384(), r)

	if err := clientA(serverAddr, localIpPort, cAPrivateKey); err != nil {
		reportError(err)
	}

	fmt.Printf("\nCONGRATULATIONS! Your implementation correctly handles the single client scenario.\n")
}
