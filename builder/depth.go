package builder

import "errors"

type DepthResponse struct {
	Sequence string      `json:"sequence"`
	Asks     [][3]string `json:"asks"`
	Bids     [][3]string `json:"bids"`
}

type FullOrderBook struct {
	Sequence uint64      `json:"sequence"`
	Asks     [][3]string `json:"asks"`
	Bids     [][3]string `json:"bids"`
}

func (b *Builder) GetAtomicFullOrderBook() (*DepthResponse, error) {
	rsp, err := b.apiService.AtomicFullOrderBook(b.symbol)
	if err != nil {
		return nil, err
	}

	c := &DepthResponse{}
	if err := rsp.ReadData(c); err != nil {
		return nil, err
	}

	if c.Sequence == "" {
		return nil, errors.New("empty key sequence")
	}

	return c, nil
}
