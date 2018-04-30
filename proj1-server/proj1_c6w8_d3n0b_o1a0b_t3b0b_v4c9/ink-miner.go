/*

Usage:

$ go run ink-miner.go
  -b int
        Heartbeat interval in ms (default 10)
  -i string
        RPC server ip:port
  -p int
        start port (default 54320)

  go run ink-miner.go [server ip:port] [pubKey] [privKey]

  go run ink-miner.go 127.0.0.1:12345 3076301006072a8648ce3d020106052b810400220362000461521b69e8fc90c3a87d194db94b61a1a09594e54b4602edb2a10f03b4d08d02016234b37ae3cc136dcef0e890786ff926acc74ad376eaeab9bf5fff92ba150685ba1a4918d2ba369b34c9b247f424c561d82f63ce43fd7e116f4871a9cdf9e5 3081a40201010430dd09bbc48d497df5fa20be98e42cc57b11705d324a1ecac4c04572897fa71accf45d69b90073bbc4f58fb67f235742c9a00706052b81040022a1640362000461521b69e8fc90c3a87d194db94b61a1a09594e54b4602edb2a10f03b4d08d02016234b37ae3cc136dcef0e890786ff926acc74ad376eaeab9bf5fff92ba150685ba1a4918d2ba369b34c9b247f424c561d82f63ce43fd7e116f4871a9cdf9e5

	go run ink-miner.go 127.0.0.1:12345 3076301006072a8648ce3d020106052b8104002203620004d15dd793e07cde2b3d892cfec2c3ea46e0a4a8da30f0ff4b21731f50743269f239a3b3c1ddeeb5920092fe9ef65fd9c46d7ddec3befdfcc7f732bd5c3f9dbe9f70aa8204e2ba21a62182576111ca66d1d575f1cafda47cb52d680629c0e9983d 3081a402010104301f3db045d12d94a49a113e18e2f808ba62e3947acf9963e2e8f81b8b3ca73296e324029a4a9c5b160f4450dd920f46eda00706052b81040022a16403620004d15dd793e07cde2b3d892cfec2c3ea46e0a4a8da30f0ff4b21731f50743269f239a3b3c1ddeeb5920092fe9ef65fd9c46d7ddec3befdfcc7f732bd5c3f9dbe9f70aa8204e2ba21a62182576111ca66d1d575f1cafda47cb52d680629c0e9983d

	go run ink-miner.go 127.0.0.1:12345 3076301006072a8648ce3d020106052b8104002203620004a765b4b29dc33e19668fbd7e2d92d38f27839082920c080cd3fd14c449f3093fe0a3c103551bd7c135426bfd089effbd61c3ccfb601373d4a26e6517648bf9ca20cc17e26c713f4fad01bc40ee4e48952e97f05f6fe98d8b4e25cacd992345b4 3081a40201010430c1114c64f035404c99fe6fd8bb0a18421c0de58310877e67ea768666adfd65b894cce83654e4002376731e080e10d33da00706052b81040022a16403620004a765b4b29dc33e19668fbd7e2d92d38f27839082920c080cd3fd14c449f3093fe0a3c103551bd7c135426bfd089effbd61c3ccfb601373d4a26e6517648bf9ca20cc17e26c713f4fad01bc40ee4e48952e97f05f6fe98d8b4e25cacd992345b4
*/

package main

import (
	"./collision"
	"./shared"
	"./verification"

	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/x509"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net"
	"net/rpc"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// temp global variables for art node - ink miner
// genesis block hash in minerNetSettings
var inkBank uint32 = 0

type ArtNodeMinerRPC int

/* ERRORS */
type NotFoundError string

func (e NotFoundError) Error() string {
	return fmt.Sprintf("Not found [%s]", string(e))
}

// Stores all the blocks in a chain format
// Need to still think of a way to implement when new block with same hash (which path do you build on?)
var existingBlockHashes map[string]shared.Block = make(map[string]shared.Block)
var blockChain shared.Node
var haveChain bool

var prevHash = minerNetSettings.GenesisBlockHash

var minerNetSettings shared.MinerNetSettings
var minerPrivateKey *ecdsa.PrivateKey

var peers = make(map[net.Addr]shared.PeerInfo)
var peerList []net.Addr
var minerInfo shared.MinerInfo
var myAddr *net.TCPAddr
var serverIP string

// Blocks that need to be verified that are not current in the chain yet
// var blocksNotInChain map[string]shared.Block = make(map[string]shared.Block)
var blocksNotInChain shared.BlockNotChain = shared.BlockNotChain{Blocks: make(map[string]shared.Block), InkUsedForBlock: make(map[string]uint32)}

var ExpectedError = errors.New("Expected error, none found")

var childrenMap map[string][]string = make(map[string][]string)

type NoopThread struct {
	sync.RWMutex
	runNoopGeneration bool
}

type BlockChainThread struct {
	sync.RWMutex
	accessToChain bool
}

type OpThread struct {
	sync.RWMutex
	operations map[string]shared.Operation
}

var noopThread = NoopThread{runNoopGeneration: true}
var blockChainThread = BlockChainThread{accessToChain: true}

// Operations that need to disseminated to other blocks
var opsNotInBlockThread = OpThread{operations: make(map[string]shared.Operation)}
var allShapes map[string]shared.Operation = make(map[string]shared.Operation)

func exitOnError(prefix string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s, err = %s\n", prefix, err.Error())
		os.Exit(1)
	}
}

type MinerRPC struct {
}

func (t *MinerRPC) Ping(message string, messageBack *string) error {
	fmt.Println("MinerRPC: Ping")
	*messageBack = message
	return nil
}

func (t *MinerRPC) Connect(fromAddr *net.TCPAddr, result *ecdsa.PublicKey) error {
	fmt.Println("MinerRPC: Connect", fromAddr)
	*result = minerInfo.Key
	return nil
}

func (t *MinerRPC) ConnectToMe(addr *net.TCPAddr, result *bool) error {
	if _, ok := peers[addr]; !ok {
		fmt.Println("Connect", addr)
		client, err := rpc.Dial("tcp", addr.String())

		// self connect to miner
		if err != nil {
			fmt.Println("Cannot connect to miner: ", addr.String())
		} else {
			defer client.Close()
			var pubKey ecdsa.PublicKey
			var peerInfo shared.PeerInfo
			err = client.Call("MinerRPC.Connect", myAddr, &pubKey)

			if err != nil {
				fmt.Println("MinerRPC.Connect RPC failed:", err)
				if _, ok := peers[addr]; ok {
					delete(peers, addr)
				}
			} else {
				peerInfo.Client = client
				peerInfo.PubKey = pubKey
				peers[addr] = peerInfo
				if !Contains(peerList, addr) {
					peerList = append(peerList, addr)
				}
				peerList = append(peerList, addr)
			}
		}
	}

	*result = true
	return nil
}

// Returns the current Canvas, along with a map of current shapes in the blockchain
// (shapeHash to Operation)
func (t *MinerRPC) GetCurrentCanvas(requestStruct shared.CanvasRequestStruct, replyStruct *shared.CanvasRequestReplyStruct) (err error) {
	reply := shared.CanvasRequestReplyStruct{Canvas: &blockChain}
	replyStruct = &reply

	return err
}

func (t *MinerRPC) FloodOperation(op shared.Operation, result *bool) error {
	op.ArtNodeKey.Curve = elliptic.P384()

	opsNotInBlockThread.Lock()
	_, ok := opsNotInBlockThread.operations[op.ShapeHash]
	if !ok {
		opsNotInBlockThread.operations[op.ShapeHash] = op
		for _, k := range peerList {
			// Flood neighnbours here
			client, err := rpc.Dial("tcp", k.String())
			// defer client.Close()
			if err != nil {
				fmt.Println("MinerRPC.FloodOperation: The miner is no longer online")
			} else {
				defer client.Close()
				var res bool
				err = client.Call("MinerRPC.FloodOperation", op, &res)

				if err != nil {
					fmt.Println("MinerRPC.FloodOperation failed:", err)
				}
			}

		}
	}
	opsNotInBlockThread.Unlock()
	*result = true
	return nil
}

// Floods block between network and waits for a positive verification from each
// peer before returning
func (t *MinerRPC) FloodBlock(block shared.Block, result *bool) error {
	block.MinerKey.Curve = elliptic.P384()
	_, ok := blocksNotInChain.Blocks[block.Hash]

	blocksNotInChain.Blocks[block.Hash] = block
	if !ok {
		if !verification.VerifyBlock(block, blockChain, minerNetSettings, existingBlockHashes, allShapes) {
			fmt.Println("FloodBlock MinerRPC.FloodBlock for verification failed")
			*result = false
			return nil
		}

		for _, k := range peerList {
			client, err := rpc.Dial("tcp", k.String())
			if err == nil {
				defer client.Close()
				var res bool
				err = client.Call("MinerRPC.FloodBlock", block, &res)
				if err == nil {
					continue
				} else {
					fmt.Println("FloodBlock MinerRPC.FloodBlock failed:", err)
					*result = false
					return nil
				}
			}
		}
	}
	*result = true
	return nil
}

// Floods block between network and waits for a positive verification from each
// peer before returning
func (t *MinerRPC) AddBlockFlood(block shared.Block, result *bool) error {
	block.MinerKey.Curve = elliptic.P384()

	_, ok := blocksNotInChain.Blocks[block.Hash]
	blocksNotInChain.Blocks[block.Hash] = block
	if !ok {
		UpdateBlockChain(block)
		for _, k := range peerList {
			client, err := rpc.Dial("tcp", k.String())
			if err == nil {
				defer client.Close()
				var res bool
				err = client.Call("MinerRPC.AddBlockFlood", block, &res)

				if err != nil {
					fmt.Println("MinerRPC.AddBlockFlood failed:", err)
					continue
				} else {
					*result = false
					return nil
				}
			}
		}
	}
	*result = true
	return nil
}

//
func (t *MinerRPC) AddBlockToBlockChain(block shared.Block, result *bool) error {
	_, ok := blocksNotInChain.Blocks[block.Hash]
	if ok {
		for _, k := range peerList {
			client, err := rpc.Dial("tcp", k.String())
			defer client.Close()
			if err != nil {
				fmt.Println("AddBlockToBlockChain: The miner is no longer online", err)
			} else {
				var res bool
				UpdateBlockChain(block)
				// delete the block and ink from maps in blocksNotInChain
				delete(blocksNotInChain.Blocks, block.Hash)
				delete(blocksNotInChain.InkUsedForBlock, block.Hash)
				err = client.Call("MinerRPC.AddBlockToBlockChain", block, &res)
			}
		}
	}
	*result = true
	return nil
}

func (t *MinerRPC) GetBlockChain(addr *net.TCPAddr, result *shared.BlockChainInit) error {
	if haveChain {
		blockChainRes := shared.BlockChainInit{blockChain, prevHash, allShapes, existingBlockHashes}
		*result = blockChainRes
		fmt.Println("Block chain obtained")
		return nil
	} else {
		for _, k := range peerList {
			if addr.String() != k.String() {
				client, err := rpc.Dial("tcp", k.String())

				if err != nil {
					fmt.Println("MinerRPC.GetBlockChain: The miner is no longer online")
				} else {
					defer client.Close()
					var res shared.BlockChainInit
					str := ""
					err = client.Call("MinerRPC.GetBlockChain", str, &res)
					if err == nil {
						haveChain = true
						*result = res
						return nil
					}
				}
			}
		}
	}
	return nil
}

func main() {

	priv1, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	fmt.Println("P256 PUB KEY: ", priv1.PublicKey)

	priv1, _ = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	fmt.Println("P384 PUB KEY: ", priv1.PublicKey)

	priv1, _ = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	fmt.Println("P521 PUB KEY: ", priv1.PublicKey)

	gob.Register(&net.TCPAddr{})
	gob.Register(&elliptic.CurveParams{})

	args := os.Args[1:]

	if len(args) != 3 {
		exitOnError("Incorrect number of arguments %d instead of 3.", nil)
	}

	serverIP = args[0]

	privateKeyBytesRestored, _ := hex.DecodeString(args[2])
	priv, _ := x509.ParseECPrivateKey(privateKeyBytesRestored)
	minerPrivateKey = priv

	addr, err := net.ResolveTCPAddr("tcp", serverIP)
	exitOnError("resolve addr", err)
	minerInfo.Key = priv.PublicKey
	minerInfo.Address = addr
	minerPrivateKey = priv

	SetupMiner(addr, serverIP, priv)

	time.Sleep(100000000000 * time.Millisecond)

	return

}

// Sets up the miner in a separate method, starts heartbeat, and gets the existing blockchain from
// other miners.
// Once the current blockchain is obtained from other miners, returns from method
func SetupMiner(localTCPAddr *net.TCPAddr, ipPort string, privKey *ecdsa.PrivateKey) {
	c, err := rpc.Dial("tcp", ipPort)
	exitOnError("rpc dial", err)
	defer c.Close()

	addrs, err := net.InterfaceAddrs()
	ln, err := net.Listen("tcp", ":0")
	exitOnError("net InterfaceAddrs", err)

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				minerIPString := ipnet.IP.String() + ":" + strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
				myAddr, _ = net.ResolveTCPAddr("tcp", minerIPString)
			}
		}
	}

	// set up miner rpc
	minerRPC := new(MinerRPC)
	rpc.Register(minerRPC)
	go rpc.Accept(ln)

	// set up artnode miner rpc
	artNodeMinerRPC := new(ArtNodeMinerRPC)
	rpc.RegisterName("ArtNodeMinerRPC", artNodeMinerRPC)
	go rpc.Accept(ln)

	err = c.Call("RServer.Register", shared.MinerInfo{Address: myAddr, Key: privKey.PublicKey}, &minerNetSettings)
	exitOnError(fmt.Sprintf("client registration for %s", localTCPAddr.String()), err)

	fmt.Println("Running miner at: ", myAddr)
	go RunHeartBeat(ipPort, privKey.PublicKey)

	// Get all Nodes
	var addrSet []net.Addr
	err = c.Call("RServer.GetNodes", privKey.PublicKey, &addrSet)
	exitOnError(fmt.Sprintf("Get nodes was unsuccessful with public key %s", localTCPAddr.String()), err)

	fmt.Println("PeerList: ", addrSet)
	for i := 0; i < len(addrSet); i++ {
		ConnectToMiners(addrSet[i], myAddr)
	}

	GetInitialBlockChain()
	go GenerateNoopBlock()

	return
}

func GetInitialBlockChain() {
	success := false
	for i := 0; i < len(peerList); i++ {
		client, err := rpc.Dial("tcp", peerList[i].String())
		if err == nil {
			var reply shared.BlockChainInit
			defer client.Close()
			err = client.Call("MinerRPC.GetBlockChain", myAddr, &reply)
			if err == nil {
				blockChain = reply.BlockChain
				prevHash = reply.PreviousHash
				allShapes = reply.AllShapes
				existingBlockHashes = reply.ExistingBlocks
				success = true
				haveChain = true
				break
			}
		}
	}
	if !success {
		genBlock := shared.Block{Hash: minerNetSettings.GenesisBlockHash}

		pHash := minerNetSettings.GenesisBlockHash
		blockChain = shared.Node{Block: genBlock}
		existingBlockHashes[pHash] = genBlock
		haveChain = true

		// Updates the previous hash here
		prevHash = genBlock.Hash
	}
}

func ConnectToMiners(peer net.Addr, minerIP *net.TCPAddr) {
	client, err := rpc.Dial("tcp", peer.String())
	// self connect to miner
	if err != nil {
		fmt.Println("Cannot connect to miner: ", peer.String())
	} else {
		defer client.Close()
		var pubKey ecdsa.PublicKey
		var peerInfo shared.PeerInfo
		err = client.Call("MinerRPC.Connect", minerIP, &pubKey)

		if err != nil {
			fmt.Println("MinerRPC.Connect RPC failed:", err)
			if _, ok := peers[peer]; ok {
				delete(peers, peer)
			}
		} else {
			peerInfo.Client = client
			peerInfo.PubKey = pubKey
			peers[peer] = peerInfo
			if !Contains(peerList, peer) {
				peerList = append(peerList, peer)
			}
		}
	}

	// tell miner to connect to self
	var result bool
	client.Call("MinerRPC.ConnectToMe", minerIP, &result)
}

func GetNodes() {
	count := uint8(0)
	var peersToRemove []net.Addr
	for i := 0; i < len(peerList); i++ {
		if count >= minerNetSettings.MinNumMinerConnections {
			break
		}
		neighbour, err := rpc.Dial("tcp", peerList[i].String())
		// self connect to miner
		if err != nil {
			fmt.Println("Cannot connect to miner: ", peerList[i].String())
			peersToRemove = append(peersToRemove, peerList[i])
			if _, ok := peers[peerList[i]]; ok {
				delete(peers, peerList[i])
			}
		} else {
			defer neighbour.Close()
			var reply string
			var message = "hi"
			err = neighbour.Call("MinerRPC.Ping", message, &reply)

			if err != nil {
				peersToRemove = append(peersToRemove, peerList[i])
				fmt.Println("MinerRPC.Ping RPC failed:", err)
				if _, ok := peers[peerList[i]]; ok {
					delete(peers, peerList[i])
				}
			} else if "hi" == reply {
				count++
			}
		}
	}

	//clean up peerList
	if len(peersToRemove) > 0 {
		var temp []net.Addr
		for _, peer := range peerList {
			remove := false
			for _, valueToRemove := range peersToRemove {
				if valueToRemove == peer {
					remove = true
					break
				}
			}
			if !remove {
				temp = append(temp, peer)
			}
		}

		peerList = temp
	}

	if count < minerNetSettings.MinNumMinerConnections {
		server, err := rpc.Dial("tcp", serverIP)
		if err == nil {
			defer server.Close()
			var addrSet []net.Addr
			err = server.Call("RServer.GetNodes", minerInfo.Key, &addrSet)
			exitOnError(fmt.Sprintf("Get nodes was unsuccessful with public key %s", serverIP), err)
			for i := 0; i < len(addrSet); i++ {
				ConnectToMiners(addrSet[i], myAddr)
			}
		}
	}
}

func RunHeartBeat(ipPort string, pubKey ecdsa.PublicKey) {
	var _ignored bool
	c, err := rpc.Dial("tcp", ipPort)
	exitOnError("rpc dial", err)
	for {
		err = c.Call("RServer.HeartBeat", pubKey, &_ignored)
		if err != nil {
			exitOnError("late heartbeat", ExpectedError)
		}
		time.Sleep(time.Duration(minerNetSettings.HeartBeat/2) * time.Millisecond)
	}
	c.Close()
}

// Function to flood the operation
func FloodOperation(op shared.Operation) bool {
	client, err := rpc.Dial("tcp", myAddr.String())

	if err != nil {
		fmt.Println("There was a problem with the connection")
	} else {
		defer client.Close()
		var reply bool
		client.Call("MinerRPC.FloodOperation", op, &reply)
		return reply
	}

	return false
}

// Function to flood the block
func FloodBlock(block shared.Block) (success bool) {
	client, err := rpc.Dial("tcp", myAddr.String())

	if err != nil {
		fmt.Println("There was a problem with the connection")
	} else {
		defer client.Close()
		var reply bool

		err = client.Call("MinerRPC.FloodBlock", block, &reply)
		if err == nil {
			if reply {
				var addBlock bool

				// Send out flooding notification to add to chain
				client.Call("MinerRPC.AddBlockFlood", block, &addBlock)
				return addBlock
			} else {
				return false
			}
		}
	}

	return false
}

func FloodBlockChain(b shared.Block) bool {
	client, err := rpc.Dial("tcp", myAddr.String())

	if err != nil {
		fmt.Println("There was a problem with the connection")
	} else {
		defer client.Close()
		var reply bool

		client.Call("MinerRPC.AddBlockToBlockChain", b, &reply)

		return reply
	}

	return false
}

// Returns the MD5 hash as a hex string for the (nonce + secret) value.
func ComputeNonceSecretHash(nonce string, secret string) string {
	h := md5.New()
	h.Write([]byte(nonce + secret))
	str := hex.EncodeToString(h.Sum(nil))
	return str
}

// Once we have a transaction, we can't carry it out
func FindSecret(nonce string, n uint8) (nonce_val uint32, hash string) {
	re := regexp.MustCompile("0{" + strconv.Itoa(int(n)) + "}")
	for i := 0; i < math.MaxInt64; i++ {
		i_to_string := strconv.Itoa(i)
		nonce_secret := ComputeNonceSecretHash(nonce, i_to_string)
		val := int(n)
		last_string := nonce_secret[len(nonce_secret)-val:]
		if re.MatchString(last_string) {
			return uint32(i), nonce_secret
		}
	}

	return 0, ""
}

func EncodePublicKey(key ecdsa.PublicKey) string {
	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(&key)
	encodedPubBytes := hex.EncodeToString(publicKeyBytes)
	return encodedPubBytes
}

// Generates a new Noop Block
func GenerateNoopBlock() (hash string, b shared.Block) {
	for {
		noopThread.Lock()
		if noopThread.runNoopGeneration {
			noopThread.Unlock()
			b = shared.Block{prevHash, true, nil, minerPrivateKey.PublicKey, 0, ""}
			blockString := ConvertBlockToString(b)
			nonce, hash := FindSecret(blockString, minerNetSettings.PoWDifficultyNoOpBlock)

			b.Nonce = nonce
			b.Hash = hash

			FloodBlock(b)

			if verification.VerifyBlock(b, blockChain, minerNetSettings, existingBlockHashes, allShapes) {
				existingBlockHashes[hash] = b
				inkBank += minerNetSettings.InkPerNoOpBlock
				prevHash = hash
				fmt.Println("Adding ink to bank, Bank:", inkBank)
			}

			UpdateBlockChain(b)

		} else {
			noopThread.Unlock()
			return "", shared.Block{}
		}
	}
	noopThread.Unlock()
	return "", shared.Block{}
}

//
func getOperationsToArrayToAddBlock() ([]shared.Operation, bool) {
	opsNotInBlockThread.Lock()

	result := make([]shared.Operation, 0)
	for k, op := range opsNotInBlockThread.operations {
		result = append(result, op)
		delete(opsNotInBlockThread.operations, k)
		intersect, _ := HasIntersection(op, opsNotInBlockThread.operations)
		if intersect {
			opsNotInBlockThread.Unlock()
			return nil, false
		}
	}
	opsNotInBlockThread.operations = make(map[string]shared.Operation)
	opsNotInBlockThread.Unlock()
	return result, false
}

// Function to update the block chain
// The block is added to the childrenMap under its parent, which is the previous block
func UpdateBlockChain(block shared.Block) {
	// Updates the block chain
	tempBlockChain := blockChain
	newBlockChain := shared.Node{Block: block, Prev: &tempBlockChain}
	blockChain = newBlockChain
	// Updates the previous hash here
	prevHash = block.Hash
	// Updates the block hashes that exist
	existingBlockHashes[block.Hash] = block

	hash_val := block.PreviousBlockHash

	childrenList, ok := childrenMap[hash_val]
	if ok {
		// update the map here
		new_list := childrenList
		contains := false
		for _, value := range childrenList {
			if value == prevHash {
				contains = true
			}
		}

		if !contains {
			new_list = append(new_list, prevHash)
		}
		childrenMap[hash_val] = new_list
	} else {
		new_list := []string{prevHash}
		childrenMap[hash_val] = new_list
	}
}

// Function to generate the operation blocks
func GenerateOpBlock() (hash string, b shared.Block, success bool) {
	fmt.Println("Generate Op Block")

	b = shared.Block{PreviousBlockHash: prevHash, MinerKey: minerPrivateKey.PublicKey}
	operationsToAdd, intersection := getOperationsToArrayToAddBlock()
	if intersection {
		return "", shared.Block{}, false
	}
	b.Operations = operationsToAdd
	blockString := ConvertBlockToString(b)
	nonce, hash := FindSecret(blockString, minerNetSettings.PoWDifficultyOpBlock)
	b.Nonce = nonce
	b.Hash = hash

	success = FloodBlock(b)

	// valid := true
	if verification.VerifyBlock(b, blockChain, minerNetSettings, existingBlockHashes, allShapes) {
		inkBank += minerNetSettings.InkPerOpBlock
	}

	// Adding the new block to the block chain
	if success {
		UpdateBlockChain(b)
	}

	fmt.Println("Done GenerateOpBlock")

	// verify block here
	return hash, b, success
}

// Helper function to validate validNum
// This function checks every child block and does a traversal
func haveChildren(blockHash string, numValidate uint8) bool {
	childrenList := childrenMap[blockHash]
	for _, child := range childrenList {
		validate := travserseToParent(blockHash, child, numValidate)
		if validate {
			return true
		}
	}
	return false
}

// Helper function to validate validNum
// This function traverses the previous blocks and checks to see if
// the number of blocks >= validNum
func travserseToParent(blockHash string, child string, numValidate uint8) bool {
	prevBlocks := 0
	childBlock := blocksNotInChain.Blocks[child]
	if blockHash == childBlock.PreviousBlockHash {
		prevBlocks += 1
		if prevBlocks >= int(numValidate) {
			return true
		} else {
			return false
		}
	}

	return travserseToParent(blockHash, childBlock.PreviousBlockHash, numValidate)
}

// Converts the entire block into a string so that it can be used in hashing of computation of the nonce
func ConvertBlockToString(b shared.Block) string {
	eMKey := EncodePublicKey(b.MinerKey)
	opString := ""
	for i := 0; i < len(b.Operations); i++ {
		opString = opString + b.Operations[i].AppShapeOp + b.Operations[i].R.String() + b.Operations[i].S.String()
	}
	return b.PreviousBlockHash + opString + eMKey
}

// Adds the r and s values which are big ints that ecdsa.Sign provides into the operation
func AddValuesToOp(op shared.Operation, pkey *ecdsa.PrivateKey) {
	data := []byte(op.AppShapeOp)
	r, s, _ := ecdsa.Sign(rand.Reader, pkey, data)
	op.R = r
	op.S = s
}

// Checks that the number of blocks following the current block that the
// operation is in is equal or more than the n (numValidateBlock value).
func CheckBlockNumValidateWithOp(n uint8, blockHash string) bool {
	// access the block for the operation
	return true
}

// Checks if there are any intersections with the shapes on the current canvas and the one to
// be added onto the canvas
func HasIntersection(op shared.Operation, shapes map[string]shared.Operation) (bool, string) {
	return collision.CollideWithOtherShapes(op, shapes)
}

// Attemps to add a new operation to the block - if valid, will return true
// Valid means that it does not intersect with any other blocks
// and that there is enough ink for the miner to draw the shape.
// Ink calculation need to consider the ink from the blockchain as well as
// the operations in the blocks that are yet to be added.
// We will check the ink when we add the operation, and then re-evaluate the miner's
// ink bank when we add the block to the blockchain
func AddOperationHelper(op shared.Operation) (valid bool, blockHash string) {
	if op.ArtNodeKey != minerPrivateKey.PublicKey {
		// It was not signed with the correct key
		return false, "Keys don't match"
	}

	// check intersections
	intersected, shapeHashCollided := HasIntersection(op, allShapes)

	if intersected {
		fmt.Println("Shape intersected", shapeHashCollided)
		return false, shapeHashCollided
	}

	// Flooding protocol
	opsNotInBlockThread.Lock()
	opsNotInBlockThread.operations[op.ShapeHash] = op
	opsNotInBlockThread.Unlock()

	FloodOperation(op)

	// Generation of the op block
	noopThread.Lock()
	noopThread.runNoopGeneration = false
	noopThread.Unlock()
	blockHash, block, success := GenerateOpBlock()

	enoughNumBlock := false
	for !enoughNumBlock {
		enoughNumBlock = CheckBlockNumValidateWithOp(op.NumBlockValidate, blockHash)
	}

	// Check to ensure that block has numValidate
	existingBlockHashes[block.Hash] = block
	allShapes[op.ShapeHash] = op

	noopThread.Lock()
	noopThread.runNoopGeneration = true
	noopThread.Unlock()

	go GenerateNoopBlock()

	return success, blockHash
}

func DeleteOperationHelper(op shared.Operation) bool {
	opsNotInBlockThread.Lock()
	opsNotInBlockThread.operations[op.ShapeHash] = op
	opsNotInBlockThread.Unlock()

	FloodOperation(op)

	// lookup block hash here
	blockHash := "temp"

	noopThread.Lock()
	noopThread.runNoopGeneration = false
	noopThread.Unlock()

	blockHash, block, success := GenerateOpBlock()
	if success {
		delete(allShapes, op.ShapeHash)

		existingBlockHashes[block.Hash] = block
		allShapes[op.ShapeHash] = op

		enoughNumBlock := false
		for !enoughNumBlock {
			enoughNumBlock = CheckBlockNumValidateWithOp(op.NumBlockValidate, blockHash)
		}

		noopThread.Lock()
		noopThread.runNoopGeneration = true
		noopThread.Unlock()

		go GenerateNoopBlock()

		return success
	} else {
		noopThread.Lock()
		noopThread.runNoopGeneration = true
		noopThread.Unlock()

		go GenerateNoopBlock()
		return false
	}

}

func Contains(peers []net.Addr, element net.Addr) bool {
	for _, value := range peers {
		if element.String() == value.String() {
			return true
		}
	}
	return false
}

// Helper function which calculates the ink usage for blocks
// that have not yet been validated in the blockchain
func CalculateInkForFutureBlocks() (inkCost uint32) {
	var totalInk uint32
	for _, block := range blocksNotInChain.Blocks {
		bHash := block.Hash
		operations := block.Operations
		inkUsed, ok := blocksNotInChain.InkUsedForBlock[bHash]
		// if ink used has been previously calculated
		if ok {
			totalInk += inkUsed
		} else {
			// calculate the ink used by the operations in the block
			for _, op := range operations {
				var inkUsage uint32
				matched := verification.EqualPublicKey(op.ArtNodeKey, minerPrivateKey.PublicKey)
				if matched && !op.IsDelete {
					inkUsage += op.InkCost
				} else if matched && op.IsDelete {
					inkUsage -= op.InkCost
				}
				blocksNotInChain.InkUsedForBlock[bHash] = inkUsage
				totalInk += inkUsage
			}
		}
	}
	return totalInk
}

// ===================== Art Node - Ink Miner RPC Functions =======================================

// args: artnode's pubkey
// reply: KeyMatched, CanvasSettings
func (t *ArtNodeMinerRPC) OpenCanvasRPC(args *shared.Args, reply *shared.OpenCanvasReply) error {
	//validate the artnode's message and signature
	reply.KeyMatched = ecdsa.Verify(&minerPrivateKey.PublicKey, args.Message, args.R, args.S)
	reply.MyCanvasSettings = minerNetSettings.CanvasSettings
	return nil
}

// args: validateNum, operation, its hash, inkRequired, artnode's publicKey
// reply: blockHash, inkRemaining, errorCode (-1 for InsufficientInkError, -2 for ShapeOverlapError)

func (t *ArtNodeMinerRPC) AddShapeRPC(args *shared.Operation, reply *shared.AddShapeReply) error {

	// check ink amount
	minerTotalInk := inkBank + CalculateInkForFutureBlocks()
	if args.InkCost > minerTotalInk {
		reply.ErrorCode = -1 // InsufficientInkError
		return nil
	}

	args.ArtNodeKey = minerPrivateKey.PublicKey
	//TODO make opSig in blockartlib
	isOK, str := AddOperationHelper(*args)

	if isOK {
		inkBank -= args.InkCost
		reply.InkRemaining = inkBank + CalculateInkForFutureBlocks()
		allShapes[args.ShapeHash] = *args
		reply.BlockHash = str

	} else if len(str) > 0 {
		reply.OverlappedShapeHash = str
		reply.ErrorCode = -2
	}

	return nil
}

// args: shapeHash
// reply: shape's svgstring and confirmation if it's found
func (t *ArtNodeMinerRPC) GetSvgStringRPC(args *shared.Args, reply *shared.GetSvgStringReply) error {
	_, ok := allShapes[args.ShapeHash]
	if !ok {
		// if does not exists return empty reply
		return nil
	}
	// grab the SVG string for this shape
	thisShape := allShapes[args.ShapeHash]
	reply.Data = thisShape.AppShapeOp
	reply.Found = true
	return nil
}

// args: none
// reply: inkRemaining
func (t *ArtNodeMinerRPC) GetInkRPC(args *shared.Args, reply *uint32) error {
	// should we take the ink used/returned in blocks not added into the blockchain into consideration?
	*reply = inkBank + CalculateInkForFutureBlocks()
	return nil
}

// args: shapeHash, validateNum
// reply: inkRemaining
func (t *ArtNodeMinerRPC) DeleteShapeRPC(args *shared.Operation, reply *uint32) error {
	// TODO make opSig in blockartlib
	args.ArtNodeKey = minerPrivateKey.PublicKey
	isOK := DeleteOperationHelper(*args)
	if isOK {
		inkBank += allShapes[args.ShapeHash].InkCost
		*reply = inkBank + CalculateInkForFutureBlocks()
		delete(allShapes, args.ShapeHash)
	}

	return nil
}

// args: blockHash
// reply: shapeHashes []string and confirmation if block's found
func (t *ArtNodeMinerRPC) GetShapesRPC(args *shared.Args, reply *shared.GetShapesReply) error {
	fmt.Println("Get Shapes RPC")
	_, ok := existingBlockHashes[args.BlockHash]
	if !ok {
		// if does not exists return empty reply
		return nil
	}
	// grab the shape hash list of this block
	thisBlockOperations := existingBlockHashes[args.BlockHash].Operations

	for _, op := range thisBlockOperations {
		if !op.IsDelete {
			reply.Data = append(reply.Data, op.ShapeHash)
		}
	}

	reply.Found = true
	return nil
}

// args: none
// reply: GenesisBlockHash from MinerNetSettings
func (t *ArtNodeMinerRPC) GetGenesisBlockRPC(args *shared.Args, reply *string) error {
	fmt.Println("Get GetGenesisBlock RPC")
	*reply = minerNetSettings.GenesisBlockHash
	return nil
}

// args: blockHash
// reply: blockHashes []string and confirmation if block's found
func (t *ArtNodeMinerRPC) GetChildrenRPC(args *shared.Args, reply *shared.GetChildrenReply) error {
	// TODO change blocks into existingBlockHashes eventually
	fmt.Println("Get Children RPC")
	fmt.Println(childrenMap)
	_, ok := childrenMap[args.BlockHash]
	if !ok {
		// if does not exists return empty reply
		return nil
	}
	// loop through list of blocks
	reply.Data = childrenMap[args.BlockHash]
	reply.Found = true
	return nil
}

// args: none
// reply: inkRemaining
func (t *ArtNodeMinerRPC) CloseCanvasRPC(args *shared.Args, reply *uint32) error {
	*reply = inkBank + CalculateInkForFutureBlocks()
	return nil
}
