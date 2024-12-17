package concurrent

import (
	"bytes"
	"testing"
)

func TestObjectPool(t *testing.T) {
	pool := NewObjectPool(func() *bytes.Buffer {
		return bytes.NewBuffer(make([]byte, 0, 25))
	})

	want := "TEST"
	var buff = pool.Get()
	if buff.Cap() != 25 || buff.Len() != 0 {
		t.Fatal("Capacity and Len must be 0")
	}
	buff.Reset()
	buff.WriteString(want)
	pool.Put(buff)
}
