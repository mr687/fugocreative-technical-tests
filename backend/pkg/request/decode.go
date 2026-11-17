package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/mr687/fugocreative-technical-tests/pkg/validation"
)

func DecodeAndValidate(w http.ResponseWriter, r *http.Request, target any) ([]validation.FieldError, error) {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return nil, err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("request contains multiple JSON values")
		}
		return nil, err
	}
	return validation.Struct(target), nil
}
