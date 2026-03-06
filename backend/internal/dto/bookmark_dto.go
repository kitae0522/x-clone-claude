package dto

type BookmarkStatusResponse struct {
	Bookmarked bool `json:"bookmarked"`
}

type BookmarkListResponse struct {
	Posts      []PostDetailResponse `json:"posts"`
	NextCursor string               `json:"nextCursor,omitempty"`
	HasMore    bool                 `json:"hasMore"`
}
