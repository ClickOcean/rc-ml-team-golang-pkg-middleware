package ginmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type rateSuite struct {
	suite.Suite
	doReq func() *httptest.ResponseRecorder
}

func TestRateLimier(t *testing.T) {
	suite.Run(t, new(rateSuite))
}

func (s *rateSuite) SetupSuite() {
	e := gin.New()
	e.Use(RateLimit(1))

	e.GET("/", func(ctx *gin.Context) {
		ctx.AbortWithStatus(http.StatusOK)
	})

	s.doReq = func() *httptest.ResponseRecorder {
		return PerformRequest(e, http.MethodGet, "/", http.NoBody)
	}
}

func (s *rateSuite) TestRate() {
	time.Sleep(time.Second)

	rec := s.doReq()
	s.Equal(http.StatusOK, rec.Result().StatusCode)

	rec = s.doReq()
	s.Equal(http.StatusTooManyRequests, rec.Result().StatusCode)

}
