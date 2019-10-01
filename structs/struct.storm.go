package structs

type (
	UsersLocationBookmark struct {
		ID           int    `storm:"id,increment"` // primary key
		LocationID   string `storm:"index"`        // this field will be indexed
		UserID       int    `storm:"index"`
		MaxWindSpeed int
		LowestTemp   int
		IsReady      bool `storm:"index"`
	}

	UserState struct {
		ID           int `storm:"id,increment"`
		UserID       int `storm:"unique"` // one user can have only one state
		CurrentState int
	}
)
