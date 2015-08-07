package render

import (
	"reflect"
)

type sorter struct {
	array []interface{}
	by    string
}

// Len is part of sort.Interface.
func (s *sorter) Len() int {
	return len(s.array)
}

// Swap is part of sort.Interface.
func (s *sorter) Swap(i, j int) {
	s.array[i], s.array[j] = s.array[j], s.array[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *sorter) Less(i, j int) bool {
	value_i, value_j := s.array[i], s.array[j]
	field_i := reflect.ValueOf(value_i).FieldByName(s.by)
	field_j := reflect.ValueOf(value_j).FieldByName(s.by)
	switch tf := field_i.Interface().(type) {
	case string:
		return tf < field_j.Interface().(string)
	case int:
		return tf < field_j.Interface().(int)
	}
	return false
}
