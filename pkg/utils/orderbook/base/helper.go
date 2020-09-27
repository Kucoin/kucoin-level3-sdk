package base

import (
	"encoding/json"
	"errors"
	"fmt"
)

func Min(x, y int) int {
	if x < y {
		return x
	}

	return y
}

// ToJsonString converts any value to JSON string.
func ToJsonString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

const (
	//Sort price from low to high
	AskSide = "asks"
	//Sort price from high to low
	BidSide = "bids"
)

func CheckSide(side string) error {
	switch side {
	case AskSide:
	case BidSide:
	default:
		return errors.New(fmt.Sprintf("error side, side should be %s or %s", AskSide, BidSide))
	}

	return nil
}
