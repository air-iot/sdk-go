package api

func (p *client) FindHandlerQuery(query, result interface{}) error {
	p.url.Path = "event/eventHandler"
	return p.Get(p.url, query, result)
}

func (p *client) FindHandlerById(id string, result interface{}) error {
	p.url.Path = "event/eventHandler"
	return p.GetById(p.url, id, result)
}

func (p *client) SaveHandler(data, result interface{}) error {
	p.url.Path = "event/eventHandler"
	return p.Post(p.url, data, result)
}

func (p *client) DelHandlerById(id string, result interface{}) error {
	p.url.Path = "event/eventHandler"
	return p.Delete(p.url, id, result)
}

func (p *client) UpdateHandlerById(id string, data, result interface{}) error {
	p.url.Path = "event/eventHandler"
	return p.Patch(p.url, id, data, result)
}

func (p *client) ReplaceHandlerById(id string, data, result interface{}) error {
	p.url.Path = "event/eventHandler"
	return p.Put(p.url, id, data, result)
}
