package test

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	url "net/url"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type DexOauth struct {
	DexURL       string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Username     string
	Password     string
}

type DexAccessToken struct {
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
}

func (d *DexOauth) GetAccessToken() (*DexAccessToken, error) {
	ldapURL, err := d.getLdapURL()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ldap url")
	}
	log.Tracef("Got ldap url: %s\n", ldapURL)

	approvalURL, err := d.getApprovalURL(ldapURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ldap approval url")
	}
	log.Tracef("Got approval url: %s\n", approvalURL)

	redirectURL, err := d.approve(approvalURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get redirect URL")
	}
	log.Tracef("Got redirect url: %s\n", redirectURL)

	uri, err := url.Parse(redirectURL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse redirect URL: %s", redirectURL)
	}
	params, err := url.ParseQuery(uri.RawQuery)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse redirect URL query: %s", redirectURL)
	}
	code := params.Get("code")
	if code == "" {
		return nil, errors.Errorf("could not find code in redirect URL: %s", redirectURL)
	}
	log.Tracef("Code is: %s\n", code)

	tokenJWT, err := d.getToken(code)
	if err != nil {
		return nil, errors.Wrap(err, "could not get token")
	}

	token := &DexAccessToken{}

	if err := json.Unmarshal([]byte(tokenJWT), token); err != nil {
		return nil, errors.Wrap(err, "failed to decode jwt")
	}

	return token, nil
}

func (d *DexOauth) getLdapURL() (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", d.DexURL, "auth"), nil)
	if err != nil {
		return "", errors.Wrap(err, "failed to create http request")
	}

	nonce, err := randomHex(16)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate nonce")
	}
	state, err := randomHex(16)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate state")
	}

	q := req.URL.Query()
	q.Add("access_type", "offline")
	q.Add("client_id", d.ClientID)
	q.Add("nonce", nonce)
	q.Add("redirect_uri", d.RedirectURI)
	q.Add("response_type", "code")
	q.Add("state", state)
	req.URL.RawQuery = q.Encode() + "&scope=offline_access+openid+profile+email+groups" // we want this scope to not be escaped

	client := d.newHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrapf(err, "failed to post url: %s", req.URL.String())
	}
	defer resp.Body.Close()

	return getRedirect(resp, 302)
}

func (d *DexOauth) getApprovalURL(ldapURL string) (string, error) {
	form := url.Values{}
	form.Set("login", d.Username)
	form.Set("password", d.Password)

	client := d.newHttpClient()
	url := fmt.Sprintf("%s%s", d.DexURL, ldapURL)
	resp, err := client.PostForm(url, form)
	if err != nil {
		return "", errors.Wrapf(err, "failed to post url: %s", url)
	}
	defer resp.Body.Close()
	return getRedirect(resp, 303)
}

func (d *DexOauth) approve(approvalURL string) (string, error) {
	form := url.Values{}
	form.Set("approval", "approve")

	url := fmt.Sprintf("%s%s", d.DexURL, approvalURL)

	client := d.newHttpClient()
	resp, err := client.PostForm(url, form)
	if err != nil {
		return "", errors.Wrapf(err, "failed to post url: %s", url)
	}
	defer resp.Body.Close()

	return getRedirect(resp, 303)
}

func getRedirect(resp *http.Response, expectedCode int) (string, error) {
	if resp.StatusCode != expectedCode {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		return "", errors.Errorf("expected redirect code, got %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	location := resp.Header.Get("location")
	if location != "" {
		return location, nil
	}
	return "", errors.New("could not find redirect url")
}

func (d *DexOauth) getToken(code string) (string, error) {
	codeVerifier, err := randomHex(16)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate code verifier")
	}
	form := url.Values{}
	form.Set("code", code)
	form.Set("code_verifier", codeVerifier)
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", d.RedirectURI)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", d.DexURL, "/token"), strings.NewReader(form.Encode()))
	if err != nil {
		return "", errors.Wrap(err, "failed to create http request")
	}
	req.SetBasicAuth(d.ClientID, d.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to post token")
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		return "", errors.Errorf("expected redirect code, got %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(bodyBytes), nil
}

func (d *DexOauth) newHttpClient() *http.Client {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	return client
}

func randomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
