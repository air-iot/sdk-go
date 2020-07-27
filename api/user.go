package api

func (p *client) FindUserQuery(query, result interface{}) error {
	p.url.Path = "core/user"
	return p.Get(p.url, query, result)
}

func (p *client) FindUserById(id string, result interface{}) error {
	p.url.Path = "core/user"
	return p.GetById(p.url, id, result)
}

func (p *client) SaveUser(data, result interface{}) error {
	p.url.Path = "core/user"
	return p.Post(p.url, data, result)
}

func (p *client) DelUserById(id string, result interface{}) error {
	p.url.Path = "core/user"
	return p.Delete(p.url, id, result)
}

func (p *client) UpdateUserById(id string, data, result interface{}) error {
	p.url.Path = "core/user"
	return p.Patch(p.url, id, data, result)
}

func (p *client) ReplaceUserById(id string, data, result interface{}) error {
	p.url.Path = "core/user"
	return p.Put(p.url, id, data, result)
}
