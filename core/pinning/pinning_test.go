package pinning_test

import (
	"context"
	"strings"
	"testing"

	"github.com/redesblock/mop/core/file/pipeline/builder"
	"github.com/redesblock/mop/core/pinning"
	statestorem "github.com/redesblock/mop/core/statestore/mock"
	"github.com/redesblock/mop/core/storage"
	storagem "github.com/redesblock/mop/core/storage/mock"
	"github.com/redesblock/mop/core/traversal"
)

func TestPinningService(t *testing.T) {
	const content = "Hello, Mop!"

	var (
		ctx        = context.Background()
		storerMock = storagem.NewStorer()
		service    = pinning.NewService(
			storerMock,
			statestorem.NewStateStore(),
			traversal.New(storerMock),
		)
	)

	pipe := builder.NewPipelineBuilder(ctx, storerMock, storage.ModePutUpload, false)
	ref, err := builder.FeedPipeline(ctx, pipe, strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("create and list", func(t *testing.T) {
		if err := service.CreatePin(ctx, ref, false); err != nil {
			t.Fatalf("CreatePin(...): unexpected error: %v", err)
		}
		refs, err := service.Pins()
		if err != nil {
			t.Fatalf("Pins(...): unexpected error: %v", err)
		}
		if have, want := len(refs), 1; have != want {
			t.Fatalf("Pins(...): have %d; want %d", have, want)
		}
		if have, want := refs[0], ref; !have.Equal(want) {
			t.Fatalf("reference mismatch: have %q; want %q", have, want)
		}
	})

	t.Run("create idempotent and list", func(t *testing.T) {
		if err := service.CreatePin(ctx, ref, false); err != nil {
			t.Fatalf("CreatePin(...): unexpected error: %v", err)
		}
		refs, err := service.Pins()
		if err != nil {
			t.Fatalf("Pins(...): unexpected error: %v", err)
		}
		if have, want := len(refs), 1; have != want {
			t.Fatalf("Pins(...): have %d; want %d", have, want)
		}
		if have, want := refs[0], ref; !have.Equal(want) {
			t.Fatalf("reference mismatch: have %q; want %q", have, want)
		}
	})

	t.Run("delete and has", func(t *testing.T) {
		err := service.DeletePin(ctx, ref)
		if err != nil {
			t.Fatalf("DeletePin(...): unexpected error: %v", err)
		}
		has, err := service.HasPin(ref)
		if err != nil {
			t.Fatalf("HasPin(...): unexpected error: %v", err)
		}
		if has {
			t.Fatalf("HasPin(...): have %t; want %t", has, !has)
		}
	})

	t.Run("delete idempotent and has", func(t *testing.T) {
		err := service.DeletePin(ctx, ref)
		if err != nil {
			t.Fatalf("DeletePin(...): unexpected error: %v", err)
		}
		has, err := service.HasPin(ref)
		if err != nil {
			t.Fatalf("HasPin(...): unexpected error: %v", err)
		}
		if has {
			t.Fatalf("HasPin(...): have %t; want %t", has, !has)
		}
	})
}
