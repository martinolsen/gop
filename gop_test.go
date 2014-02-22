package gop

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestGop(t *testing.T) {
	type Foo struct {
		Id   int
		Name string
	}

	type Bar struct {
		Id  int
		Foo *Foo
	}

	var buffer bytes.Buffer

	for i := 0; i < 2; i++ {
		var getFoo = func(k interface{}) (interface{}, time.Duration, error) {
			ttl := time.Hour * 100

			if i > 0 {
				t.Fatalf("i > 0 = %d > 0", i)
			}

			switch k.(int) {
			case 0:
				return nil, ttl, nil
			case 1:
				return &Foo{k.(int), "One"}, ttl, nil
			case 2:
				return &Foo{k.(int), "Two"}, ttl, nil
			default:
				return nil, time.Duration(0), fmt.Errorf("unknown")
			}
		}

		gop := New(&buffer)
		if gop == nil {
			t.Fatalf("*Gop is nil")
		}

		for id := 0; id < 3; id++ {
			var foo *Foo
			if err := gop.Get(&foo, id, getFoo); err != nil {
				t.Fatalf("could not retrieve *Foo{Id: %d}: %s", id, err)
			} else if id != 0 && foo == nil {
				t.Fatalf("retrieved *Foo{Id: %d} is nil", id)
			} else if id == 0 && foo != nil {
				t.Fatalf("retrieved *Foo{Id: %d} is not nil", id)
			} else if (id == 1 && foo.Name != "One") || (id == 2 && foo.Name != "Two") {
				t.Fatalf("retrieved *Foo{Id: %d} has unexpected Name: %q", id, foo.Name)
			}
		}

		gop.Close()
	}

	gop := New(&buffer)
	bar := &Bar{Id: 1}
	if err := gop.Get(&bar.Foo, 1, nil); err != nil {
		t.Fatalf("could not retrieve *Bar.Foo: %s", err)
	} else if bar.Foo == nil {
		t.Fatalf("retrieved Bar{Id: %d}.Foo is nil", 1)
	}
	err := gop.Get(&bar, 1, func(k interface{}) (interface{}, time.Duration, error) {
		bar := &Bar{Id: k.(int)}
		if err := gop.Get(&bar.Foo, 1, nil); err != nil {
			return nil, time.Hour * 100, err
		}
		return bar, time.Hour * 100, nil
	})
	if err != nil {
		t.Fatalf("could not retrieve Bar{Id: %d}: %s", 1, err)
	} else if bar == nil {
		t.Fatalf("retrieved Bar{Id: %d} is nil", 1)
	}
	gop.Close()
}
