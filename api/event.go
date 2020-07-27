package api

func (p *client) FindEventQuery(query, result interface{}) error {
	p.url.Path = "event/event"
	return p.Get(p.url, query, result)
}

func (p *client) FindEventById(id string, result interface{}) error {
	p.url.Path = "event/event"
	return p.GetById(p.url, id, result)
}

func (p *client) SaveEvent(data, result interface{}) error {
	p.url.Path = "event/event"
	return p.Post(p.url, data, result)
}

func (p *client) DelEventById(id string, result interface{}) error {
	p.url.Path = "event/event"
	return p.Delete(p.url, id, result)
}

func (p *client) UpdateEventById(id string, data, result interface{}) error {
	p.url.Path = "event/event"
	return p.Patch(p.url, id, data, result)
}

func (p *client) ReplaceEventById(id string, data, result interface{}) error {
	p.url.Path = "event/event"
	return p.Put(p.url, id, data, result)
}
