package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client calls the Barong Admin API using session cookie authentication.
type Client struct {
	baseURL    string
	cookies    []*http.Cookie
	httpClient *http.Client
}

func NewAuthenticatedClient(baseURL string, cookies []*http.Cookie) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		cookies:    cookies,
		httpClient: &http.Client{},
	}
}

// --- User operations ---

func (c *Client) ListUsers(params url.Values) ([]User, error) {
	body, status, err := c.get("/api/v1/auth/admin/users", params)
	return decodeList[User](body, status, err)
}

func (c *Client) GetUser(uid string) (*UserWithKYC, error) {
	body, status, err := c.get("/api/v1/auth/admin/users/"+uid, nil)
	return decodeOne[UserWithKYC](body, status, err)
}

func (c *Client) UpdateUser(uid, email, state string, otp *bool) error {
	form := url.Values{"uid": {uid}}
	setFormIfNotEmpty(form, "email", email)
	setFormIfNotEmpty(form, "state", state)
	if otp != nil {
		if *otp {
			form.Set("otp", "true")
		} else {
			form.Set("otp", "false")
		}
	}
	body, status, err := c.put("/api/v1/auth/admin/users", form)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("API error %d: %s", status, body)
	}
	return nil
}

func (c *Client) UpdateUserRole(uid, role string) error {
	form := url.Values{"uid": {uid}, "role": {role}}
	body, status, err := c.post("/api/v1/auth/admin/users/role", form)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("API error %d: %s", status, body)
	}
	return nil
}

func (c *Client) UpdateUserAttrs(uid, state string, otp *bool) error {
	form := url.Values{"uid": {uid}}
	setFormIfNotEmpty(form, "state", state)
	if otp != nil {
		if *otp {
			form.Set("otp", "true")
		} else {
			form.Set("otp", "false")
		}
	}
	body, status, err := c.post("/api/v1/auth/admin/users/update", form)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("API error %d: %s", status, body)
	}
	return nil
}

func (c *Client) ListPendingDocUsers(params url.Values) ([]User, error) {
	body, status, err := c.get("/api/v1/auth/admin/users/documents/pending", params)
	return decodeList[User](body, status, err)
}

func (c *Client) DeleteDataStorage(uid, title string) (*UserWithKYC, error) {
	params := url.Values{"uid": {uid}, "title": {title}}
	body, status, err := c.deleteReq("/api/v1/auth/admin/users/data_storage", params)
	return decodeOne[UserWithKYC](body, status, err)
}

// --- Label operations ---

func (c *Client) ListLabelKeys() (RawJSON, error) {
	body, status, err := c.get("/api/v1/auth/admin/users/labels/list", nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("API error %d: %s", status, body)
	}
	return body, nil
}

func (c *Client) FilterUsersByLabel(key, value string, page, limit int64) ([]User, error) {
	params := url.Values{"key": {key}, "value": {value}}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	body, status, err := c.get("/api/v1/auth/admin/users/labels", params)
	return decodeList[User](body, status, err)
}

func (c *Client) AddLabel(uid, key, value, description, scope string) error {
	form := url.Values{"uid": {uid}, "key": {key}, "value": {value}}
	setFormIfNotEmpty(form, "description", description)
	setFormIfNotEmpty(form, "scope", scope)
	body, status, err := c.post("/api/v1/auth/admin/users/labels", form)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("API error %d: %s", status, body)
	}
	return nil
}

func (c *Client) UpdateLabel(uid, key, scope, value, description string) error {
	form := url.Values{"uid": {uid}, "key": {key}, "scope": {scope}, "value": {value}}
	setFormIfNotEmpty(form, "description", description)
	body, status, err := c.put("/api/v1/auth/admin/users/labels", form)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("API error %d: %s", status, body)
	}
	return nil
}

func (c *Client) UpdateLabelValue(uid, key, scope, value, description string, replace bool) error {
	form := url.Values{"uid": {uid}, "key": {key}, "scope": {scope}, "value": {value}}
	setFormIfNotEmpty(form, "description", description)
	if replace {
		form.Set("replace", "true")
	}
	body, status, err := c.post("/api/v1/auth/admin/users/labels/update", form)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("API error %d: %s", status, body)
	}
	return nil
}

func (c *Client) DeleteLabel(uid, key, scope string) error {
	params := url.Values{"uid": {uid}, "key": {key}, "scope": {scope}}
	body, status, err := c.deleteReq("/api/v1/auth/admin/users/labels", params)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("API error %d: %s", status, body)
	}
	return nil
}

// --- Comment operations ---

func (c *Client) AddComment(uid, title, data string) (*UserWithKYC, error) {
	form := url.Values{"uid": {uid}, "title": {title}, "data": {data}}
	body, status, err := c.post("/api/v1/auth/admin/users/comments", form)
	return decodeOne[UserWithKYC](body, status, err)
}

func (c *Client) UpdateComment(id int, title, data string) (*UserWithKYC, error) {
	form := url.Values{"id": {fmt.Sprintf("%d", id)}}
	setFormIfNotEmpty(form, "title", title)
	setFormIfNotEmpty(form, "data", data)
	body, status, err := c.put("/api/v1/auth/admin/users/comments", form)
	return decodeOne[UserWithKYC](body, status, err)
}

func (c *Client) DeleteComment(id int) (*UserWithKYC, error) {
	params := url.Values{"id": {fmt.Sprintf("%d", id)}}
	body, status, err := c.deleteReq("/api/v1/auth/admin/users/comments", params)
	return decodeOne[UserWithKYC](body, status, err)
}

// --- API Key operations ---

func (c *Client) ListAPIKeys(uid, ordering, orderBy string, page, limit int64) ([]APIKey, error) {
	params := url.Values{"uid": {uid}}
	setParamIfNotEmpty(params, "ordering", ordering)
	setParamIfNotEmpty(params, "order_by", orderBy)
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	body, status, err := c.get("/api/v1/auth/admin/api_keys", params)
	return decodeList[APIKey](body, status, err)
}

// --- Permission operations ---

func (c *Client) ListPermissions(page, limit int64) ([]Permission, error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	body, status, err := c.get("/api/v1/auth/admin/permissions", params)
	return decodeList[Permission](body, status, err)
}

func (c *Client) CreatePermission(role, verb, path, action, topic string) error {
	form := url.Values{"role": {role}, "verb": {verb}, "path": {path}, "action": {action}}
	setFormIfNotEmpty(form, "topic", topic)
	body, status, err := c.post("/api/v1/auth/admin/permissions", form)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("API error %d: %s", status, body)
	}
	return nil
}

func (c *Client) UpdatePermission(id int, role, verb, path, action, topic string) error {
	form := url.Values{"id": {fmt.Sprintf("%d", id)}}
	setFormIfNotEmpty(form, "role", role)
	setFormIfNotEmpty(form, "verb", verb)
	setFormIfNotEmpty(form, "path", path)
	setFormIfNotEmpty(form, "action", action)
	setFormIfNotEmpty(form, "topic", topic)
	body, status, err := c.put("/api/v1/auth/admin/permissions", form)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("API error %d: %s", status, body)
	}
	return nil
}

func (c *Client) DeletePermission(id int) error {
	params := url.Values{"id": {fmt.Sprintf("%d", id)}}
	body, status, err := c.deleteReq("/api/v1/auth/admin/permissions", params)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("API error %d: %s", status, body)
	}
	return nil
}

// --- Activity operations ---

func (c *Client) ListActivities(params url.Values) ([]Activity, error) {
	body, status, err := c.get("/api/v1/auth/admin/activities", params)
	return decodeList[Activity](body, status, err)
}

func (c *Client) ListAdminActivities(params url.Values) ([]AdminActivity, error) {
	body, status, err := c.get("/api/v1/auth/admin/activities/admin", params)
	return decodeList[AdminActivity](body, status, err)
}

// --- Metrics ---

func (c *Client) GetMetrics(createdFrom, createdTo string) (RawJSON, error) {
	params := url.Values{}
	setParamIfNotEmpty(params, "created_from", createdFrom)
	setParamIfNotEmpty(params, "created_to", createdTo)
	body, status, err := c.get("/api/v1/auth/admin/metrics", params)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("API error %d: %s", status, body)
	}
	return body, nil
}

// --- Restriction operations ---

func (c *Client) ListRestrictions(scope, category, rangeStr string, page, limit int64) ([]Restriction, error) {
	params := url.Values{}
	setParamIfNotEmpty(params, "scope", scope)
	setParamIfNotEmpty(params, "category", category)
	setParamIfNotEmpty(params, "range", rangeStr)
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	body, status, err := c.get("/api/v1/auth/admin/restrictions", params)
	return decodeList[Restriction](body, status, err)
}

func (c *Client) CreateRestriction(scope, value, category, state string, code int) error {
	form := url.Values{"scope": {scope}, "value": {value}, "category": {category}}
	setFormIfNotEmpty(form, "state", state)
	if code != 0 {
		form.Set("code", fmt.Sprintf("%d", code))
	}
	body, status, err := c.post("/api/v1/auth/admin/restrictions", form)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("API error %d: %s", status, body)
	}
	return nil
}

func (c *Client) UpdateRestriction(id int, scope, category, value, state string, code int) error {
	form := url.Values{"id": {fmt.Sprintf("%d", id)}}
	setFormIfNotEmpty(form, "scope", scope)
	setFormIfNotEmpty(form, "category", category)
	setFormIfNotEmpty(form, "value", value)
	setFormIfNotEmpty(form, "state", state)
	if code != 0 {
		form.Set("code", fmt.Sprintf("%d", code))
	}
	body, status, err := c.put("/api/v1/auth/admin/restrictions", form)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("API error %d: %s", status, body)
	}
	return nil
}

func (c *Client) DeleteRestriction(id int) error {
	params := url.Values{"id": {fmt.Sprintf("%d", id)}}
	body, status, err := c.deleteReq("/api/v1/auth/admin/restrictions", params)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("API error %d: %s", status, body)
	}
	return nil
}

func (c *Client) CreateWhitelink(expireTime int, rangeStr string) (RawJSON, error) {
	form := url.Values{}
	if expireTime > 0 {
		form.Set("expire_time", fmt.Sprintf("%d", expireTime))
	}
	setFormIfNotEmpty(form, "range", rangeStr)
	body, status, err := c.post("/api/v1/auth/admin/restrictions/whitelink", form)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("API error %d: %s", status, body)
	}
	return body, nil
}

// --- Profile operations ---

func (c *Client) ListProfiles(page, limit int64) ([]Profile, error) {
	params := url.Values{}
	if page > 0 {
		params.Set("page", fmt.Sprintf("%d", page))
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	body, status, err := c.get("/api/v1/auth/admin/profiles", params)
	return decodeList[Profile](body, status, err)
}

func (c *Client) CreateProfile(uid, firstName, lastName, dob, address, postcode, city, country, metadata string) (*Profile, error) {
	form := url.Values{"uid": {uid}}
	setFormIfNotEmpty(form, "first_name", firstName)
	setFormIfNotEmpty(form, "last_name", lastName)
	setFormIfNotEmpty(form, "dob", dob)
	setFormIfNotEmpty(form, "address", address)
	setFormIfNotEmpty(form, "postcode", postcode)
	setFormIfNotEmpty(form, "city", city)
	setFormIfNotEmpty(form, "country", country)
	setFormIfNotEmpty(form, "metadata", metadata)
	body, status, err := c.post("/api/v1/auth/admin/profiles", form)
	return decodeOne[Profile](body, status, err)
}

func (c *Client) VerifyProfile(uid, state string) (*Profile, error) {
	form := url.Values{"uid": {uid}, "state": {state}}
	body, status, err := c.put("/api/v1/auth/admin/profiles", form)
	return decodeOne[Profile](body, status, err)
}

// --- Levels ---

func (c *Client) ListLevels() ([]Level, error) {
	body, status, err := c.get("/api/v1/auth/admin/levels", nil)
	return decodeList[Level](body, status, err)
}

// --- Abilities ---

func (c *Client) GetAbilities() (RawJSON, error) {
	body, status, err := c.get("/api/v1/auth/admin/abilities", nil)
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("API error %d: %s", status, body)
	}
	return body, nil
}

// --- HTTP helpers ---

func (c *Client) get(path string, params url.Values) ([]byte, int, error) {
	fullURL := c.baseURL + path
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, 0, err
	}
	c.setAuth(req)
	return c.do(req)
}

func (c *Client) post(path string, form url.Values) ([]byte, int, error) {
	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.setAuth(req)
	return c.do(req)
}

func (c *Client) put(path string, form url.Values) ([]byte, int, error) {
	req, err := http.NewRequest(http.MethodPut, c.baseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.setAuth(req)
	return c.do(req)
}

func (c *Client) deleteReq(path string, params url.Values) ([]byte, int, error) {
	fullURL := c.baseURL + path
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}
	req, err := http.NewRequest(http.MethodDelete, fullURL, nil)
	if err != nil {
		return nil, 0, err
	}
	c.setAuth(req)
	return c.do(req)
}

func (c *Client) setAuth(req *http.Request) {
	for _, cookie := range c.cookies {
		req.AddCookie(cookie)
	}
}

func (c *Client) do(req *http.Request) ([]byte, int, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

// --- Decode helpers ---

func decodeOne[T any](body []byte, status int, err error) (*T, error) {
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("API error %d: %s", status, body)
	}
	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &result, nil
}

func decodeList[T any](body []byte, status int, err error) ([]T, error) {
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("API error %d: %s", status, body)
	}
	var result []T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return result, nil
}

func setFormIfNotEmpty(form url.Values, key, value string) {
	if value != "" {
		form.Set(key, value)
	}
}

func setParamIfNotEmpty(params url.Values, key, value string) {
	if value != "" {
		params.Set(key, value)
	}
}
