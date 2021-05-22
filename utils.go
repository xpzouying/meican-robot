package main

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

func postWithDecode(ctx context.Context, url string, body io.Reader, v interface{}) error {

	data, err := post200(ctx, url, body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &v)
}

func post200(ctx context.Context, url string, body io.Reader) ([]byte, error) {
	return send(ctx, http.MethodPost, url, http.StatusOK, body)
}

func send(ctx context.Context, method, url string, wantStatus int, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	resp, err := httpclient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != wantStatus {
		return nil, errors.Errorf("HTTP %s: %s (expected %v)", method, url, wantStatus)
	}

	return data, nil
}
