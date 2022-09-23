package validator_test

import (
	"testing"

	"github.com/redesblock/mop/core/storage/mock/validator"
	"github.com/redesblock/mop/core/swarm"
)

func TestMockValidator(t *testing.T) {
	validAddr := swarm.NewAddress([]byte("foo"))
	invalidAddr := swarm.NewAddress([]byte("bar"))

	validContent := []byte("xyzzy")
	invalidContent := []byte("yzzyx")

	validator := validator.NewMockValidator(validAddr, validContent)

	ch := swarm.NewChunk(validAddr, validContent)
	if !validator.Validate(ch) {
		t.Fatalf("chunk '%v' should be valid", ch)
	}

	ch = swarm.NewChunk(invalidAddr, validContent)
	if validator.Validate(ch) {
		t.Fatalf("chunk '%v' should be invalid", ch)
	}

	ch = swarm.NewChunk(validAddr, invalidContent)
	if validator.Validate(ch) {
		t.Fatalf("chunk '%v' should be invalid", ch)
	}

	ch = swarm.NewChunk(invalidAddr, invalidContent)
	if validator.Validate(ch) {
		t.Fatalf("chunk '%v' should be invalid", ch)
	}
}
