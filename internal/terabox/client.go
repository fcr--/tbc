package terabox

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"tbc/internal/util"

	"resty.dev/v3"
)

const (
	userAgent = "okhttp/7.4"
	baseUrl   = "https://www.terabox.com"
	appId     = "250528"
)

type Client struct {
	client   *resty.Client
	cookies  []*http.Cookie
	jsToken  string
	bdsToken string
	cwd      string
}

func NewClient(cookieStr string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating cookie jar: %w", err)
	}

	restyClient := resty.New()
	restyClient.SetCookieJar(jar)
	restyClient.SetHeader("User-Agent", userAgent)
	restyClient.SetHeader("Referer", baseUrl+"/")
	restyClient.SetHeader("Origin", baseUrl)
	restyClient.SetBaseURL(baseUrl)

	cookies := strings.Split(cookieStr, "; ")
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, fmt.Errorf("Error parsing base URL: %w", err)
	}

	var cookieSlice []*http.Cookie
	for _, cookie := range cookies {
		parts := strings.SplitN(cookie, "=", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			httpCookie := &http.Cookie{
				Name:     name,
				Value:    value,
				Domain:   "terabox.com",
				Secure:   u.Scheme == "https",
				HttpOnly: true,
			}
			restyClient.SetCookie(httpCookie)
			cookieSlice = append(cookieSlice, httpCookie)
		}
	}

	c := &Client{
		client:  restyClient,
		cookies: cookieSlice,
		cwd:     "/",
	}

	c.jsToken, c.bdsToken, err = c.getToken()
	if err != nil {
		return nil, fmt.Errorf("Cannot obtain tokens: %w", err)
	}

	return c, nil
}

func (c *Client) getToken() (jsToken, bdsToken string, err error) {
	body, err := util.GetResponse(c.client.R().Get("/main"))
	if err != nil {
		return "", "", fmt.Errorf("error getting initial response: %w", err)
	}

	bodyStr := string(body)

	reBdsToken := regexp.MustCompile(`"bdstoken":"([^"]+)"`)
	matchBdsToken := reBdsToken.FindStringSubmatch(bodyStr)
	if len(matchBdsToken) > 1 {
		bdsToken = matchBdsToken[1]
	} else {
		err = fmt.Errorf("bdstoken not found in response")
		return
	}

	reJsToken := regexp.MustCompile(`window\.jsToken%20%3D%20a%7D%3Bfn%28%22([^%]+)`)
	matchJsToken := reJsToken.FindSubmatch(body)
	if len(matchJsToken) > 1 {
		jsToken = string(matchJsToken[1])
	} else {
		err = fmt.Errorf("jsToken not found in response")
		return
	}

	return jsToken, bdsToken, nil
}
