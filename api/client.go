package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/sirupsen/logrus"
	"gopkg.in/resty.v1"

	"github.com/air-iot/sdk-go/model"
	"github.com/air-iot/sdk-go/traefik"
)

// 根据 app appkey和appsecret 获取token
func FindToken() string {

	if traefik.AppKey == "" || traefik.AppSecret == "" {
		logrus.Warn("app key或者secret为空")
		return ""
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
		return ""
	}
	if resp.StatusCode() != 200 {
		logrus.Warnf("token查询错误,状态:%d,信息:%s", resp.StatusCode(), resp.String())
		return ""
	}
	return auth.Token
}

func Get(url1 url.URL, token string, query, result interface{}) error {
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
		SetHeader("Authorization", token).
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

func GetById(url url.URL, token, id string, result interface{}) error {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
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

func Post(url url.URL, token string, data, result interface{}) error {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
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

func Delete(url url.URL, token, id string, result interface{}) error {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
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

func Put(url url.URL, token, id string, data, result interface{}) error {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
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

func Patch(url url.URL, token, id string, data, result interface{}) error {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", token).
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
