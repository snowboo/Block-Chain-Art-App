package createhtml

// Expects blockartlib.go to be in the ./blockartlib/ dir, relative to
// this art-app.go file
import "../blockartlib"

func main() {
    // NOTE: make sure that canvas settings are setup
    shapes := []string{}
    // empty shapes
    blockartlib.CreateHtmlFile(shapes)

    shape1 := "<path d=\"M150 0 L75 200 L225 200 Z\" stroke=\"black\" fill=\"blue\"/>"

    shape2 := "<path d=\"M650 600 l75 200 l225 200 Z\" stroke=\"red\" fill=\"white\"/>"
    shapes = append(shapes, shape1)
    shapes = append(shapes, shape2)

    // one shape
    blockartlib.CreateHtmlFile(shapes)
}
