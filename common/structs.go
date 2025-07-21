package common

type NewUserInfo struct {
	Login       string `json:"login"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	SecondName  string `json:"second_name"`
	DateOfBirth string `json:"date_of_birth"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

type UserInfo struct {
	FirstName   string `json:"first_name"`
	SecondName  string `json:"second_name"`
	DateOfBirth string `json:"date_of_birth"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

type AuthInfo struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type PostInfo struct {
	Headline string `json:"headline"`
	Content  string `json:"content"`
	Location string `json:"location"`
	Value    uint64 `json:"value"`
}

type PostId struct {
	Id uint64 `json:"postId"`
}

type Post struct {
	PostId           uint64 `json:"PostId"`
	Author           string `json:"author"`
	DateOfCreation   string `json:"dateOfCreation"`
	Content          string `json:"content"`
	CommentSectionId uint64 `json:"commentSectionId"`
}

type PaginationInfo struct {
	PageNumber uint64 `json:"pageNumber"`
	BatchSize  uint32 `json:"batchSize"`
	SortBy     string `json:"sort_by"`       // "Date", "Value"
	Order      string `json:"sorting_order"` // "Ascending", "Descending" or represents price range "10-100" (if SortBy == Value)
}

type PostIsLike struct {
	IsLike uint64 `json:"isLike"`
}

type ReactionInfo struct {
	PostId uint64 `json:"postId"`
}

type Reaction struct {
	User   string `json:"user"`
	Author string `json:"author"`
	PostId uint64 `json:"postId"`
}
