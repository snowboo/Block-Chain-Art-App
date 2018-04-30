/*

A trivial application to illustrate how the blockartlib library can be
used from an application in project 1 for UBC CS 416 2017W2.

Usage:
$ go run art-app.go [minerAddr] [minerPrivKey]

$ go run art-app.go [minerAddr] 3081a40201010430dd09bbc48d497df5fa20be98e42cc57b11705d324a1ecac4c04572897fa71accf45d69b90073bbc4f58fb67f235742c9a00706052b81040022a1640362000461521b69e8fc90c3a87d194db94b61a1a09594e54b4602edb2a10f03b4d08d02016234b37ae3cc136dcef0e890786ff926acc74ad376eaeab9bf5fff92ba150685ba1a4918d2ba369b34c9b247f424c561d82f63ce43fd7e116f4871a9cdf9e5
*/

package main

// Expects blockartlib.go to be in the ./blockartlib/ dir, relative to
// this art-app.go file
import (
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"os"

	"./blockartlib"
)

func main() {
	args := os.Args[1:]

	if len(args) != 2 {
		fmt.Println("Incorrect number of arguments, need 2")
		return
	}

	minerAddr := args[0]
	privateKeyBytesRestored, _ := hex.DecodeString(args[1])
	privKey, _ := x509.ParseECPrivateKey(privateKeyBytesRestored)

	// Open a canvas.
	canvas, _, err := blockartlib.OpenCanvas(minerAddr, *privKey)
	fmt.Printf("Canvas %v\n", canvas)
	if checkError(err) != nil {
		fmt.Println(err)
	}

	inkRemaining, err := canvas.GetInk()
	fmt.Printf("Ink %v %v \n", inkRemaining, err)
	if checkError(err) != nil {
		fmt.Println(err)
	}

	blockHash, err := canvas.GetGenesisBlock()
	fmt.Printf("GenesisBlock %v %v \n", blockHash, err)
	if checkError(err) != nil {
		fmt.Println(err)
		return
	}

	// Valid case
	blockHashChildren, err := canvas.GetChildren(blockHash)
	fmt.Printf("Children %v %v \n", blockHashChildren, err)
	if checkError(err) != nil {
		fmt.Println(err)
	}

	blockHashChildrenInvalid, err := canvas.GetChildren("")
	fmt.Printf("Children %v %v \n", blockHashChildrenInvalid, err)

	// Valid case - nothing
	shapeList, err := canvas.GetShapes(blockHashChildren[0])
	fmt.Printf("Shape %v %v \n", shapeList, err)
	if checkError(err) != nil {
		fmt.Println(err)
	}

	shapeListInvalid, err := canvas.GetShapes("hi there")
	fmt.Printf("Shape %v %v \n", shapeListInvalid, err)

	// Valid case
	fmt.Println("TINA: shapeList length ", shapeList)
	if len(shapeList) > 0 {
		shapeSvgString, err := canvas.GetSvgString(shapeList[0])
		fmt.Printf("SVG String %v %v \n", shapeSvgString, err)
		if checkError(err) != nil {
			fmt.Println(err)
		}
	}

	shapeSvgStringInvalid, err := canvas.GetSvgString("123")
	fmt.Printf("SVG String %v %v \n", shapeSvgStringInvalid, err)

	var validateNum uint8 = 2

	// Add a line.
	shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 0 5", "transparent", "red")
	fmt.Printf("%v %v %v %v \n", shapeHash, blockHash, ink, err)
	if checkError(err) != nil {
		fmt.Println(err)
	}

	// Valid case - something
	fmt.Println("TINA: shapeList2 length ", shapeList)
	shapeList2, err := canvas.GetShapes(blockHash)
	if len(shapeList2) > 0 {
		fmt.Printf("Shape %v %v \n", shapeList2, err)
		if checkError(err) != nil {
			fmt.Println(err)
		}
	}

	// Add another line.
	shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 5 0", "transparent", "blue")
	fmt.Printf("%v %v %v %v \n", shapeHash2, blockHash2, ink2, err)
	if checkError(err) != nil {
		fmt.Println(err)
	}

	// Delete the first line.
	ink3, err := canvas.DeleteShape(validateNum, shapeHash)
	fmt.Printf("%v %v \n", ink3, err)
	if checkError(err) != nil {
		fmt.Println(err)
	}

	// assert ink3 > ink2

	// Close the canvas.
	ink4, err := canvas.CloseCanvas()
	fmt.Printf("%v %v \n", ink4, err)
	if checkError(err) != nil {
		return
	}
}

// If error is non-nil, print it out and return it.
func checkError(err error) error {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		return err
	}
	return nil
}
