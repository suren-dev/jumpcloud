# jumpcloud-exercise
JumpCloud coding exercise repository.

HashService 

This is a simple HashService which starts a http server on port 8081. 
It exposes the following REST endpoints

/hash
	This endpoint is used to create a hash for the given password. It accepts POST request with a form field "password". When invoked, it returns the 'RequestId' of the hash request. (Please make a note of this 'RequestId'. It is used to get the actual hash value of the provided password).

	NOTE: Hash for the given password is not created immediately. Hash will be created after 5 seconds. 

	Example:
	> curl -d password=abc http://localhost:8081/hash
	{"RequestId":1}


/hash/{RequestId}
	This endpoint is used to get the hash value using the 'RequestId'. 'RequestId' is the idetifier retured from the POST call to /hash. If the hash with a particular 'RequestId' is accessed before 5 seconds OR if the 'RequestId' does not exist, it will return a error message.

	Example:
	> curl http://localhost:8081/hash/1
	"3a81oZNherrMQXNJriBBMRLm+k6JqX6iCp7u5ktV05ohkpkqJ0/BqDa6PCOj/uu9RU1EI2Q86A4qmslPpUyknw=="

/stats
	This endpoint is used to get the current statistics of the hash service. It will return the total number of POST requests that has been made to hash service and the average time taken to create the hash value in microseconds.

	Example:
	> curl http://localhost:8081/stats
	{"Total":"1","Average":"23"}

/shutdown
	This endpoint is used to gracefully shutdown the hash http service. If there are any pending request, they will be completed before shutting down.

	Example:
	> curl http://localhost:8081/shutdown

Usage

1. Execute the main.go (go run main.go) to run the http server on port 8081 (default port)


