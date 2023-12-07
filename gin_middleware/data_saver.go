package ginmiddleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
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
	dataSaverURL string,
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
		lenDataBefore := data.Len()

		tee := io.TeeReader(c.Request.Body, &data)
		c.Request.Body = io.NopCloser(tee)
		c.Next()

		if lenDataBefore == data.Len() {
			data.WriteString(`null`)
		}

		data.WriteString(`,"response":`)
		written, err := data.ReadFrom(&respBuf)
		if err != nil {
			logger.Printf("Reading response buffer caused an error: %s", err.Error())
		}
		if written == 0 {
			data.WriteString(`null`)
		}
		data.WriteString(`}}`)

		go func() {
			if !json.Valid(data.Bytes()) {
				logger.Printf("JSON isn't valid: %s", data.String())
				return
			}

			req, err := http.NewRequest(http.MethodPost, dataSaverURL, &data)
			if err != nil {
				logger.Printf("Building http request caused an error: %s", err.Error())
				return
			}

			resp, err := client.Do(req)
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
