package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/Chillobass/fork_template/internal/domain"
	api "github.com/Chillobass/fork_template/internal/generated"
	"github.com/Chillobass/fork_template/internal/service"
)

const maxRequestBodyBytes = 1 << 20

var (
	tripDataRequiredFields    = []string{"user_id", "driver_id", "start_point", "end_point", "price"}
	coordinatesRequiredFields = []string{"latitude", "longitude"}
)

func decodeCreateTripRequest(w http.ResponseWriter, r *http.Request) (service.CreateTripInput, error) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes))
	if err != nil {
		return service.CreateTripInput{}, errors.New("request body is too large or could not be read")
	}

	var req api.TripData
	if err := decodeStrict(body, &req); err != nil {
		return service.CreateTripInput{}, err
	}

	// после Unmarshal не отличить "price": 0 от отсутствующего price, поэтому наличие полей проверяем отдельно
	if err := checkRequiredFields(body); err != nil {
		return service.CreateTripInput{}, err
	}

	if err := validateTripData(req); err != nil {
		return service.CreateTripInput{}, err
	}

	return service.CreateTripInput{
		UserID:     req.UserId,
		DriverID:   req.DriverId,
		StartPoint: domain.Point{Latitude: req.StartPoint.Latitude, Longitude: req.StartPoint.Longitude},
		EndPoint:   domain.Point{Latitude: req.EndPoint.Latitude, Longitude: req.EndPoint.Longitude},
		Price:      req.Price,
	}, nil
}

func decodeStrict(body []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return describeJSONError(err)
	}
	if dec.More() {
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
		return fmt.Errorf("invalid request body: %w", err)
	}
}

func checkRequiredFields(body []byte) error {
	var trip map[string]json.RawMessage
	if err := json.Unmarshal(body, &trip); err != nil || trip == nil {
		return errors.New("request body must be a JSON object")
	}
	if err := requireFields(trip, "", tripDataRequiredFields); err != nil {
		return err
	}

	for _, pointField := range []string{"start_point", "end_point"} {
		var point map[string]json.RawMessage
		if err := json.Unmarshal(trip[pointField], &point); err != nil {
			return fmt.Errorf("field %s must be an object", pointField)
		}
		if err := requireFields(point, pointField+".", coordinatesRequiredFields); err != nil {
			return err
		}
	}

	return nil
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

func validateTripData(req api.TripData) error {
	if req.UserId == uuid.Nil {
		return errors.New("user_id must be a non-empty UUID")
	}
	if req.DriverId == uuid.Nil {
		return errors.New("driver_id must be a non-empty UUID")
	}
	if err := validateCoordinates("start_point", req.StartPoint); err != nil {
		return err
	}
	if err := validateCoordinates("end_point", req.EndPoint); err != nil {
		return err
	}
	if req.Price < 0 {
		return errors.New("price must be greater than or equal to 0")
	}

	return nil
}

func validateCoordinates(field string, c api.Coordinates) error {
	if c.Latitude < -90 || c.Latitude > 90 {
		return fmt.Errorf("%s.latitude must be between -90 and 90", field)
	}
	if c.Longitude < -180 || c.Longitude > 180 {
		return fmt.Errorf("%s.longitude must be between -180 and 180", field)
	}

	return nil
}
