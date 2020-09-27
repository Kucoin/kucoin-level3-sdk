package api

type AddEventClientOidsMessage struct {
	Data map[string][]string `json:"data"`
	TokenMessage
}

func (s *Server) AddEventClientOidsToChannels(message *AddEventClientOidsMessage, reply *Response) error {
	if errResp := s.checkToken(message.Token); errResp != nil {
		*reply = *errResp
		return nil
	}

	if len(message.Data) == 0 {
		*reply = s.failure(ServerErrorCode, "empty event data")
		return nil
	}

	if err := s.app.AddEventClientOidsToChannels(message.Data); err != nil {
		*reply = s.failure(ServerErrorCode, err.Error())
		return nil
	}

	*reply = s.success("")
	return nil
}
