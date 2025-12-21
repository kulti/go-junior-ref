package httpapi_test

import (
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	listID := createList(t, jsonShouldMarshal(t, map[string]any{
		"name": "My List",
	}))

	itemID1 := createItem(t, listID, jsonShouldMarshal(t, map[string]any{
		"name": "My Item 1",
	}))

	itemID2 := createItem(t, listID, jsonShouldMarshal(t, map[string]any{
		"name": "My Item 2",
	}))

	listBody := getList(t, listID)

	require.JSONEq(t, string(jsonShouldMarshal(t, map[string]any{
		"list": map[string]any{
			"id":   listID,
			"name": "My List",
			"items": []map[string]any{
				{
					"id":   itemID1,
					"name": "My Item 1",
					"done": false,
				},
				{
					"id":   itemID2,
					"name": "My Item 2",
					"done": false,
				},
			},
		},
	})), string(listBody))
}

func createList(t *testing.T, body []byte) string {
	t.Helper()
	resp := doStep(t, httpReq{
		method: http.MethodPost,
		url:    "http://127.0.0.1:8090/v1/lists",
		body:   body,
	})
	require.Equal(t, http.StatusOK, resp.statusCode)
	var respData struct {
		List struct {
			ID string `json:"id"`
		} `json:"list"`
	}
	jsonShouldUnmarshal(t, resp.body, &respData)
	return respData.List.ID
}

func createItem(t *testing.T, listID string, body []byte) string {
	t.Helper()
	resp := doStep(t, httpReq{
		method: http.MethodPost,
		url:    "http://127.0.0.1:8090/v1/lists/" + listID + "/items",
		body:   body,
	})
	require.Equal(t, http.StatusOK, resp.statusCode)
	var respData struct {
		Item struct {
			ID string `json:"id"`
		} `json:"item"`
	}
	jsonShouldUnmarshal(t, resp.body, &respData)
	return respData.Item.ID
}

func getList(t *testing.T, listID string) []byte {
	t.Helper()
	resp := doStep(t, httpReq{
		method: http.MethodGet,
		url:    "http://127.0.0.1:8090/v1/lists/" + listID,
	})
	require.Equal(t, http.StatusOK, resp.statusCode)
	return resp.body
}

type httpReq struct {
	method  string
	url     string
	headers http.Header
	body    []byte
}

type httpResp struct {
	statusCode int
	headers    http.Header
	body       []byte
}

func doStep(t *testing.T, r httpReq) httpResp {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), r.method, r.url, bytes.NewReader(r.body))
	require.NoError(t, err)

	maps.Copy(req.Header, r.headers)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return httpResp{
		statusCode: resp.StatusCode,
		headers:    resp.Header,
		body:       respBodyBytes,
	}
}

func jsonShouldMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func jsonShouldUnmarshal(t *testing.T, data []byte, v any) {
	t.Helper()
	err := json.Unmarshal(data, v)
	require.NoError(t, err)
}
