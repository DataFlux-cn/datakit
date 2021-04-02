package redis

import (
	"testing"
	"fmt"
	"github.com/go-redis/redis"
	// "github.com/stretchr/testify/assert"
)

func TestDBMeasurement(t *testing.T) {
	cli := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "dev", // no password set
        DB:       0,  // use default DB
    })

	info := CollectDBMeasurement(cli)

	fmt.Println(info.fields)
}