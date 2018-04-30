/*

A trivial application to illustrate how the blockartlib library can be
used from an application in project 1 for UBC CS 416 2017W2.

Usage:
go run art-app.go miner_ip privateKey
*/

package main

import "./blockartlib"

import "fmt"
import "os"
import "crypto/x509"
import "encoding/hex"

func main() {
    args := os.Args[1:]

    if len(args) != 2 {
        fmt.Println("Incorrect number of arguments, need 2")
        return
    }

    minerAddr := args[0]
    privateKeyBytesRestored, _ := hex.DecodeString(args[1])
    priv, _ := x509.ParseECPrivateKey(privateKeyBytesRestored)

    canvas, _, err := blockartlib.OpenCanvas(minerAddr, *priv)
    if checkError(err) != nil {
        return
    }

    inkRemaining, err := canvas.GetInk()
    if checkError(err) != nil {
        return
    }

    validateNum := uint8(3)

    // Expect invalid svg
    shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "550 550 h 50 v 50 Z", "green", "black")
    fmt.Printf("%v %v %v %v \n", shapeHash, blockHash, ink, err)
    if checkError(err) != nil {
        fmt.Println("expected an invalid svg error: ", err)
    }

    // Expect svg too long
    shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 500 500 v 50 h 50 v 50 h 50 v 50 h 50 v 50 h 50 v 50 h 50 v 50 h 50 v 50 h 50 v 50 h 50 v 50 h 50 v 50 h 50 v 50 h 50 v 50 h v 50 h 50 v 50 h 50 v 50 h Z", "green", "black")
    fmt.Printf("%v %v %v %v \n", shapeHash2, blockHash2, ink2, err)
    if checkError(err) != nil {
        fmt.Println("expected an svg too long error", err)
    }

    // Expect out of bounds
    shapeHash3, blockHash3, ink3, err := canvas.AddShape(validateNum, blockartlib.PATH, "M -1 -1 h 50 v 50 Z", "green", "black")
    fmt.Printf("%v %v %v %v \n", shapeHash3, blockHash3, ink3, err)
    if checkError(err) != nil {
        fmt.Println("expected an out of bounds svg error: ", err)
    }

    // Close the canvas.
    inkRemaining, err = canvas.CloseCanvas()
    fmt.Printf("%v %v \n", inkRemaining, err)
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
