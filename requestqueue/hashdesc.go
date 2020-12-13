package requestqueue

import (
	"fmt"
	"time"
	"crypto/sha512"
    "encoding/base64"
)

// HashDesc stores the user provided password to be encrypted along with its Id and created time.
type HashDesc struct {
	Id string
	Pass string
	CreatedTime time.Time
}

// Returns the base64 encoded string of SHA512 checksum created for the given password string.
func (hash *HashDesc) Encode() (string, time.Duration) {
	fmt.Println("Encoding password ", hash.Pass);
	startTime := time.Now()
    shaEncoded := sha512.Sum512([]byte(hash.Pass));
	encoded := base64.StdEncoding.EncodeToString(shaEncoded[:]);
	timeToEncode := time.Now().Sub(startTime)
    return encoded, timeToEncode;
}