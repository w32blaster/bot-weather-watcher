package structs

type (
	UsersLocationBookmark struct {
		ID           int    // primary key
		LocationID   string `storm:"index"` // this field will be indexed
		UserID       string
		MaxWindSpeed int
		LowestTemp   int
	}

	UserState struct {
		ID           int
		UserID       int `storm:"unique"` // one user can have only one state
		CurrentState int
	}
)
