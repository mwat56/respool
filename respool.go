/*
Copyright © 2023 M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package respool

//lint:file-ignore ST1017 - I prefer Yoda conditions
//lint:file-ignore ST1005 - Allow any error text

import (
	"errors"
	"io"
	"log"
	"sync"
)

type (
	// `TResPool` manages a set of resources that can be shared
	// safely by multiple goroutines.
	// The resource being managed must implement the `io.Closer`
	// interface.
	TResPool struct {
		factory   func() (io.Closer, error)
		mtx       sync.Mutex
		resources chan io.Closer
		closed    bool
	}

	// `TPoolErr` is the base error for all error conditions
	// returned by this package.
	TPoolErr error
)

var (
	// `ErrPoolClose` is returned if `Close()` is called multiple times.
	ErrPoolClose TPoolErr = errors.New("Pool already closed.")

	// `ErrPoolGetClosed` is returned when a `Get()` call returns on
	// a closed pool.
	ErrPoolGetClosed TPoolErr = errors.New("Pool is closed.")

	// `ErrPoolCapacity` is returned when `New()` is called with
	// an invalid capacity argument.
	ErrPoolCapacity TPoolErr = errors.New("Capacity value too small.")

	// `ErrPoolInit` is returned if `New()` has problems initialising
	// the first `aLen` pool items.
	ErrPoolInit TPoolErr = errors.New("Can't init Len pool elements.")
)

// `DEBUG` activates some screen output (if set `true`).
var DEBUG = false

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `Cap` returns the resource pool's capacity, in units of elements.
func (pool *TResPool) Cap() int {
	// Since the pool's capacity never changes after its
	// initialisation we don't need a mutex lock here.
	return cap(pool.resources)
} // Cap()

// `Close` will shutdown the pool and close all existing resources.
func (pool *TResPool) Close() error {
	// Sync this operation with the Put operation.
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	// If the pool is already close, don't do anything more.
	if pool.closed {
		if DEBUG {
			log.Println("Close:", "Pool already closed")
		}
		// While we want to close the pool anyway we return
		// an error to signal that the program's current logic
		// causes the closing attempt multiple times.
		return ErrPoolClose
	}

	// Set the pool as closed.
	pool.closed = true

	if DEBUG {
		log.Println("Close:", "Closing Pool")
	}

	// Close the channel before we drain it of its resources.
	// If we don't do this, we will get a deadlock.
	close(pool.resources)

	// Close the resources …
	var err error
	for r := range pool.resources {
		if e2 := r.Close(); nil == err {
			err = e2
		}
	}
	return err
} // Close()

// `Get` retrieves a resource from the pool.
func (pool *TResPool) Get() (io.Closer, error) {
	select {
	// Check for a free resource.
	case r, ok := <-pool.resources:
		if DEBUG {
			log.Println("Get:", "Shared Resource")
		}
		if !ok {
			return nil, ErrPoolGetClosed
		}
		return r, nil

	// Provide a new resource since there are none available.
	default:
		if DEBUG {
			log.Println("Get:", "New Resource")
		}
		return pool.factory()
	}
} // Get()

// `IsClosed` tells whether the pool is already closed.
func (pool *TResPool) IsClosed() bool {
	// Sync this operation with Close/Put operations.
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	return pool.closed
} // IsClosed()

// `Len` returns the number of currently unused elements in
// the resources pool.
func (pool *TResPool) Len() int {
	// Sync this operation with the Get/Put operations.
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	return len(pool.resources)
} // Len()

// `Put` places a resource into the pool.
func (pool *TResPool) Put(aResource io.Closer) {
	// Sync this operation with the Close operation.
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	// If the pool is closed, discard the resource.
	if pool.closed {
		if DEBUG {
			log.Println("Put:", "Queue already closed")
		}
		aResource.Close()
		return
	}

	select {
	// Try to place the resource on the queue.
	case pool.resources <- aResource:
		if DEBUG {
			log.Println("Put:", "Into Queue")
		}
		return

	// If the queue is already at cap we close the resource.
	default:
		select {
		case res, ok := <-pool.resources:
			// Get the first/oldest pool element.
			if ok {
				if err := res.Close(); (nil == err) && DEBUG {
					log.Println("Put:", "Closed oldest")
				}
			}

		case pool.resources <- aResource:
			// This time we succeeded …
			return

		default:
			if err := aResource.Close(); (nil == err) && DEBUG {
				log.Println("Put:", "Closed newest")
			}
		}
	} // select
} // Put()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `New` creates a pool that manages resources.
// A pool requires a function that can allocate a new resource and the
// initial Len and the final capacity of the pool.
//
//	`aFunc` A user provided function that can allocate a new resource.
//	`aLen` The number of elements to initialise at startup.
//	`aCap` Tha maximal number of elements in the pool.
func New(aFunc func() (io.Closer, error), aLen, aCap int) (*TResPool, TPoolErr) {
	if 0 >= aCap {
		if DEBUG {
			log.Println("New:", "Invalid pool capacity:", aCap)
		}
		return nil, ErrPoolCapacity
	}
	if (0 > aLen) || (aLen > aCap) {
		if DEBUG {
			log.Println("New:", "Invalid pool len:", aLen)
		}
		return nil, ErrPoolInit
	}

	rPool := TResPool{
		factory:   aFunc,
		resources: make(chan io.Closer, aCap),
	}

	if 0 < aLen {
		if DEBUG {
			log.Println("New:", "Initialising pool elements:", aLen)
		}

		for i := 0; i < aLen; i++ {
			if r, err := aFunc(); nil == err {
				select {
				// Try placing a new resource on the queue.
				case rPool.resources <- r:
					// Success: go on with the loop.
					break

				// If the queue is already at cap we close the resource.
				default:
					r.Close()
					// Should never, ever happen …
					return nil, ErrPoolInit
				}
			}
		}
	}

	return &rPool, nil
} // New()

/* _EoF_ */
