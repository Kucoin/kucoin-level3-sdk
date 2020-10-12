package api

type GetPartOrderBookMessage struct {
	Number int `json:"number"`
	TokenMessage
}

func (s *Server) GetOrderBook(message *GetPartOrderBookMessage, reply *Response) error {
	if errResp := s.checkToken(message.Token); errResp != nil {
		*reply = *errResp
		return nil
	}

	*reply = s.success(s.app.PartOrderBook(message.Number))
	return nil
}
