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

	canvas, canvasSettings, err := blockartlib.OpenCanvas(minerAddr, *priv)
	if checkError(err) != nil {
		fmt.Println(err)
	}

	inkRemaining, err := canvas.GetInk()
	if checkError(err) != nil {
		fmt.Println(err)
	}

	validateNum := uint8(3)
	svgString := []string{}

	// Add a line
	shapeHash, blockHash, ink, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 200 150 h 80", "transparent", "red")
	fmt.Printf("%v %v %v %v \n", shapeHash, blockHash, ink, err)
	if checkError(err) != nil {
		fmt.Println(err)
	}

	time.Sleep(1000 * time.Millisecond)

	inkRemaining, err = canvas.GetInk()
	if checkError(err) != nil {
		fmt.Println(err)
	}

	shapeHash, blockHash, ink, err = canvas.AddShape(validateNum, blockartlib.PATH, "M 300 200 v -100", "transparent", "red")
	fmt.Printf("%v %v %v %v \n", shapeHash, blockHash, ink, err)
	if checkError(err) != nil {
		fmt.Println(err)
	}

	shapeHash, blockHash, ink, err = canvas.AddShape(validateNum, blockartlib.PATH, "M 325 100 v 100 h 50 v -50 h -50", "transparent", "red")
	fmt.Printf("%v %v %v %v \n", shapeHash, blockHash, ink, err)
	if checkError(err) != nil {
		fmt.Println(err)
	}

	if inkRemaining >= 10956 {
		// Add a weird polygon
		shapeHash2, blockHash2, ink2, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 50 50 h -40 l 20 50 h 60 v 30 H 300 z", "red", "black")
		fmt.Printf("%v %v %v %v \n", shapeHash2, blockHash2, ink2, err)
		if checkError(err) != nil {
			fmt.Println(err)
		}

		svg, err := canvas.GetSvgString(shapeHash2)
		svgString = append(svgString, svg)
	}
	if inkRemaining >= 480 {
		// Add a square
		shapeHash3, blockHash3, ink3, err := canvas.AddShape(validateNum, blockartlib.PATH, "M 750 100 h 20 v 20 h -20 z", "red", "black")
		fmt.Printf("%v %v %v %v \n", shapeHash3, blockHash3, ink3, err)
		if checkError(err) != nil {
			fmt.Println(err)
		}
		if err == nil {
			inkRemaining, err = canvas.DeleteShape(validateNum, shapeHash3)
			if checkError(err) != nil {
				fmt.Println("Error in deleting shape")
			}
		}
	}

	genesisHash, err := canvas.GetGenesisBlock()
	fmt.Printf("GenesisBlock %v %v \n", genesisHash, err)
	if checkError(err) != nil {
	}

	time.Sleep(5000 * time.Millisecond)

	childrenHash, err := canvas.GetChildren(genesisHash)
	svgString = []string{}

	s := make(stack, 0)
	s = s.Push(stackElement{genesisHash, 0})
	parentHashMap := make(map[string]string)
	depth, maxDepth := 0, 0
	deepestLeafHash := genesisHash

	s, p := s.Pop()
	for {
		childrenHash, _ = canvas.GetChildren(p.Hash)
		if p.Depth >= maxDepth {
			maxDepth = p.Depth
			deepestLeafHash = p.Hash
		}
		if len(childrenHash) > 0 {
			for _, child := range childrenHash {
				parentHashMap[child] = p.Hash
				s = s.Push(stackElement{child, p.Depth + 1})
			}
		}
		if len(s) == 0 {
			break
		}
		s, p = s.Pop()
	}
	fmt.Println(depth, maxDepth, s, deepestLeafHash)

	for {
		if deepestLeafHash == genesisHash {
			break
		} else {
			shapeHashes, err := canvas.GetShapes(deepestLeafHash)
			if err == nil {
				for i := 0; i < len(shapeHashes); i++ {
					svg, err := canvas.GetSvgString(shapeHashes[i])
					if err == nil {
						svgString = append(svgString, svg)
					}
				}
			}
			deepestLeafHash = parentHashMap[deepestLeafHash]
		}
	}

	fmt.Println("Done GetChildren", svgString)
	CreateHtmlFile(svgString, canvasSettings.CanvasXMax, canvasSettings.CanvasYMax)

	// Close the canvas.
	inkRemaining, err = canvas.CloseCanvas()
	fmt.Printf("%v %v \n", inkRemaining, err)
	if checkError(err) != nil {
		fmt.Println(err)
		return
	}
}

type stackElement struct {
	Hash  string
	Depth int
}

type stack []stackElement

func (s stack) Push(v stackElement) stack {
	return append(s, v)
}

// dont call on empty stack :)
func (s stack) Pop() (stack, stackElement) {
	l := len(s)
	return s[:l-1], s[l-1]
}

func CreateHtmlFile(svgString []string, width uint32, height uint32) (success bool) {
	f, err := os.Create("./art_app1_miner1_output.html")
	if err == nil {
		_, err = f.WriteString("<svg height=" + fmt.Sprint(height) + " width=" + fmt.Sprint(width) + ">\n")
		for i := 0; i < len(svgString); i++ {
			_, err = f.WriteString(svgString[i] + "\n")
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

// If error is non-nil, print it out and return it.
func checkError(err error) error {
	if err != nil {
		fmt.Println("Error ", err)
		return err
	}
	return nil
}
