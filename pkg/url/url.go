package url

import (
	"context"
	"fmt"
	"github.com/vaishakdinesh/tiny-url-svc/types"
	"go.uber.org/zap"
	"math/big"
	"math/rand/v2"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	urlFormatWithPort = "%s://%s:%s/tinyurlsvc/%s"
	urlFormat         = "%s://%s/tinyurlsvc/%s"
)

type urlSVC struct {
	l    *zap.Logger
	repo types.URLRepo
}

var base58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

func NewTinyURLService(l *zap.Logger, r types.URLRepo) types.URLService {
	return &urlSVC{
		l:    l,
		repo: r,
	}
}

func (u *urlSVC) GenerateTinyURL(ctx context.Context, longUrl string) (types.URLDocument, error) {
	tinyURL, err := formTinyURL(longUrl)
	if err != nil {
		u.l.Error("failed to form tiny url", zap.Error(err))
		return types.URLDocument{}, err
	}
	err = u.repo.Put(ctx, longUrl, tinyURL)
	if err != nil {
		u.l.Error("failed to store url", zap.Error(err))
		return types.URLDocument{}, err
	}
	u.l.Info("added url to db", zap.String(longUrl, strconv.FormatInt(tinyURL.ID, 10)))
	return tinyURL, nil
}

func formTinyURL(longURL string) (types.URLDocument, error) {
	parsedURL, err := url.Parse(longURL)
	if err != nil {
		return types.URLDocument{}, err
	}
	host := parsedURL.Host
	port := parsedURL.Port()
	if port != "" {
		host, port, err = net.SplitHostPort(parsedURL.Host)
		if err != nil {
			return types.URLDocument{}, err
		}
	}
	cTime := time.Now()
	id := rand.Int64N(cTime.Unix())
	base58String := encode(id)
	urlObj := types.URLDocument{
		ID:         id,
		URL:        longURL,
		ExpireTime: cTime.Add(time.Hour * 24 * 362),
	}
	if port != "" {
		urlObj.TinyURL = fmt.Sprintf(urlFormatWithPort, parsedURL.Scheme, host, port, base58String)
		return urlObj, nil
	}
	urlObj.TinyURL = fmt.Sprintf(urlFormat, parsedURL.Scheme, host, base58String)
	return urlObj, nil
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
