package ref

// UUID represents UUID of a resource
type UUID string

func (u UUID) String() string {
	return string(u)
}

// IsZero returns true if UUID has zero value
func (u UUID) IsZero() bool {
	return u == ""
}
