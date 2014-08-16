package main

import (
	"code.google.com/p/go.net/context"
	"errors"
	"fmt"
	"github.com/tncardoso/artreepie/twik"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"strings"
)

const (
	// Width of generated image. This value is passed as 'w' to twik
	WIDTH = 1024
	// Height of generated image. Passed as 'h' to twik
	HEIGHT = 1024
)

// Sin function for twik
func sinFn(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.New("sin expects one value")
	}

	switch arg := args[0].(type) {
	case int64:
		return math.Sin(float64(arg)), nil
	case float64:
		return math.Sin(arg), nil
	default:
		return nil, fmt.Errorf("cannot sin %#v", arg)
	}
}

// Cos function for twik
func cosFn(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.New("cos expects one value")
	}

	switch arg := args[0].(type) {
	case int64:
		return math.Cos(float64(arg)), nil
	case float64:
		return math.Cos(arg), nil
	default:
		return nil, fmt.Errorf("cannot cos %#v", arg)
	}
}

// int Mod function for twik
func modFn(args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, errors.New("% expects two values")
	}

	var dividend int64
	var divisor int64

	switch arg := args[0].(type) {
	case int64:
		dividend = arg
	case float64:
		dividend = int64(arg)
	default:
		return nil, fmt.Errorf("%% cannot operate with %#v", arg)
	}

	switch arg := args[1].(type) {
	case int64:
		divisor = arg
	case float64:
		divisor = int64(arg)
	default:
		return nil, fmt.Errorf("%% cannot operate with %#v", arg)
	}

	return dividend % divisor, nil
}

// square root function for twik
func sqrtFn(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.New("sqrt expects one values")
	}

	switch arg := args[0].(type) {
	case int64:
		return math.Sqrt(float64(arg)), nil
	case float64:
		return math.Sqrt(arg), nil
	default:
		return nil, fmt.Errorf("sqrt cannot operate with %#v", arg)
	}
}

// random returns a random float in [0.0, 1.0)
func randFn(args []interface{}) (interface{}, error) {
	return rand.Float64(), nil
}

// & function for twik
func bitwiseAndFn(args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, errors.New("& expects at least one value")
	}

	var res int64 = 0

	switch v := args[0].(type) {
	case int64:
		res = v
	case float64:
		res = int64(v)
	default:
		return nil, fmt.Errorf("cannot & %#v", v)
	}

	for _, arg := range args[1:] {
		switch v := arg.(type) {
		case int64:
			res = res & v
		case float64:
			res = res & int64(v)
		default:
			return nil, fmt.Errorf("cannot & %#v", arg)
		}
	}

	return res, nil
}

// | function for twik
func bitwiseOrFn(args []interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, errors.New("| expects at least one value")
	}

	var res int64 = 0

	switch v := args[0].(type) {
	case int64:
		res = v
	case float64:
		res = int64(v)
	default:
		return nil, fmt.Errorf("cannot | %#v", v)
	}

	for _, arg := range args[1:] {
		switch v := arg.(type) {
		case int64:
			res = res | v
		case float64:
			res = res | int64(v)
		default:
			return nil, fmt.Errorf("cannot | %#v", arg)
		}
	}

	return res, nil
}

// Struct that respresents the result of a pixel calculation. This is
// used in order to use channels and allow early termination.
type pixelResult struct {
	value interface{}
	err   error
}

// Actually calculate one pixel value. This function starts a twik scope
// sets the necessary variables and register available mathematical
// functions. The values returned is an uint8 with the value generated
// by the provided code.
func calcPixel(ctx context.Context, i, j int, code string) (uint8, error) {
	fileSet := twik.NewFileSet()
	scope := twik.NewScope(fileSet)

	// Additional functions
	scope.Create("sin", sinFn)
	scope.Create("cos", cosFn)
	scope.Create("&", bitwiseAndFn)
	scope.Create("|", bitwiseOrFn)
	scope.Create("%", modFn)
	scope.Create("rnd", randFn)
	scope.Create("sqrt", sqrtFn)

	// Available values
	scope.Create("i", int64(i))
	scope.Create("j", int64(j))
	scope.Create("w", int64(WIDTH))
	scope.Create("h", int64(HEIGHT))

	node, err := twik.ParseString(fileSet, "artreepie.twik", code)
	if err != nil {
		return 0, err
	}

	// Start the actual evaluation in a goroutine
	c := make(chan pixelResult, 1)
	go func() {
        defer func() {
            if r := recover(); r != nil {
                c <- pixelResult{int64(0), fmt.Errorf("Panic %s", r)}
            }
        }()
		ret, err := scope.Eval(node)
		c <- pixelResult{ret, err}
	}()

	// Wait for eval
	select {
	case <-ctx.Done():
		// Request taking too long, abort
		scope.Abort()
		return 0, ctx.Err()
	case ret := <-c:
		// Request finished, return result
		if ret.err != nil {
			return 0, ret.err
		}

		switch v := (ret.value).(type) {
		case int64:
			return uint8(v), nil
		case float64:
			return uint8(v), nil
		default:
			// Invalid return type, returning error
			return 0, fmt.Errorf("Invalid type returned: %#v",
				v)
		}
	}
}

// Generate image using one piece of code for each color. If one pixel
// yields an invalid result then the whole image is discarded.
func plot(ctx context.Context, r, g, b string) (*image.RGBA, error) {
	printStep := WIDTH * 100
	img := image.NewRGBA(image.Rect(0, 0, WIDTH, HEIGHT))
	totalPixels := WIDTH * HEIGHT
	proc := 0
	var err error = nil
	for j := 0; j < HEIGHT; j++ {
		for i := 0; i < WIDTH; i++ {
			cl := color.RGBA{A: 0xff}
			cl.R, err = calcPixel(ctx, i, j, r)
			if err != nil {
				return nil, err
			}
			cl.G, err = calcPixel(ctx, i, j, g)
			if err != nil {
				return nil, err
			}
			cl.B, err = calcPixel(ctx, i, j, b)
			if err != nil {
				return nil, err
			}
			img.Set(i, j, cl)
			proc++
			if proc%printStep == 0 {
				p := float64(proc) / float64(totalPixels) * 100
				log.Printf("        Processed %d/%d [%.2f]\n",
					proc, totalPixels, p)
			}
		}
	}

	return img, nil
}

// isCode returns true if post contains a piece of code. Currently this
// only checks if post starts with parenthesis. If an invalid post is
// considered code then it will be responsible for throwing the whole
// computation away.
func isCode(tweet string) bool {
	if strings.HasPrefix(strings.TrimSpace(tweet), "(") {
		return true
	} else {
		return false
	}
}
