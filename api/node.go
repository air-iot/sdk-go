package api

func (p *client) FindNodeQuery(query, result interface{}) error {
	p.url.Path = "core/node"
	return p.Get(p.url, query, result)
}

func (p *client) FindNodeById(id string, result interface{}) error {
	p.url.Path = "core/node"
	return p.GetById(p.url, id, result)
}

func (p *client) SaveNode(data, result interface{}) error {
	p.url.Path = "core/node"
	return p.Post(p.url, data, result)
}

func (p *client) DelNodeById(id string, result interface{}) error {
	p.url.Path = "core/node"
	return p.Delete(p.url, id, result)
}

func (p *client) UpdateNodeById(id string, data, result interface{}) error {
	p.url.Path = "core/node"
	return p.Patch(p.url, id, data, result)
}

func (p *client) ReplaceNodeById(id string, data, result interface{}) error {
	p.url.Path = "core/node"
	return p.Put(p.url, id, data, result)
}
