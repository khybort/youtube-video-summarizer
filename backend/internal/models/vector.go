package models

import (
	"database/sql/driver"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pgvector/pgvector-go"
)

// Vector is a custom type for pgvector that implements GORM interfaces
type Vector struct {
	Data []float32
}

// Scan implements the sql.Scanner interface for Vector
func (v *Vector) Scan(value interface{}) error {
	if value == nil {
		v.Data = nil
		return nil
	}

	// Handle pgvector.Vector type
	if vec, ok := value.(pgvector.Vector); ok {
		v.Data = vec.Slice()
		return nil
	}

	// Handle string (JSON or text representation)
	if str, ok := value.(string); ok {
		// Try to parse as JSON array first (GORM might return JSON string)
		var floatArray []float32
		if err := json.Unmarshal([]byte(str), &floatArray); err == nil {
			v.Data = floatArray
			return nil
		}
		
		// Try to parse as pgvector text format (e.g., "[1,2,3]")
		// Remove brackets if present
		cleaned := strings.TrimSpace(str)
		cleaned = strings.TrimPrefix(cleaned, "[")
		cleaned = strings.TrimSuffix(cleaned, "]")
		
		// Split by comma and parse floats
		parts := strings.Split(cleaned, ",")
		floatArray = make([]float32, 0, len(parts))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			var f float64
			if _, err := fmt.Sscanf(part, "%f", &f); err == nil {
				floatArray = append(floatArray, float32(f))
			}
		}
		if len(floatArray) > 0 {
			v.Data = floatArray
			return nil
		}
		
		return fmt.Errorf("cannot parse vector from string: %s", str)
	}

	// Handle byte array (raw PostgreSQL vector format or JSON)
	if bytes, ok := value.([]byte); ok {
		// Try to parse as JSON first (GORM might return JSON bytes)
		var floatArray []float32
		if err := json.Unmarshal(bytes, &floatArray); err == nil {
			v.Data = floatArray
			return nil
		}
		
		// Fallback to raw binary format (PostgreSQL vector binary format)
		if len(bytes) < 4 {
			return fmt.Errorf("invalid vector data: too short")
		}
		dim := int(binary.BigEndian.Uint16(bytes[0:2]))
		unused := int(binary.BigEndian.Uint16(bytes[2:4]))
		if unused != 0 {
			return fmt.Errorf("invalid vector data: unused bits not zero")
		}
		if len(bytes) < 4+dim*4 {
			return fmt.Errorf("invalid vector data: incomplete")
		}
		v.Data = make([]float32, dim)
		for i := 0; i < dim; i++ {
			v.Data[i] = float32(binary.BigEndian.Uint32(bytes[4+i*4:8+i*4]))
		}
		return nil
	}

	return fmt.Errorf("cannot scan %T into Vector", value)
}

// Value implements the driver.Valuer interface for Vector
func (v Vector) Value() (driver.Value, error) {
	if v.Data == nil {
		return nil, nil
	}
	vec := pgvector.NewVector(v.Data)
	return vec, nil
}

// GormDataType returns the GORM data type
func (Vector) GormDataType() string {
	return "vector(768)"
}

// Slice returns the underlying float32 slice
func (v Vector) Slice() []float32 {
	return v.Data
}

