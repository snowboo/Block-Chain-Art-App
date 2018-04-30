package blockartlib

import "fmt"
import "testing"

func TestIsValidSvgShape(t *testing.T) {
    canvasSettings.CanvasXMax = 1024
    canvasSettings.CanvasYMax = 1024
    err, success := IsValidSvgShape(PATH, "M 0 0", "transparent", "red")
    if !success {
        fmt.Println("Error ", err)
        t.Error("Test fail expected: '%s', got: '%s'", "true", "false")
    }

    // InvalidShapeSvgStringError
    err, success = IsValidSvgShape(PATH, "Lasd 0 0", "transparent", "red")
    if success {
        fmt.Println("Error ", err)
        t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
    }

    err, success = IsValidSvgShape(PATH, "M 0 0 L 3 Z", "transparent", "red")
    if success {
        fmt.Println("Error ", err)
        t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
    }

    err, success = IsValidSvgShape(PATH, " M 0 0 L 3 Z", "transparent", "red")
    if success {
        fmt.Println("Error ", err)
        t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
    }

    err, success = IsValidSvgShape(PATH, "M 0 0 L 3 Z ", "transparent", "red")
    if success {
        fmt.Println("Error ", err)
        t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
    }

    err, success = IsValidSvgShape(PATH, "M 0 0 L 3 Z ", "transparent", "transparent")
    if success {
        fmt.Println("Error ", err)
        t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
    }

    err, success = IsValidSvgShape(PATH, "M 0 0 M 3 6 Z ", "red", "transparent")
    if success {
        fmt.Println("Error ", err)
        t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
    }
}

func TestSimpleLine(t *testing.T) {
    canvasSettings.CanvasXMax = 1024
    canvasSettings.CanvasYMax = 1024
    err, success := IsValidSvgShape(PATH, "M 0 0 L 3 4", "transparent", "red")
    if success {
        area := CalculateInkUsed(PATH, "M 0 0 L 3 4", "transparent", "red")
        if area != 5 {
            t.Error("Test fail expected: '%d', got: '%d'", 5, area)
        }
    } else {
        fmt.Println("Error ", err)
        t.Error("Test fail expected: '%s', got: '%s'", "true", "false")
    }

    err, success = IsValidSvgShape(PATH, "M 0 0 L 20 20", "transparent", "red")
    if success {
        area := CalculateInkUsed(PATH, "M 0 0 L 20 20", "transparent", "red")
        if area != 29 {
            t.Error("Test fail expected: '%d', got: '%d'", 29, area)
        }
    } else {
        fmt.Println("Error ", err)
        t.Error("Test fail expected: '%s', got: '%s'", "true", "false")
    }
}

func TestSimpleAreas(t *testing.T) {
    canvasSettings.CanvasXMax = 1024
    canvasSettings.CanvasYMax = 1024
    err, success := IsValidSvgShape(PATH, "M 0 0 h 20 v 20 h -20 z", "red", "red")
    if success {
        area := CalculateInkUsed(PATH, "M 0 0 h 20 v 20 h -20 z", "red", "red")
        if area != 480 {
            t.Error("Test fail expected: '%d', got: '%d'", 480, area)
        }
    } else {
        fmt.Println("Error ", err)
        t.Error("Test fail expected: '%s', got: '%s'", "true", "false")
    }

    err, success = IsValidSvgShape(PATH, "M 0 0 H 50 V 40 h -20 Z", "red", "red")
    if success {
        area := CalculateInkUsed(PATH, "M 0 0 H 50 V 40 h -20 Z", "red", "red")
        if area != 1560 {
            t.Error("Test fail expected: '%d', got: '%d'", 1560, area)
        }
    } else {
        fmt.Println("Error ", err)
        t.Error("Test fail expected: '%s', got: '%s'", "true", "false")
    }
}

func TestOutOfBounds(t *testing.T) {
    canvasSettings.CanvasXMax = 5
    canvasSettings.CanvasYMax = 5

    err, success := IsValidSvgShape(PATH, "M 5 6", "red", "red")

    if success {
        fmt.Println("Error ", err)
        t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
    }

    _, success = IsValidSvgShape(PATH, "M -1 1", "red", "red")

    if success {
        fmt.Println("Error ", err)
        t.Error("Test fail expected: '%s', got: '%s'", "false", "true")
    }
}

func TestComplicatedInkUsed(t *testing.T) {
    canvasSettings.CanvasXMax = 1024
    canvasSettings.CanvasYMax = 1024
    err, success := IsValidSvgShape(PATH, "M 50 50 h -40 l 20 50 h 60 v 30 h 210 z", "transparent", "red")
    if success {
        area := CalculateInkUsed(PATH, "M 50 50 h -40 l 20 50 h 60 v 30 h 210 z", "transparent", "red")
        if area != 657 {
            t.Error("Test fail expected: '%d', got: '%d'", 657, area)
        }

        area = CalculateInkUsed(PATH, "M 50 50 h -40 l 20 50 h 60 v 30 H 300 z", "transparent", "red")
        if area != 657 {
            t.Error("Test fail expected: '%d', got: '%d'", 657, area)
        }

        area = CalculateInkUsed(PATH, "M 50 50 h -40 l 20 50 h 60 v 30 H 300 z", "red", "red")
        if area != 10956 {
            t.Error("Test fail expected: '%d', got: '%d'", 10956, area)
        }
    } else {
        fmt.Println("Error ", err)
        t.Error("Test fail expected: '%s', got: '%s'", "true", "false")
    }
}

func TestStar(t *testing.T) {
    canvasSettings.CanvasXMax = 1024
    canvasSettings.CanvasYMax = 1024
    err, success := IsValidSvgShape(PATH, "M 250 350 l 100 -200 l 100 200 l -200 -150 h 200 z", "transparent", "red")
    if !success {
        t.Error("Test fail expected: '%d', got: '%d'", "true", "false")
    }

    err, success = IsValidSvgShape(PATH, "M 250 350 l 100 -200 l 100 200 l -200 -150 h 200 z", "transparent", "red")
    if success {
        area := CalculateInkUsed(PATH, "M 250 350 l 100 -200 l 100 200 l -200 -150 h 200 z", "transparent", "red")
        if area != 1148 {
            t.Error("Test fail expected: '%d', got: '%d'", 1148, area)
        }
    } else {
        fmt.Println("Error ", err)
        t.Error("Test fail expected: '%s', got: '%s'", "true", "false")
    }
}
