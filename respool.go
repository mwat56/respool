/*
Copyright © 2023 M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package respool

//lint:file-ignore ST1017 - I prefer Yoda conditions
//lint:file-ignore ST1005 - Allow any error text

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"
)

type (
	// `TCreateFunc` is the function type that has to be passed
	// to the `New()` function.
	// The function is supposed the create the type of resource
	// the pool is supposed to handle.
	// It should return an `io.Closer` which is able to
	// close/free/release the resource created by this function.
	TCreateFunc func() (io.Closer, error)

	// `TResPool` manages a set of resources that can be shared
	// safely by multiple goroutines.
	// The resource being managed must implement the `io.Closer`
	// interface.
	TResPool struct {
		factory   TCreateFunc
		mtx       sync.Mutex
		resources chan io.Closer
		closed    bool
	}

	// `TPoolErr` is the base error for all error conditions
	// returned by this package.
	TPoolErr error
)

var (
	// `ErrPoolCapacity` is returned when `New()` is called with
	// an invalid capacity argument.
	ErrPoolCapacity TPoolErr = errors.New("Capacity value too small.")

	// `ErrPoolClosed` is returned when a `Get()` call returns on
	// a closed pool, or `Close()` is called multiple times.
	ErrPoolClosed TPoolErr = errors.New("Pool is closed.")

	// `ErrPoolDone` is returned if a given context is done.
	ErrPoolDone TPoolErr = errors.New("Context is done.")

	// `ErrPoolInit` is returned if `New()` has problems initialising
	// the first `aLen` pool items.
	ErrPoolInit TPoolErr = errors.New("Can't init Len pool elements.")
)

// `DEBUG` activates some screen output (if set `true`);
// default: `false`.
var DEBUG = false

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `Cap` returns the resource pool's capacity, i.e. the number of elements.
func (pool *TResPool) Cap() (rCap int) {
	// Since the pool's capacity never changes after its
	// initialisation we don't need a mutex lock here.

	rCap = cap(pool.resources)
	if DEBUG {
		log.Println("Cap:", rCap)
	}
	return
} // Cap()

// `Close` will shutdown the pool and close all existing resources.
func (pool *TResPool) Close() error {
	// We don't expect a `context` here because we have
	// to close/free our resources in any case.

	// Sync this operation with the Get/Put operation.
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
		return ErrPoolClosed
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
			//TODO: wrap the error(s)
			err = e2
		}
	}
	return err
} // Close()

// `Get` retrieves a resource from the pool.
//
//	`aContext` A (possibly canceled) context.
func (pool *TResPool) Get(aContext context.Context) (io.Closer, error) {
	// Sync this operation with Close/Put operations.
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	select {
	// Check whether we're already done.
	case <-aContext.Done():
		if DEBUG {
			log.Println("Get:", "Context is done.")
		}
		return nil, ErrPoolDone

	// Check for a free resource.
	case r, ok := <-pool.resources:
		if DEBUG {
			log.Println("Get:", "Shared Resource -- ", ok)
		}
		if ok {
			return r, nil
		}
		return nil, ErrPoolClosed

	// Provide a new resource since there are none available.
	default:
		if DEBUG {
			log.Println("Get:", "New Resource")
		}
		return pool.factory()
	} // select
} // Get()

// `IsClosed` tells whether the pool is already closed.
func (pool *TResPool) IsClosed() (rClosed bool) {
	// Sync this operation with Close/Put operations.
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	rClosed = pool.closed
	if DEBUG {
		log.Println("IsClosed:", rClosed)
	}

	return
} // IsClosed()

// `Len` returns the number of currently unused elements in
// the resources pool.
func (pool *TResPool) Len() (rLen int) {
	// Sync this operation with the Get/Put operations.
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	rLen = len(pool.resources)
	if DEBUG {
		log.Println("Len:", rLen)
	}
	return
} // Len()

// `Put` places a resource into the pool.
//
//	`aContext` A (possibly canceled) context.
//	`aResource` The resource to put back into the pool.
func (pool *TResPool) Put(aContext context.Context, aResource io.Closer) error {
	// Sync this operation with the Close/Get operation.
	pool.mtx.Lock()
	defer pool.mtx.Unlock()

	// If the pool is closed, discard the resource.
	if pool.closed {
		if DEBUG {
			log.Println("Put:", "Queue already closed")
		}
		return ErrPoolClosed
	}

	// Flag for closing the very first resource.
	killedOldest := false

	select {
	// Check whether we're already done.
	case <-aContext.Done():
		if DEBUG {
			log.Println("Put:", "Context is done (I).")
		}
		return ErrPoolDone

	// Try to place the resource on the queue.
	case pool.resources <- aResource:
		if DEBUG {
			log.Println("Put:", "Into Queue (I)")
		}
		return nil

	// If the queue is already at capacity we close a resource.
	// First we try to close the very first one (making room
	// for the new resource).
	default:
		select {
		// Check whether we're already done.
		case <-aContext.Done():
			if DEBUG {
				log.Println("Put:", "Context is done (II).")
			}
			return ErrPoolDone

		case res, ok := <-pool.resources:
			// Get the first/oldest pool element.
			if ok {
				err := res.Close()
				killedOldest = true
				if DEBUG {
					log.Println("Put:", "Closed oldest -- ", err)
				}
			}
		} // select
	}

	if killedOldest {
		// Okay, we freed a place in the pool by removing
		// the very first/oldest resource and now we try
		// again to put the new resource into the pool.
		select {
		// Again, check whether we're already done.
		case <-aContext.Done():
			if DEBUG {
				log.Println("Put:", "Context is done (III).")
			}
			return ErrPoolDone

		case pool.resources <- aResource:
			// This time we succeeded …
			if DEBUG {
				log.Println("Put:", "Into Queue (II)")
			}

		default:
			err := aResource.Close()
			if DEBUG {
				log.Println("Put:", "Closed newest -- ", err)
			}
			return err
		} // select
	}

	return nil
} // Put()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `New` creates a pool that manages resources.
// A pool requires a function that can allocate a new resource and
// the initial Len and the final Capacity of the pool.
//
//	`aFunc` A user provided function that can allocate a new resource.
//	`aLen` The number of elements to initialise at startup.
//	`aCap` Tha maximal number of elements in the pool.
func New(aFunc TCreateFunc, aLen int, aCap int) (*TResPool, TPoolErr) {
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
				// Try placing a new resource in the queue.
				case rPool.resources <- r:
					// Success: go on with the loop.
					continue

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
