package grpcx

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

const JSONCodecName = "json"

func init() {
	encoding.RegisterCodec(jsonCodec{})
}

type jsonCodec struct{}

func (jsonCodec) Name() string {
	return JSONCodecName
}

func (jsonCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
