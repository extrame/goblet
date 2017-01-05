package session

type sessionStore interface {
	parseConfig(prefix string)
	storeForUser(key string, itemKey string, item interface{})
	get(userKey, itemKey string) (interface{}, bool)
	getInt(userKey, itemKey string) (int, bool)
	init() error
}

type localStore struct {
	store map[string]map[string]interface{}
}

func (l *localStore) parseConfig(prefix string) {
}

func (l *localStore) init() error {
	l.store = make(map[string]map[string]interface{})
	return nil
}

func (l *localStore) storeForUser(userKey, itemKey string, item interface{}) {
	if _, ok := l.store[userKey]; !ok {
		l.store[userKey] = make(map[string]interface{})
	}
	l.store[userKey][itemKey] = item
}

func (l *localStore) get(user, item string) (interface{}, bool) {
	if s, ok := l.store[user]; ok {
		if item, ok := s[item]; ok {
			return item, true
		}
	}
	return nil, false
}

func (l *localStore) getInt(user, item string) (int, bool) {
	if s, ok := l.store[user]; ok {
		if item, ok := s[item]; ok {
			return item.(int), true
		}
	}
	return 0, false
}
