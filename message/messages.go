package message

import (
	"errors"
)

var (
	ErrMessageNotFound = errors.New("message not found")
)

type Messages []*Message

func NewMessages() *Messages {
	return &Messages{}
}

func (m *Messages) Reset() {
	for _, mes := range *m {
		ReleaseMessage(mes)
	}

	*m = (*m)[:0]
}

func (m *Messages) GetByName(name string) (*Message, error) {
	for _, mes := range *m {
		if mes.Name == name {
			return mes, nil
		}
	}

	return nil, ErrMessageNotFound
}

func (m *Messages) GetByIndex(idx int) (*Message, error) {
	if idx >= len(*m) {
		return nil, ErrMessageNotFound
	}

	return (*m)[idx], nil
}

func (m *Messages) Len() int {
	return len(*m)
}
