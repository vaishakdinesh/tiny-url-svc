package url

import (
	"context"
	"encoding/json"
	"math/big"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/vaishakdinesh/tiny-url-svc/types"
)

const (
	defaultExpiryTime = time.Hour * 24 * 362 // 1 year
)

type urlSVC struct {
	l     *zap.Logger
	repo  types.URLRepo
	cache types.CacheService
}

var base58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func NewTinyURLService(l *zap.Logger, r types.URLRepo, c types.CacheService) types.URLService {
	return &urlSVC{
		l:     l,
		repo:  r,
		cache: c,
	}
}

func (u *urlSVC) GenerateTinyURL(ctx context.Context, longUrl string) (types.URLDocument, error) {
	tinyURL := formTinyURL(longUrl)

	err := u.repo.Put(ctx, tinyURL)
	if err != nil {
		u.l.Error("failed to store tiny url in db", zap.Error(err), zap.String("db-key", longUrl))
		return types.URLDocument{}, err
	}
	u.l.Info("added url to db", zap.String(tinyURL.LongURL, strconv.FormatInt(tinyURL.Base10ID, 10)))
	return tinyURL, u.cacheURLs(ctx, tinyURL)
}

func (u *urlSVC) GetTinyURL(ctx context.Context, urlKey string) (types.URLDocument, error) {
	cachedURL, err := u.checkCacheForTinyURLDocument(ctx, urlKey)
	if err != nil {
		u.l.Warn("failed to get cache for long url", zap.Error(err))
	}
	if cachedURL != nil {
		return *cachedURL, nil
	}
	return u.repo.GetDocument(ctx, urlKey)
}

func (u *urlSVC) DeleteTinyURL(ctx context.Context, urlKey string) error {
	err := u.repo.Delete(ctx, urlKey)
	if err != nil {
		u.l.Error("failed to delete from db", zap.Error(err), zap.String("db-key", urlKey))
		return err
	}
	err = u.cache.Delete(ctx, urlKey)
	if err != nil {
		u.l.Warn("failed to delete from cache", zap.Error(err), zap.String("cache-key", urlKey))
	}
	return nil
}

func (u *urlSVC) cacheURLs(ctx context.Context, tinyURL types.URLDocument) error {
	keyBytes, err := json.Marshal(tinyURL)
	if err != nil {
		return err
	}
	// cache generatedKey -> URLDocument
	err = u.cache.Cache(ctx, tinyURL.URLKey, keyBytes)
	if err != nil {
		u.l.Error("failed to cache generated tiny url", zap.Error(err), zap.String("cache-key", tinyURL.URLKey))
	}
	return nil
}

func (u *urlSVC) checkCacheForTinyURLDocument(ctx context.Context, urlKey string) (*types.URLDocument, error) {
	cached, err := u.cache.GetCachedValue(ctx, urlKey)
	if err != nil {
		return nil, err
	}
	if cached == "" {
		return nil, types.ErrCacheNotFound
	}
	cachedURL := new(types.URLDocument)
	err = json.Unmarshal([]byte(cached), cachedURL)
	if err != nil {
		return nil, err
	}
	return cachedURL, err
}

func formTinyURL(longURL string) types.URLDocument {
	cTime := time.Now()
	id := rand.Int64N(cTime.Unix())
	base58String := encode(id)
	urlObj := types.URLDocument{
		Base10ID:   id,
		LongURL:    longURL,
		URLKey:     base58String,
		ExpireTime: cTime.Add(defaultExpiryTime),
	}
	return urlObj
}

func encode(num int64) string {
	bigNum := big.NewInt(num)
	var result strings.Builder

	zero := big.NewInt(0)
	base := big.NewInt(58)
	for bigNum.Cmp(zero) > 0 {
		mod := new(big.Int)
		bigNum.DivMod(bigNum, base, mod)
		result.WriteByte(base58Alphabet[mod.Int64()])
	}
	bytes := []byte(result.String())
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
	return string(bytes)
}
