package api

import (
	"errors"
	"fmt"

	"gopkg.in/resty.v1"
)

func (p *client) FindExtQuery(collection string, query, result interface{}) error {
	p.url.Path = fmt.Sprintf("core/ext/%s", collection)
	return p.Get(p.url, query, result)
}

func (p *client) SaveExt(collection string, data, result interface{}) error {
	p.url.Path = fmt.Sprintf("core/ext/%s", collection)
	return p.Post(p.url, data, result)
}

func (p *client) SaveManyExt(collection string, data, result interface{}) error {
	p.url.Path = fmt.Sprintf("core/ext/%s", collection)
	p.CheckToken()
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", p.Token).
		SetResult(result).
		SetBody(data).
		Post(fmt.Sprintf(`%s/many`, p.url.String()))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return errors.New(resp.String())
	}
	return json.Unmarshal(resp.Body(), result)
}

func (p *client) FindExtById(collection, id string, result interface{}) error {
	p.url.Path = fmt.Sprintf("core/ext/%s", collection)
	return p.GetById(p.url, id, result)
}

func (p *client) DelExtById(collection, id string, result interface{}) error {
	p.url.Path = fmt.Sprintf("core/ext/%s", collection)
	return p.Delete(p.url, id, result)
}

func (p *client) UpdateExtById(collection, id string, data, result interface{}) error {
	p.url.Path = fmt.Sprintf("core/ext/%s", collection)
	return p.Patch(p.url, id, data, result)
}

func (p *client) ReplaceExtById(collection, id string, data, result interface{}) error {
	p.url.Path = fmt.Sprintf("core/ext/%s", collection)
	return p.Put(p.url, id, data, result)
}
