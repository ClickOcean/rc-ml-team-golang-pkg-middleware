package ginmiddleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type multiWriter struct {
	gin.ResponseWriter
	w io.Writer
}

type dataSaverReq struct {
	ServiceName string  `json:"service_name"`
	Payload     payload `json:"payload"`
}

type payload struct {
	Request            json.RawMessage `json:"request,omitempty"`
	Response           json.RawMessage `json:"response,omitempty"`
	RequestDateTime    string          `json:"request_date_time,omitempty"`
	Headers            http.Header     `json:"headers,omitempty"`
	URI                string          `json:"uri"`
	ResponseStatusCode int             `json:"response_status_code"`
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

		var (
			respBuff, reqBuff bytes.Buffer
		)

		reqTime := time.Now().UTC().Format(time.DateTime)

		mWriter := &multiWriter{
			ResponseWriter: c.Writer,
			w:              io.MultiWriter(c.Writer, &respBuff),
		}
		c.Writer = mWriter

		tee := io.TeeReader(c.Request.Body, &reqBuff)
		c.Request.Body = io.NopCloser(tee)

		c.Next()

		reqBody := dataSaverReq{
			ServiceName: serviceName,
			Payload: payload{
				Request:            reqBuff.Bytes(),
				Response:           respBuff.Bytes(),
				RequestDateTime:    reqTime,
				Headers:            c.Request.Header.Clone(),
				URI:                c.Request.RequestURI,
				ResponseStatusCode: c.Writer.Status(),
			},
		}

		go func() {
			body, err := json.Marshal(reqBody)
			if err != nil {
				logger.Printf("Marshaling request caused an error: %s", err.Error())
				return
			}

			req, err := http.NewRequest(http.MethodPost, dataSaverURL, bytes.NewBuffer(body))
			if err != nil {
				logger.Printf("Building http request caused an error: %s", err.Error())
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				logger.Printf("Sending to the DataSaver caused an error: %s", err.Error())
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode > 299 {
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
