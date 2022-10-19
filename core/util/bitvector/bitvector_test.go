package bitvector

import "testing"

// TestBitvectorNew checks that enforcements of argument length works in the constructors
func TestBitvectorNew(t *testing.T) {
	_, err := New(0)
	if err != errInvalidLength {
		t.Errorf("expected err %v, got %v", errInvalidLength, err)
	}

	_, err = NewFromBytes(nil, 0)
	if err != errInvalidLength {
		t.Errorf("expected err %v, got %v", errInvalidLength, err)
	}

	_, err = NewFromBytes([]byte{0}, 9)
	if err != errInvalidLength {
		t.Errorf("expected err %v, got %v", errInvalidLength, err)
	}

	_, err = NewFromBytes(make([]byte, 8), 8)
	if err != nil {
		t.Error(err)
	}
}

// TestBitvectorGetSet tests correctness of individual Set and Get commands
func TestBitvectorGetSet(t *testing.T) {
	for _, length := range []int{
		1,
		2,
		4,
		8,
		9,
		15,
		16,
	} {
		bv, err := New(length)
		if err != nil {
			t.Errorf("error for length %v: %v", length, err)
		}

		for i := 0; i < length; i++ {
			if bv.Get(i) {
				t.Errorf("expected false for element on index %v", i)
			}
		}

		func() {
			defer func() {
				if err := recover(); err == nil {
					t.Error("expecting panic")
				}
			}()
			bv.Get(length + 8)
		}()

		for i := 0; i < length; i++ {
			bv.Set(i)
			for j := 0; j < length; j++ {
				if j == i {
					if !bv.Get(j) {
						t.Errorf("element on index %v is not set to true", i)
					}
				}
			}
		}
	}
}

// TestBitvectorNewFromBytesGet tests that bit vector is initialized correctly from underlying byte slice
func TestBitvectorNewFromBytesGet(t *testing.T) {
	bv, err := NewFromBytes([]byte{8}, 8)
	if err != nil {
		t.Error(err)
	}
	if !bv.Get(3) {
		t.Fatalf("element 3 is not set to true: state %08b", bv.b[0])
	}
}
