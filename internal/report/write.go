package report

import (
    "encoding/json"
    "os"
)

func WriteJSON(path string, v any) error {
    f, err := os.Create(path)
    if err != nil {
        return err
    }
    defer f.Close()

    enc := json.NewEncoder(f)
    enc.SetIndent("", "  ")
    return enc.Encode(v)
}