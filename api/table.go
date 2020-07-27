package api

func (p *client) FindTableQuery(query, result interface{}) error {
	p.url.Path = "core/table"
	return p.Get(p.url, query, result)
}

func (p *client) FindTableById(id string, result interface{}) error {
	p.url.Path = "core/table"
	return p.GetById(p.url, id, result)
}

func (p *client) SaveTable(data, result interface{}) error {
	p.url.Path = "core/table"
	return p.Post(p.url, data, result)
}

func (p *client) DelTableById(id string, result interface{}) error {
	p.url.Path = "core/table"
	return p.Delete(p.url, id, result)
}

func (p *client) UpdateTableById(id string, data, result interface{}) error {
	p.url.Path = "core/table"
	return p.Patch(p.url, id, data, result)
}

func (p *client) ReplaceTableById(id string, data, result interface{}) error {
	p.url.Path = "core/table"
	return p.Put(p.url, id, data, result)
}
