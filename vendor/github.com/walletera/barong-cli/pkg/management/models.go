package management

type User struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
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

type UserWithProfile struct {
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	UID         string    `json:"uid"`
	Role        string    `json:"role"`
	Level       int       `json:"level"`
	OTP         bool      `json:"otp"`
	State       string    `json:"state"`
	ReferralUID string    `json:"referral_uid"`
	Data        string    `json:"data"`
	Profiles    []Profile `json:"profiles"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
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

type Label struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Scope     string `json:"scope"`
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
	Username     string           `json:"username"`
	UID          string           `json:"uid"`
	Role         string           `json:"role"`
	Level        int              `json:"level"`
	OTP          bool             `json:"otp"`
	State        string           `json:"state"`
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

type ServiceAccount struct {
	Email     string `json:"email"`
	UID       string `json:"uid"`
	Role      string `json:"role"`
	Level     int    `json:"level"`
	State     string `json:"state"`
	User      *User  `json:"user"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
