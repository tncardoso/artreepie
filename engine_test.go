package main

import (
	"testing"
)

var bitwiseAndTests = []struct {
	Args   []interface{}
	Result int64
}{
	{
		[]interface{}{
			int64(0x123f35fad8dcbac3),
		},
		int64(0x123f35fad8dcbac3),
	},
	{
		[]interface{}{
			int64(0x123f35fad8dcbac3),
			int64(0x1289298329732998),
		},
		int64(0x123f35fad8dcbac3) & int64(0x1289298329732998),
	},
	{
		[]interface{}{
			int64(0x123f35fad8dcbac3),
			int64(0x1289298329732998),
			int64(0x898ac98d98e89f),
		},
		int64(0x123f35fad8dcbac3) & int64(0x1289298329732998) & int64(0x898ac98d98e89f),
	},
}

func TestBitwiseAnd(t *testing.T) {
	for _, testCase := range bitwiseAndTests {
		res, err := bitwiseAndFn(testCase.Args)
		if err != nil {
			t.Errorf("Error calculating bitwise and: %s", err.Error())
		}
		if res != testCase.Result {
			t.Errorf("Error calculating bitwiseAnd for %#v, %d",
				testCase.Args, res)
		}
	}
}

var bitwiseOrTests = []struct {
	Args   []interface{}
	Result int64
}{
	{
		[]interface{}{
			int64(0x123f35fad8dcbac3),
		},
		int64(0x123f35fad8dcbac3),
	},
	{
		[]interface{}{
			int64(0x123f35fad8dcbac3),
			int64(0x1289298329732998),
		},
		int64(0x123f35fad8dcbac3) | int64(0x1289298329732998),
	},
	{
		[]interface{}{
			int64(0x123f35fad8dcbac3),
			int64(0x1289298329732998),
			int64(0x898ac98d98e89f),
		},
		int64(0x123f35fad8dcbac3) | int64(0x1289298329732998) | int64(0x898ac98d98e89f),
	},
}

func TestBitwiseOr(t *testing.T) {
	for _, testCase := range bitwiseOrTests {
		res, err := bitwiseOrFn(testCase.Args)
		if err != nil {
			t.Errorf("Error calculating bitwise or: %s", err.Error())
		}
		if res != testCase.Result {
			t.Errorf("Error calculating bitwiseOr for %#v, %d",
				testCase.Args, res)
		}
	}
}

var modTests = []struct {
	Args   []interface{}
	Result int64
}{
	{[]interface{}{int64(12), int64(2)}, int64(0)},
	{[]interface{}{int64(12), int64(5)}, int64(2)},
}

func TestMod(t *testing.T) {
	for _, testCase := range modTests {
		res, err := modFn(testCase.Args)
		if err != nil {
			t.Errorf("Error calculating mod: %s", err.Error())
		}
		if res != testCase.Result {
			t.Errorf("Error calculating mod for %#v, %d",
				testCase.Args, res)
		}
	}
}
