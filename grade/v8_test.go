package grade

import (
	"testing"

	"github.com/ditto-assistant/dittobench-datagen/protocol"
)

func v8Case(kind string) protocol.MemoryCase {
	return protocol.MemoryCase{BenchVersion: protocol.BenchVersionV8, AnswerKind: kind}
}

func TestV8StrictKindsDoNotChangeV7(t *testing.T) {
	chitchat := v8Case(protocol.AnswerChitchat)
	if got := Memory(chitchat, protocol.RunResponse{FinalText: "hello"}).Score; got != 0.25 {
		t.Fatalf("v8 chitchat liveness credit = %v, want .25", got)
	}
	chitchat.BenchVersion = protocol.BenchVersionV7
	if got := Memory(chitchat, protocol.RunResponse{FinalText: "hello"}).Score; got != 1 {
		t.Fatalf("v7 chitchat changed: %v", got)
	}

	ack := v8Case(protocol.AnswerAcknowledge)
	if got := Memory(ack, protocol.RunResponse{FinalText: "I don't have that."}).Score; got != 0 {
		t.Fatalf("v8 decline laundered as acknowledgement: %v", got)
	}
	ack.BenchVersion = protocol.BenchVersionV7
	if got := Memory(ack, protocol.RunResponse{FinalText: "I don't have that."}).Score; got != 1 {
		t.Fatalf("v7 acknowledgement changed: %v", got)
	}
}

func TestV8StructuredAnswerIsAuthoritative(t *testing.T) {
	mc := v8Case(protocol.AnswerValue)
	mc.ExpectedAnswer = "Lisbon"
	resp := protocol.RunResponse{Answer: "Porto", FinalText: "The answer is Lisbon."}
	if got := Memory(mc, resp).Score; got != 0 {
		t.Fatalf("v8 wrong slot was laundered by prose: %v", got)
	}
	mc.BenchVersion = protocol.BenchVersionV7
	if got := Memory(mc, resp).Score; got != 1 {
		t.Fatalf("v7 slot/prose fallback changed: %v", got)
	}
}

func TestV8PersistenceRejectsIncidentalGenericWords(t *testing.T) {
	mc := v8Case(protocol.AnswerPersistence)
	mc.AnswerItems = []string{"tennis"}
	if got := Memory(mc, protocol.RunResponse{FinalText: "I would still like to check tennis later."}).Score; got != 0 {
		t.Fatalf("generic persistence words credited: %v", got)
	}
	if got := Memory(mc, protocol.RunResponse{FinalText: "You remain a big fan of tennis."}).Score; got != 1 {
		t.Fatalf("explicit persistence stance missed: %v", got)
	}
}
