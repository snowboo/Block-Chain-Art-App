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
import "time"

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

    // Add a line
    shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 200 150 L 248 102 v 100", "transparent", "blue")
    fmt.Printf("%v %v %v %v \n", shapeHash, blockHash, ink, err)
    if checkError(err) != nil {
        fmt.Println("expected shape overlap error: ", err)
    }

    inkRemaining, err = canvas.GetInk()
    if checkError(err) != nil {
        fmt.Println(err)
        return
    }

    if inkRemaining >= 1560 {
        // Add a weird polygon
        shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 H 50 V 40 h -20 Z", "blue", "black")
        fmt.Printf("%v %v %v %v \n", shapeHash2, blockHash2, ink2, err)
        if checkError(err) != nil {
            fmt.Println("expected shape overlap error: ", err)
        }
    }

    inkRemaining, err = canvas.GetInk()
    if checkError(err) != nil {
        fmt.Println(err)
        return
    }

    if inkRemaining >= 480 {
        // Add a square
        shapeHash3, blockHash3, ink3, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 1000 1000 h 20 v 20 h -20 z", "blue", "black")
        fmt.Printf("%v %v %v %v \n", shapeHash3, blockHash3, ink3, err)
        if checkError(err) != nil {
            fmt.Println("expected shape overlap error: ", err)
        }
    }

    time.Sleep(5000 * time.Millisecond)

    genesisHash, err := canvas.GetGenesisBlock()
    fmt.Printf("GenesisBlock %v %v \n", genesisHash, err)
    if checkError(err) != nil {
        return
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
		fmt.Println("Error ", err)
		return err
	}
	return nil
}
