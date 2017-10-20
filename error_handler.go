package coap

import (
	"log"

	"github.com/ironzhang/coap/internal/stack/base"
)

type errorHandler struct {
	name              string
	conRequestHandler func(s *session, m base.Message, e error) error
}

func (h errorHandler) handle(s *session, m base.Message, e error) {
	switch m.Type {
	case base.CON, base.NON:
		h.handleMSG(s, m, e)
	default:
		log.Printf("[%s] ignore %s", h.name, m.String())
	}
}

func (h errorHandler) handleMSG(s *session, m base.Message, e error) {
	if m.Code == 0 {
		return
	}

	c := m.Code >> 5
	switch {
	case c == 0:
		h.handleRequest(s, m, e)
	case c >= 2 && c <= 5:
		h.handleResponse(s, m, e)
	default:
		log.Printf("[%s] reserved code: %d.%d", h.name, c, m.Code&0x1f)
	}
}

func (h errorHandler) handleRequest(s *session, m base.Message, e error) {
	if m.Type == base.CON {
		if err := h.conRequestHandler(s, m, e); err != nil {
			log.Printf("[%s] handle con request: %v", h.name, err)
		}
	} else {
		if err := s.directSendRST(m.MessageID); err != nil {
			log.Printf("[%s] handle non request: %v", h.name, err)
		}
	}
}

func (h errorHandler) handleResponse(s *session, m base.Message, e error) {
	if m.Type == base.CON {
		if err := s.directSendRST(m.MessageID); err != nil {
			log.Printf("[%s] handle con response: %v", h.name, err)
		}
	} else {
		log.Printf("[%s] ignore non response: %s", h.name, m.String())
	}
}

func sendRSTHandler(s *session, m base.Message, e error) error {
	return s.directSendRST(m.MessageID)
}

func sendBadOptionACKHandler(s *session, m base.Message, e error) error {
	return s.directSendBadOptionACK(m.MessageID, m.Token)
}

var messageFormatErrorHandler = errorHandler{
	name:              "messageFormatErrorHandler",
	conRequestHandler: sendRSTHandler,
}

var badOptionsErrorHandler = errorHandler{
	name:              "badOptionsErrorHandler",
	conRequestHandler: sendBadOptionACKHandler,
}

func handleError(s *session, m base.Message, err error) {
	if e, ok := err.(base.MessageFormatError); ok && e.FormatError() {
		messageFormatErrorHandler.handle(s, m, err)
	} else if e, ok := err.(base.BadOptionsError); ok && e.BadOptions() {
		badOptionsErrorHandler.handle(s, m, err)
	}
}
