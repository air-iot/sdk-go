package api

import (
	"errors"
	"fmt"

	"gopkg.in/resty.v1"
)

func (p *client) ChangeCommand(id string, data, result interface{}) error {
	p.url.Path = "driver/driver"
	p.CheckToken()
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", p.Token).
		SetResult(result).
		SetBody(data).
		Post(fmt.Sprintf(`%s/%s/command`, p.url.String(), id))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return errors.New(resp.String())
	}
	return nil
}
