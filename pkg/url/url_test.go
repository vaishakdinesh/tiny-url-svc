package url

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/vaishakdinesh/tiny-url-svc/types"
)

func TestEncoder(t *testing.T) {
	testCases := map[string]struct {
		input    int64
		expected string
	}{
		"encode 2468135791013": {
			input:    2468135791013,
			expected: "27qMi57J",
		},
		"encode 7489135791013": {
			input:    7489135791013,
			expected: "4PjAHW6Y",
		},
		"encode 5638910482": {
			input:    5638910482,
			expected: "9bHtdX",
		},
		"encode 00000000000": {
			input:    87452840931,
			expected: "3JEufoG",
		},
	}
	a := assert.New(t)
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			a.Equal(testCase.expected, encode(testCase.input))
		})
	}
}

func TestGenerateTinyURL(t *testing.T) {
	ctx := context.Background()
	a := assert.New(t)
	l := zap.NewNop()
	r := &types.MockRepo{Data: make(map[string]types.URLDocument)}
	c := &types.MockCache{Data: make(map[string]string)}
	testCases := map[string]struct {
		lURL          string
		liveForever   bool
		expectedError bool
	}{
		"gen url": {
			lURL: "https://abc.io",
		},
		"gen url with query param ": {
			lURL: "https://abc.efg.io?page=1&id=100",
		},
		"gen url live forever ": {
			lURL:        "https://abc.efg.io?page=1&id=100",
			liveForever: true,
		},
		"gen localhost url ": {
			lURL: "http://localhost:9090/api/v1/foo?id=100",
		},
		"error": {
			lURL:          "",
			expectedError: true,
		},
		"db put fail": {
			lURL:          types.StoreFail,
			expectedError: true,
		},
		"db get fail": {
			lURL:          types.GetFail,
			expectedError: true,
		},
	}

	svc := NewTinyURLService(l, r, c)
	a.NotNil(svc)

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			tURL, err := svc.GenerateTinyURL(ctx, testCase.lURL, testCase.liveForever)
			if !testCase.expectedError {
				a.Nil(err)
				a.NotEmpty(tURL.Base10ID)
				a.NotEmpty(tURL.URLKey)
				if !testCase.liveForever {
					a.NotEmpty(tURL.ExpireTime)
				}
			} else {
				a.Error(err)
			}
		})
	}
}

func TestGetTinyURL(t *testing.T) {
	ctx := context.Background()
	a := assert.New(t)
	l := zap.NewNop()
	r := &types.MockRepo{Data: make(map[string]types.URLDocument)}
	c := &types.MockCache{Data: make(map[string]string)}
	testCases := map[string]struct {
		urlKey      string
		expectError bool
		expectedErr error
		pre         func(a *assert.Assertions)
	}{
		"failed to get tiny url": {
			urlKey:      "JV5pY",
			expectError: true,
			expectedErr: types.ErrDocumentNotFound,
		},
		"get tiny url": {
			urlKey: "342dLy",
			pre: func(a *assert.Assertions) {
				r.Data["342dLy"] = types.URLDocument{
					Base10ID:   1029208386,
					URLKey:     "342dLy",
					LongURL:    "https://foo.com?id=1",
					ExpireTime: time.Now(),
				}
			},
		},
		"get tiny url from cache": {
			urlKey: "GdMuR",
			pre: func(a *assert.Assertions) {
				doc := types.URLDocument{
					Base10ID:   1029208386,
					URLKey:     "GdMuR",
					LongURL:    "https://foo.com?id=1",
					ExpireTime: time.Now().Add(time.Hour * 1),
				}
				bytes, err := json.Marshal(doc)
				a.Nil(err)
				c.Data["GdMuR"] = string(bytes)
			},
		},
		"expired url, delete from cache": {
			urlKey: "GdMuR",
			pre: func(a *assert.Assertions) {
				doc := types.URLDocument{
					Base10ID:   1029208386,
					URLKey:     "GdMuR",
					LongURL:    "https://foo.com?id=1",
					ExpireTime: time.Now(),
				}
				bytes, err := json.Marshal(doc)
				a.Nil(err)
				c.Data["GdMuR"] = string(bytes)
			},
			expectError: true,
			expectedErr: types.ErrDocumentNotFound,
		},
	}

	svc := NewTinyURLService(l, r, c)
	a.NotNil(svc)
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			if testCase.pre != nil {
				testCase.pre(a)
			}
			tURL, err := svc.GetTinyURL(ctx, testCase.urlKey)
			if !testCase.expectError {
				a.Nil(err)
				a.NotEmpty(tURL.Base10ID)
				a.NotEmpty(tURL.URLKey)
				a.NotEmpty(tURL.ExpireTime)
			} else {
				a.Error(err)
				a.Equal(err, testCase.expectedErr)
			}
		})
	}
}

func TestDeleteTinyURL(t *testing.T) {
	ctx := context.Background()
	a := assert.New(t)
	l := zap.NewNop()
	r := &types.MockRepo{Data: make(map[string]types.URLDocument)}
	c := &types.MockCache{Data: make(map[string]string)}
	testCases := map[string]struct {
		urlKey        string
		expectedError bool
		pre           func(a *assert.Assertions)
	}{
		"failed to get tiny url": {
			urlKey:        "VpPmN",
			expectedError: true,
		},
		"get tiny url": {
			urlKey: "CFght",
			pre: func(a *assert.Assertions) {
				r.Data["CFght"] = types.URLDocument{
					Base10ID:   1029208386,
					URLKey:     "CFght",
					LongURL:    "https://foo.com?id=1",
					ExpireTime: time.Now(),
				}
			},
		},
		"delete tiny url from cache and db": {
			urlKey: "Yc651",
			pre: func(a *assert.Assertions) {
				doc := types.URLDocument{
					Base10ID:   1029208386,
					URLKey:     "Yc651",
					LongURL:    "https://foo.com?id=1",
					ExpireTime: time.Now(),
				}
				bytes, err := json.Marshal(doc)
				a.Nil(err)
				c.Data["Yc651"] = string(bytes)
				r.Data["Yc651"] = doc
			},
		},
		"cache get fail": {
			urlKey:        types.CacheGetFail,
			expectedError: true,
		},
	}

	svc := NewTinyURLService(l, r, c)
	a.NotNil(svc)

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			if testCase.pre != nil {
				testCase.pre(a)
			}
			err := svc.DeleteTinyURL(ctx, testCase.urlKey)
			if !testCase.expectedError {
				a.Nil(err)
			} else {
				a.Error(err)
			}
		})
	}
}
