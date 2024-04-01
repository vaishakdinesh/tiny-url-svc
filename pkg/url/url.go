package url

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"math/big"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/vaishakdinesh/tiny-url-svc/types"
)

const (
	defaultExpiryTime = time.Hour * 24 * 365 // 1 year
)

type urlSVC struct {
	l       *zap.Logger
	repo    types.URLRepo
	cache   types.CacheService
	counter *prometheus.CounterVec
}

var base58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func NewTinyURLService(l *zap.Logger, r types.URLRepo, c types.CacheService) types.URLService {
	svc := &urlSVC{
		l:     l,
		repo:  r,
		cache: c,
		counter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:      "tiny_url_usage",
			Namespace: "tiny_url_svc",
		},
			[]string{"url_key"}), // labels for cardinality
	}
	return svc
}

func (u *urlSVC) RegisterProm() error {
	return prometheus.Register(u.counter)
}

func (u *urlSVC) GenerateTinyURL(ctx context.Context, longUrl string, liveForever bool) (types.URLDocument, error) {
	tinyURL := formTinyURL(longUrl, liveForever)
	err := u.repo.Put(ctx, tinyURL)
	if err != nil {
		u.l.Error("failed to store tiny url in db", zap.Error(err), zap.String("db-key", longUrl))
		return types.URLDocument{}, err
	}
	u.l.Info("added url to db", zap.String(tinyURL.LongURL, strconv.FormatInt(tinyURL.Base10ID, 10)))
	return tinyURL, u.cacheTinyURL(ctx, tinyURL)
}

func (u *urlSVC) GetTinyURL(ctx context.Context, urlKey string) (types.URLDocument, error) {
	var cacheAgain bool
	u.counter.WithLabelValues(urlKey).Inc()
	cachedURL, err := u.checkCacheForTinyURLDocument(ctx, urlKey)
	if err != nil {
		cacheAgain = errors.Is(err, redis.Nil)
		u.l.Warn("failed to get cache for long url", zap.Error(err))
	}
	if cachedURL != nil {
		if !cachedURL.LiveForever && cachedURL.ExpireTime.Before(time.Now()) {
			if cErr := u.cache.Delete(ctx, urlKey); cErr != nil && !errors.Is(cErr, types.ErrCacheNotFound) {
				u.l.Error("failed to delete cache", zap.Error(cErr), zap.String("db-key", urlKey))
			}
			return types.URLDocument{}, types.ErrDocumentNotFound
		}
		return *cachedURL, nil
	}
	doc, err := u.repo.GetDocument(ctx, urlKey)
	if err != nil {
		u.l.Error("failed to get tiny url", zap.Error(err), zap.String("db-key", urlKey))
		return types.URLDocument{}, err
	}
	if cacheAgain {
		err = u.cacheTinyURL(ctx, doc)
		if err != nil {
			u.l.Error("failed to cache tiny url", zap.Error(err), zap.String("cache-key", urlKey))
		}
	}
	return doc, nil
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
	u.counter.DeleteLabelValues(urlKey)
	return nil
}

func (u *urlSVC) cacheTinyURL(ctx context.Context, tinyURL types.URLDocument) error {
	keyBytes, err := json.Marshal(tinyURL)
	if err != nil {
		return err
	}
	// cache generatedKey -> URLDocument
	return u.cache.Cache(ctx, tinyURL.URLKey, keyBytes)
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

func formTinyURL(longURL string, liveForever bool) types.URLDocument {
	cTime := time.Now()
	id := rand.Int64N(cTime.Unix())
	base58String := encode(id)
	urlObj := types.URLDocument{
		Base10ID: id,
		LongURL:  longURL,
		URLKey:   base58String,
	}
	if liveForever {
		urlObj.LiveForever = true
		urlObj.ExpireTime = cTime.Add(time.Hour * 24 * 365 * 250) // arbitrary 250 years
	} else {
		urlObj.ExpireTime = cTime.Add(defaultExpiryTime)
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
