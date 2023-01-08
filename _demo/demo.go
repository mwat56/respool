// This sample program demonstrates how to use the respool package
// to share a simulated set of database connections.
package main

import (
	"io"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mwat56/respool"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	maxGoroutines = 23 // the number of routines to use.
	capResources  = 13 // max. number of resources in the pool
	lenResources  = 7  // init number of resources in the pool
)

// `dbConnection` simulates a resource to share.
type dbConnection struct {
	ID int32
}

// `Close` implements the io.Closer interface so dbConnection
// can be managed by the pool.
// Close performs any resource release management.
func (aDbConn *dbConnection) Close() error {
	log.Println("Close: Connection", aDbConn.ID)

	return nil
} // Close()

// `idCounter` provides support for giving each connection a unique id.
var idCounter int32

// `createConnection` is a factory method that will be called
// by the pool when a new connection is needed.
func createConnection() (io.Closer, error) {
	id := atomic.AddInt32(&idCounter, 1)
	log.Println("Create: New Connection", id)

	return &dbConnection{id}, nil
} // createConnection()

// `main` is the entry point for all Go programs.
func main() {
	var wg sync.WaitGroup
	wg.Add(maxGoroutines)

	// Create the pool to manage our connections.
	pool, err := respool.New(createConnection, lenResources, capResources)
	if nil != err {
		log.Println(err)
	}

	// Perform queries using connections from the pool.
	for query := 0; query < maxGoroutines; query++ {
		// Each goroutine needs its own copy of the query
		// value else they will all be sharing the same
		// query variable.
		go func(aQuery int) {
			performQueries(aQuery, pool)
			wg.Done()
		}(query)
	} // for

	// Wait for the goroutines to finish.
	wg.Wait()

	// Close the pool.
	log.Println("Shutdown Program.")
	pool.Close()
} // main()

// performQueries tests the resource pool of connections.
func performQueries(aQuery int, aPool *respool.TResPool) {
	// Get a connection from the pool.
	conn, err := aPool.Get()
	if nil != err {
		log.Println(err)
		return
	}

	// Put the connection back into the pool.
	defer aPool.Put(conn)

	// Wait to simulate a query response.
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	log.Printf("Query: QID[%d] CID[%d] CAP[%d] LEN[%d]\n", aQuery, conn.(*dbConnection).ID, aPool.Cap(), aPool.Len())
} // performQueries()

/* _EoF_ */
