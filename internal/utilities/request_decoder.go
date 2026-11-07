package utility

import (
    "encoding/json"
    "net/http"
    "net/url"
)

type RequestData[T any] struct {
    Body    T
    Params  map[string]string
    Query   url.Values
    Headers http.Header
}

func DecodeRequest[T any](
    r *http.Request,
    pathParams ...string,
) (RequestData[T], error) {

    var data RequestData[T]

    // Decode body (if present)
    if r.Body != nil && r.Body != http.NoBody {
        err := json.NewDecoder(r.Body).Decode(&data.Body)
        if err != nil {
            return data, err
        }
    }

    data.Params = make(map[string]string)

    for _, key := range pathParams {
        data.Params[key] = r.PathValue(key)
    }

    data.Query = r.URL.Query()


    return data, nil
}