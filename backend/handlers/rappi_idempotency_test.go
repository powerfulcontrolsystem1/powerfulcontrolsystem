package handlers

import (
	"testing"
	"time"
)

func TestRappiSignatureTimestampFresh(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	if !rappiSignatureTimestampFresh("1800000000", now) {
		t.Fatal("timestamp actual fue rechazado")
	}
	if !rappiSignatureTimestampFresh("1800000000000", now) {
		t.Fatal("timestamp en milisegundos fue rechazado")
	}
	if rappiSignatureTimestampFresh("1799999000", now) {
		t.Fatal("timestamp antiguo fue aceptado")
	}
	if rappiSignatureTimestampFresh("invalido", now) {
		t.Fatal("timestamp invalido fue aceptado")
	}
}
