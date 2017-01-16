package session

type sessionStore interface {
	parseConfig(prefix string)
	storeForUser(key string, itemKey string, item interface{})
	get(userKey, itemKey string) (interface{}, bool)

	getInt(userKey, itemKey string) (int, bool)
	getUint64(userKey, itemKey string) (uint64, bool)
	getInts(userKey, itemKey string) ([]int, bool)
	getInt64(userKey, itemKey string) (int64, bool)
	getIntMap(userKey, itemKey string) (map[string]int, bool)
	getInt64Map(userKey, itemKey string) (map[string]int64, bool)

	getFloat64(userKey, itemKey string) (float64, bool)

	getString(userKey, itemKey string) (string, bool)
	getStrings(userKey, itemKey string) ([]string, bool)
	getStringMap(userKey, itemKey string) (map[string]string, bool)

	getBool(userKey, itemKey string) (bool, bool)
	getBytes(userKey, itemKey string) ([]byte, bool)
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

func (l *localStore) getBool(user, item string) (bool, bool) {
	if s, ok := l.store[user]; ok {
		if item, ok := s[item]; ok {
			return item.(bool), true
		}
	}
	return false, false
}

func (l *localStore) getBytes(user, item string) ([]byte, bool) {
	if s, ok := l.store[user]; ok {
		if item, ok := s[item]; ok {
			return item.([]byte), true
		}
	}
	return nil, false
}

func (l *localStore) getFloat64(user, item string) (float64, bool) {
	if s, ok := l.store[user]; ok {
		if item, ok := s[item]; ok {
			return item.(float64), true
		}
	}
	return 0, false
}

func (l *localStore) getInt(user, item string) (int, bool) {
	if s, ok := l.store[user]; ok {
		if item, ok := s[item]; ok {
			return item.(int), true
		}
	}
	return 0, false
}

func (l *localStore) getInt64(user, item string) (int64, bool) {
	if s, ok := l.store[user]; ok {
		if item, ok := s[item]; ok {
			return item.(int64), true
		}
	}
	return 0, false
}

func (l *localStore) getUint64(user, item string) (uint64, bool) {
	if s, ok := l.store[user]; ok {
		if item, ok := s[item]; ok {
			return item.(uint64), true
		}
	}
	return 0, false
}

func (l *localStore) getInts(user, item string) ([]int, bool) {
	if s, ok := l.store[user]; ok {
		if item, ok := s[item]; ok {
			return item.([]int), true
		}
	}
	return nil, false
}

func (l *localStore) getIntMap(user, item string) (map[string]int, bool) {
	if s, ok := l.store[user]; ok {
		if item, ok := s[item]; ok {
			return item.(map[string]int), true
		}
	}
	return nil, false
}

func (l *localStore) getInt64Map(user, item string) (map[string]int64, bool) {
	if s, ok := l.store[user]; ok {
		if item, ok := s[item]; ok {
			return item.(map[string]int64), true
		}
	}
	return nil, false
}

func (l *localStore) getString(user, item string) (string, bool) {
	if s, ok := l.store[user]; ok {
		if item, ok := s[item]; ok {
			return item.(string), true
		}
	}
	return "", false
}

func (l *localStore) getStrings(user, item string) ([]string, bool) {
	if s, ok := l.store[user]; ok {
		if item, ok := s[item]; ok {
			return item.([]string), true
		}
	}
	return nil, false
}

func (l *localStore) getStringMap(user, item string) (map[string]string, bool) {
	if s, ok := l.store[user]; ok {
		if item, ok := s[item]; ok {
			return item.(map[string]string), true
		}
	}
	return nil, false
}
