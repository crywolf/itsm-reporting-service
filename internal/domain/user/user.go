package user

type User struct {
	ChannelID string
	UserID    string
	Email     string
	Name      string
	Type      string
	OrgName   string
}

type List []User
