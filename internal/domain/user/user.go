package user

type User struct {
	ChannelID string
	UserID    string
	Email     string
	Name      string
}

type List []User
