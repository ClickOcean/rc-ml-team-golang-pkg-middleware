package ginmiddleware

import (
	"context"
	"encoding/json"
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

type serviceData struct {
	ServiceName string                 `json:"service_name"`
	Payload     map[string]interface{} `json:"payload"`
}

func DataSaver(
	client HTTPClient,
	serviceName string,
	cfg DataSaverCfg,
) gin.HandlerFunc {

	logger := log.Default() //TODO: replace custom logger when will be implemented
	logger.SetOutput(os.Stdout)

	return func(c *gin.Context) {
		var req, resp map[string]interface{}

		data := serviceData{
			ServiceName: serviceName,
			Payload:     make(map[string]interface{}),
		}

		err := json.NewDecoder(c.Request.Body).Decode(&req)
		if err != nil {
			logger.Printf("Decoding request caused an error: %s", err.Error())
		} else {
			data.Payload["request"] = req
		}

		c.Next()

		err = json.NewDecoder(c.Request.Response.Body).Decode(&resp)
		if err != nil {
			logger.Printf("Decoding response caused an error: %s", err.Error())
		} else {
			data.Payload["response"] = resp
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
		go func() {
			defer cancel()

			req := requestParams{
				URL:  cfg.URL,
				Body: data,
			}
			resp, err := client.Post(ctx, req)
			switch {
			case err != nil:
				logger.Printf("Sending to the DataSaver caused an error: %s", err.Error())
			case resp.StatusCode > 299:
				var body any

				err := json.NewDecoder(resp.Body).Decode(body)
				if err != nil {
					logger.Printf("Decoding response from the DataSaver caused an error: %s", err.Error())
				}
				logger.Printf("Response from the DataSaver has status: %d, body: %v", resp.StatusCode, body)
			}
		}()
	}
}
