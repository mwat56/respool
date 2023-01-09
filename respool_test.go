/*
Copyright © 2023 M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package respool

import (
	"context"
	"io"
	"log"
	"reflect"
	"testing"
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

type testCloser struct{}

func (cl testCloser) Close() error {
	return nil
} // Close()

var testClose testCloser

func testFactory() (io.Closer, error) {
	return testClose, nil
} // testFunc()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

func TestNew(t *testing.T) {
	type tArgs struct {
		aFunc TCreateFunc
		aLen  int
		aCap  int
	}
	poolDummy := &TResPool{}
	tests := []struct {
		name  string
		args  tArgs
		want  *TResPool
		want1 TPoolErr
	}{
		{"1", tArgs{testFactory, 0, 0}, nil, ErrPoolCapacity},
		{"2", tArgs{testFactory, 0, 2}, poolDummy, nil},
		{"3", tArgs{testFactory, 3, 2}, poolDummy, ErrPoolInit},
		{"4", tArgs{testFactory, 2, 3}, poolDummy, nil},
		// TODO: Add test cases.
	}

	DEBUG = true
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poolQueue, poolErr := New(tt.args.aFunc, tt.args.aLen, tt.args.aCap)
			if nil != poolErr {
				if nil == tt.want1 {
					t.Errorf("New() poolErr = `%v`, want `%v`", poolErr, tt.want1)
				} else if poolErr != tt.want1 {
					t.Errorf("New() poolErr = `%v`, want `%v`", poolErr, tt.want1)
				}
			} else if poolErr != tt.want1 {
				t.Errorf("New() poolErr = `%v`, want `%v`", poolErr, tt.want1)
			}
			if nil == tt.want {
				if poolQueue != nil {
					t.Errorf("New() poolQueue = `%v`, want `%v`", poolQueue, tt.want)
				}
			}
			if nil != poolQueue {
				log.Println("TestNew:", "len:", poolQueue.Len())
				poolQueue.Close()
			}
		})
	}
} // TestNew()

func TestTResPool_Cap(t *testing.T) {
	DEBUG = true
	p0, _ := New(testFactory, 0, 0)
	p1, _ := New(testFactory, 1, 2)
	p2, _ := New(testFactory, 2, 4)
	p3, _ := New(testFactory, 3, 6)
	tests := []struct {
		name string
		pool *TResPool
		want int
	}{
		{"0", p0, 0},
		{"1", p1, 2},
		{"2", p2, 4},
		{"3", p3, 6},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		if nil != tt.pool {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.pool.Cap(); got != tt.want {
					t.Errorf("TResPool.Cap() = `%v`, want `%v`", got, tt.want)
				}
			})
		}
	}
} // TestTResPool_Cap()

func TestTResPool_Close(t *testing.T) {
	DEBUG = true
	p1, _ := New(testFactory, 0, 1)
	p3, _ := New(testFactory, 0, 1)
	p3.Close()
	p4, _ := New(testFactory, 0, 1)

	tests := []struct {
		name    string
		pool    *TResPool
		wantErr bool
	}{
		// TODO: Add test cases.
		{"1", p1, false},
		{"2", p1, true},
		{"3", p3, true},
		{"4", p4, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.pool.Close(); (err != nil) != tt.wantErr {
				t.Errorf("TResPool.Close() error = `%v`, wantErr `%v`", err, tt.wantErr)
			}
		})
	}
} // TestTResPool_Close()

func TestTResPool_Get(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	DEBUG = true
	p0, e0 := New(testFactory, 0, 0)
	p1, e1 := New(testFactory, 1, 1)
	p2, e2 := New(testFactory, 2, 4)
	p3, e3 := New(testFactory, 3, 2)
	p4, e4 := New(testFactory, 0, 4)

	type tArgs struct {
		aContext context.Context
	}
	tests := []struct {
		name    string
		pool    *TResPool
		args    tArgs
		want    io.Closer
		wantErr bool
	}{
		{"0", p0, tArgs{ctx}, testClose, nil != e0},
		{"1", p1, tArgs{ctx}, testClose, nil != e1},
		{"2", p2, tArgs{ctx}, testClose, nil != e2},
		{"3", p3, tArgs{ctx}, testClose, nil != e3},
		{"4", p4, tArgs{ctx}, testClose, nil != e4},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		if nil == tt.pool {
			if !tt.wantErr {
				t.Errorf("TResPool.Get() error = `%v`, wantErr `%v`", true, tt.wantErr)
			}
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.pool.Get(tt.args.aContext)
			if (err != nil) != tt.wantErr {
				t.Errorf("TResPool.Get() error = `%v`, wantErr `%v`", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TResPool.Get() = `%v`, want `%v`", got, tt.want)
			}
		})
		log.Println("TestTResPool_Get:", "len:", tt.pool.Len())
	}
} // TestTResPool_Get()

func TestTResPool_IsClosed(t *testing.T) {
	DEBUG = true
	p1, _ := New(testFactory, 0, 1)
	p2, _ := New(testFactory, 0, 1)
	p2.closed = true
	p3, _ := New(testFactory, 0, 1)
	p3.Close()

	tests := []struct {
		name string
		pool *TResPool
		want bool
	}{
		// TODO: Add test cases.
		{"1", p1, false},
		{"2", p2, true},
		{"3", p3, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if nil != tt.pool {
				if got := tt.pool.IsClosed(); got != tt.want {
					t.Errorf("TResPool.IsClosed() = `%v`, want `%v`", got, tt.want)
				}
				tt.pool.Close()
			}
		})
	}
} // TestTResPool_IsClosed()

func TestTResPool_Len(t *testing.T) {
	DEBUG = true
	p0, _ := New(testFactory, 0, 1)
	p1, _ := New(testFactory, 1, 2)
	p2, _ := New(testFactory, 2, 3)
	p3, _ := New(testFactory, 3, 3)
	tests := []struct {
		name string
		pool *TResPool
		want int
	}{
		{"0", p0, 0},
		{"1", p1, 1},
		{"2", p2, 2},
		{"3", p3, 3},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		if nil != tt.pool {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.pool.Len(); got != tt.want {
					t.Errorf("TResPool.Len() = `%v`, want `%v`", got, tt.want)
				}
			})
		}
	}
} // TestTResPool_Len()

func TestTResPool_Put(t *testing.T) {
	DEBUG = true
	p1, _ := New(testFactory, 2, 4)

	c1 := testClose
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type tArgs struct {
		aContext  context.Context
		aResource io.Closer
	}
	tests := []struct {
		name    string
		pool    *TResPool
		args    tArgs
		wantErr bool
	}{
		{"1", p1, tArgs{ctx, c1}, false},
		{"2", p1, tArgs{ctx, c1}, false},
		{"3", p1, tArgs{ctx, c1}, false},
		{"4", p1, tArgs{ctx, c1}, false},
		{"5", p1, tArgs{ctx, c1}, false},
		{"6", p1, tArgs{ctx, c1}, false},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.pool.Put(tt.args.aContext, tt.args.aResource); (err != nil) != tt.wantErr {
				t.Errorf("TResPool.Put() error = `%v`, wantErr `%v`", err, tt.wantErr)
			}
		})
		log.Println("TestTResPool_Put:", "len:", tt.pool.Len())
	}
} // TestTResPool_Put()

/* _EoF_ */
