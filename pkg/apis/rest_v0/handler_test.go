package rest_v0

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/vaishakdinesh/tiny-url-svc/pkg/url"
	"github.com/vaishakdinesh/tiny-url-svc/types"
	v0 "github.com/vaishakdinesh/tiny-url-svc/types/api/rest/v0"
)

var apiURL = "/tinyurlsvc"

func TestGenerateURL(t *testing.T) {
	a := assert.New(t)
	l := zap.NewNop()
	r := &types.MockRepo{Data: make(map[string]types.URLDocument)}
	c := &types.MockCache{Data: make(map[string]string)}
	svc := url.NewTinyURLService(l, r, c)
	h, err := NewHandler(l, svc)
	a.NotNil(h)
	a.Nil(err)

	testCases := map[string]struct {
		req           *v0.GenerateURLRequest
		expectedError bool
		pre           func()
		validate      func(a *assert.Assertions, rec *httptest.ResponseRecorder, err error)
	}{
		"successfully create tiny url": {
			req: &v0.GenerateURLRequest{Url: "http://foo.com"},
			validate: func(a *assert.Assertions, rec *httptest.ResponseRecorder, err error) {
				a.Nil(err)
				res := rec.Result()
				defer res.Body.Close()
				a.Equal(http.StatusCreated, res.StatusCode)
				body, err := io.ReadAll(res.Body)
				a.Nil(err)

				tinyURLRes := &v0.GenerateURLResponse{}
				err = json.Unmarshal(body, &tinyURLRes)
				a.Nil(err)
				a.NotEmpty(tinyURLRes.GeneratedTinyURL)
				a.NotEmpty(tinyURLRes.ExpireTime)
			},
		},
		"successfully create tiny url no expire": {
			req: &v0.GenerateURLRequest{Url: "http://foo.com", LiveForever: true},
			validate: func(a *assert.Assertions, rec *httptest.ResponseRecorder, err error) {
				a.Nil(err)
				res := rec.Result()
				defer res.Body.Close()
				a.Equal(http.StatusCreated, res.StatusCode)
				body, err := io.ReadAll(res.Body)
				a.Nil(err)

				tinyURLRes := &v0.GenerateURLResponse{}
				err = json.Unmarshal(body, &tinyURLRes)
				a.Nil(err)
				a.NotEmpty(tinyURLRes.GeneratedTinyURL)
			},
		},
		"successfully create tiny url https": {
			req: &v0.GenerateURLRequest{Url: "https://foo.com"},
			validate: func(a *assert.Assertions, rec *httptest.ResponseRecorder, err error) {
				a.Nil(err)
				res := rec.Result()
				defer res.Body.Close()
				a.Equal(http.StatusCreated, res.StatusCode)
				body, err := io.ReadAll(res.Body)
				a.Nil(err)

				tinyURLRes := &v0.GenerateURLResponse{}
				err = json.Unmarshal(body, &tinyURLRes)
				a.Nil(err)
				a.NotEmpty(tinyURLRes.GeneratedTinyURL)
				a.NotEmpty(tinyURLRes.ExpireTime)
			},
		},
		"tiny url input error": {
			req: &v0.GenerateURLRequest{Url: ""},
			validate: func(a *assert.Assertions, rec *httptest.ResponseRecorder, err error) {
				a.Nil(err)
				res := rec.Result()
				defer res.Body.Close()
				a.Equal(http.StatusBadRequest, res.StatusCode)
				body, err := io.ReadAll(res.Body)
				a.Nil(err)

				tinyURLRes := &v0.APIError{}
				err = json.Unmarshal(body, &tinyURLRes)
				a.Nil(err)
				a.Equal(tinyURLRes.Code, types.InputError)
			},
		},
		"tiny url invalid url": {
			req: &v0.GenerateURLRequest{Url: "?skhasdpasp"},
			validate: func(a *assert.Assertions, rec *httptest.ResponseRecorder, err error) {
				a.Nil(err)
				res := rec.Result()
				defer res.Body.Close()
				a.Equal(http.StatusBadRequest, res.StatusCode)
				body, err := io.ReadAll(res.Body)
				a.Nil(err)

				tinyURLRes := &v0.APIError{}
				err = json.Unmarshal(body, &tinyURLRes)
				a.Nil(err)
				a.Equal(tinyURLRes.Code, types.InputError)
			},
		},
		"tiny url unsupported scheme": {
			req: &v0.GenerateURLRequest{Url: "grpc://service:19081/FooExample/GrpcHello"},
			validate: func(a *assert.Assertions, rec *httptest.ResponseRecorder, err error) {
				a.Nil(err)
				res := rec.Result()
				defer res.Body.Close()
				a.Equal(http.StatusBadRequest, res.StatusCode)
				body, err := io.ReadAll(res.Body)
				a.Nil(err)

				tinyURLRes := &v0.APIError{}
				err = json.Unmarshal(body, &tinyURLRes)
				a.Nil(err)
				a.Equal(tinyURLRes.Code, types.InputError)
				a.Equal(tinyURLRes.Message, types.ErrInvalidScheme.Error())
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			bytes, err := json.Marshal(testCase.req)
			a.Nil(err)

			req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(string(bytes)))
			a.Nil(err)

			ctx, rec := getCTX(req)
			testCase.validate(a, rec, h.GenerateURL(ctx))
		})
	}
}

func TestGetURL(t *testing.T) {
	a := assert.New(t)
	l := zap.NewNop()
	r := &types.MockRepo{Data: make(map[string]types.URLDocument)}
	c := &types.MockCache{Data: make(map[string]string)}
	svc := url.NewTinyURLService(l, r, c)
	h, err := NewHandler(l, svc)
	a.NotNil(h)
	a.Nil(err)

	testCases := map[string]struct {
		urlKey        string
		expectedError bool
		pre           func()
		validate      func(a *assert.Assertions, rec *httptest.ResponseRecorder, err error)
	}{
		"successfully get tiny url": {
			urlKey: "f56Cd",
			pre: func() {
				r.Data["f56Cd"] = types.URLDocument{}
			},
			validate: func(a *assert.Assertions, rec *httptest.ResponseRecorder, err error) {
				a.Nil(err)
				res := rec.Result()
				defer res.Body.Close()
				a.Equal(http.StatusMovedPermanently, res.StatusCode)
			},
		},
		"tiny url not found": {
			urlKey: "6hgtEs",
			validate: func(a *assert.Assertions, rec *httptest.ResponseRecorder, err error) {
				a.Nil(err)
				res := rec.Result()
				defer res.Body.Close()
				a.Equal(http.StatusNotFound, res.StatusCode)
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			if testCase.pre != nil {
				testCase.pre()
			}
			req, err := http.NewRequest(http.MethodDelete, apiURL, nil)
			a.Nil(err)

			ctx, rec := getCTX(req)
			testCase.validate(a, rec, h.GetURL(ctx, testCase.urlKey))
		})
	}
}

func TestDeleteURL(t *testing.T) {
	a := assert.New(t)
	l := zap.NewNop()
	r := &types.MockRepo{Data: make(map[string]types.URLDocument)}
	c := &types.MockCache{Data: make(map[string]string)}
	svc := url.NewTinyURLService(l, r, c)
	h, err := NewHandler(l, svc)
	a.NotNil(h)
	a.Nil(err)

	testCases := map[string]struct {
		urlKey        string
		expectedError bool
		pre           func()
		validate      func(a *assert.Assertions, rec *httptest.ResponseRecorder, err error)
	}{
		"successfully delete tiny url": {
			urlKey: "f56Cd",
			pre: func() {
				r.Data["f56Cd"] = types.URLDocument{}
			},
			validate: func(a *assert.Assertions, rec *httptest.ResponseRecorder, err error) {
				a.Nil(err)
				res := rec.Result()
				defer res.Body.Close()
				a.Equal(http.StatusNoContent, res.StatusCode)
			},
		},
		"tiny url not found": {
			urlKey: "6hgtEs",
			validate: func(a *assert.Assertions, rec *httptest.ResponseRecorder, err error) {
				a.Nil(err)
				res := rec.Result()
				defer res.Body.Close()
				a.Equal(http.StatusNotFound, res.StatusCode)
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			if testCase.pre != nil {
				testCase.pre()
			}
			req, err := http.NewRequest(http.MethodDelete, apiURL, nil)
			a.Nil(err)

			ctx, rec := getCTX(req)
			testCase.validate(a, rec, h.DeleteURL(ctx, testCase.urlKey))
		})
	}
}

func getCTX(r *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	s := echo.New()
	rec := httptest.NewRecorder()
	ctx := s.NewContext(r, rec)
	return ctx, rec
}
