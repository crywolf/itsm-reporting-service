package validators

// PayloadValidator provides input payload validation functionality
type PayloadValidator interface {
	Validate(payload interface{}) error
}
