package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// contextKey is a custom type to avoid context key collisions in request context.
type contextKey string

// UserIDKey is the key used to store and retrieve the user's ID from the request context.
// The handler will use this exact constant.
const UserIDKey contextKey = "userID"

// snakeToPascal converts a snake_case string to PascalCase (e.g., "default_region" -> "DefaultRegion").
func snakeToPascal(snake string) string {
	parts := strings.Split(snake, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// ApplyUpdates dynamically applies updates from a map to a struct using reflection.
// This is the missing function that your service layer needs.
func ApplyUpdates(target interface{}, updates map[string]interface{}) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to a struct")
	}
	val = val.Elem() // Get the struct value itself

	for key, value := range updates {
		pascalKey := snakeToPascal(key)
		field := val.FieldByName(pascalKey)

		if !field.IsValid() {
			continue // Field does not exist in struct, skip it
		}
		if !field.CanSet() {
			continue // Field is not settable (e.g., unexported), skip it
		}

		updateVal := reflect.ValueOf(value)

		// Handle special case for datatypes.JSON
		if field.Type() == reflect.TypeOf(datatypes.JSON{}) {
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal field '%s' to JSON: %w", key, err)
			}
			field.Set(reflect.ValueOf(datatypes.JSON(jsonBytes)))
			continue
		}

		// Handle nil for pointer types
		if value == nil {
			if field.Kind() == reflect.Ptr || field.Kind() == reflect.Interface || field.Kind() == reflect.Slice || field.Kind() == reflect.Map {
				field.Set(reflect.Zero(field.Type())) // Set to nil
				continue
			}
		}

		// If field is a pointer, create a new value and set its pointer
		if field.Kind() == reflect.Ptr {
			if updateVal.IsValid() {
				elemType := field.Type().Elem()
				newPtr := reflect.New(elemType)
				if updateVal.Type().ConvertibleTo(elemType) {
					newPtr.Elem().Set(updateVal.Convert(elemType))
					field.Set(newPtr)
				} else {
					return fmt.Errorf("cannot convert update value for field '%s' to type %s", key, elemType)
				}
			}
		} else if updateVal.IsValid() && updateVal.Type().ConvertibleTo(field.Type()) {
			// For non-pointer fields, convert and set
			field.Set(updateVal.Convert(field.Type()))
		}
	}

	// Specific logic: if the name is updated, regenerate the slug
	if nameUpdate, ok := updates["name"]; ok {
		if nameStr, isStr := nameUpdate.(string); isStr && nameStr != "" {
			slugField := val.FieldByName("Slug")
			if slugField.IsValid() && slugField.CanSet() {
				slugField.SetString(GenerateSlug(nameStr))
			}
		}
	}

	return nil
}

// GenerateRandomString creates a random alphanumeric string of a given length.
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// GenerateSlug creates a URL-friendly slug from a string.
func GenerateSlug(s string) string {
	s = strings.ToLower(s)
	reg := regexp.MustCompile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 50 {
		s = s[:50]
	}
	return s
}

// ParseUUID validates if a string is a valid UUID.
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// StringInSlice checks if a given string exists in a slice
func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
