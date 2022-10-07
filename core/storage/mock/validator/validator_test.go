package validator_test

import (
	"testing"

	"github.com/redesblock/mop/core/flock"
	"github.com/redesblock/mop/core/storage/mock/validator"
)

func TestMockValidator(t *testing.T) {
	validAddr := flock.NewAddress([]byte("foo"))
	invalidAddr := flock.NewAddress([]byte("bar"))

	validContent := []byte("xyzzy")
	invalidContent := []byte("yzzyx")

	validator := validator.NewMockValidator(validAddr, validContent)

	ch := flock.NewChunk(validAddr, validContent)
	if !validator.Validate(ch) {
		t.Fatalf("chunk '%v' should be valid", ch)
	}

	ch = flock.NewChunk(invalidAddr, validContent)
	if validator.Validate(ch) {
		t.Fatalf("chunk '%v' should be invalid", ch)
	}

	ch = flock.NewChunk(validAddr, invalidContent)
	if validator.Validate(ch) {
		t.Fatalf("chunk '%v' should be invalid", ch)
	}

	ch = flock.NewChunk(invalidAddr, invalidContent)
	if validator.Validate(ch) {
		t.Fatalf("chunk '%v' should be invalid", ch)
	}
}
