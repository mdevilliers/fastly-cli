package fastlyext

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-fastly/fastly"
)

// Token maps to a Fastly Token
type Token struct {
	ID          string    `mapstructure:"id"`
	UserID      string    `mapstructure:"user_id"`
	Services    []string  `mapstructure:"services"`
	AccessToken string    `mapstructure:"access_token"`
	Name        string    `mapstructure:"name"`
	Scope       string    `mapstructure:"scope"`
	CreatedAt   time.Time `mapstructure:"created_at"`
	LastUsedAt  time.Time `mapstructure:"last_used_at"`
	ExpiresAt   time.Time `mapstructure:"expires_at"`
	IP          string    `mapstructure:"ip"`
	UserAgent   string    `mapstructure:"user_agent"`
}

// GetTokensInput exists to follow the same pattern as the go-fastly library
type GetTokensInput struct{}

// GetTokens returns all of a users tokens or an error.
// A fastly.Client is passed in to the function as a shim for inclusion in the fastly-go library.
// If merged with fastly-go removing this parameter is trivial.
func GetTokens(c *fastly.Client, i *GetTokensInput) ([]*Token, error) {

	resp, err := c.Get("/tokens", nil)
	if err != nil {
		return nil, err
	}

	var s []*Token
	if err := decodeJSON(&s, resp.Body); err != nil {
		return nil, err
	}
	return s, nil
}

// CreateTokenInput contains all of the data for a CreateToken request
type CreateTokenInput struct {
	Name       string    `url:"name"`
	Username   string    `url:"username"`
	Password   string    `url:"password"`
	Scope      string    `url:"scope"`
	TwoFAToken string    `url:"-"` // NOTE: don't serialise
	Services   []string  `url:"services"`
	ExpiresAt  time.Time `url:"-"` // NOTE : not implemented
}

// CreateToken returns the new Token or an error
func CreateToken(i *CreateTokenInput) (*Token, error) {

	// create a client with an empty API Key as the POST /tokens endpoint
	// doesn't require an API Key. If merged with fastly-go this weirdness
	// can be delt with as an implementation detail
	client, err := fastly.NewClientForEndpoint("", fastly.DefaultEndpoint)

	if err != nil {
		return nil, errors.Wrap(err, "error creating client with no apiKey")
	}

	ro := &fastly.RequestOptions{}

	if i.TwoFAToken != "" {
		ro.Headers = map[string]string{"Fastly-OTP": i.TwoFAToken}
	}

	resp, err := RequestForm(client, "POST", "/tokens", i, ro)
	if err != nil {
		return nil, err
	}

	var s *Token
	if err := decodeJSON(&s, resp.Body); err != nil {
		return nil, err
	}
	return s, nil

}

// RequestForm makes an HTTP request with the given interface being encoded as
// form data
// NOTE : this has been patched from the go-fastly to URL encode the tags correctly (or as Fastly are expecting them....)
func RequestForm(c *fastly.Client, verb, p string, i interface{}, ro *fastly.RequestOptions) (*http.Response, error) {
	if ro == nil {
		ro = new(fastly.RequestOptions)
	}

	if ro.Headers == nil {
		ro.Headers = make(map[string]string)
	}
	ro.Headers["Content-Type"] = "application/x-www-form-urlencoded"

	v, err := query.Values(i)

	if err != nil {
		return nil, err
	}

	body := v.Encode()
	ro.Body = strings.NewReader(body)
	ro.BodyLength = int64(len(body))

	return c.Request(verb, p, ro)
}

// Below are copied from the go-fastly client
// decodeJSON is used to decode an HTTP response body into an interface as JSON.
func decodeJSON(out interface{}, body io.ReadCloser) error {
	defer body.Close() // nolint

	var parsed interface{}
	dec := json.NewDecoder(body)
	if err := dec.Decode(&parsed); err != nil {
		return err
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapToHTTPHeaderHookFunc(),
			stringToTimeHookFunc(),
		),
		WeaklyTypedInput: true,
		Result:           out,
	})
	if err != nil {
		return err
	}
	return decoder.Decode(parsed)
}

// mapToHTTPHeaderHookFunc returns a function that converts maps into an
// http.Header value.
func mapToHTTPHeaderHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.Map {
			return data, nil
		}
		if t != reflect.TypeOf(new(http.Header)) {
			return data, nil
		}

		typed, ok := data.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot convert %T to http.Header", data)
		}

		n := map[string][]string{}
		for k, v := range typed {
			switch v.(type) {
			case string:
				n[k] = []string{v.(string)}
			case []string:
				n[k] = v.([]string)
			case int, int8, int16, int32, int64:
				n[k] = []string{fmt.Sprintf("%d", v.(int))}
			case float32, float64:
				n[k] = []string{fmt.Sprintf("%f", v.(float64))}
			default:
				return nil, fmt.Errorf("cannot convert %T to http.Header", v)
			}
		}

		return n, nil
	}
}

// stringToTimeHookFunc returns a function that converts strings to a time.Time
// value.
func stringToTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(time.Now()) {
			return data, nil
		}

		// Convert it by parsing
		return time.Parse(time.RFC3339, data.(string))
	}
}
