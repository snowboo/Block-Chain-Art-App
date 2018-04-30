// https://stackoverflow.com/questions/13172539/polygon-collision-detection-implementation
// Collision Detection Using the Separating Axis Theorem
// File to include helper function to check for collisions in convex shapes
// https://gamedevelopment.tutsplus.com/tutorials/collision-detection-using-the-separating-axis-theorem--gamedev-169

package collision

import (
	// "../blockartlib"
	"math"
	"strconv"
	"strings"

	"../shared"
)

func CollideWithShape(svgShape string, svgShape2 string) bool {
	new_x, new_y := SvgToPoints(svgShape)
	// new_x, new_y := blockartlib.SvgToPoints(svgShape)
	new_length := len(new_x)

	shape_x, shape_y := SvgToPoints(svgShape2)
	// shape_x, shape_y := blockartlib.SvgToPoints(svgShape2)
	shape_length := len(shape_x)

	new_i := 0
	if !isLine(svgShape) {
		new_i = 1
	}

	new_sk := 0
	if !isLine(svgShape2) {
		new_sk = 1
	}

	// Loop through all the axis in the new shape
	j := new_length - 1
	for i := new_i; i < new_length; i++ {
		vx := float64(new_x[j] - new_x[i])
		vy := float64(-(new_y[j] - new_y[i]))
		vx_vy_len := math.Sqrt(vx*vx + vx*vy)

		if vx_vy_len == 0 {
			vx = 0
			vy = 0
		} else {
			vx = vx / vx_vy_len
			vy = vy / vx_vy_len
		}

		// Project new shape
		maxValN := (float64(new_x[new_i]) * vx) + (float64(new_y[new_i]) * vy)
		minValN := maxValN
		for k := new_i + 1; k < new_length; k++ {
			proj_new := (float64(new_x[k]) * vx) + (float64(new_y[k]) * vy)

			if proj_new > maxValN {
				maxValN = proj_new
			} else if proj_new < minValN {
				minValN = proj_new
			}
		}

		// Now check for the original shape
		maxValS := (float64(shape_x[new_sk]) * vx) + (float64(shape_y[new_sk]) * vy)
		minValS := maxValS
		for k := new_sk + 1; k < shape_length; k++ {
			proj_shape := (float64(shape_x[k]) * vx) + (float64(shape_y[k]) * vy)

			if proj_shape > maxValS {
				maxValS = proj_shape
			} else if proj_shape < minValS {
				minValS = proj_shape
			}
		}

		// The axis don't overlap
		if !axisOverlap(minValN, maxValN, minValS, maxValS) && !checkAllZeros(minValN, maxValN, minValS, maxValS) {
			return false
		}
		j = i + 1
	}

	return true
}

func CollideWithOtherShapes(shape shared.Operation, shapes map[string]shared.Operation) (bool, string) {
	for _, op := range shapes {
		if shape.ArtNodeKey.X.Cmp(op.ArtNodeKey.X) == 0 && shape.ArtNodeKey.Y.Cmp(op.ArtNodeKey.Y) == 0 {
			continue
		}
		if isLine(shape.DAttribute) && isLine(op.DAttribute) {
			if CollideWithLines(shape.DAttribute, op.DAttribute) {
				return true, op.ShapeHash
			}
		} else {
			if CollideWithShape(shape.DAttribute, op.DAttribute) {
				return true, op.ShapeHash
			}
		}
	}
	return false, ""
}

func CollideWithLines(svgLine1 string, svgLine2 string) bool {
	x_1, y_1 := SvgToPoints(svgLine1)
	x_2, y_2 := SvgToPoints(svgLine2)
	length_1 := len(x_1)
	length_2 := len(x_2)

	for i := 0; i < length_1-1; i++ {
		willIntersect := false
		for j := 0; j < length_2-1; j++ {
			willIntersect = intersect(x_1[i], y_1[i], x_1[i+1], y_1[i+1], x_2[j], y_2[j], x_2[j+1], y_2[j+1])
			if willIntersect {
				return true
			}
		}
	}
	return false
}

func isLine(svgString string) bool {
	x_coord, y_coord := SvgToPoints(svgString)
	return !(x_coord[0] == x_coord[len(x_coord)-1] && y_coord[0] == y_coord[len(y_coord)-1])
}

func SelfIntersection(svgLine string) bool {
	x_1, y_1 := SvgToPoints(svgLine)
	length_1 := len(x_1)

	for i := 0; i < length_1-1; i++ {
		willIntersect := false
		for j := i + 1; j < length_1-1; j++ {
			willIntersect = s_intersect(x_1[i], y_1[i], x_1[i+1], y_1[i+1], x_1[j], y_1[j], x_1[j+1], y_1[j+1])
			if willIntersect {
				return true
			}
		}
	}
	return false

}

func s_intersect(ax, ay, bx, by, cx, cy, dx, dy int) bool {
	return (s_ccw(ax, ay, cx, cy, dx, dy) != s_ccw(bx, by, cx, cy, dx, dy) && s_ccw(ax, ay, bx, by, cx, cy) != s_ccw(ax, ay, bx, by, dx, dy))
}

func s_ccw(ax, ay, bx, by, cx, cy int) bool {
	return (cy-ay)*(bx-ax) >= (by-ay)*(cx-ax)
}

// Return true if line segments AB and CD intersec
func intersect(ax, ay, bx, by, cx, cy, dx, dy int) bool {
	return (ccw(ax, ay, cx, cy, dx, dy) != ccw(bx, by, cx, cy, dx, dy) && ccw(ax, ay, bx, by, cx, cy) != ccw(ax, ay, bx, by, dx, dy))
}

func ccw(ax, ay, bx, by, cx, cy int) bool {
	return (cy-ay)*(bx-ax) > (by-ay)*(cx-ax)
}

func checkAllZeros(a, b, c, d float64) bool {
	return (a == 0 && b == 0 && c == 0 && d == 0)
}

func axisOverlap(minN, maxN, minS, maxS float64) bool {
	return !(minN >= maxS || minS >= maxN)
}

// Temp - remove later and import the actual function
func SvgToPoints(shapeSvgString string) (x_coor []int, y_coor []int) {
	sliceWithSpace := strings.Split(shapeSvgString, " ")

	// remove slice that contains white spaces
	var slice []string
	for _, str := range sliceWithSpace {
		if str != "" {
			slice = append(slice, str)
		}
	}

	for i := 0; i < len(slice); i++ {
		// 2 args
		if slice[i] == "M" || slice[i] == "L" {
			i++
			x, _ := strconv.Atoi(string(slice[i]))
			i++
			y, _ := strconv.Atoi(string(slice[i]))
			x_coor = append(x_coor, x)
			y_coor = append(y_coor, y)
		} else if string(slice[i]) == "H" {
			i++
			x, _ := strconv.Atoi(string(slice[i]))
			x_coor = append(x_coor, x)
			y_coor = append(y_coor, y_coor[len(y_coor)-1])
		} else if slice[i] == "V" {
			i++
			y, _ := strconv.Atoi(string(slice[i]))
			x_coor = append(x_coor, x_coor[len(x_coor)-1])
			y_coor = append(y_coor, y)
		} else if slice[i] == "m" || slice[i] == "l" {
			i++
			x, _ := strconv.Atoi(string(slice[i]))
			i++
			y, _ := strconv.Atoi(string(slice[i]))
			x_coor = append(x_coor, x+x_coor[len(x_coor)-1])
			y_coor = append(y_coor, y+y_coor[len(y_coor)-1])
		} else if slice[i] == "h" {
			i++
			x, _ := strconv.Atoi(string(slice[i]))
			x_coor = append(x_coor, x+x_coor[len(x_coor)-1])
			y_coor = append(y_coor, y_coor[len(y_coor)-1])
		} else if slice[i] == "v" {
			i++
			y, _ := strconv.Atoi(string(slice[i]))
			x_coor = append(x_coor, x_coor[len(x_coor)-1])
			y_coor = append(y_coor, y+y_coor[len(y_coor)-1])
		} else if slice[i] == "Z" || slice[i] == "z" {
			// get the first point
			if len(x_coor) > 0 && len(y_coor) > 0 {
				x_coor = append(x_coor, x_coor[0])
				y_coor = append(y_coor, y_coor[0])
			}
		}
	}

	return x_coor, y_coor
}

func lineIntersectionHelper(ax, ay, bx, by, cx, cy, dx, dy int) bool {
	if (ax == bx && ay == by) || (cx == dx && cy == dy) {
		return false
	}

	if (ax == cx && ay == cy) || (bx == cx && by == cy) || (ax == dx && ay == dy) || (bx == dx && by == dy) {
		return false
	}

	bx_f := float64(bx) - float64(ax)
	by_f := float64(by) - float64(ay)
	cx_f := float64(cx) - float64(ax)
	cy_f := float64(cx) - float64(ay)
	dx_f := float64(dx) - float64(ax)
	dy_f := float64(dx) - float64(ay)

	abDist := math.Sqrt(float64(bx*bx + by*by))

	cosine := bx_f / abDist
	sin := by_f / abDist
	new_cx := cx_f*cosine + cy_f*sin
	new_cy := cy_f*cosine - cx_f*sin
	new_dx := dx_f*cosine + dy_f*sin
	new_dy := dy_f*cosine - dx_f*sin

	if new_cy < 0 && new_dy < 0 || new_cy >= 0 && new_dy >= 0 {
		return false
	}

	posAB := new_dx + (new_cx-new_dx)*new_dy/(new_dy-new_cy)

	if posAB < 0 || posAB > abDist {
		return false
	}

	return true
}
