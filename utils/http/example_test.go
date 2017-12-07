// http包测试 author chenweidong

package http

import (
	"fmt"
	"testing"
	"net/http"
	"net/url"
)

func TestParamToMap(t *testing.T) {
	req := &http.Request{URL: &url.URL{RawQuery: "a=12&b=something"}}
	fmt.Println(ParamToMap(req))
}
