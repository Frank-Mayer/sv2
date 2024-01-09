package save

var (
	store = make(map[string][]float32)
)

func Add(key string, value float32) {
	if _, ok := store[key]; !ok {
		store[key] = make([]float32, 0)
	}
	store[key] = append(store[key], value)
}

func Get(key string) (data []float32, ok bool) {
	data, ok = store[key]
	return
}

func Keys() []string {
	keys := make([]string, 0, len(store))
	for key := range store {
		keys = append(keys, key)
	}
	return keys
}
