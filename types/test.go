package types

import (
	"context"
	"errors"
)

type (
	MockRepo struct {
		Data map[string]URLDocument
	}
	MockCache struct {
		Data map[string]string
	}
)

const (
	StoreFail      = "StoreFail"
	GetFail        = "GetFail"
	CacheStoreFail = "CacheStoreFail"
	CacheGetFail   = "CacheGetFail"
	Empty          = ""
)

func (mr *MockRepo) Put(_ context.Context, document any) error {
	switch o := document.(type) {
	case URLDocument:
		switch {
		case o.LongURL == StoreFail:
			return errorCondition(StoreFail)
		case o.LongURL == GetFail:
			return errorCondition(GetFail)
		case o.LongURL == Empty:
			return errorCondition(Empty)
		}
		mr.Data[o.URLKey] = o
	}
	return nil
}

func (mr *MockRepo) GetDocument(_ context.Context, urlKey string) (URLDocument, error) {
	doc, ok := mr.Data[urlKey]
	if !ok {
		return URLDocument{}, ErrDocumentNotFound
	}
	return doc, nil
}

func (mr *MockRepo) Delete(_ context.Context, urlKey string) error {
	doc, ok := mr.Data[urlKey]
	if !ok {
		return ErrDocumentNotFound
	}
	delete(mr.Data, doc.URLKey)
	return nil
}

func (mc *MockCache) Cache(_ context.Context, key string, val any) error {
	switch o := val.(type) {
	case string:
		mc.Data[key] = o
	}
	return nil
}

func (mc *MockCache) Delete(_ context.Context, key string) error {
	switch key {
	case CacheStoreFail:
		return errorCondition(StoreFail)
	case CacheGetFail:
		return errorCondition(StoreFail)
	case Empty:
		return errorCondition(Empty)
	default:
		_, ok := mc.Data[key]
		if !ok {
			return ErrDocumentNotFound
		}
		delete(mc.Data, key)
		return nil
	}
}

func (mc *MockCache) GetCachedValue(_ context.Context, key string) (string, error) {
	switch key {
	case CacheGetFail:
		return "", errorCondition(StoreFail)
	case Empty:
		return "", errorCondition(Empty)
	default:
		val, ok := mc.Data[key]
		if !ok {
			return "", nil
		}
		return val, nil
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
