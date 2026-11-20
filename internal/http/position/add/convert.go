package add

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Chillobass/fork_template/internal/domain/trip"
	api "github.com/Chillobass/fork_template/internal/generated"
)

const maxRequestBodyBytes = 1 << 20

var (
	positionRequiredFields    = []string{"coordinates", "recorded_at"}
	coordinatesRequiredFields = []string{"latitude", "longitude"}
)

func decodeAddPositionRequest(w http.ResponseWriter, r *http.Request) (trip.PositionData, error) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes))
	if err != nil {
		return trip.PositionData{}, errors.New("request body is too large or could not be read")
	}

	var req api.PositionData
	if err := decodeStrict(body, &req); err != nil {
		return trip.PositionData{}, err
	}
	if err := checkRequiredFields(body); err != nil {
		return trip.PositionData{}, err
	}

	return trip.PositionData{
		Coordinates: trip.Point{
			Latitude:  req.Coordinates.Latitude,
			Longitude: req.Coordinates.Longitude,
		},
		RecordedAt: req.RecordedAt,
	}, nil
}

func decodeStrict(body []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return describeJSONError(err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}

func describeJSONError(err error) error {
	var (
		syntaxErr *json.SyntaxError
		typeErr   *json.UnmarshalTypeError
	)

	switch {
	case errors.As(err, &typeErr) && typeErr.Field == "":
		return errors.New("request body must be a JSON object")
	case errors.As(err, &typeErr):
		return fmt.Errorf("field %s has invalid type", typeErr.Field)
	case errors.As(err, &syntaxErr), errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		return errors.New("request body must be a valid JSON object")
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		return errors.New(strings.TrimPrefix(err.Error(), "json: "))
	default:
		return errors.New("invalid request body")
	}
}

func checkRequiredFields(body []byte) error {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil || payload == nil {
		return errors.New("request body must be a JSON object")
	}
	if err := requireFields(payload, "", positionRequiredFields); err != nil {
		return err
	}

	var coords map[string]json.RawMessage
	if err := json.Unmarshal(payload["coordinates"], &coords); err != nil || coords == nil {
		return errors.New("field coordinates must be an object")
	}
	return requireFields(coords, "coordinates.", coordinatesRequiredFields)
}

func requireFields(object map[string]json.RawMessage, prefix string, names []string) error {
	for _, name := range names {
		value, ok := object[name]
		if !ok || string(value) == "null" {
			return fmt.Errorf("field %s%s is required", prefix, name)
		}
	}
	return nil
}
