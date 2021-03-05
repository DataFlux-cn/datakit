package httpProb

import (
	"testing"
	"net/http"
	"bytes"
	"net/http/httptest"
)

func TestMain(t *testing.T) {
	var prob = &HttpProb{
		Bind:  "0.0.0.0",
		Port: 8090,
		Source: "test-app",
		Url:    []*Url{
			{
				Uri: "/",
			},
		},
	}

	// data := `
	//    {
	//    	  "data": "test data",
	//    }
	// `

	data := "aaaaaaaaaa"

	request, _ := http.NewRequest(http.MethodPost, "/test?key1=value1&key2=value2&key3=value3", bytes.NewBuffer([]byte(data)))
	request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8");
	request.Header.Add("Accept-Encoding", "gzip, deflate");
	request.Header.Add("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3");
	request.Header.Add("Connection", "keep-alive");
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0");

    response := httptest.NewRecorder()

    prob.ServeHTTP(response, request)

    got := response.Body.String()
    want := "ok"

    if got != want {
    	t.Errorf("got '%s', want '%s'", got, want)
    }
}
