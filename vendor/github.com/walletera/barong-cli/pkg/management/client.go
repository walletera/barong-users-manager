package management

import (
    "bytes"
    "crypto"
    "crypto/rand"
    "crypto/rsa"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"
)

// Client calls the Barong Management API using JWT multisig authentication.
// Every request is a POST (or PUT for label updates) with a JSON body containing
// the signed JWT in the JWS JSON serialization format (RFC 7515).
type Client struct {
    baseURL    string
    keyID      string
    privateKey *rsa.PrivateKey
    httpClient *http.Client
}

func NewClient(baseURL, keyID string, privateKey *rsa.PrivateKey) *Client {
    return &Client{
        baseURL:    strings.TrimRight(baseURL, "/"),
        keyID:      keyID,
        privateKey: privateKey,
        httpClient: &http.Client{},
    }
}

// --- User operations ---

func (c *Client) CreateUser(email, password, referralUID string) (*UserWithProfile, error) {
    data := map[string]interface{}{"email": email, "password": password}
    if referralUID != "" {
        data["referral_uid"] = referralUID
    }
    return decodeUserWithProfile(c.postJSON("/api/v2/management/users", data))
}

func (c *Client) GetUser(uid, email, phoneNum string) (*UserWithKYC, error) {
    data := map[string]interface{}{}
    if uid != "" {
        data["uid"] = uid
    }
    if email != "" {
        data["email"] = email
    }
    if phoneNum != "" {
        data["phone_num"] = phoneNum
    }
    return decodeUserWithKYC(c.postJSON("/api/v2/management/users/get", data))
}

func (c *Client) ListUsers(extended bool, from, to, page, limit int64) ([]User, error) {
    data := map[string]interface{}{}
    if extended {
        data["extended"] = true
    }
    if from > 0 {
        data["from"] = from
    }
    if to > 0 {
        data["to"] = to
    }
    if page > 0 {
        data["page"] = page
    }
    if limit > 0 {
        data["limit"] = limit
    }
    return decodeUsers(c.postJSON("/api/v2/management/users/list", data))
}

func (c *Client) UpdateUser(uid, role, userData string) (*UserWithProfile, error) {
    data := map[string]interface{}{"uid": uid}
    if role != "" {
        data["role"] = role
    }
    if userData != "" {
        data["data"] = userData
    }
    return decodeUserWithProfile(c.postJSON("/api/v2/management/users/update", data))
}

func (c *Client) ImportUser(email, passwordDigest, referralUID, phone, firstName, lastName, dob, address, postcode, city, country, state string) (*UserWithProfile, error) {
    data := map[string]interface{}{
        "email":           email,
        "password_digest": passwordDigest,
    }
    setIfNotEmpty(data, "referral_uid", referralUID)
    setIfNotEmpty(data, "phone", phone)
    setIfNotEmpty(data, "first_name", firstName)
    setIfNotEmpty(data, "last_name", lastName)
    setIfNotEmpty(data, "dob", dob)
    setIfNotEmpty(data, "address", address)
    setIfNotEmpty(data, "postcode", postcode)
    setIfNotEmpty(data, "city", city)
    setIfNotEmpty(data, "country", country)
    setIfNotEmpty(data, "state", state)
    return decodeUserWithProfile(c.postJSON("/api/v2/management/users/import", data))
}

// --- Label operations ---

func (c *Client) CreateLabel(userUID, key, value, description string) (*Label, error) {
    data := map[string]interface{}{
        "user_uid": userUID,
        "key":      key,
        "value":    value,
    }
    setIfNotEmpty(data, "description", description)
    return decodeLabel(c.postJSON("/api/v2/management/labels", data))
}

func (c *Client) UpdateLabel(userUID, key, value, description string, replace bool) (*Label, error) {
    data := map[string]interface{}{
        "user_uid": userUID,
        "key":      key,
        "value":    value,
    }
    setIfNotEmpty(data, "description", description)
    if replace {
        data["replace"] = true
    }
    return decodeLabel(c.putJSON("/api/v2/management/labels", data))
}

func (c *Client) DeleteLabel(userUID, key string) error {
    data := map[string]interface{}{
        "user_uid": userUID,
        "key":      key,
    }
    body, status, err := c.postJSON("/api/v2/management/labels/delete", data)
    if err != nil {
        return err
    }
    if status >= 400 {
        return fmt.Errorf("API error %d: %s", status, body)
    }
    return nil
}

func (c *Client) ListLabels(userUID string) ([]AdminLabelView, error) {
    data := map[string]interface{}{"user_uid": userUID}
    body, status, err := c.postJSON("/api/v2/management/labels/list", data)
    if err != nil {
        return nil, err
    }
    if status >= 400 {
        return nil, fmt.Errorf("API error %d: %s", status, body)
    }
    var result []AdminLabelView
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return result, nil
}

func (c *Client) FilterUsersByLabel(key, value, scope string, extended bool, page, limit int64) ([]User, error) {
    data := map[string]interface{}{"key": key}
    setIfNotEmpty(data, "value", value)
    setIfNotEmpty(data, "scope", scope)
    if extended {
        data["extended"] = true
    }
    if page > 0 {
        data["page"] = page
    }
    if limit > 0 {
        data["limit"] = limit
    }
    return decodeUsers(c.postJSON("/api/v2/management/labels/filter/users", data))
}

// --- Profile operations ---

func (c *Client) ImportProfile(uid, firstName, lastName, dob, address, postcode, city, country, state, metadata string) (*UserWithProfile, error) {
    data := map[string]interface{}{"uid": uid}
    setIfNotEmpty(data, "first_name", firstName)
    setIfNotEmpty(data, "last_name", lastName)
    setIfNotEmpty(data, "dob", dob)
    setIfNotEmpty(data, "address", address)
    setIfNotEmpty(data, "postcode", postcode)
    setIfNotEmpty(data, "city", city)
    setIfNotEmpty(data, "country", country)
    setIfNotEmpty(data, "state", state)
    setIfNotEmpty(data, "metadata", metadata)
    return decodeUserWithProfile(c.postJSON("/api/v2/management/profiles", data))
}

// --- Phone operations ---

func (c *Client) CreatePhone(uid, number string) (*Phone, error) {
    data := map[string]interface{}{"uid": uid, "number": number}
    return decodePhone(c.postJSON("/api/v2/management/phones", data))
}

func (c *Client) GetPhones(uid string) ([]Phone, error) {
    data := map[string]interface{}{"uid": uid}
    body, status, err := c.postJSON("/api/v2/management/phones/get", data)
    if err != nil {
        return nil, err
    }
    if status >= 400 {
        return nil, fmt.Errorf("API error %d: %s", status, body)
    }
    var result []Phone
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return result, nil
}

func (c *Client) DeletePhone(uid, number string) (*Phone, error) {
    data := map[string]interface{}{"uid": uid, "number": number}
    return decodePhone(c.postJSON("/api/v2/management/phones/delete", data))
}

// --- Document operations ---

func (c *Client) PushDocument(uid, docType, docNumber, filename, fileExt, upload, docExpire string, updateLabels bool, metadata string) error {
    data := map[string]interface{}{
        "uid":        uid,
        "doc_type":   docType,
        "doc_number": docNumber,
        "filename":   filename,
        "file_ext":   fileExt,
        "upload":     upload,
    }
    setIfNotEmpty(data, "doc_expire", docExpire)
    if !updateLabels {
        data["update_labels"] = false
    }
    setIfNotEmpty(data, "metadata", metadata)
    body, status, err := c.postJSON("/api/v2/management/documents", data)
    if err != nil {
        return err
    }
    if status >= 400 {
        return fmt.Errorf("API error %d: %s", status, body)
    }
    return nil
}

// --- Service account operations ---

func (c *Client) CreateServiceAccount(ownerUID, role, serviceAccountUID, serviceAccountEmail string) (*ServiceAccount, error) {
    data := map[string]interface{}{
        "owner_uid":            ownerUID,
        "service_account_role": role,
    }
    setIfNotEmpty(data, "service_account_uid", serviceAccountUID)
    setIfNotEmpty(data, "service_account_email", serviceAccountEmail)
    return decodeServiceAccount(c.postJSON("/api/v2/management/service_accounts/create", data))
}

func (c *Client) GetServiceAccount(uid, email string) (*ServiceAccount, error) {
    data := map[string]interface{}{}
    setIfNotEmpty(data, "uid", uid)
    setIfNotEmpty(data, "email", email)
    return decodeServiceAccount(c.postJSON("/api/v2/management/service_accounts/get", data))
}

func (c *Client) ListServiceAccounts(ownerUID, ownerEmail string, page, limit int64) ([]ServiceAccount, error) {
    data := map[string]interface{}{}
    setIfNotEmpty(data, "owner_uid", ownerUID)
    setIfNotEmpty(data, "owner_email", ownerEmail)
    if page > 0 {
        data["page"] = page
    }
    if limit > 0 {
        data["limit"] = limit
    }
    body, status, err := c.postJSON("/api/v2/management/service_accounts/list", data)
    if err != nil {
        return nil, err
    }
    if status >= 400 {
        return nil, fmt.Errorf("API error %d: %s", status, body)
    }
    var result []ServiceAccount
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return result, nil
}

func (c *Client) DeleteServiceAccount(uid string) (*ServiceAccount, error) {
    data := map[string]interface{}{"uid": uid}
    return decodeServiceAccount(c.postJSON("/api/v2/management/service_accounts/delete", data))
}

// --- OTP operations ---

func (c *Client) SignOTP(userUID, otpCode string) error {
    data := map[string]interface{}{
        "user_uid": userUID,
        "otp_code": otpCode,
    }
    body, status, err := c.postJSON("/api/v2/management/otp/sign", data)
    if err != nil {
        return err
    }
    if status >= 400 {
        return fmt.Errorf("API error %d: %s", status, body)
    }
    return nil
}

// --- Timestamp ---

func (c *Client) GetTimestamp() (int64, error) {
    body, status, err := c.postJSON("/api/v2/management/timestamp", map[string]interface{}{})
    if err != nil {
        return 0, err
    }
    if status >= 400 {
        return 0, fmt.Errorf("API error %d: %s", status, body)
    }
    var ts int64
    if err := json.Unmarshal(body, &ts); err != nil {
        return 0, fmt.Errorf("failed to decode response: %w", err)
    }
    return ts, nil
}

// --- HTTP helpers ---

func (c *Client) postJSON(path string, data map[string]interface{}) ([]byte, int, error) {
    return c.doRequest(http.MethodPost, path, data)
}

func (c *Client) putJSON(path string, data map[string]interface{}) ([]byte, int, error) {
    return c.doRequest(http.MethodPut, path, data)
}

func (c *Client) doRequest(method, path string, data map[string]interface{}) ([]byte, int, error) {
    jwtBody, err := c.buildJWT(data)
    if err != nil {
        return nil, 0, err
    }
    req, err := http.NewRequest(method, c.baseURL+path, bytes.NewReader(jwtBody))
    if err != nil {
        return nil, 0, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, 0, err
    }
    defer resp.Body.Close()
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, resp.StatusCode, err
    }
    return respBody, resp.StatusCode, nil
}

// buildJWT constructs a JWT in JWS JSON serialization format (RFC 7515).
// The API parameters are placed in the "data" field of the JWT payload.
// The JWT is signed with RS256 using the configured private key.
func (c *Client) buildJWT(data map[string]interface{}) ([]byte, error) {
    jtiBytes := make([]byte, 16)
    if _, err := rand.Read(jtiBytes); err != nil {
        return nil, err
    }
    jti := base64.RawURLEncoding.EncodeToString(jtiBytes)

    now := time.Now()
    payload := map[string]interface{}{
        "iat":  now.Unix(),
        "exp":  now.Add(30 * time.Second).Unix(),
        "jti":  jti,
        "data": data,
    }

    payloadJSON, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }
    encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)

    headerJSON, err := json.Marshal(map[string]string{"alg": "RS256"})
    if err != nil {
        return nil, err
    }
    protected := base64.RawURLEncoding.EncodeToString(headerJSON)

    signingInput := protected + "." + encodedPayload
    hash := sha256.Sum256([]byte(signingInput))
    sig, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, crypto.SHA256, hash[:])
    if err != nil {
        return nil, err
    }

    jwt := map[string]interface{}{
        "payload": encodedPayload,
        "signatures": []map[string]interface{}{
            {
                "protected": protected,
                "header":    map[string]string{"kid": c.keyID},
                "signature": base64.RawURLEncoding.EncodeToString(sig),
            },
        },
    }
    return json.Marshal(jwt)
}

// --- Decode helpers ---

func decodeUserWithProfile(body []byte, status int, err error) (*UserWithProfile, error) {
    if err != nil {
        return nil, err
    }
    if status >= 400 {
        return nil, fmt.Errorf("API error %d: %s", status, body)
    }
    var result UserWithProfile
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return &result, nil
}

func decodeUserWithKYC(body []byte, status int, err error) (*UserWithKYC, error) {
    if err != nil {
        return nil, err
    }
    if status >= 400 {
        return nil, fmt.Errorf("API error %d: %s", status, body)
    }
    var result UserWithKYC
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return &result, nil
}

func decodeUsers(body []byte, status int, err error) ([]User, error) {
    if err != nil {
        return nil, err
    }
    if status >= 400 {
        return nil, fmt.Errorf("API error %d: %s", status, body)
    }
    var result []User
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return result, nil
}

func decodeLabel(body []byte, status int, err error) (*Label, error) {
    if err != nil {
        return nil, err
    }
    if status >= 400 {
        return nil, fmt.Errorf("API error %d: %s", status, body)
    }
    var result Label
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return &result, nil
}

func decodePhone(body []byte, status int, err error) (*Phone, error) {
    if err != nil {
        return nil, err
    }
    if status >= 400 {
        return nil, fmt.Errorf("API error %d: %s", status, body)
    }
    var result Phone
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return &result, nil
}

func decodeServiceAccount(body []byte, status int, err error) (*ServiceAccount, error) {
    if err != nil {
        return nil, err
    }
    if status >= 400 {
        return nil, fmt.Errorf("API error %d: %s", status, body)
    }
    var result ServiceAccount
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return &result, nil
}

func setIfNotEmpty(m map[string]interface{}, key, value string) {
    if value != "" {
        m[key] = value
    }
}
