package ext

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	jsoniter "github.com/json-iterator/go"
	"gopkg.in/resty.v1"

	"github.com/air-iot/sdk-go/api"
	"github.com/air-iot/sdk-go/traefik"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type ExtClient interface {
	FindQuery(query, result interface{}) error
	FindById(id string, result interface{}) error
	Save(data, result interface{}) error
	SaveMany(data, result interface{}) error
	DelById(id string, result interface{}) error
	UpdateById(id string, data, result interface{}) error
	ReplaceById(id string, data, result interface{}) error
}

type extClient struct {
	url   url.URL
	token string
}

func NewExtClient(collection string) ExtClient {
	cli := new(extClient)
	u := url.URL{Host: net.JoinHostPort(traefik.Host, strconv.Itoa(traefik.Port)), Path: fmt.Sprintf("core/ext/%s", collection)}
	u.Scheme = traefik.Proto
	cli.url = u
	cli.token = api.FindToken()

	return cli
}

func (p *extClient) FindQuery(query, result interface{}) error {
	return api.Get(p.url, p.token, query, result)
}

func (p *extClient) Save(data, result interface{}) error {
	return api.Post(p.url, p.token, data, result)
}

func (p *extClient) SaveMany(data, result interface{}) error {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", p.token).
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

func (p *extClient) FindById(id string, result interface{}) error {
	return api.GetById(p.url, p.token, id, result)
}

func (p *extClient) DelById(id string, result interface{}) error {
	return api.Delete(p.url, p.token, id, result)
}

func (p *extClient) UpdateById(id string, data, result interface{}) error {
	return api.Patch(p.url, p.token, id, data, result)
}

func (p *extClient) ReplaceById(id string, data, result interface{}) error {
	return api.Put(p.url, p.token, id, data, result)
}
