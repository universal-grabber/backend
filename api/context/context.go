package context

var injections = make(map[string]interface{})

func Register(name string, object interface{}) {
	injections[name] = object
}

func Get(name string) interface{} {
	return injections[name]
}
