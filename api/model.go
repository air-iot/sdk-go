package api

func (p *client) FindModelQuery(query, result interface{}) error {
	p.url.Path = "core/model"
	return p.Get(p.url, query, result)
}

func (p *client) FindModelById(id string, result interface{}) error {
	p.url.Path = "core/model"
	return p.GetById(p.url, id, result)
}

func (p *client) SaveModel(data, result interface{}) error {
	p.url.Path = "core/model"
	return p.Post(p.url, data, result)
}

func (p *client) DelModelById(id string, result interface{}) error {
	p.url.Path = "core/model"
	return p.Delete(p.url, id, result)
}

func (p *client) UpdateModelById(id string, data, result interface{}) error {
	p.url.Path = "core/model"
	return p.Patch(p.url, id, data, result)
}

func (p *client) ReplaceModelById(id string, data, result interface{}) error {
	p.url.Path = "core/model"
	return p.Put(p.url, id, data, result)
}
