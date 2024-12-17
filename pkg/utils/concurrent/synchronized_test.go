package concurrent

import "testing"

func TestSynchronizedAccess(t *testing.T) {
	initialValue := "AAAA"
	vs := NewSynchronized(&initialValue)

	expected := "BBBBB"
	vs.WithLock(func(value *string) {
		if *value != initialValue {
			t.Fatalf("unexpected value; got %v", *value)
		}
		*value = expected
	})

	vs.WithLock(func(value *string) {
		if *value != expected {
			t.Fatalf("unexpected value; wanted %v; got %v", expected, *value)
		}
	})
}
