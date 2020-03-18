package utils

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/baetyl/baetyl-function/common"
	routing "github.com/qiangxue/fasthttp-routing"
)

func ResolveAddress(serviceName string) string {
	// TODO: using serviceName to get ip:port
	return "0.0.0.0:50080"
}

func SetHeaders(c *routing.Context, metadata map[string]string) {
	var headers []string
	c.RequestCtx.Request.Header.VisitAll(func(key, value []byte) {
		k := string(key)
		v := string(value)

		headers = append(headers, fmt.Sprintf("%s%s%s", k, common.HeaderEquals, v))
	})
	if len(headers) > 0 {
		metadata["headers"] = strings.Join(headers, common.HeaderDelim)
	}
}

func SetHeadersOnRequest(metadata map[string]string, c *routing.Context) {
	if metadata == nil {
		return
	}
	if val, ok := metadata["headers"]; ok {
		headers := strings.Split(val, common.HeaderDelim)
		for _, h := range headers {
			kv := strings.Split(h, common.HeaderEquals)
			c.RequestCtx.Response.Header.Set(kv[0], kv[1])
		}
	}
}

// GetStatusCodeFromMetadata extracts the http status code from the metadata if it exists
func GetStatusCodeFromMetadata(metadata map[string]string) int {
	code := metadata[common.HTTPStatusCode]
	if code != "" {
		statusCode, err := strconv.Atoi(code)
		if err == nil {
			return statusCode
		}
	}

	return http.StatusOK
}
