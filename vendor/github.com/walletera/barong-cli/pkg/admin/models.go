package admin

import "encoding/json"

type User struct {
	Email       string `json:"email"`
	UID         string `json:"uid"`
	Role        string `json:"role"`
	Level       int    `json:"level"`
	OTP         bool   `json:"otp"`
	State       string `json:"state"`
	ReferralUID string `json:"referral_uid"`
	Data        string `json:"data"`
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

type Document struct {
	Upload    string `json:"upload"`
	DocType   string `json:"doc_type"`
	DocNumber string `json:"doc_number"`
	DocExpire string `json:"doc_expire"`
	Metadata  string `json:"metadata"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type DataStorage struct {
	Title     string `json:"title"`
	Data      string `json:"data"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Comment struct {
	ID        int    `json:"id"`
	AuthorUID string `json:"author_uid"`
	Title     string `json:"title"`
	Data      string `json:"data"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type AdminLabelView struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Scope       string `json:"scope"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Phone struct {
	Country     string `json:"country"`
	Number      string `json:"number"`
	ValidatedAt string `json:"validated_at"`
}

type UserWithKYC struct {
	Email        string           `json:"email"`
	UID          string           `json:"uid"`
	Role         string           `json:"role"`
	Level        int              `json:"level"`
	OTP          bool             `json:"otp"`
	State        string           `json:"state"`
	ReferralUID  string           `json:"referral_uid"`
	Data         string           `json:"data"`
	Profiles     []Profile        `json:"profiles"`
	Labels       []AdminLabelView `json:"labels"`
	Phones       []Phone          `json:"phones"`
	Documents    []Document       `json:"documents"`
	DataStorages []DataStorage    `json:"data_storages"`
	Comments     []Comment        `json:"comments"`
	CreatedAt    string           `json:"created_at"`
	UpdatedAt    string           `json:"updated_at"`
}

type APIKey struct {
	Kid       string `json:"kid"`
	Algorithm string `json:"algorithm"`
	Scope     string `json:"scope"`
	State     string `json:"state"`
	Secret    string `json:"secret"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Permission struct {
	ID        int    `json:"id"`
	Action    string `json:"action"`
	Role      string `json:"role"`
	Verb      string `json:"verb"`
	Path      string `json:"path"`
	Topic     string `json:"topic"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Activity struct {
	UserIP    string `json:"user_ip"`
	UserAgent string `json:"user_agent"`
	Topic     string `json:"topic"`
	Action    string `json:"action"`
	Result    string `json:"result"`
	Data      string `json:"data"`
	User      *User  `json:"user"`
	CreatedAt string `json:"created_at"`
}

type AdminActivity struct {
	UserIP    string `json:"user_ip"`
	UserAgent string `json:"user_agent"`
	Topic     string `json:"topic"`
	Action    string `json:"action"`
	Result    string `json:"result"`
	Data      string `json:"data"`
	Admin     *User  `json:"admin"`
	Target    *User  `json:"target"`
	CreatedAt string `json:"created_at"`
}

type Restriction struct {
	ID        int    `json:"id"`
	Category  string `json:"category"`
	Scope     string `json:"scope"`
	Value     string `json:"value"`
	Code      int    `json:"code"`
	State     string `json:"state"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type Level struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RawJSON is used for endpoints that return an unstructured JSON response.
type RawJSON = json.RawMessage
