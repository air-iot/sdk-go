package api

import (
	"errors"
	"fmt"
	"net/url"

	"gopkg.in/resty.v1"
)

type (
	RealTimeData struct {
		TagId string      `json:"tagId"`
		Uid   string      `json:"uid"`
		Time  int64       `json:"time"`
		Value interface{} `json:"value"`
	}

	QueryData struct {
		Results []Results `json:"results"`
	}

	Series struct {
		Name    string          `json:"name"`
		Columns []string        `json:"columns"`
		Values  [][]interface{} `json:"values"`
	}

	Results struct {
		Series []Series `json:"series"`
	}
)

func (p *client) GetLatest(query interface{}) (result []RealTimeData, err error) {
	p.url.Path = "core/data"
	p.CheckToken()
	b, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	v := url.Values{}
	v.Set("query", string(b))
	result = make([]RealTimeData, 0)
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", p.Token).
		SetResult(&result).
		Get(fmt.Sprintf(`%s/latest?%s`, p.url.String(), v.Encode()))

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}

	return result, nil
}

func (p *client) PostLatest(data interface{}) (result []RealTimeData, err error) {
	p.url.Path = "core/data"
	p.CheckToken()
	result = make([]RealTimeData, 0)
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", p.Token).
		SetResult(&result).
		SetBody(data).
		Post(fmt.Sprintf(`%s/latest`, p.url.String()))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}
	return result, nil
}

func (p *client) GetQuery(query interface{}) (result *QueryData, err error) {
	p.url.Path = "core/data"
	p.CheckToken()
	result = new(QueryData)
	b, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	v := url.Values{}
	v.Set("query", string(b))
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", p.Token).
		SetResult(result).
		Get(fmt.Sprintf(`%s/query?%s`, p.url.String(), v.Encode()))

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}

	return result, nil
}

func (p *client) PostQuery(query interface{}) (result *QueryData, err error) {
	p.url.Path = "core/data"
	p.CheckToken()
	result = new(QueryData)
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", p.Token).
		SetResult(result).
		SetBody(query).
		Post(fmt.Sprintf(`%s/query`, p.url.String()))

	if err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.String())
	}

	return result, nil
}
