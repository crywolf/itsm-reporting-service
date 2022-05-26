package validators

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/KompiTech/itsm-reporting-service/internal/domain/job"
	"github.com/KompiTech/itsm-reporting-service/internal/http/rest/presenters"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

// NewPayloadValidator returns new payload validator
func NewPayloadValidator() PayloadValidator {
	validate := validator.New()
	validate.RegisterStructValidation(validateJobType, job.Type{})

	return &payloadValidator{
		validator: validate,
	}
}

type payloadValidator struct {
	validator *validator.Validate
}

func (v payloadValidator) Validate(payload interface{}) error {
	english := en.New()
	uni := ut.New(english, english)
	trans, _ := uni.GetTranslator("en")
	_ = en_translations.RegisterDefaultTranslations(v.validator, trans)

	// use the names which have been specified for JSON representations of structs, rather than normal Go field names
	v.validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return fmt.Sprintf("'%s'", name)
	})

	err := v.validator.Struct(payload)
	if err != nil {
		errs := translateError(err, trans)
		var errStrings []string
		for _, transErr := range errs {
			errStrings = append(errStrings, transErr.Error())
		}
		translErr := strings.Join(errStrings, ", ")
		return presenters.NewErrorf(http.StatusBadRequest, translErr)
	}

	return nil
}

func translateError(err error, trans ut.Translator) (errs []error) {
	if err == nil {
		return nil
	}

	validatorErrs := err.(validator.ValidationErrors)

	validatorErrs.Translate(trans)

	for _, e := range validatorErrs {
		translatedErr := fmt.Errorf(e.Translate(trans))
		errs = append(errs, translatedErr)
	}

	return errs
}

// validateJobType is a custom implementation of job.Type field validation
func validateJobType(sl validator.StructLevel) {
	s := sl.Current().FieldByName("v").String()
	if s == "" {
		sl.ReportError(s, "type", "v", "required", "")
		return
	}

	if s != job.TypeFE.String() && s != job.TypeSD.String() && s != job.TypeAll.String() {
		param := fmt.Sprintf("'%s' '%s' '%s'", job.TypeFE, job.TypeSD, job.TypeAll)
		sl.ReportError(s, "type", "v", "oneof", param)
	}
}
