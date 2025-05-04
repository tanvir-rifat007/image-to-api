package main

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	pngstructure "github.com/dsoprea/go-png-image-structure"
)

type ResponsePayload struct {
	Success bool                   `json:"success"`
	Data    map[string]any `json:"data"`
	Message string                 `json:"message"`
}


func extractJSONFromPNG(data []byte) (map[string]any, error) {
	pmp := pngstructure.NewPngMediaParser()
	chunkSliceInterface, err := pmp.ParseBytes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PNG: %w", err)
	}

	sl := chunkSliceInterface.(*pngstructure.ChunkSlice)
	for _, chunk := range sl.Chunks() {
		switch chunk.Type {
		case "tEXt", "iTXt":
			// PNG text chunks are often in format: key\0value
			parts := bytes.SplitN(chunk.Data, []byte{0}, 2)
			if len(parts) == 2 {
				text := string(parts[1])
				if idx := strings.Index(text, "{"); idx >= 0 {
					var result map[string]any
					if err := json.Unmarshal([]byte(text[idx:]), &result); err == nil {
						return result, nil
					}
				}
			}
		case "zTXt":
			// zTXt is compressed, decompress it
			parts := bytes.SplitN(chunk.Data, []byte{0}, 2)
			if len(parts) != 2 || len(parts[1]) < 2 {
				continue
			}
			compressedData := parts[1][1:] // skip compression method byte
			r, err := zlib.NewReader(bytes.NewReader(compressedData))
			if err != nil {
				continue
			}
			defer r.Close()

			decompressed, err := io.ReadAll(r)
			if err != nil {
				continue
			}
			text := string(decompressed)
			if idx := strings.Index(text, "{"); idx >= 0 {
				var result map[string]any
				if err := json.Unmarshal([]byte(text[idx:]), &result); err == nil {
					return result, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("no JSON found in PNG metadata")
}


func (app *application) uploadImageHandler(w http.ResponseWriter, r *http.Request) {
	type RequestPayload struct {
		ImageBase64 string `json:"imageBase64"`
	}

	var req RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	base64Data := req.ImageBase64
	if strings.HasPrefix(base64Data, "data:image/png;base64,") {
		base64Data = strings.TrimPrefix(base64Data, "data:image/png;base64,")
	}

	imgData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		http.Error(w, "invalid base64", http.StatusBadRequest)
		return
	}

	jsonData, err := extractJSONFromPNG(imgData)
	if err != nil {
		http.Error(w, "failed to extract JSON: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	response := ResponsePayload{
		Success: true,
		Data:    jsonData,
		Message: "Successfully extracted JSON from image",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
