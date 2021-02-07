package errors

import (
	"log"
	"testing"
)

func TestNewError(t *testing.T) {
	err := New("some error")

	log.Println("Something go wrong", "err:", err)

	err = WrapStack(err)

	log.Println("Something go wrong", "err:", err, GetStack(err))
}
