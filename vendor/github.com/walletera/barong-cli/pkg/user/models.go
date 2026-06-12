package user

type UserWithFullInfo struct {
    Email       string        `json:"email"`
    UID         string        `json:"uid"`
    Role        string        `json:"role"`
    Level       int           `json:"level"`
    OTP         bool          `json:"otp"`
    State       string        `json:"state"`
    ReferralUID string        `json:"referral_uid"`
    Data        string        `json:"data"`
    CSRFToken   string        `json:"csrf_token"`
    Labels      []Label       `json:"labels"`
    Phones      []Phone       `json:"phones"`
    Profiles    []Profile     `json:"profiles"`
    DataStorage []DataStorage `json:"data_storages"`
    CreatedAt   string        `json:"created_at"`
    UpdatedAt   string        `json:"updated_at"`
}

type Label struct {
    Key       string `json:"key"`
    Value     string `json:"value"`
    Scope     string `json:"scope"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}

type Phone struct {
    Country     string `json:"country"`
    Number      string `json:"number"`
    ValidatedAt string `json:"validated_at"`
}

type Profile struct {
    FirstName string      `json:"first_name"`
    LastName  string      `json:"last_name"`
    DOB       string      `json:"dob"`
    Address   string      `json:"address"`
    Postcode  string      `json:"postcode"`
    City      string      `json:"city"`
    Country   string      `json:"country"`
    State     string      `json:"state"`
    Metadata  interface{} `json:"metadata"`
    CreatedAt string      `json:"created_at"`
    UpdatedAt string      `json:"updated_at"`
}

type DataStorage struct {
    Title     string `json:"title"`
    Data      string `json:"data"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}

type OTPQRCode struct {
    Barcode string `json:"barcode"`
    URL     string `json:"url"`
}

type ServiceAccount struct {
    Email     string `json:"email"`
    UID       string `json:"uid"`
    Role      string `json:"role"`
    Level     int    `json:"level"`
    State     string `json:"state"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}

type APIKey struct {
    Kid       string   `json:"kid"`
    Algorithm string   `json:"algorithm"`
    Scope     []string `json:"scope"`
    State     string   `json:"state"`
    Secret    string   `json:"secret"`
    CreatedAt string   `json:"created_at"`
    UpdatedAt string   `json:"updated_at"`
}
