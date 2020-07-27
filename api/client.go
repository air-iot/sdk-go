package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"

	"github.com/air-iot/sdk-go/model"
	"github.com/air-iot/sdk-go/traefik"
)

type AuthToken struct {
	Token   string
	Expires int64
}

// 根据 app appkey和appsecret 获取token
func (p *AuthToken) FindToken() {

	if traefik.AppKey == "" || traefik.AppSecret == "" {
		logrus.Warn("app key或者secret为空")
		return
	}
	// 生成要访问的url
	u := url.URL{Host: net.JoinHostPort(traefik.Host, strconv.Itoa(traefik.Port)), Path: "core/auth/token"}
	v := url.Values{}
	v.Set("appkey", traefik.AppKey)
	v.Set("appsecret", traefik.AppSecret)
	u.Scheme = traefik.Proto
	u.RawQuery = v.Encode()
	auth := new(model.Auth)
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetResult(auth).
		Get(u.String())
	if err != nil {
		logrus.Warnf("token查询错误,", err.Error())
		return
	}
	if resp.StatusCode() != 200 {
		logrus.Warnf("token查询错误,状态:%d,信息:%s", resp.StatusCode(), resp.String())
		return
	}
	p.Token = auth.Token
	p.Expires = auth.Expires + time.Now().UnixNano()

	return
}

func (p *AuthToken) Get(url1 url.URL, query, result interface{}) error {
	p.CheckToken()
	b, err := json.Marshal(query)
	if err != nil {
		return err
	}
	v := url.Values{}
	v.Set("query", string(b))
	u := fmt.Sprintf(`%s?%s`, url1.String(), v.Encode())
	logrus.Debugf("查询请求url:%s", u)
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", p.Token).
		SetResult(result).
		Get(u)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return errors.New(resp.String())
	}

	return nil
}

func (p *AuthToken) GetById(url url.URL, id string, result interface{}) error {
	p.CheckToken()
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", p.Token).
		SetResult(result).
		Get(fmt.Sprintf(`%s/%s`, url.String(), id))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return errors.New(resp.String())
	}
	return nil
}

func (p *AuthToken) Post(url url.URL, data, result interface{}) error {
	p.CheckToken()
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", p.Token).
		SetResult(result).
		SetBody(data).
		Post(url.String())
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return errors.New(resp.String())
	}
	return nil
}

func (p *AuthToken) Delete(url url.URL, id string, result interface{}) error {
	p.CheckToken()
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", p.Token).
		SetResult(result).
		Delete(fmt.Sprintf(`%s/%s`, url.String(), id))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return errors.New(resp.String())
	}
	return nil
}

func (p *AuthToken) Put(url url.URL, id string, data, result interface{}) error {
	p.CheckToken()
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", p.Token).
		SetResult(result).
		SetBody(data).
		Put(fmt.Sprintf(`%s/%s`, url.String(), id))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return errors.New(resp.String())
	}
	return nil
}

func (p *AuthToken) Patch(url url.URL, id string, data, result interface{}) error {
	p.CheckToken()
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", p.Token).
		SetResult(result).
		SetBody(data).
		Patch(fmt.Sprintf(`%s/%s`, url.String(), id))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return errors.New(resp.String())
	}
	return nil
}

func (p *AuthToken) CheckToken() {
	if p.Expires < time.Now().UnixNano() {
		p.FindToken()
	}
	return
}
