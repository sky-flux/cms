package dashboard

// DashboardStats aggregates site-wide statistics for the admin dashboard.
type DashboardStats struct {
	Posts    PostStats    `json:"posts"`
	Users    UserStats    `json:"users"`
	Comments CommentStats `json:"comments"`
	Media    MediaStats   `json:"media"`
}

type PostStats struct {
	Total     int64 `json:"total"`
	Published int64 `json:"published"`
	Draft     int64 `json:"draft"`
	Scheduled int64 `json:"scheduled"`
}

type UserStats struct {
	Total    int64 `json:"total"`
	Active   int64 `json:"active"`
	Inactive int64 `json:"inactive"`
}

type CommentStats struct {
	Total    int64 `json:"total"`
	Pending  int64 `json:"pending"`
	Approved int64 `json:"approved"`
	Spam     int64 `json:"spam"`
}

type MediaStats struct {
	Total       int64 `json:"total"`
	StorageUsed int64 `json:"storage_used"`
}
