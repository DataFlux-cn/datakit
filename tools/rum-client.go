// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"net/http"
)

var (
	fHost   = flag.String("host", "", "")
	fReqCnt = flag.Int("req-cnt", 100, "")
)

func rangeRand(min, max int64) int64 {
	if min > max {
		panic("the min is greater than max!")
	}

	if min < 0 {
		f64Min := math.Abs(float64(min))
		i64Min := int64(f64Min)
		result, err := rand.Int(rand.Reader, big.NewInt(max+1+i64Min))
		if err != nil {
			return -1
		}

		return result.Int64() - i64Min
	}

	result, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		return -1
	}

	return min + result.Int64()
}

func main() {
	flag.Parse()
	cli := http.Client{}

	for i := 0; i < *fReqCnt; i++ {
		req, err := http.NewRequest("GET", *fHost, nil)
		if err != nil {
			log.Fatal(err)
		}

		i := rangeRand(1000000000000000000, 8888888888888888888)
		log.Println(i)
		req.Header.Set("X-Datadog-Origin", "rum")
		req.Header.Set("X-Datadog-Trace-ID", fmt.Sprintf("%d", i))
		req.Header.Set("X-Datadog-Sampled", "0")
		req.Header.Set("X-Datadog-Origin", "rum")
		req.Header.Set("x-datadog-sampling-priority", "0")

		resp, err := cli.Do(req)
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			continue
		}

		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}
}
