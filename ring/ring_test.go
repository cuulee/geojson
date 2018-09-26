package ring

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/tidwall/lotsa"
)

func init() {
	seed := time.Now().UnixNano()
	println(seed)
	rand.Seed(seed)
	if os.Getenv("PIPBENCH") != "1" {
		println("use PIPBENCH=1 for point-in-polygon benchmarks")
	}
}

func S(ax, ay, bx, by float64) Segment {
	return Segment{Point{ax, ay}, Point{bx, by}}
}
func R(minX, minY, maxX, maxY float64) Rect {
	return Rect{Point{minX, minY}, Point{maxX, maxY}}
}
func P(x, y float64) Point {
	return Point{x, y}
}

var (
	rectangle = []Point{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 0}}
	pentagon  = []Point{{2, 2}, {8, 0}, {10, 6}, {5, 10}, {0, 6}, {2, 2}}
	triangle  = []Point{{0, 0}, {10, 0}, {5, 10}, {0, 0}}
	trapezoid = []Point{{0, 0}, {10, 0}, {8, 10}, {2, 10}, {0, 0}}
	octagon   = []Point{
		{3, 0}, {7, 0}, {10, 3}, {10, 7},
		{7, 10}, {3, 10}, {0, 7}, {0, 3}, {3, 0},
	}
	concave1  = []Point{{5, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 5}, {5, 5}, {5, 0}}
	concave2  = []Point{{0, 0}, {5, 0}, {5, 5}, {10, 5}, {10, 10}, {0, 10}, {0, 0}}
	concave3  = []Point{{0, 0}, {10, 0}, {10, 5}, {5, 5}, {5, 10}, {0, 10}, {0, 0}}
	concave4  = []Point{{0, 0}, {10, 0}, {10, 10}, {5, 10}, {5, 5}, {0, 5}, {0, 0}}
	bowtie    = []Point{{0, 0}, {5, 4}, {10, 0}, {10, 10}, {5, 6}, {0, 10}, {0, 0}}
	notClosed = []Point{{0, 0}, {10, 0}, {10, 10}, {0, 10}}
)

func expect(t testing.TB, what bool) {
	t.Helper()
	if !what {
		t.Fatal("expection failure")
	}
}

func TestRingScan(t *testing.T) {
	test := func(t *testing.T, indexed bool) {
		rectangleRing := NewRing(rectangle, indexed)
		var segs []Segment
		rectangleRing.Scan(func(seg Segment) bool {
			segs = append(segs, seg)
			return true
		})
		segsExpect := []Segment{
			S(0, 0, 10, 0),
			S(10, 0, 10, 10),
			S(10, 10, 0, 10),
			S(0, 10, 0, 0),
		}
		expect(t, len(segs) == len(segsExpect))
		for i := 0; i < len(segs); i++ {
			expect(t, segs[i] == segsExpect[i])
		}

		segs = nil
		notClosedRing := NewRing(rectangle, indexed)
		notClosedRing.Scan(func(seg Segment) bool {
			segs = append(segs, seg)
			return true
		})
		expect(t, len(segs) == len(segsExpect))
		for i := 0; i < len(segs); i++ {
			expect(t, segs[i] == segsExpect[i])
		}
	}
	t.Run("Indexed", func(t *testing.T) {
		test(t, true)
	})
	t.Run("Simple", func(t *testing.T) {
		test(t, false)
	})
}

func TestRingSearch(t *testing.T) {
	test := func(t *testing.T, indexed bool) {
		octagonRing := NewRing(octagon, indexed)
		var segs []Segment
		octagonRing.Search(R(0, 0, 0, 0), func(seg Segment, _ int) bool {
			segs = append(segs, seg)
			return true
		})
		segsExpect := []Segment{
			S(0, 3, 3, 0),
		}
		for i := 0; i < len(segs); i++ {
			expect(t, segs[i] == segsExpect[i])
		}
		segs = nil
		octagonRing.Search(R(0, 0, 0, 10), func(seg Segment, _ int) bool {
			segs = append(segs, seg)
			return true
		})
		segsExpect = []Segment{
			S(3, 10, 0, 7),
			S(0, 7, 0, 3),
			S(0, 3, 3, 0),
		}
		for i := 0; i < len(segs); i++ {
			expect(t, segs[i] == segsExpect[i])
		}
		segs = nil
		octagonRing.Search(R(0, 0, 5, 10), func(seg Segment, _ int) bool {
			segs = append(segs, seg)
			return true
		})
		segsExpect = []Segment{
			S(3, 0, 7, 0),
			S(7, 10, 3, 10),
			S(3, 10, 0, 7),
			S(0, 7, 0, 3),
			S(0, 3, 3, 0),
		}
		for i := 0; i < len(segs); i++ {
			expect(t, segs[i] == segsExpect[i])
		}
	}
	t.Run("Indexed", func(t *testing.T) {
		test(t, true)
	})
	t.Run("Simple", func(t *testing.T) {
		test(t, false)
	})
}

func TestBigRandomPIP(t *testing.T) {
	simple := NewRing(az, false)
	tree := NewRing(az, true)
	expect(t, simple.Rect() == tree.Rect())
	rect := tree.Rect()
	start := time.Now()
	for time.Since(start) < time.Second/2 {
		point := P(
			rand.Float64()*(rect.Max.X-rect.Min.X)+rect.Min.X,
			rand.Float64()*(rect.Max.Y-rect.Min.Y)+rect.Min.Y,
		)
		expect(t, tree.ContainsPoint(point, true) ==
			simple.ContainsPoint(point, true))
	}
}

func TestBigArizona(t *testing.T) {
	simple := NewRing(az, false)
	tree := NewRing(az, true)
	pointIn := P(-112, 33)
	pointOut := P(-114.47753906249999, 33.99802726234877)
	pointOn := P(-114.604715, 35.061744)

	expect(t, simple.ContainsPoint(pointIn, true))
	expect(t, tree.ContainsPoint(pointIn, true))

	expect(t, simple.ContainsPoint(pointOn, true))
	expect(t, tree.ContainsPoint(pointOn, true))

	expect(t, !simple.ContainsPoint(pointOn, false))
	expect(t, !tree.ContainsPoint(pointOn, false))

	expect(t, !simple.ContainsPoint(pointOut, true))
	expect(t, !tree.ContainsPoint(pointOut, true))
	if os.Getenv("PIPBENCH") == "1" {
		lotsa.Output = os.Stderr
		fmt.Printf("az/tree/in  ")
		lotsa.Ops(1000, 1, func(_, _ int) {
			simple.ContainsPoint(pointIn, true)
		})
		fmt.Printf("az/simp/in  ")
		lotsa.Ops(1000, 1, func(_, _ int) {
			tree.ContainsPoint(pointIn, true)
		})
		fmt.Printf("az/simp/on  ")
		lotsa.Ops(1000, 1, func(_, _ int) {
			simple.ContainsPoint(pointOn, true)
		})
		fmt.Printf("az/tree/on  ")
		lotsa.Ops(1000, 1, func(_, _ int) {
			tree.ContainsPoint(pointOn, true)
		})
		fmt.Printf("az/simp/out ")
		lotsa.Ops(1000, 1, func(_, _ int) {
			simple.ContainsPoint(pointOut, true)
		})
		fmt.Printf("az/tree/out ")
		lotsa.Ops(1000, 1, func(_, _ int) {
			tree.ContainsPoint(pointOut, true)
		})
	}
}

func TestBigTexas(t *testing.T) {
	simple := NewRing(tx, false)
	tree := NewRing(tx, true)
	pointIn := P(-98.525390625, 29.36302703778376)
	pointOut := P(-101.953125, 29.32472016151103)
	pointOn := P(-100.402214, 28.532657)

	expect(t, simple.ContainsPoint(pointIn, true))
	expect(t, tree.ContainsPoint(pointIn, true))

	expect(t, simple.ContainsPoint(pointOn, true))
	expect(t, tree.ContainsPoint(pointOn, true))

	expect(t, !simple.ContainsPoint(pointOn, false))
	expect(t, !tree.ContainsPoint(pointOn, false))

	expect(t, !simple.ContainsPoint(pointOut, true))
	expect(t, !tree.ContainsPoint(pointOut, true))
	if os.Getenv("PIPBENCH") == "1" {
		lotsa.Output = os.Stderr
		fmt.Printf("tx/simp/in  ")
		lotsa.Ops(1000, 1, func(_, _ int) {
			simple.ContainsPoint(pointIn, true)
		})
		fmt.Printf("tx/tree/in  ")
		lotsa.Ops(1000, 1, func(_, _ int) {
			tree.ContainsPoint(pointIn, true)
		})
		fmt.Printf("tx/simp/on  ")
		lotsa.Ops(1000, 1, func(_, _ int) {
			simple.ContainsPoint(pointOn, true)
		})
		fmt.Printf("tx/tree/on  ")
		lotsa.Ops(1000, 1, func(_, _ int) {
			tree.ContainsPoint(pointOn, true)
		})
		fmt.Printf("tx/simp/out ")
		lotsa.Ops(1000, 1, func(_, _ int) {
			simple.ContainsPoint(pointOut, true)
		})
		fmt.Printf("tx/tree/out ")
		lotsa.Ops(1000, 1, func(_, _ int) {
			tree.ContainsPoint(pointOut, true)
		})
	}
}