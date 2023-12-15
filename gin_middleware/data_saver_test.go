package ginmiddleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
	client *MockHTTPClient
	doReq  func(method, path string, body io.Reader) *httptest.ResponseRecorder
	cfg    testParams
}

type testParams struct {
	dataSaverURL string
	serviceName  string
}

func TestDataSaver(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) SetupSuite() {
	engine := gin.New()

	s.client = &MockHTTPClient{}
	s.cfg = testParams{
		dataSaverURL: "http://example.com",
		serviceName:  "test",
	}

	engine.Use(DataSaver(s.client, s.cfg.serviceName, s.cfg.dataSaverURL))

	engine.GET("/", func(ctx *gin.Context) {
		ctx.AbortWithStatus(http.StatusNoContent)
	})

	engine.POST("/", func(ctx *gin.Context) {
		var inp map[string]any
		err := json.NewDecoder(ctx.Request.Body).Decode(&inp)
		if s.NoError(err) {
			delete(inp, "input")
			inp["output"] = "data"
			ctx.AbortWithStatusJSON(http.StatusOK, inp)
			return
		}
		ctx.AbortWithStatus(http.StatusInternalServerError)
	})

	s.doReq = func(method, path string, body io.Reader) *httptest.ResponseRecorder {
		return PerformRequest(engine, method, path, body)
	}

}

func (s *Suite) TearDownSuite() {
	s.True(
		mock.AssertExpectationsForObjects(
			s.T(),
			s.client,
		),
	)
}

func (s *Suite) TestDataSaverMiddleware() {
	reader := strings.NewReader(`{"input":"data"}`)

	s.client.On(
		"Do",
		mock.AnythingOfType("*http.Request"),
	).Run(func(args mock.Arguments) {
		var data map[string]any

		req := args.Get(0).(*http.Request)
		err := json.NewDecoder(req.Body).Decode(&data)
		s.NoError(err)
	}).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer([]byte{})),
	}, nil).Once()

	recorder := s.doReq(http.MethodPost, "/", reader)
	resp := recorder.Result()
	if s.Equal(http.StatusOK, resp.StatusCode) {
		var respBody map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&respBody)
		s.Equal("data", respBody["output"])
	}

	time.Sleep(50 * time.Millisecond)
}

func (s *Suite) TestDataSaverMiddlewareWOBody() {

	s.client.On(
		"Do",
		mock.AnythingOfType("*http.Request"),
	).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer([]byte{})),
	}, nil).Once()

	recorder := s.doReq(http.MethodGet, "/", http.NoBody)
	resp := recorder.Result()
	s.Equal(http.StatusNoContent, resp.StatusCode)

	time.Sleep(50 * time.Millisecond)
}

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func PerformRequest(r http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
