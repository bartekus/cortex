package gov

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type XrayIndex struct {
	Root   string         `json:"root"`
	Digest string         `json:"digest"`
	Files  []XrayFileNode `json:"files"`
	// Other fields are ignored for validation logic if not strictly required by invariants
	// But we need to keep them for digest calculation, so we use map for digest.
}

type XrayFileNode struct {
	Path string `json:"path"`
}

func CheckXrayDrift(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read XRAY fixture %s: %w", path, err)
	}

	// 1. Validate Schema & Structure
	var index XrayIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return fmt.Errorf("failed to parse XRAY fixture JSON: %w", err)
	}

	if index.Root == "" {
		return fmt.Errorf("XRAY fixture missing required field: root")
	}
	if index.Files == nil {
		return fmt.Errorf("XRAY fixture missing required field: files")
	}

	// 2. Validate Sorting of files
	if !sort.SliceIsSorted(index.Files, func(i, j int) bool {
		return index.Files[i].Path < index.Files[j].Path
	}) {
		// Find the unsorted one for better error message
		for i := 0; i < len(index.Files)-1; i++ {
			if index.Files[i].Path > index.Files[i+1].Path {
				return fmt.Errorf("XRAY fixture files are not sorted: %s > %s", index.Files[i].Path, index.Files[i+1].Path)
			}
		}
		return fmt.Errorf("XRAY fixture files are not sorted (unknown position)")
	}

	// Check for uniqueness
	for i := 0; i < len(index.Files)-1; i++ {
		if index.Files[i].Path == index.Files[i+1].Path {
			return fmt.Errorf("XRAY fixture contains duplicate file path: %s", index.Files[i].Path)
		}
	}

	// 3. Verify Digest
	var rawMap map[string]interface{}
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return fmt.Errorf("failed to parse JSON map: %w", err)
	}

	if _, ok := rawMap["digest"]; !ok {
		return fmt.Errorf("XRAY fixture missing digest field")
	}

	// Remove digest for calculation
	delete(rawMap, "digest")

	canonicalJSON, err := json.Marshal(rawMap)
	if err != nil {
		return fmt.Errorf("failed to marshal canonical JSON: %w", err)
	}

	hash := sha256.Sum256(canonicalJSON)
	calculatedDigest := hex.EncodeToString(hash[:])

	if index.Digest != calculatedDigest {
		return fmt.Errorf("XRAY fixture digest mismatch!\nExpected: %s\nCalculated: %s\n(Note: Calculated over canonical JSON without 'digest' field)", index.Digest, calculatedDigest)
	}

	return nil
}
