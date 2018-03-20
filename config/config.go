package config

var (
	GlobalConfiger     Configer
)



// 每一个配置需实现如下接口
// key的形式 a::b::c  用::表示Map的层级,
// 如果key不存在需要反回error
type Configer interface {
	String(key string) (string, error)
	Strings(key string) ([]string, error)
	Int(key string) (int, error)
	Ints(key string) ([]int, error)
	Float(key string) (float64, error)
	Floats(key string) ([]float64, error)
	Bool(key string) (bool, error)
	Bools(key string) ([]bool, error)
	Map(key string) (map[string]interface{}, error)
	Maps(key string) ([]map[string]interface{}, error)
}

func String(key string) (string, error) {
	return GlobalConfiger.String(key)
}

func Strings(key string)([]string, error) {
	return GlobalConfiger.Strings(key)
}

func Int(key string)(int, error) {
	return GlobalConfiger.Int(key)
}

func Ints(key string)([]int, error) {
	return GlobalConfiger.Ints(key)
}

func Float(key string)(float64, error) {
	return GlobalConfiger.Float(key)
}

func Floats(key string)([]float64, error) {
	return GlobalConfiger.Floats(key)
}

func Bool(key string)(bool, error) {
	return GlobalConfiger.Bool(key)
}

func Bools(key string)([]bool, error) {
	return GlobalConfiger.Bools(key)
}

func Map(key string)(map[string]interface{}, error) {
	return GlobalConfiger.Map(key)
}

func Maps(key string)([]map[string]interface{}, error) {
	return GlobalConfiger.Maps(key)
}


func InitYamlGlobalConfiger(filename string) {
	GlobalConfiger = NewYamlConfig(filename)
}
