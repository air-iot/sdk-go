package api

func (p *client) FindSettingQuery(query, result interface{}) error {
	p.url.Path = "core/setting"
	return p.Get(p.url, query, result)
}

func (p *client) FindSettingById(id string, result interface{}) error {
	p.url.Path = "core/setting"
	return p.GetById(p.url, id, result)
}

func (p *client) SaveSetting(data, result interface{}) error {
	p.url.Path = "core/setting"
	return p.Post(p.url, data, result)
}

func (p *client) DelSettingById(id string, result interface{}) error {
	p.url.Path = "core/setting"
	return p.Delete(p.url, id, result)
}

func (p *client) UpdateSettingById(id string, data, result interface{}) error {
	p.url.Path = "core/setting"
	return p.Patch(p.url, id, data, result)
}

func (p *client) ReplaceSettingById(id string, data, result interface{}) error {
	p.url.Path = "core/setting"
	return p.Put(p.url, id, data, result)
}
