/*

This package specifies the application's interface to the the BlockArt
library (blockartlib) to be used in project 1 of UBC CS 416 2017W2.

*/

package blockartlib

import (
	"../shared"
	"crypto/ecdsa"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"net/rpc"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Represents a type of shape in the BlockArt system.
type ShapeType int

var mAddr string

const (
	// Path shape.
	PATH ShapeType = iota

	// Circle shape (extra credit).
	// CIRCLE
)

// Settings for a canvas in BlockArt.
type CanvasSettings struct {
	// Canvas dimensions
	CanvasXMax uint32
	CanvasYMax uint32
}

// Settings for an instance of the BlockArt project/network.
type MinerNetSettings struct {
	// Hash of the very first (empty) block in the chain.
	GenesisBlockHash string

	// The minimum number of ink miners that an ink miner should be
	// connected to. If the ink miner dips below this number, then
	// they have to retrieve more nodes from the server using
	// GetNodes().
	MinNumMinerConnections uint8

	// Mining ink reward per op and no-op blocks (>= 1)
	InkPerOpBlock   uint32
	InkPerNoOpBlock uint32

	// Number of milliseconds between heartbeat messages to the server.
	HeartBeat uint32

	// Proof of work difficulty: number of zeroes in prefix (>=0)
	PoWDifficultyOpBlock   uint8
	PoWDifficultyNoOpBlock uint8

	// Canvas settings
	canvasSettings CanvasSettings
}

// For shape creation
type Shape struct {
	SvgString  string
	ShapeType  ShapeType
	DAttribute string
	Fill       string
	Stroke     string
	InkCost    uint32
}

var myShapes = make(map[string]Shape)
var canvasSettings CanvasSettings

////////////////////////////////////////////////////////////////////////////////////////////
// <ERROR DEFINITIONS>

// These type definitions allow the application to explicitly check
// for the kind of error that occurred. Each API call below lists the
// errors that it is allowed to raise.
//
// Also see:
// https://blog.golang.org/error-handling-and-go
// https://blog.golang.org/errors-are-values

// Contains address IP:port that art node cannot connect to.
type DisconnectedError string

func (e DisconnectedError) Error() string {
	return fmt.Sprintf("BlockArt: cannot connect to [%s]", string(e))
}

// Contains amount of ink remaining.
type InsufficientInkError uint32

func (e InsufficientInkError) Error() string {
	return fmt.Sprintf("BlockArt: Not enough ink to addShape [%d]", uint32(e))
}

// Contains the offending svg string.
type InvalidShapeSvgStringError string

func (e InvalidShapeSvgStringError) Error() string {
	return fmt.Sprintf("BlockArt: Bad shape svg string [%s]", string(e))
}

// Contains the offending svg string.
type ShapeSvgStringTooLongError string

func (e ShapeSvgStringTooLongError) Error() string {
	return fmt.Sprintf("BlockArt: Shape svg string too long [%s]", string(e))
}

// Contains the bad shape hash string.
type InvalidShapeHashError string

func (e InvalidShapeHashError) Error() string {
	return fmt.Sprintf("BlockArt: Invalid shape hash [%s]", string(e))
}

// Contains the bad shape hash string.
type ShapeOwnerError string

func (e ShapeOwnerError) Error() string {
	return fmt.Sprintf("BlockArt: Shape owned by someone else [%s]", string(e))
}

// Empty
type OutOfBoundsError struct{}

func (e OutOfBoundsError) Error() string {
	return fmt.Sprintf("BlockArt: Shape is outside the bounds of the canvas")
}

// Contains the hash of the shape that this shape overlaps with.
type ShapeOverlapError string

func (e ShapeOverlapError) Error() string {
	return fmt.Sprintf("BlockArt: Shape overlaps with a previously added shape [%s]", string(e))
}

// Contains the invalid block hash.
type InvalidBlockHashError string

func (e InvalidBlockHashError) Error() string {
	return fmt.Sprintf("BlockArt: Invalid block hash [%s]", string(e))
}

type InvalidArtNodeMinerKeyPairError struct{}

func (e InvalidArtNodeMinerKeyPairError) Error() string {
	return fmt.Sprintf("BlockArt: Art Node and Miner Key Pairs did not match")
}

// </ERROR DEFINITIONS>
////////////////////////////////////////////////////////////////////////////////////////////

// Represents a canvas in the system.
type Canvas interface {
	// Adds a new shape to the canvas.
	// Can return the following errors:
	// - DisconnectedError
	// - InsufficientInkError
	// - InvalidShapeSvgStringError
	// - ShapeSvgStringTooLongError
	// - ShapeOverlapError
	// - OutOfBoundsError
	AddShape(validateNum uint8, shapeType ShapeType, shapeSvgString string, fill string, stroke string) (shapeHash string, blockHash string, inkRemaining uint32, err error)

	// Returns the encoding of the shape as an svg string.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidShapeHashError
	GetSvgString(shapeHash string) (svgString string, err error)

	// Returns the amount of ink currently available.
	// Can return the following errors:
	// - DisconnectedError
	GetInk() (inkRemaining uint32, err error)

	// Removes a shape from the canvas.
	// Can return the following errors:
	// - DisconnectedError
	// - ShapeOwnerError
	DeleteShape(validateNum uint8, shapeHash string) (inkRemaining uint32, err error)

	// Retrieves hashes contained by a specific block.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidBlockHashError
	GetShapes(blockHash string) (shapeHashes []string, err error)

	// Returns the block hash of the genesis block.
	// Can return the following errors:
	// - DisconnectedError
	GetGenesisBlock() (blockHash string, err error)

	// Retrieves the children blocks of the block identified by blockHash.
	// Can return the following errors:
	// - DisconnectedError
	// - InvalidBlockHashError
	GetChildren(blockHash string) (blockHashes []string, err error)

	// Closes the canvas/connection to the BlockArt network.
	// - DisconnectedError
	CloseCanvas() (inkRemaining uint32, err error)
}

// SVG Helper Function
// Check whether Svg is valid
// Return ShapeSvgStringTooLongError if len is more than 128
// Return InvalidShapeSvgStringError if it is an invalid Svg Path
func IsValidSvgShape(shapeType ShapeType, shapeSvgString string, fill string, stroke string) (err error, success bool) {
	if shapeType == PATH {
		if len(shapeSvgString) > 128 {
			return ShapeSvgStringTooLongError(shapeSvgString), false
		}

		if fill == "transparent" && stroke == "transparent" {
			return InvalidShapeSvgStringError(shapeSvgString), false
		}

		// add an additional space as our regex needs to end with a white space
		testString := shapeSvgString + " "
		regex := `^[M](\s-?\d+){2}\s(([m](\s-?\d+){2}|[Ll](\s-?\d+){2}|[Hh](\s-?\d+){1}|[Vv](\s-?\d+){1}|[Zz])\s)*$`
		r, _ := regexp.Compile(regex)
		valid := r.MatchString(testString)
		if !valid {
			return InvalidShapeSvgStringError(shapeSvgString), false
		}

		// check bounds
		x_coor, y_coor := SvgToPoints(shapeSvgString)
		if len(x_coor) == len(y_coor) {
			for i := 0; i < len(x_coor); i++ {
				if uint32(x_coor[i]) > canvasSettings.CanvasXMax || uint32(x_coor[i]) < 0 || uint32(y_coor[i]) > canvasSettings.CanvasYMax || uint32(y_coor[i]) < 0 {
					return new(OutOfBoundsError), false
				}
			}
		}
	}
	return nil, true
}

// Add Shape to map of local shapes
func AddShape(inkCost uint32, shapeHash string, shapeType ShapeType, shapeSvgString string, fill string, stroke string) (err error, success bool) {
	shape := Shape{
		SvgString:  "<path d=\"" + shapeSvgString + "\" stroke=\"" + stroke + "\" fill=\"" + fill + "\"/>",
		DAttribute: shapeSvgString,
		ShapeType:  shapeType,
		Fill:       fill,
		Stroke:     stroke,
		InkCost:    inkCost}
	myShapes[shapeHash] = shape
	return nil, true
}

func CalculateInkUsed(shapeType ShapeType, shapeSvgString string, fill string, stroke string) (inkUsed uint32) {
	if shapeType == PATH {
		x_coor, y_coor := SvgToPoints(shapeSvgString)
		var inkUsedFloat float64 = 0
		if len(x_coor) == len(y_coor) {
			if fill == "transparent" && stroke != "transparent" {
				if len(x_coor) == 0 {
					return 0
				} else if len(x_coor) == 1 {
					return 1
				} else if len(x_coor) > 0 {
					x := x_coor[0]
					y := y_coor[0]
					for i := 1; i < len(x_coor); i++ {
						// Pythagoras
						inkUsedFloat += math.Sqrt(math.Pow(math.Abs(float64(x_coor[i]-x)), 2) + math.Pow(math.Abs(float64(y_coor[i]-y)), 2))
						x = x_coor[i]
						y = y_coor[i]
					}
					return uint32(math.Ceil(inkUsedFloat))
				}
			} else if fill != "transparent" && stroke == "transparent" {
				// get area of polygon
				j := len(y_coor) - 1
				for i := 0; i < len(x_coor); i++ {
					inkUsedFloat += float64((x_coor[j] + x_coor[i]) * (y_coor[j] - y_coor[i]))
					j = i
				}
				inkUsed = uint32(math.Ceil(math.Abs(inkUsedFloat / 2)))
				return inkUsed
			} else if fill != "transparent" && stroke != "transparent" {
				// length
				if len(x_coor) == 0 {
					return 0
				} else if len(x_coor) == 1 {
					return 1
				} else if len(x_coor) > 0 {
					x := x_coor[0]
					y := y_coor[0]
					for i := 1; i < len(x_coor); i++ {
						// Pythagoras
						inkUsedFloat += math.Sqrt(math.Pow(math.Abs(float64(x_coor[i]-x)), 2) + math.Pow(math.Abs(float64(y_coor[i]-y)), 2))
						x = x_coor[i]
						y = y_coor[i]
					}
				}

				inkUsed = uint32(inkUsedFloat)
				// area
				inkUsedFloat = 0
				j := len(y_coor) - 1
				for i := 0; i < len(x_coor); i++ {
					inkUsedFloat += float64((x_coor[j] + x_coor[i]) * (y_coor[j] - y_coor[i]))
					j = i
				}

				// total
				inkUsed += uint32(math.Ceil(math.Abs(inkUsedFloat / 2)))
				return inkUsed
			}
		}

	}
	return 0
}

func SvgToPoints(shapeSvgString string) (x_coor []int, y_coor []int) {
	sliceWithSpace := strings.Split(shapeSvgString, " ")

	// remove slice that contains white spaces
	var slice []string
	for _, str := range sliceWithSpace {
		if str != "" {
			slice = append(slice, str)
		}
	}

	for i := 0; i < len(slice); i++ {
		// 2 args
		if slice[i] == "M" || slice[i] == "L" {
			i++
			x, _ := strconv.Atoi(string(slice[i]))
			i++
			y, _ := strconv.Atoi(string(slice[i]))
			x_coor = append(x_coor, x)
			y_coor = append(y_coor, y)
		} else if string(slice[i]) == "H" {
			i++
			x, _ := strconv.Atoi(string(slice[i]))
			x_coor = append(x_coor, x)
			y_coor = append(y_coor, y_coor[len(y_coor)-1])
		} else if slice[i] == "V" {
			i++
			y, _ := strconv.Atoi(string(slice[i]))
			x_coor = append(x_coor, x_coor[len(x_coor)-1])
			y_coor = append(y_coor, y)
		} else if slice[i] == "m" || slice[i] == "l" {
			i++
			x, _ := strconv.Atoi(string(slice[i]))
			i++
			y, _ := strconv.Atoi(string(slice[i]))
			x_coor = append(x_coor, x+x_coor[len(x_coor)-1])
			y_coor = append(y_coor, y+y_coor[len(y_coor)-1])
		} else if slice[i] == "h" {
			i++
			x, _ := strconv.Atoi(string(slice[i]))
			x_coor = append(x_coor, x+x_coor[len(x_coor)-1])
			y_coor = append(y_coor, y_coor[len(y_coor)-1])
		} else if slice[i] == "v" {
			i++
			y, _ := strconv.Atoi(string(slice[i]))
			x_coor = append(x_coor, x_coor[len(x_coor)-1])
			y_coor = append(y_coor, y+y_coor[len(y_coor)-1])
		} else if slice[i] == "Z" || slice[i] == "z" {
			// get the first point
			if len(x_coor) > 0 && len(y_coor) > 0 {
				x_coor = append(x_coor, x_coor[0])
				y_coor = append(y_coor, y_coor[0])
			}
		}
	}

	return x_coor, y_coor
}

// Remove Shape from the list of local shape
func DeleteShape(shapeHash string) (success bool) {
	if _, ok := myShapes[shapeHash]; ok {
		delete(myShapes, shapeHash)
		return true
	}
	return false
}

func CreateHtmlFile(shapes []string) (success bool) {
	f, err := os.Create("./output.html")
	if err == nil {
		_, err = f.WriteString("<svg height=" + fmt.Sprint(canvasSettings.CanvasYMax) + " width=" + fmt.Sprint(canvasSettings.CanvasYMax) + ">\n")
		for i := 0; i < len(shapes); i++ {
			_, err = f.WriteString(shapes[i] + "\n")
		}
		_, err = f.WriteString("</svg>\n")
		f.Sync()
		f.Close()
		return true
	}
	f.Sync()
	f.Close()
	fmt.Println(err)
	return false
}

type canvasStruct struct {
	// CanvasSettings?
	// shapes belonging to this instance - use map with shapeHash as key and {svg string and ink cost} as value?
	// RPC connection similar to dfslib
	MyCanvasSettings *shared.CanvasSettings
	Miner            *rpc.Client
	PrivKey          ecdsa.PrivateKey
}

// needs to be accessed by miner
type ConnectionReplyArgs struct {
	MyCanvasSettings *CanvasSettings
	KeyMatched       bool
}

type ConnectionArgs struct {
	PrivKey ecdsa.PrivateKey
}

// Shape creation struct - Shape

// The constructor for a new Canvas object instance. Takes the miner's
// IP:port address string and a public-private key pair (ecdsa private
// key type contains the public key). Returns a Canvas instance that
// can be used for all future interactions with blockartlib.
//
// The returned Canvas instance is a singleton: an application is
// expected to interact with just one Canvas instance at a time.
//
// Can return the following errors:
// - DisconnectedError
func OpenCanvas(minerAddr string, privKey ecdsa.PrivateKey) (canvas Canvas, setting CanvasSettings, err error) {

	// establish RPC connection with miner using minerAddr for connection and privkey as args (Connect)
	// return disconnected error if cannot connect to miner
	// validate that this miner mines on behalf of the right key pair (whatever that means)
	// initialize or get existing instance of canvas, I assume Canvas instance is singleton per artnode
	// wait for miner to respond with settings and then return
	mAddr = minerAddr
	miner, err := rpc.Dial("tcp", minerAddr)
	if err != nil {
		fmt.Println("Dialing Error", err)
	}
	// verify miner
	msg := []byte("Hello")
	r, s, _ := ecdsa.Sign(rand.Reader, &privKey, msg)
	reply := shared.OpenCanvasReply{}
	args := shared.Args{R: r, S: s, Message: msg}

	err = miner.Call("ArtNodeMinerRPC.OpenCanvasRPC", &args, &reply)
	if err != nil {
		fmt.Println(err)
		return nil, CanvasSettings{}, DisconnectedError(minerAddr)
	}

	fmt.Println("Keys matched?:", reply.KeyMatched)

	//Key pair did not match with the miner
	if reply.KeyMatched != true {
		fmt.Println("Keys did not matched: ", reply.KeyMatched)
		return nil, CanvasSettings{}, new(InvalidArtNodeMinerKeyPairError)
	}

	canvasSettings = CanvasSettings(reply.MyCanvasSettings)
	canvasInstance := canvasStruct{&reply.MyCanvasSettings, miner, privKey}

	return canvasInstance, canvasSettings, nil
}

// Adds a new shape to the canvas.
// Can return the following errors:
// - DisconnectedError
// - InsufficientInkError
// - InvalidShapeSvgStringError
// - ShapeSvgStringTooLongError
// - ShapeOverlapError
// - OutOfBoundsError
func (canvas canvasStruct) AddShape(validateNum uint8, shapeType ShapeType, shapeSvgString string, fill string, stroke string) (shapeHash string, blockHash string, inkRemaining uint32, err error) {
	// if length of shapeSvgString > 128, return ShapeSvgStringTooLongError

	// for now all svg is PATH, so care about shapeType if we do extra
	// send fill, stroke and shapeSvgString to svg package to get ink amount needed to draw and an actual operation like <path d="M 0 0 L 20 20" stroke="red" fill="transparent"/>
	// svg package will check those things:
	// - if shapeSvgString is invalid, return InvalidShapeSvgStringError (shapeSvgString represents 'd' attribute of svg. how to check if svg is invalid?)
	// - if the path needs to go beyond CanvasSettings x and y, return OutOfBoundsError

	// hash the operation, then send validateNum, operation, its hash, inkRequired and publicKey to miner (AddShape)
	// if cant connect to miner return DisconnectedError
	// miner checks the following:
	// - if not enough ink for inkRequired, return InsufficientInkError
	// - if shape overlaps with other shapes on global canvas that don't belong to this node, return ShapeOverlapError
	// miner waits for validation, i.e. there must be validateNum blocks after the block with the operation
	// on success miner adds shape to global canvas saveing operation and its ink cost
	// 		and returns blockHash and inkRemaining after validation

	// save operation and its ink cost to canvas struct shapes with its hash as key
	// return operation hash, blockHash, inkRemaining and nil error

	err, _ = IsValidSvgShape(shapeType, shapeSvgString, fill, stroke) // will return ShapeSvgStringTooLongError, InvalidShapeSvgStringError, OutOfBoundsError
	if err != nil {
		return "", "", 0, err
	}

	// TODO check more than just empty string of stroke and fill
	if len(stroke) == 0 || len(fill) == 0 {
		return "", "", 0, InvalidShapeSvgStringError(shapeSvgString)
	}

	fullSvgString := "<path d=\"" + shapeSvgString + "\" stroke=\"" + stroke + "\" fill=\"" + fill + "\"/>"
	inkUsed := CalculateInkUsed(shapeType, shapeSvgString, fill, stroke)
	shapeHash = computeHash(fullSvgString + time.Now().String())

	//sign the operation with node's private key
	reply := shared.AddShapeReply{"", 0, 0, ""}

	r, s, _ := ecdsa.Sign(rand.Reader, &canvas.PrivKey, []byte(shapeSvgString))
	args := shared.Operation{Fill: fill, Stroke: stroke, NumBlockValidate: validateNum, AppShapeOp: fullSvgString, InkCost: inkUsed, ShapeHash: shapeHash, R: r, S: s, IsDelete: false, DAttribute: shapeSvgString, ShapeType: int(shapeType)}

	err = canvas.Miner.Call("ArtNodeMinerRPC.AddShapeRPC", &args, &reply)
	if err != nil {
		fmt.Println("ERR", err)
		return "", "", 0, DisconnectedError(mAddr)
	}
	if reply.ErrorCode == -1 {
		return "", "", 0, InsufficientInkError(inkUsed)
	}
	if reply.ErrorCode == -2 {
		return "", "", 0, ShapeOverlapError(reply.OverlappedShapeHash)
	}

	// TODO AddShape should take fullSvgString
	AddShape(inkUsed, shapeHash, shapeType, shapeSvgString, fill, stroke)

	return shapeHash, reply.BlockHash, reply.InkRemaining, nil
}

type AddShapeReply struct {
	BlockHash           string
	InkRemaining        uint32
	ErrorCode           int
	OverlappedShapeHash string
}

// Returns the MD5 hash as a hex string for the str
func computeHash(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	retVal := hex.EncodeToString(h.Sum(nil))
	return retVal
}

type GetSvgStringReply struct {
	Data  string
	Found bool
}

// Returns the encoding of the shape as an svg string.
// Can return the following errors:
// - DisconnectedError
// - InvalidShapeHashError
func (canvas canvasStruct) GetSvgString(shapeHash string) (svgString string, err error) {
	// if the shape belongs to this artnode return it right away
	// otherwise ask miner for that shape by sending shapeHash (GetSvgString)
	// if cant connect to miner return DisconnectedError
	// if miner couldnt return shape return InvalidShapeHashError

	// there are two suggestions how miners store shapes:
	//		1. store all shapes in each miner
	//		2. store shapes only produced by the art nodes of the miner
	// I think 1. is better, because if other miners fail, their shapes are still available, even though we duplicate lots of data

	//var reply string
	reply := shared.GetSvgStringReply{"", false}
	args := shared.Args{ShapeHash: shapeHash}
	err = canvas.Miner.Call("ArtNodeMinerRPC.GetSvgStringRPC", &args, &reply)
	if err != nil {
		return "", DisconnectedError("")
	}

	if !reply.Found {
		fmt.Println("ERR GetSvgStringReply", reply.Found)
		return "", InvalidShapeHashError(shapeHash)
	}

	return reply.Data, nil
}

// Returns the amount of ink currently available.
// Can return the following errors:
// - DisconnectedError
func (canvas canvasStruct) GetInk() (inkRemaining uint32, err error) {
	// just ask miner for inkRemaining (GetInk)
	// if cant connect to miner return DisconnectedError

	var reply uint32
	args := shared.Args{}
	err = canvas.Miner.Call("ArtNodeMinerRPC.GetInkRPC", &args, &reply)
	if err != nil {

		fmt.Println(err)
		return 0, DisconnectedError("")
	}

	return reply, nil
}

// Removes a shape from the canvas.
// Can return the following errors:
// - DisconnectedError
// - ShapeOwnerError
func (canvas canvasStruct) DeleteShape(validateNum uint8, shapeHash string) (inkRemaining uint32, err error) {
	// first check if this artnode owns the shape - check the map, if there is entry, it means it belongs to this artnode
	// otherwise return ShapeOwnerError

	// send shapeHash and validateNum to miner (DeleteShape)
	// if cant connect to miner return DisconnectedError
	// if there is no shape in miner network with this shapeHash, return InvalidShapeHashError (not specified in interface?)
	// miner waits for validation, i.e. there must be validateNum blocks after the block with the operation
	// on success miner removes this shape from the global canvas and refunds the ink cost specified, and returns inkRemaining

	isOwner := true
	// TODO validate that local shapeMap has the shape
	if !isOwner {
		return 0, ShapeOwnerError(shapeHash)
	}

	myShape := myShapes[shapeHash]
	stroke := myShape.Stroke
	fill := myShape.Fill
	dAttribute := myShape.DAttribute
	shapeType := myShape.ShapeType

	var reply uint32
	//sign the operation with node's private key
	r, s, _ := ecdsa.Sign(rand.Reader, &canvas.PrivKey, []byte(dAttribute))
	args := shared.Operation{NumBlockValidate: validateNum, Stroke: stroke, Fill: fill, ShapeHash: shapeHash, R: r, S: s, DAttribute: dAttribute, ShapeType: int(shapeType), IsDelete: true}
	err = canvas.Miner.Call("ArtNodeMinerRPC.DeleteShapeRPC", &args, &reply)
	if err != nil {
		fmt.Println(err)
		return 0, DisconnectedError("")
	}

	// TODO remove the shape from local shapeMap

	return 0, nil
}

type GetShapesReply struct {
	Data  []string
	Found bool
}

// Retrieves hashes contained by a specific block.
// Can return the following errors:
// - DisconnectedError
// - InvalidBlockHashError
func (canvas canvasStruct) GetShapes(blockHash string) (shapeHashes []string, err error) {
	// send blockHash to miner
	// if cant connect to miner return DisconnectedError
	// if miner cant return shapeHashes return InvalidBlockHashError

	// I assume miner will have access to list of shapeHashes for each block, since they have access to the whole tree of blocks
	// if there is no block with blockhash that's InvalidBlockHashError

	//var reply []string
	reply := shared.GetShapesReply{[]string{}, false}
	args := shared.Args{BlockHash: blockHash}
	err = canvas.Miner.Call("ArtNodeMinerRPC.GetShapesRPC", &args, &reply)
	if err != nil {
		return []string{}, DisconnectedError("")
	}

	if !reply.Found {
		fmt.Println("ERR GetShapesReply", reply.Found)
		return []string{}, InvalidBlockHashError(blockHash)
	}
	return reply.Data, nil
}

// Returns the block hash of the genesis block.
// Can return the following errors:
// - DisconnectedError
func (canvas canvasStruct) GetGenesisBlock() (blockHash string, err error) {
	// send request to miner
	// if cant connect to miner return DisconnectedError

	// should be simple, because that is actually in miner settings

	var reply string
	args := shared.Args{}
	err = canvas.Miner.Call("ArtNodeMinerRPC.GetGenesisBlockRPC", &args, &reply)
	if err != nil {
		fmt.Println("GetGenesisBlock failed: ", err)
		return "", DisconnectedError("")
	}

	return reply, nil
}

// Retrieves the children blocks of the block identified by blockHash.
// Can return the following errors:
// - DisconnectedError
// - InvalidBlockHashError
func (canvas canvasStruct) GetChildren(blockHash string) (blockHashes []string, err error) {
	// send blockHash to miner
	// if cant connect to miner return DisconnectedError
	// if there is no block with blockHash return InvalidBlockHashError
	// miner returns all blocks that have blockHash as their prevHash

	//var reply []string
	reply := shared.GetChildrenReply{[]string{}, false}
	args := shared.Args{BlockHash: blockHash}
	err = canvas.Miner.Call("ArtNodeMinerRPC.GetChildrenRPC", &args, &reply)
	if err != nil {
		return []string{}, DisconnectedError("")
	}

	if !reply.Found {
		fmt.Println("ERR GetChildren", reply.Found)
		return []string{}, InvalidBlockHashError(blockHash)
	}

	return reply.Data, nil
}

// Closes the canvas/connection to the BlockArt network.
// - DisconnectedError
func (canvas canvasStruct) CloseCanvas() (inkRemaining uint32, err error) {
	// let miner know that this art node is disconnecting
	// also ask for inkRemaining
	// if cant connect to miner return DisconnectedError
	// close the RPC connection
	var reply uint32
	args := shared.Args{}
	err = canvas.Miner.Call("ArtNodeMinerRPC.CloseCanvasRPC", &args, &reply)
	if err != nil {
		return 0, DisconnectedError("")
	}

	canvas.Miner.Close()
	canvas.Miner = nil

	return reply, nil
}
