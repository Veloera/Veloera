package common

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"strings"
)

const KeyRequestBody = "key_request_body"

func GetRequestBody(c *gin.Context) ([]byte, error) {
	requestBody, _ := c.Get(KeyRequestBody)
	if requestBody != nil {
		return requestBody.([]byte), nil
	}
	requestBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	_ = c.Request.Body.Close()
	c.Set(KeyRequestBody, requestBody)
	return requestBody.([]byte), nil
}

func UnmarshalBodyReusable(c *gin.Context, v any) error {
	requestBody, err := GetRequestBody(c)
	if err != nil {
		return err
	}
	contentType := c.Request.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		err = json.Unmarshal(requestBody, &v)
	} else {
		// skip for now
		// TODO: someday non json request have variant model, we will need to implementation this
	}
	if err != nil {
		return err
	}
	// Reset request body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	return nil
}

// ResponseWriter is a wrapper for gin.ResponseWriter to capture the status code.
type ResponseWriter struct {
	gin.ResponseWriter
	status int
}

// WriteHeader captures the status code before calling the original WriteHeader.
func (w *ResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Status returns the captured status code.
func (w *ResponseWriter) Status() int {
	return w.status
}
