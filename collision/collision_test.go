package collision

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"testing"

	"../shared"
)

func TestCollisionLines(t *testing.T) {

	// Valid test case - should pass
	success := CollideWithLines("M 0 0 L 100 100", "M 10 0 L 50 70")
	if !success {
		fmt.Println("Error - they should have collided")
		t.Error("Test fail expected: %s, got: %s", "true", "false")
	}

	// TODO: Use separate function for testing line detection
	success = CollideWithLines("M 0 0 L 100 100", "M 0 1 L 50 120")
	if success {
		fmt.Println("Error - they should not have collided")
		t.Error("Test fail expected: %s, got: %s", "false", "true")
	}
}

func TestCollisionShapes(t *testing.T) {

	// Valid test case - should pass
	success := CollideWithShape("M 0 0 h 20 v 20 h -20 z", "M 0 0 H 50 V 40 h -20 Z")
	if !success {
		fmt.Println("Error - they should have collided")
		t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
	}

	success = CollideWithShape("M 0 0 H 50 V 40 h -20 Z", "M 70 0 h 20 v 20 h -20 Z")
	if success {
		fmt.Println("Error - they should have not collided")
		t.Error("Test fail expected: '%s', got: '%s'", "true", "false")
	}

	success = CollideWithShape("M 0 0 H 50 V 40 h -20 Z", "M 45 0 h 20 v 20 h -20 Z")
	if !success {
		fmt.Println("Error - they should have collided")
		t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
	}

	success = CollideWithShape("M 0 0 h 20 v 20 h -20 Z", "M 0 10 H 50 V 40 h -20 Z")
	if !success {
		fmt.Println("Error - they should have collided")
		t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
	}

	success = CollideWithShape("M 0 0 h 20 v 20 h -20 z", "M 0 0 h 20 v 20 h -20 z")
	if !success {
		fmt.Println("Error - they should have collided")
		t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
	}

	s := "M 250 350 l 100 -200 l 100 200 l -200 -150 h 200 z"
	success = CollideWithShape(s, s)
	if !success {
		fmt.Println("Error - they should have collided")
		t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
	}
}

func TestCollisionShapeLine(t *testing.T) {

	// Should collide with line
	success := CollideWithShape("M 0 0 h 20 v 20 h -20 z", "M 0 0 L 100 100")
	if !success {
		fmt.Println("Error - they should have collided")
		t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
	}

	// Should not collide with line
	success = CollideWithShape("M 100 0 h 20 v 20 h -20 z", "M 0 0 L 100 100")
	if success {
		fmt.Println("Error - they should not have collided")
		t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
	}
}

func TestCollisionOpShapes(t *testing.T) {
	gob.Register(&net.TCPAddr{})
	gob.Register(&elliptic.CurveParams{})

	r, _ := os.Open("/dev/urandom")
	defer r.Close()

	priv1, _ := ecdsa.GenerateKey(elliptic.P384(), r)
	// public key
	op1 := shared.Operation{AppShapeOp: "M 0 0 h 20 v 20 h -20 z", ShapeHash: "op1", ArtNodeKey: priv1.PublicKey}
	op2 := shared.Operation{AppShapeOp: "M 100 100 L 200 200", ShapeHash: "op2", ArtNodeKey: priv1.PublicKey}

	allShapes := make(map[string]shared.Operation)
	allShapes["0"] = op1
	allShapes["1"] = op2

	success, shapeHash := CollideWithOtherShapes("M 0 0 H 50 V 40 h -20 Z", allShapes)
	if !success && shapeHash != "op1" {
		fmt.Println("Error - they should have collided")
		t.Error("Test fail expected: '%s', got: '%s'", "true", "false")
	}

	success, shapeHash = CollideWithOtherShapes("M 80 0 H 50 V 40 h -20 Z", allShapes)
	if success {
		fmt.Println("Error - they should have not collided")
		t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
	}

	success, shapeHash = CollideWithOtherShapes("M 100 120 L 160 200", allShapes)
	if !success && shapeHash != "op2" {
		fmt.Println("Error - they should have collided")
		t.Error("Test fail expected: '%s', got: '%s'", "true", "false")
	}

	success, shapeHash = CollideWithOtherShapes("M 200 300 L 400 400", allShapes)
	if success {
		fmt.Println("Error - they should have not collided")
		t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
	}

}

func TestSelfIntersection(t *testing.T) {
	s := "M 250 350 l 100 -200 l 100 200 l -200 -150 h 200 z"
	success := SelfIntersection(s)
	if !success {
		fmt.Println("Error - it should have self intersected")
		t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
	}

	success = SelfIntersection("M 0 0 h 20 v 20 h -20 z")
	if success {
		fmt.Println("Error - it should have not self intersected")
		t.Error("Test fail expected: '%s', got: '%s'", "true", "false")
	}

	success = SelfIntersection("M 500 400 l -100 50 h 50 v -50 z")
	if !success {
		fmt.Println("Error - it should have self intersected")
		t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
	}

	success = SelfIntersection("M 50 50 h -40 l 20 50 h 60 v 30 h 210 z")
	if success {
		fmt.Println("Error - it should have not self intersected")
		t.Error("Test fail expected: '%s', got: '%s'", "true", "false")
	}

}
