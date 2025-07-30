package utils

type Object map[string]any

func (o *Object) Get(key string) (any, bool) {
	v, ok := (*o)[key]
	return v, ok
}

func (o *Object) Set(key string, v any) {
	(*o)[key] = v
}

func (o *Object) Delete(key string) {
	delete((*o), key)
}
