package main 

import (
	"fmt"
	"log"
	"encoding/json"
	"sync"
	"strings"
	"regexp"
	"sync/atomic"
	"time"
	"context"
	"requestqueue"
	"net/http"
)

// create response json object
type CreateResponse struct {
	RequestId int32 `json:"RequestId"`
}

// To send error message
type ErrorResponse struct {
	Error string `json:"Error"`
}

// To send stats response
type StatsResponse struct {
	Total string `json:"Total"`
	Average string `json:"Average"`
}

// http server
var server *http.Server

/* 
Queue to store the request list
When user request to create Hash, it is immediately placed in this 
queue for processing.
*/
var requestList *requestqueue.Queue

// Place to store all generated hash
var listOfHash map[string]string

// Total time to generate hash 
var totalTimeToGenerateHash int64

// Request counter - Idetifier returned to the user
type count int32
var pageCount count

// Returns the next available number to be used for queuing the request.
func (c *count) getNextCount() int32 {
	return atomic.AddInt32((*int32)(c), 1)
}

// Returns the current number that has been used for queuing the request.
// This is the total number of passwords which are encrypted so far.
func (c *count) getCurrentCount() int32 {
	return atomic.LoadInt32((*int32)(c))
}

/* 	
	Add the request to create hash in the QUEUE. Add the password to the queue 
 	and returns the id for the hash request in the response.
 	/hash?password=abc
*/
func createHash(w http.ResponseWriter, r *http.Request) {
	pass := r.FormValue("password")
	
	if pass == "" {
		json.NewEncoder(w).Encode(ErrorResponse{Error: "'password' parameter is required."})
	} else {
		requestId := pageCount.getNextCount()

		desc := requestqueue.HashDesc{Id:fmt.Sprintf("%d",requestId), Pass:pass, CreatedTime: time.Now()}
		requestList.Enqueue(desc)

		var createResponse = CreateResponse{RequestId: requestId}
		json.NewEncoder(w).Encode(createResponse)
	}
}

// Returns the hash for the given id. If id does not exist, returns error message.
func getHash(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimLeft(r.URL.Path, "/hash/")

	if id == "" {
		json.NewEncoder(w).Encode(ErrorResponse{Error: "'id' parameter is required to get the hash."})
	} else {
		hash, ok := listOfHash[id]

		if ok {
			json.NewEncoder(w).Encode(hash)
		} else {			
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Hash password not found for "+ id +". Please check the id or retry after 5 seconds."})
		}
	}
}

// Returns the stats of the hash service.
func getStats(w http.ResponseWriter, r *http.Request) {
	noOfHash := int64(len(listOfHash))
	average := totalTimeToGenerateHash/noOfHash

	json.NewEncoder(w).Encode(StatsResponse{Total:fmt.Sprintf("%d",noOfHash), Average:fmt.Sprintf("%d",average)})
}

// Gracefully stops the server. Wait till all user requests are executed.
func stopServer(w http.ResponseWriter, r *http.Request) {
	log.Println("Stopping Hash Service.")

	for len(requestList.Values) > 0 {
		log.Println("SHUTDOWN - waiting for queue to empty")
		time.Sleep(time.Second * 2)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("ERROR WHILE SHUTDOWN.")
    }

	log.Println("Hash Service stopped.")
}

/*
Request handler part based on regular expression
*/
type route struct {
    pattern *regexp.Regexp
    handler http.Handler
}

type RegexpHandler struct {
    routes []*route
}

func (h *RegexpHandler) Handler(pattern *regexp.Regexp, handler http.Handler) {
    h.routes = append(h.routes, &route{pattern, handler})
}

func (h *RegexpHandler) HandleFunc(pattern *regexp.Regexp, handler func(http.ResponseWriter, *http.Request)) {
    h.routes = append(h.routes, &route{pattern, http.HandlerFunc(handler)})
}

func (h *RegexpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    for _, route := range h.routes {
        if route.pattern.MatchString(r.URL.Path) {
            route.handler.ServeHTTP(w, r)
            return
        }
    }
    // no pattern matched; send 404 response
    http.NotFound(w, r)
}

// Main request handlers for all the APIs.
func handleRequest() {
	handler := &RegexpHandler{}

    hashWithID, _ := regexp.Compile("/hash/\\d+")
    handler.HandleFunc(hashWithID, getHash)

	hashRequest, _ := regexp.Compile("/hash")
	handler.HandleFunc(hashRequest, createHash)

	statsRequest, _ := regexp.Compile("/stats")
	handler.HandleFunc(statsRequest, getStats)

	shutdown, _ := regexp.Compile("/shutdown")
	handler.HandleFunc(shutdown, stopServer)
	
	server = &http.Server{Addr: ":8081", Handler: handler}
	log.Fatal(server.ListenAndServe())
	//log.Fatal(http.ListenAndServe(":8081", handler))
}

/* 
 	Function which keeps listening to the Queue where user request to create hash is stored.
 	This will get the entry from the Queue and create Hash for the given password when its 
 	created time has elapsed 5 seconds.
*/
func listenQueue() {
	for {		
		if len(requestList.Values) > 0 {
			hashDesc := requestList.Peek()
			diff := float64(5.0) - time.Since(hashDesc.CreatedTime).Seconds()
			
			if diff > 0 {
				log.Printf("Sleeping before encoding ............... %s \n", time.Second * time.Duration(diff))
				time.Sleep(time.Second * time.Duration(diff))
			}

			hashDesc = requestList.Dequeue()
			hash, duration := hashDesc.Encode()
			listOfHash[hashDesc.Id] = hash
			totalTimeToGenerateHash += duration.Microseconds()

			log.Printf("Created hash for ID %s in %d microseconds \n", hashDesc.Id, duration.Microseconds())
		} else {
			//fmt.Println("Waiting for queue entry....")
			time.Sleep(time.Second * 1)
		}
	}
}

func main() {
	requestList = requestqueue.Init()
	listOfHash = make(map[string]string)
	go listenQueue()

	handleRequest()
	//test(10)
}

func test(n int) {
	// var c count
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			fmt.Println(pageCount.getNextCount()) // increment and get may return different values
			wg.Done()
		}()
	}
	wg.Wait()
}