/*
Structs for the Ink Miner class
*/

package shared

import (
	"crypto"
	"crypto/ecdsa"
	"math/big"
	"net"
	"net/rpc"
)

type Node struct {
	Block Block
	Prev  *Node
}

type Block struct {
	PreviousBlockHash string `json:"previous-block-hash"`
	IsNoopBlock       bool
	// Operations in the Block
	Operations []Operation

	// Key of miner who computed this block
	MinerKey ecdsa.PublicKey

	// Nonce computed for the block
	Nonce uint32

	// TODO: add the ink miner info here for validation
	// TODO: add the sercret string here for easy validation?
	Hash string
}

// Settings for a canvas in BlockArt.
type CanvasSettings struct {
	// Canvas dimensions
	CanvasXMax uint32 `json:"canvas-x-max"`
	CanvasYMax uint32 `json:"canvas-y-max"`
}

type MinerInfo struct {
	Address net.Addr
	Key     ecdsa.PublicKey
}

type BlockChainInit struct {
	BlockChain     Node
	PreviousHash   string
	AllShapes      map[string]Operation
	ExistingBlocks map[string]Block
}

// Settings for an instance of the BlockArt project/network.
type MinerNetSettings struct {
	// Hash of the very first (empty) block in the chain.
	GenesisBlockHash string `json:"genesis-block-hash"`

	// The minimum number of ink miners that an ink miner should be
	// connected to.
	MinNumMinerConnections uint8 `json:"min-num-miner-connections"`

	// Mining ink reward per op and no-op blocks (>= 1)
	InkPerOpBlock   uint32 `json:"ink-per-op-block"`
	InkPerNoOpBlock uint32 `json:"ink-per-no-op-block"`

	// Number of milliseconds between heartbeat messages to the server.
	HeartBeat uint32 `json:"heartbeat"`

	// Proof of work difficulty: number of zeroes in prefix (>=0)
	PoWDifficultyOpBlock   uint8 `json:"pow-difficulty-op-block"`
	PoWDifficultyNoOpBlock uint8 `json:"pow-difficulty-no-op-block"`

	// Canvas settings
	CanvasSettings CanvasSettings `json:"canvas-settings"`
}

type MinerSettings struct {
	// Hash of the very first (empty) block in the chain.
	GenesisBlockHash string `json:"genesis-block-hash"`

	// The minimum number of ink miners that an ink miner should be
	// connected to.
	MinNumMinerConnections uint8 `json:"min-num-miner-connections"`

	// Mining ink reward per op and no-op blocks (>= 1)
	InkPerOpBlock   uint32 `json:"ink-per-op-block"`
	InkPerNoOpBlock uint32 `json:"ink-per-no-op-block"`

	// Number of milliseconds between heartbeat messages to the server.
	HeartBeat uint32 `json:"heartbeat"`

	// Proof of work difficulty: number of zeroes in prefix (>=0)
	PoWDifficultyOpBlock   uint8 `json:"pow-difficulty-op-block"`
	PoWDifficultyNoOpBlock uint8 `json:"pow-difficulty-no-op-block"`
}

type PeerInfo struct {
	PubKey ecdsa.PublicKey
	Client *rpc.Client
}

type Operation struct {
	// An application shape operation (op)
	AppShapeOp string
	Fill       string // for ink calculation
	Stroke     string // for ink calculation
	DAttribute string // for ink calculation
	ShapeType  int    // 0 is PATH, 1 is CIRCLE
	IsDelete   bool   // true is delete, false is add operation

	// A public key of the art node that generated the op (used to validate op/op-sig)
	ArtNodeKey ecdsa.PublicKey

	NumBlockValidate uint8
	InkCost          uint32
	ShapeHash        string

	// A signature of the operation (op-sig)
	R *big.Int
	S *big.Int
}

type BlockNotChain struct {
	Blocks          map[string]Block
	InkUsedForBlock map[string]uint32
}

// These types and structs are for artminer

type Args struct {
	BlockHash string
	ShapeHash string
	// validateNum, operation, its hash, inkRequired and publicKey to miner (AddShape)
	ValidateNum     uint8
	OperationString string
	InkCost         uint32
	PublicKey       crypto.PublicKey
	R               *big.Int
	S               *big.Int
	Message         []byte
}

type OpenCanvasReply struct {
	KeyMatched       bool
	MyCanvasSettings CanvasSettings
}

type AddShapeReply struct {
	BlockHash           string
	InkRemaining        uint32
	ErrorCode           int
	OverlappedShapeHash string
}

type GetChildrenReply struct {
	Data  []string
	Found bool
}

type GetShapesReply struct {
	Data  []string
	Found bool
}

type GetSvgStringReply struct {
	Data  string
	Found bool
}

// TODO: This is Block struct
type BlockStruct struct {
	PrevHash      string
	ShapeHashList []string
}

// For shape map - key: shapeHash, value: ShapeStruct
// TODO: This is Operation struct
type ShapeStruct struct {
	OwnerKey  crypto.PublicKey
	SvgString string
	InkCost   uint32
	// ShapeInfo blockartlib.Shape
}

type CanvasRequestStruct struct {
}

type CanvasRequestReplyStruct struct {
	Canvas   *Node
	ShapeMap map[string]Operation
}
