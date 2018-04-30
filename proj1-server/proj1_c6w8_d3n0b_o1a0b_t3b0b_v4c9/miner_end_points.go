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

    if len(args) != 1 {
        fmt.Println("Incorrect number of arguments, need 1 minerAddr")
        return
    }

    minerAddr := args[0]
    privateKeyBytesRestored, _:= hex.DecodeString("3081a40201010430dd09bbc48d497df5fa20be98e42cc57b11705d324a1ecac4c04572897fa71accf45d69b90073bbc4f58fb67f235742c9a00706052b81040022a1640362000461521b69e8fc90c3a87d194db94b61a1a09594e54b4602edb2a10f03b4d08d02016234b37ae3cc136dcef0e890786ff926acc74ad376eaeab9bf5fff92ba150685ba1a4918d2ba369b34c9b247f424c561d82f63ce43fd7e116f4871a9cdf9e5")
    priv, _ := x509.ParseECPrivateKey(privateKeyBytesRestored)

    canvas, _, err := blockartlib.OpenCanvas(minerAddr, *priv)
    if err != nil {
        return
    }

    genesisHash, err := canvas.GetGenesisBlock()
    _, err = canvas.GetChildren(genesisHash)

    validateNum := uint8(3)
    // Add a line
    shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 0 0 L 50 50", "transparent", "black")
    fmt.Println(err)
    if err == nil {
        shapes, _ := canvas.GetShapes(blockHash)
        fmt.Println(shapes)
        for i:=0; i < len(shapes); i++ {
            svgString, _ := canvas.GetSvgString(shapes[0])
            fmt.Println(svgString)
        }
        ink, err = canvas.DeleteShape(validateNum, shapeHash)
    }

    ink, err = canvas.GetInk()
    ink, err = canvas.CloseCanvas()

    fmt.Println(ink)
}
