package user

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
)

type Client struct {
    baseURL    string
    httpClient *http.Client
    cookies    []*http.Cookie // non-nil only for authenticated clients
}

func NewClient(baseURL string) *Client {
    return &Client{
        baseURL:    strings.TrimRight(baseURL, "/"),
        httpClient: &http.Client{},
    }
}

func NewAuthenticatedClient(baseURL string, cookies []*http.Cookie) *Client {
    return &Client{
        baseURL:    strings.TrimRight(baseURL, "/"),
        httpClient: &http.Client{},
        cookies:    cookies,
    }
}

func (c *Client) CreateUser(email, password, username, refid string) (*UserWithFullInfo, error) {
    form := url.Values{
        "email":    {email},
        "password": {password},
    }
    if username != "" {
        form.Set("username", username)
    }
    if refid != "" {
        form.Set("refid", refid)
    }
    resp, err := c.post("/api/v1/auth/identity/users", form, false)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    return decodeUserWithFullInfo(resp)
}

func (c *Client) Login(email, password, otpCode string) (*UserWithFullInfo, []*http.Cookie, error) {
    form := url.Values{
        "email":    {email},
        "password": {password},
    }
    if otpCode != "" {
        form.Set("otp_code", otpCode)
    }
    req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1/auth/identity/sessions", strings.NewReader(form.Encode()))
    if err != nil {
        return nil, nil, err
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, nil, err
    }
    defer resp.Body.Close()
    u, err := decodeUserWithFullInfo(resp)
    if err != nil {
        return nil, nil, err
    }
    return u, resp.Cookies(), nil
}

func (c *Client) Logout(cookies []*http.Cookie) error {
    req, err := http.NewRequest(http.MethodDelete, c.baseURL+"/api/v1/auth/identity/sessions", nil)
    if err != nil {
        return err
    }
    for _, cookie := range cookies {
        req.AddCookie(cookie)
    }
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("logout failed (%d): %s", resp.StatusCode, body)
    }
    return nil
}

func (c *Client) GetMe() (*UserWithFullInfo, error) {
    resp, err := c.get("/api/v1/auth/resource/users/me", nil, true)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    return decodeUserWithFullInfo(resp)
}

func (c *Client) GenerateOTPQRCode() (*OTPQRCode, error) {
    resp, err := c.post("/api/v1/auth/resource/otp/generate_qrcode", url.Values{}, true)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
    }
    var result struct {
        Data OTPQRCode `json:"data"`
    }
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return &result.Data, nil
}

func (c *Client) EnableOTP(code string) error {
    resp, err := c.post("/api/v1/auth/resource/otp/enable", url.Values{"code": {code}}, true)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("API error %d: %s", resp.StatusCode, body)
    }
    return nil
}

func (c *Client) DisableOTP(code string) error {
    resp, err := c.post("/api/v1/auth/resource/otp/disable", url.Values{"code": {code}}, true)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("API error %d: %s", resp.StatusCode, body)
    }
    return nil
}

func (c *Client) ListServiceAccounts() ([]ServiceAccount, error) {
    resp, err := c.get("/api/v1/auth/resource/service_accounts", nil, true)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
    }
    fmt.Printf("%s\n", string(body))
    var accounts []ServiceAccount
    if err := json.Unmarshal(body, &accounts); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return accounts, nil
}

func (c *Client) ListAPIKeys(page, limit int, orderBy, ordering, serviceAccountUID string) ([]APIKey, error) {
    params := url.Values{}
    if page > 0 {
        params.Set("page", fmt.Sprintf("%d", page))
    }
    if limit > 0 {
        params.Set("limit", fmt.Sprintf("%d", limit))
    }
    if orderBy != "" {
        params.Set("order_by", orderBy)
    }
    if ordering != "" {
        params.Set("ordering", ordering)
    }
    path := "/api/v1/auth/resource/api_keys"
    if serviceAccountUID != "" {
        params.Set("service_account_uid", serviceAccountUID)
        path = "/api/v1/auth/resource/service_accounts/api_keys"
    }
    resp, err := c.get(path, params, true)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
    }
    var keys []APIKey
    if err := json.Unmarshal(body, &keys); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return keys, nil
}

func (c *Client) CreateAPIKey(algorithm, scope, totpCode, serviceAccountUID string) (*APIKey, error) {
    form := url.Values{
        "algorithm": {algorithm},
        "totp_code": {totpCode},
    }
    if scope != "" {
        form.Set("scope", scope)
    }
    path := "/api/v1/auth/resource/api_keys"
    if serviceAccountUID != "" {
        form.Set("service_account_uid", serviceAccountUID)
        path = "/api/v1/auth/resource/service_accounts/api_keys"
    }
    resp, err := c.post(path, form, true)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
    }
    fmt.Printf("%s\n", string(body))
    var key APIKey
    if err := json.Unmarshal(body, &key); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return &key, nil
}

func (c *Client) UpdateAPIKey(kid, scope, state, totpCode, serviceAccountUID string) (*APIKey, error) {
    form := url.Values{
        "totp_code": {totpCode},
    }
    if scope != "" {
        form.Set("scope", scope)
    }
    if state != "" {
        form.Set("state", state)
    }
    var (
        resp *http.Response
        err  error
    )
    if serviceAccountUID != "" {
        form.Set("service_account_uid", serviceAccountUID)
        resp, err = c.put("/api/v1/auth/resource/service_accounts/api_keys/"+kid, form)
    } else {
        resp, err = c.patch("/api/v1/auth/resource/api_keys/"+kid, form)
    }
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
    }
    var key APIKey
    if err := json.Unmarshal(body, &key); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return &key, nil
}

func (c *Client) DeleteAPIKey(kid, totpCode, serviceAccountUID string) error {
    params := url.Values{"totp_code": {totpCode}}
    path := "/api/v1/auth/resource/api_keys/" + kid
    if serviceAccountUID != "" {
        params.Set("service_account_uid", serviceAccountUID)
        path = "/api/v1/auth/resource/service_accounts/api_keys/" + kid
    }
    resp, err := c.deleteReq(path, params)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusNoContent {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("API error %d: %s", resp.StatusCode, body)
    }
    return nil
}

// --- helpers ---

func (c *Client) post(path string, form url.Values, authenticated bool) (*http.Response, error) {
    req, err := http.NewRequest(http.MethodPost, c.baseURL+path, strings.NewReader(form.Encode()))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    if authenticated {
        for _, cookie := range c.cookies {
            req.AddCookie(cookie)
        }
    }
    return c.httpClient.Do(req)
}

func (c *Client) put(path string, form url.Values) (*http.Response, error) {
    req, err := http.NewRequest(http.MethodPut, c.baseURL+path, strings.NewReader(form.Encode()))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    for _, cookie := range c.cookies {
        req.AddCookie(cookie)
    }
    return c.httpClient.Do(req)
}

func (c *Client) patch(path string, form url.Values) (*http.Response, error) {
    req, err := http.NewRequest(http.MethodPatch, c.baseURL+path, strings.NewReader(form.Encode()))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    for _, cookie := range c.cookies {
        req.AddCookie(cookie)
    }
    return c.httpClient.Do(req)
}

func (c *Client) deleteReq(path string, queryParams url.Values) (*http.Response, error) {
    fullURL := c.baseURL + path
    if len(queryParams) > 0 {
        fullURL += "?" + queryParams.Encode()
    }
    req, err := http.NewRequest(http.MethodDelete, fullURL, nil)
    if err != nil {
        return nil, err
    }
    for _, cookie := range c.cookies {
        req.AddCookie(cookie)
    }
    return c.httpClient.Do(req)
}

func (c *Client) get(path string, queryParams url.Values, authenticated bool) (*http.Response, error) {
    fullURL := c.baseURL + path
    if len(queryParams) > 0 {
        fullURL += "?" + queryParams.Encode()
    }
    req, err := http.NewRequest(http.MethodGet, fullURL, nil)
    if err != nil {
        return nil, err
    }
    if authenticated {
        for _, cookie := range c.cookies {
            req.AddCookie(cookie)
        }
    }
    return c.httpClient.Do(req)
}

func decodeUserWithFullInfo(resp *http.Response) (*UserWithFullInfo, error) {
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    if resp.StatusCode >= 400 {
        return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, body)
    }
    var u UserWithFullInfo
    if err := json.Unmarshal(body, &u); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return &u, nil
}
