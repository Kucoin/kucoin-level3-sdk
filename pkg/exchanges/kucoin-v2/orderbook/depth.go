package orderbook

//[5]interface{}{"orderId", "price", "size", "ts"}
type DepthResponse struct {
	Sequence uint64 `json:"sequence"`
	//todo update
	Asks [][4]interface{} `json:"asks"` //Sort price from low to high
	Bids [][4]interface{} `json:"bids"` //Sort price from high to low
}

func (b *Builder) GetAtomicFullOrderBook() (*DepthResponse, error) {
	resp, err := b.apiService.AtomicFullOrderBook(b.symbol)
	if err != nil {
		return nil, err
	}

	var fullOrderBook DepthResponse
	if err := resp.ReadJson(&fullOrderBook); err != nil {
		return nil, err
	}

	return &fullOrderBook, nil
}
