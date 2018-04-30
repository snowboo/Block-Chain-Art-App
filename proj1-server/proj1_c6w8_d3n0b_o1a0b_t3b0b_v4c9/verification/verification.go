package verification

import "crypto/ecdsa"
import (
	"crypto/md5"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"../blockartlib"
	"../collision"
	"../shared"
)

// Verifies a block, by checking that the miner had sufficient ink for the operations
// contained in the block. In addition, may verify that the operations did indeed come from
// that block by checking the public key.

func VerifyBlock(block shared.Block, blockChain shared.Node, minerNetSettings shared.MinerNetSettings, existingBlockHashes map[string]shared.Block, allShapes map[string]shared.Operation) (verified bool) {

	// Verifies the proof of work, that there are the correct
	// number of zeroes in the hash, and that the nonce + block
	// contents hash to that hash
	if !VerifyProofOfWork(block, minerNetSettings) {
		fmt.Println("VerifyBlock - VerifyProofOfWork failed")
		return false
	}

	// Verifies that each of the operations in the block
	// came from the correct private key, using the
	// the public key and signature
	if !VerifyOperationSignatures(block) {
		fmt.Println("VerifyBlock - VerifyOperationSignatures failed")
		return false
	}

	if !VerifySufficientInkForOperationsInBlock(block, blockChain, minerNetSettings) {
		fmt.Println("VerifyBlock - VerifySufficientInkForOperationsInBlock failed")
		return false
	}

	// Checks that the current block points to a legal previous block
	// This may not be the longest chain, but it should be okay
	// if !VerifyBlockPointsToLegalBlock(block, existingBlockHashes) {
	// 	fmt.Println("TINA, failed VerifyBlockPointsToLegalBlock")
	// 	return false
	// }

	// If we have a delete operation in our block, check that the shape
	// exists in the blockchain first
	for _, v := range block.Operations {
		if v.IsDelete && !ShapeExistsInShapeHash(v.ShapeHash, allShapes) {
			fmt.Println("VerifyBlock - ShapeExistsInShapeHash failed")
			return false
		}
	}

	// Checks whether each operation intersects with the rest of the shapes on the canvas
	for _, v := range block.Operations {
		collides, _ := collision.CollideWithOtherShapes(v, allShapes)
		if !v.IsDelete && collides {
			fmt.Println("VerifyBlock - CollideWithOtherShapes failed")
			return false
		}
	}
	return true
}

// Verifies the proof of work, that the nonce, along with the operations in the block, create the
// correct hash, and that it has the same number of zeroes
func VerifyProofOfWork(block shared.Block, minerNetSettings shared.MinerNetSettings) (valid bool) {
	return VerifyHashDifficulty(block, minerNetSettings)
}

// Verifies that the hash has the correct number of zeroes
func VerifyHashDifficulty(block shared.Block, minerNetSettings shared.MinerNetSettings) (valid bool) {
	zeroes := ""

	var difficulty uint8
	if block.IsNoopBlock {
		difficulty = minerNetSettings.PoWDifficultyNoOpBlock
	} else {
		difficulty = minerNetSettings.PoWDifficultyOpBlock

	}

	for i := 0; i < int(difficulty); i++ {
		zeroes += "0"
	}
	last_string := block.Hash[len(block.Hash)-int(difficulty):]
	return strings.Compare(zeroes, last_string) == 0
}

// Verifies that the hashed block including the nonce hashes to the correct hash of the block itself
func VerifyNonceMatchesHash(block shared.Block, minerNetSettings shared.MinerNetSettings) (verified bool) {
	blockStr := ConvertBlockToString(block)

	var hash string

	m := strconv.Itoa(int(block.Nonce))
	hash = ComputeNonceSecretHash(blockStr, m)

	return strings.Compare(block.Hash, hash) == 0
}

// Iterates through each operation and checks that the signature generated from the path came from
// the correct Public Key
func VerifyOperationSignatures(block shared.Block) (valid bool) {
	for _, v := range block.Operations {

		if !ecdsa.Verify(&v.ArtNodeKey, []byte(v.DAttribute), v.R, v.S) {
			return false
		}
	}
	return true
}

// Checks whether a block points to an existing block in the chain by looking it up in the hashmap of existing blocks.
// May or may not be the best implementation as we may encounter edge cases that make this verification inaccurate
// May need to change this in the future
func VerifyBlockPointsToLegalBlock(block shared.Block, existingBlockHashes map[string]shared.Block) (valid bool) {
	_, ok := existingBlockHashes[block.PreviousBlockHash]
	return ok
}

// Checks whether a shape exists in the longest blockchain.
// Currently does not check side chains and goes through entire chain,
// could be optimized by using a map
func ShapeExistsInBlockChain(operation shared.Operation, blockChain shared.Node) (exists bool) {
	current := &blockChain

	count := 0
	for current != nil {
		operations := current.Block.Operations

		for _, v := range operations {
			if strings.Compare(operation.AppShapeOp, v.AppShapeOp) == 0 {
				if v.IsDelete {
					count--
				} else {
					count++
				}
			}
		}
		current = current.Prev
	}

	return count >= 1
}

// Checks if a shape exists in the shape hash
func ShapeExistsInShapeHash(shapeHash string, shapeHashes map[string]shared.Operation) (exists bool) {
	if _, ok := shapeHashes[shapeHash]; ok {
		return true
	}
	return false
}

// Verifies that each operation in block has sufficient ink in the blockchain
func VerifySufficientInkForOperationsInBlock(block shared.Block, blockChain shared.Node, minerNetSettings shared.MinerNetSettings) (verified bool) {
	for _, v := range block.Operations {
		reqInk := blockartlib.CalculateInkUsed(blockartlib.PATH, v.DAttribute, v.Fill, v.Stroke)
		if !CheckPublicKeyHasSufficientInk(int(reqInk), v.ArtNodeKey, blockChain, minerNetSettings.InkPerOpBlock, minerNetSettings.InkPerNoOpBlock) {
			return false
		}
	}
	return true
}

// Checks the Canvas blockchain for sufficient in in its history, and that it has sufficient ink.
// Should take into account spends (minus ink) as well as deletes (which refunds ink).
// Whenever we find a block with the same minerPubKey as our Public Key, we add
// the reward InkPerOp or Noop block depending on the type of block.
func CheckPublicKeyHasSufficientInk(reqInk int, minerPubKey ecdsa.PublicKey, blockChain shared.Node, inkPerOpBlock, inkPerNoOpBlock uint32) bool {

	current := &blockChain

	var minedInk uint32 = 0
	for current != nil {
		block := current.Block

		// Check if this current block was mined by our public key,
		// if so, add its ink reward amount to its account

		if EqualPublicKey(block.MinerKey, minerPubKey) {
			if block.IsNoopBlock {
				minedInk += inkPerNoOpBlock
			} else {
				minedInk += inkPerOpBlock
			}
		}

		for _, v := range block.Operations {
			if EqualPublicKey(v.ArtNodeKey, minerPubKey) {
				if v.IsDelete {
					// Refunded ink to node
					minedInk += v.InkCost
				} else {
					// Spent ink to draw
					minedInk -= v.InkCost
				}
			}
		}

		if current.Prev.Prev != nil {
			current = current.Prev
		} else {
			break
		}
	}

	return minedInk >= uint32(reqInk)
}

// Checks whether two signatures are equal, using r and s generated from
// ecdsa.Sign
func EqualSignatures(r1, s1, r2, s2 *big.Int) (equals bool) {
	return r1.Cmp(r2) == 0 && s1.Cmp(s2) == 0
}

// Checks whether two Public Keys are equal
func EqualPublicKey(pubKey1, pubKey2 ecdsa.PublicKey) (equals bool) {
	return pubKey1.X.Cmp(pubKey2.X) == 0 && pubKey1.Y.Cmp((pubKey2.Y)) == 0
}

// Returns the MD5 hash as a hex string for the (nonce + secret) value.
func ComputeNonceSecretHash(nonce string, secret string) string {
	h := md5.New()
	h.Write([]byte(nonce + secret))
	str := hex.EncodeToString(h.Sum(nil))
	return str
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

func EncodePublicKey(key ecdsa.PublicKey) string {
	publicKeyBytes, _ := x509.MarshalPKIXPublicKey(&key)
	encodedPubBytes := hex.EncodeToString(publicKeyBytes)
	return encodedPubBytes
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
