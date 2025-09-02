package util

import (
	"log"

	"github.com/go-viper/mapstructure/v2"
)

func DecodeExtraTo[T any](m map[string]any) (T, error) {
	var meta mapstructure.Metadata
	var out T
	if m == nil {
		return out, nil
	}

	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &out,
		Metadata:         &meta,
		ErrorUnset:       false,
		WeaklyTypedInput: true,
		ZeroFields:       true,
	})
	if err != nil {
		return out, err
	}

	if err := dec.Decode(m); err != nil {
		return out, err
	}
	if len(meta.Unused) > 0 {
		log.Printf("unused fields: %v", meta.Unused)
	}

	return out, nil
}
