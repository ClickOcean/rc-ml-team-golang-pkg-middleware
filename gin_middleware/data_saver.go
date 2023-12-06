package ginmiddleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type HTTPClient interface {
	Post(ctx context.Context, req requestParams) (*http.Response, error)
}

type DataSaverCfg struct {
	URL     string
	Timeout int //Second
}

type requestParams struct {
	URL           string
	Headers       map[string]string
	Body          any
	ErrorResult   any
	SuccessResult any
}

type multiWriter struct {
	gin.ResponseWriter
	w io.Writer
}

func (mw multiWriter) Write(b []byte) (int, error) {
	return mw.w.Write(b)
}

func DataSaver(
	client HTTPClient,
	serviceName string,
	cfg DataSaverCfg,
) gin.HandlerFunc {

	logger := log.Default() //TODO: replace custom logger when will be implemented
	logger.SetOutput(os.Stdout)

	return func(c *gin.Context) {
		//data has format
		//{"service_name":"some_name","payload":{"request":any,"response":any}}

		var (
			data, respBuf bytes.Buffer
		)

		mWriter := &multiWriter{
			ResponseWriter: c.Writer,
			w:              io.MultiWriter(c.Writer, &respBuf),
		}
		c.Writer = mWriter

		data.WriteString(fmt.Sprintf(`{"service_name":"%s",`, serviceName))
		data.WriteString(`"payload":{"request":`)

		tee := io.TeeReader(c.Request.Body, &data)
		c.Request.Body = io.NopCloser(tee)
		c.Next()

		data.WriteString(`,"response":`)

		_, err := data.ReadFrom(&respBuf)
		if err != nil {
			logger.Printf("Reading response buffer caused an error: %s", err.Error())
		}
		data.WriteString(`}}`)

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
		go func() {
			defer cancel()

			if !json.Valid(data.Bytes()) {
				logger.Printf("JSON isn't valid: %s", data.String())
				return
			}

			req := requestParams{
				URL:  cfg.URL,
				Body: data,
			}

			resp, err := client.Post(ctx, req)
			switch {
			case err != nil:
				logger.Printf("Sending to the DataSaver caused an error: %s", err.Error())
				return
			case resp.StatusCode > 299:
				var body any

				err := json.NewDecoder(resp.Body).Decode(&body)
				if err != nil {
					logger.Printf("Decoding response from the DataSaver caused an error: %s", err.Error())
					return
				}
				logger.Printf("Response from the DataSaver has status: %d, body: %v", resp.StatusCode, body)
			}
		}()
	}
}
