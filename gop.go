package gop

import (
	"bytes"
	"encoding/gob"
	"io"
	"reflect"
	"time"
)

type Gop struct {
	rw io.ReadWriter
	m  map[string]map[string]map[interface{}]interface{}
}

func New(rw io.ReadWriter) *Gop {
	if rw == nil {
		return new(Gop)
	}

	gop := new(Gop)
	if err := gob.NewDecoder(rw).Decode(gop); err != nil && err != io.EOF {
		panic(err)
	}
	gop.rw = rw

	return gop
}

func (gop *Gop) Close() {
	if gop.rw != nil {
		if err := gob.NewEncoder(gop.rw).Encode(gop); err != nil {
			panic(err)
		}
	}
}

type Getter func(interface{}) (interface{}, time.Duration, error)

func (gop *Gop) Get(e interface{}, k interface{}, g Getter) error {
	v := reflect.ValueOf(e).Elem()

	if gop.m == nil {
		gop.m = make(map[string]map[string]map[interface{}]interface{})
	}

	pkg, ok := gop.m[v.Type().Elem().PkgPath()]
	if !ok {
		pkg = make(map[string]map[interface{}]interface{})
		gop.m[v.Type().Elem().PkgPath()] = pkg
	}

	typ, ok := pkg[v.Type().Elem().Name()]
	if !ok {
		gob.Register(e)

		typ = make(map[interface{}]interface{})
		pkg[v.Type().Elem().Name()] = typ
	}

	o, ok := typ[k]
	if !ok {
		if r, _, err := g(k); err != nil {
			return err
		} else {
			o = r
			typ[k] = o
		}
	}

	if o == nil {
		v.Set(reflect.Zero(v.Type()))
	} else if ov := reflect.ValueOf(o); ov.Elem().Kind() == reflect.Struct {
		v.Set(ov)
	} else if ov.Elem().Kind() == reflect.Ptr {
		v.Set(ov.Elem())
	} else {
		panic("unknown object")
	}

	return nil
}

func (gop *Gop) MarshalBinary() ([]byte, error) {
	var buffer bytes.Buffer
	if err := gob.NewEncoder(&buffer).Encode(gop.m); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (gop *Gop) UnmarshalBinary(data []byte) error {
	var buffer = bytes.NewBuffer(data)
	if err := gob.NewDecoder(buffer).Decode(&gop.m); err != nil {
		return err
	}
	return nil
}
