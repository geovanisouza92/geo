package object

import (
	"testing"
)

func TestHashKeys(t *testing.T) {
	t.Run("hash keys", func(t *testing.T) {
		hello1 := NewString("Hello world")
		hello2 := NewString("Hello world")
		diff := NewString("My name is johnny")

		if hello1.HashKey() != hello2.HashKey() {
			t.Errorf("hash key for hello1 and hello2 should be equal; got hello1(%v) and hello2(%v)", hello1.HashKey(), hello2.HashKey())
		}
		if hello1.HashKey() == diff.HashKey() {
			t.Errorf("hash key for hello1 and diff should be different; got hello1(%v) and diff(%v)", hello1.HashKey(), diff.HashKey())
		}
	})
}

func TestObjectTypeToString(t *testing.T) {
	tt := []struct {
		union ObjectType
		str   string
	}{
		{TypeArray | TypeString, "TypeString, TypeArray"},
		{TypeString | TypeArray, "TypeString, TypeArray"},
	}

	for _, tc := range tt {
		t.Run("", func(t *testing.T) {
			actual := ObjectTypesToString(tc.union)

			if actual != tc.str {
				t.Errorf("order of ObjectType union is not garanteed to be %q; got %q", tc.str, actual)
			}
		})
	}
}
