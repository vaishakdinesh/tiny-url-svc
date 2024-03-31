package url

import (
	"context"
	"errors"
	"github.com/vaishakdinesh/tiny-url-svc/types"
	"go.uber.org/zap"
	"testing"

	"github.com/stretchr/testify/assert"
)

type (
	mockRepo struct {
		data map[string]types.URLDocument
	}
	mockCache struct {
		data map[string]string
	}
)

const (
	StoreFail      = "StoreFail"
	GetFail        = "GetFail"
	CacheStoreFail = "CacheStoreFail"
	CacheGetFail   = "CacheGetFail"
	Empty          = ""
)

func (mr *mockRepo) Put(_ context.Context, document any) error {
	switch o := document.(type) {
	case types.URLDocument:
		switch {
		case o.LongURL == StoreFail:
			return errorCondition(StoreFail)
		case o.LongURL == GetFail:
			return errorCondition(GetFail)
		case o.LongURL == Empty:
			return errorCondition(Empty)
		}
		mr.data[o.URLKey] = o
	}
	return nil
}

func (mr *mockRepo) GetDocument(_ context.Context, urlKey string) (types.URLDocument, error) {
	switch urlKey {
	case StoreFail:
		return types.URLDocument{}, errorCondition(StoreFail)
	case GetFail:
		return types.URLDocument{}, errorCondition(StoreFail)
	case CacheStoreFail:
		return types.URLDocument{}, errorCondition(StoreFail)
	case CacheGetFail:
		return types.URLDocument{}, errorCondition(StoreFail)
	case Empty:
		return types.URLDocument{}, errorCondition(Empty)
	default:
		doc, ok := mr.data[urlKey]
		if !ok {
			return types.URLDocument{}, types.ErrDocumentNotFound
		}
		return doc, nil
	}
}

func (mr *mockRepo) Delete(_ context.Context, urlKey string) error {
	switch urlKey {
	case StoreFail:
		return errorCondition(StoreFail)
	case GetFail:
		return errorCondition(StoreFail)
	case CacheStoreFail:
		return errorCondition(StoreFail)
	case CacheGetFail:
		return errorCondition(StoreFail)
	case Empty:
		return errorCondition(Empty)
	default:
		doc, ok := mr.data[urlKey]
		if !ok {
			return types.ErrDocumentNotFound
		}
		delete(mr.data, doc.URLKey)
		return nil
	}
}

func (mc *mockCache) Cache(_ context.Context, key string, val any) error {
	switch key {
	case StoreFail:
		return errorCondition(StoreFail)
	case GetFail:
		return errorCondition(StoreFail)
	case CacheStoreFail:
		return errorCondition(StoreFail)
	case CacheGetFail:
		return errorCondition(StoreFail)
	case Empty:
		return errorCondition(Empty)
	default:
		switch o := val.(type) {
		case string:
			mc.data[key] = o
		}
		return nil
	}
}

func (mc *mockCache) Delete(_ context.Context, key string) error {
	switch key {
	case StoreFail:
		return errorCondition(StoreFail)
	case GetFail:
		return errorCondition(StoreFail)
	case CacheStoreFail:
		return errorCondition(StoreFail)
	case CacheGetFail:
		return errorCondition(StoreFail)
	case Empty:
		return errorCondition(Empty)
	default:
		_, ok := mc.data[key]
		if !ok {
			return types.ErrDocumentNotFound
		}
		delete(mc.data, key)
		return nil
	}
}

func (mc *mockCache) GetCachedValue(_ context.Context, key string) (string, error) {
	switch key {
	case StoreFail:
		return "", errorCondition(StoreFail)
	case GetFail:
		return "", errorCondition(StoreFail)
	case CacheStoreFail:
		return "", errorCondition(StoreFail)
	case CacheGetFail:
		return "", errorCondition(StoreFail)
	case Empty:
		return "", errorCondition(Empty)
	default:
		val, ok := mc.data[key]
		if !ok {
			return "", nil
		}
		return val, nil
	}
}

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
	testCases := map[string]struct {
		lURL          string
		expectedError bool
		pre           func()
		testRepoMap   map[string]types.URLDocument
		testCache     map[string]string
	}{
		"gen url": {
			lURL: "https://abc.io",
		},
		"gen url with query param ": {
			lURL: "https://abc.efg.io?page=1&id=100",
		},
		"gen localhost url ": {
			lURL: "http://localhost:9090/api/v1/foo?id=100",
		},
		"error": {
			lURL:          "",
			expectedError: true,
		},
		"db put fail": {
			lURL:          StoreFail,
			expectedError: true,
		},
		"db get fail": {
			lURL:          GetFail,
			expectedError: true,
		},
	}
	a := assert.New(t)
	l := zap.NewNop()
	r := &mockRepo{data: make(map[string]types.URLDocument)}
	c := &mockCache{data: make(map[string]string)}
	svc := NewTinyURLService(l, r, c)
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			tURL, err := svc.GenerateTinyURL(ctx, testCase.lURL)
			if !testCase.expectedError {
				a.Nil(err)
				a.NotEmpty(tURL.Base10ID)
				a.NotEmpty(tURL.URLKey)
				a.NotEmpty(tURL.ExpireTime)
			} else {
				a.Error(err)
			}
		})
	}
}

func errorCondition(e string) error {
	switch e {
	case StoreFail:
		return errors.New("failed insert object")
	case GetFail:
		return errors.New("failed get object")
	case CacheStoreFail:
		return errors.New("failed to cache")
	case CacheGetFail:
		return errors.New("failed to get cache")
	default:
		return errors.New("something went wrong")
	}
}
