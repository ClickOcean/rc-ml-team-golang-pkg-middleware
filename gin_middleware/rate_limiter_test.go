package ginmiddleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type rateSuite struct {
	suite.Suite
	doReq func() *httptest.ResponseRecorder
	limit int
}

func TestRateLimier(t *testing.T) {
	suite.Run(t, new(rateSuite))
}

func (s *rateSuite) SetupTest() {
	s.limit = 10
	e := gin.New()
	e.Use(RateLimit(s.limit))

	e.GET("/", func(ctx *gin.Context) {
		ctx.AbortWithStatus(http.StatusOK)
	})

	s.doReq = func() *httptest.ResponseRecorder {
		return PerformRequest(e, http.MethodGet, "/", http.NoBody)
	}
}

func (s *rateSuite) TestRate() {
	time.Sleep(time.Second)

	for i := 0; i < s.limit; i++ {
		rec := s.doReq()
		s.Equal(http.StatusOK, rec.Result().StatusCode)
	}

	rec := s.doReq()
	headers := rec.Result().Header
	s.Equal(http.StatusTooManyRequests, rec.Result().StatusCode)
	s.Equal(strconv.Itoa(s.limit), headers.Get("X-RateLimit-Limit"))
}

func (s *rateSuite) TestConcurrencyReq() {
	time.Sleep(time.Second)

	statusCh := make(chan int)
	statuses := []int{}

	increased := 5

	go func() {
		for s := range statusCh {
			statuses = append(statuses, s)
		}
	}()

	wg := new(sync.WaitGroup)
	for i := 0; i < s.limit+increased; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rec := s.doReq()
			statusCh <- rec.Result().StatusCode
		}()
	}

	wg.Wait()
	close(statusCh)
	time.Sleep(10 * time.Millisecond)

	count := 0
	for _, status := range statuses {
		if status == http.StatusTooManyRequests {
			count = count + 1
		}
	}

	s.Equal(increased, count)

}
