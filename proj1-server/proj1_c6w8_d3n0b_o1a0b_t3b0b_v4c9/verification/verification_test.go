package verification

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
	"testing"
	"../shared"
	"fmt"
)

func TestSignAndVerify(t *testing.T) {
	path := "M 0 0 L 20 20"
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	bytes := []byte(path)

	r, s, err := ecdsa.Sign(rand.Reader, priv, bytes)

	if err != nil {
		t.Errorf("Error signing: %s", err)
		return
	}

	converted := []byte(path)
	if !ecdsa.Verify(&priv.PublicKey, converted, r, s) {
		t.Error("Test verify failed")
	}
}

func TestEqualPublicKey(t *testing.T) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	pubKey1 := priv.PublicKey
	pubKey2 := priv.PublicKey

	equals := EqualPublicKey(pubKey1, pubKey2)

	if !equals {
		t.Error("Public keys equality failed")
	}
}

func TestUnEqualPublicKey(t *testing.T) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	priv2, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	pubKey1 := priv.PublicKey
	pubKey2 := priv2.PublicKey

	equals := EqualPublicKey(pubKey1, pubKey2)

	if equals {
		t.Error("Public keys equality failed, keys were equal when they shouldn't have been")
	}
}

func TestEqualPublicKeySameCurve(t *testing.T) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	pubKey1 := priv.PublicKey

	pubKey2 := ecdsa.PublicKey{elliptic.P256(), pubKey1.X, pubKey1.Y}

	equals := EqualPublicKey(pubKey1, pubKey2)

	if !equals {
		t.Error("Public keys equality failed, same keys with same elliptical should have succeeded")
	}
}

func TestEqualPublicKeyDifferentCurve(t *testing.T) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	pubKey1 := priv.PublicKey

	pubKey2 := ecdsa.PublicKey{elliptic.P384(), pubKey1.X, pubKey1.Y}

	equals := EqualPublicKey(pubKey1, pubKey2)

	if equals {
		t.Error("Public keys equality failed, same keys with different ellipticals should have failed")
	}
}

func TestVerifyHashDifficultyCorrect(t *testing.T) {
	hash := "DSKJFSDFKJEWRJEWR0000000"
	var numZeroes uint8 = 7

	correct := VerifyHashDifficulty(shared.Block{Hash:hash}, shared.MinerNetSettings{PoWDifficultyOpBlock: numZeroes, PoWDifficultyNoOpBlock: numZeroes})

	if !correct {
		t.Error("Should have correct identified the number of zeroes at end of hash to be 7")
	}
}

func TestVerifyHashDifficultyTooFewZeroes(t *testing.T) {
	hash := "DSKJFSDFKJEWRJEWR000000"
	var numZeroes uint8 = 7

	correct := VerifyHashDifficulty(shared.Block{Hash:hash}, shared.MinerNetSettings{PoWDifficultyOpBlock: numZeroes, PoWDifficultyNoOpBlock: numZeroes})

	if correct {
		t.Error("Number of zeroes requested was 7, but passed despite only 6 zeroes")
	}
}

func TestVerifyHashDifficultyNoZeroes(t *testing.T) {
	hash := "DSKJFSDFKJEWRJEWR"
	var numZeroes uint8 = 7

	correct := VerifyHashDifficulty(shared.Block{Hash:hash}, shared.MinerNetSettings{PoWDifficultyOpBlock: numZeroes, PoWDifficultyNoOpBlock: numZeroes})

	if correct {
		t.Error("Number of zeroes requested was 7, but passed despite no zeroes")
	}
}

func TestVerifyHashDifficultyTooManyZeroes(t *testing.T) {
	hash := "DSKJFSDFKJEWRJEWR0000000000"
	var numZeroes uint8 = 7

	correct := VerifyHashDifficulty(shared.Block{Hash:hash}, shared.MinerNetSettings{PoWDifficultyOpBlock: numZeroes, PoWDifficultyNoOpBlock: numZeroes})

	if !correct {
		t.Error("Number of zeroes requested was 7, but did not pass when difficulty was higher than needed")
	}
}

func TestVerifyNonceMatchesHashCorrectNoOp(t *testing.T) {

	var X *big.Int = big.NewInt(1)
	var Y *big.Int = big.NewInt(2)

	var r *big.Int = big.NewInt(3)
	var s *big.Int = big.NewInt(4)

	pubKey := ecdsa.PublicKey{elliptic.P384(), X, Y}
	operation := shared.Operation{"shape", "", "", "", 1, false, ecdsa.PublicKey{}, 1, 5, "hash", r, s}
	operations := []shared.Operation{operation}

	var nonce uint32 = 123456
	var difficulty uint8 = 4

	block := shared.Block{"ABCD", true, operations, pubKey, nonce, ""}

	blockStr := ConvertBlockToString(block)

	_, expected := FindSecret(blockStr, difficulty)
	fmt.Println("EXPECTED: " + expected)
	block.Hash = expected

	if !VerifyNonceMatchesHash(block, shared.MinerNetSettings{PoWDifficultyOpBlock: difficulty, PoWDifficultyNoOpBlock: difficulty}) {
		t.Error("Test incorrectly returned false when block and nonce correctly hashed to correct hash")
	}
}

func TestVerifyNonceMatchesHashIncorrectNoOp(t *testing.T) {

	var X *big.Int = big.NewInt(1)
	var Y *big.Int = big.NewInt(2)

	var r *big.Int = big.NewInt(3)
	var s *big.Int = big.NewInt(4)

	pubKey := ecdsa.PublicKey{elliptic.P384(), X, Y}
	operation := shared.Operation{"shape", "", "", "", 1, false, ecdsa.PublicKey{}, 1, 5, "hash", r, s}
	operations := []shared.Operation{operation}

	var nonce uint32 = 123456
	var difficulty uint8 = 4

	block := shared.Block{"ABCD", true, operations, pubKey, nonce, ""}

	block.Hash = "ABCD"

	if VerifyNonceMatchesHash(block, shared.MinerNetSettings{PoWDifficultyOpBlock: difficulty, PoWDifficultyNoOpBlock: difficulty}) {
		t.Error("Test incorrectly returned true when block and nonce incorrect hashed to correct hash")
	}
}


func TestVerifyOperationSignaturesCorrect(t *testing.T) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	appShapeOp := "shape"

	r, s, _ := ecdsa.Sign(rand.Reader, priv, []byte(appShapeOp))

	operation := shared.Operation{appShapeOp, "", "", "", 1, false, priv.PublicKey, 1, 5, "hash", r, s}

	operations := []shared.Operation{operation}

	block := shared.Block{Operations:operations}

	valid := VerifyOperationSignatures(block)

	if !valid {
		t.Error("Test returned incorrect when the operations in the block were valid")
	}
}

func TestVerifyOperationSignaturesTwoOperationsCorrect(t *testing.T) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	appShapeOp := "shape"
	appShapeOp2 := "hello"


	r, s, _ := ecdsa.Sign(rand.Reader, priv, []byte(appShapeOp))

	operation := shared.Operation{"shape", "", "", "", 1, false, priv.PublicKey, 1, 5, "hash", r, s}

	r, s, _ = ecdsa.Sign(rand.Reader, priv, []byte(appShapeOp2))

	operation2 := shared.Operation{appShapeOp2, "", "", "", 1, false, priv.PublicKey, 1, 5, "hash", r, s}


	operations := []shared.Operation{operation, operation2}

	block := shared.Block{Operations:operations}

	valid := VerifyOperationSignatures(block)

	if !valid {
		t.Error("Test returned incorrect when the operations in the block were valid")
	}
}


func TestVerifyOperationSignaturesIncorrect(t *testing.T) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	r := big.NewInt(5)
	s := big.NewInt(7)

	operation := shared.Operation{"shape", "", "", "", 1, false, priv.PublicKey, 1, 5, "hash", r, s}

	operations := []shared.Operation{operation}

	block := shared.Block{Operations:operations}

	valid := VerifyOperationSignatures(block)

	if valid {
		t.Error("Test returned correct when the operations in the block were invalid")
	}
}

func TestVerifyOperationSignaturesOneCorrectOneIncorrect(t *testing.T) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	r := big.NewInt(5)
	s := big.NewInt(7)

	appShapeOp2 := "hello"

	r, s, _ = ecdsa.Sign(rand.Reader, priv, []byte(appShapeOp2))

	operation := shared.Operation{"shape", "", "", "", 1, false, priv.PublicKey, 1, 5, "hash", r, s}
	operation2 := shared.Operation{appShapeOp2, "", "", "", 1, false, priv.PublicKey, 1, 5, "hash", r, s}

	operations := []shared.Operation{operation, operation2}

	block := shared.Block{Operations:operations}

	valid := VerifyOperationSignatures(block)

	if valid {
		t.Error("Test returned correct when the operations in the block were invalid")
	}
}

func TestVerifySufficientInkForOperationsInBlockEnoughInk (t *testing.T){
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	minerNetSettings := shared.MinerNetSettings{InkPerNoOpBlock: 1560}

	block1 := shared.Block{MinerKey: priv.PublicKey, IsNoopBlock: true}
	node1 := shared.Node{Block:block1, Prev: nil}

	block2 := shared.Block{MinerKey: priv.PublicKey}

	appShapeOp := "M 0 0 H 50 V 40 h -20 Z"

	operation := shared.Operation{appShapeOp, "red", "red", "M 0 0 H 50 V 40 h -20 Z", 1, false, priv.PublicKey, 1, 1560, "shapeHash", big.NewInt(5), big.NewInt(6)}

	operations := []shared.Operation{operation}
	block2.Operations = operations

	verified := VerifySufficientInkForOperationsInBlock(block2, node1, minerNetSettings)

	// 1560 ink used
	if !verified {
		t.Error("Test returned incorrect when there was sufficient ink the blockchain for that public key/miner")
	}
}

func TestVerifySufficientInkForOperationsInBlockNotEnoughInk (t *testing.T){
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	minerNetSettings := shared.MinerNetSettings{InkPerNoOpBlock: 1559}

	block1 := shared.Block{MinerKey: priv.PublicKey, IsNoopBlock: true}
	node1 := shared.Node{Block:block1, Prev: nil}

	block2 := shared.Block{MinerKey: priv.PublicKey}

	appShapeOp := "M 0 0 H 50 V 40 h -20 Z"

	operation := shared.Operation{appShapeOp, "red", "red", "M 0 0 H 50 V 40 h -20 Z", 1, false, priv.PublicKey, 1, 1560, "shapeHash", big.NewInt(5), big.NewInt(6)}

	operations := []shared.Operation{operation}
	block2.Operations = operations

	verified := VerifySufficientInkForOperationsInBlock(block2, node1, minerNetSettings)

	// 1560 ink used

	if verified {
		t.Error("Test returned correct when there was insufficient ink the blockchain for that public key/miner")
	}
}
