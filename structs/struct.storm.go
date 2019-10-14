package structs

type (
	UsersLocationBookmark struct {
		ID           int    `storm:"id,increment"` // primary key
		LocationID   string `storm:"index"`        // this field will be indexed
		UserID       int    `storm:"index"`
		ChatID       int64  // chat ID where to send notifications
		MaxWindSpeed int
		LowestTemp   int
		IsReady      bool `storm:"index"`
		CheckPeriod  int
	}

	UserState struct {
		ID           int `storm:"id,increment"`
		UserID       int `storm:"unique"` // one user can have only one state
		CurrentState int
	}
)
