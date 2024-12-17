package color

// ANSI 颜色码
const (
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Reset   = "\033[0m"
)

// Colorize 为文本添加颜色
func Colorize(text, color string) string {
	return color + text + Reset
}

// Method 根据 HTTP 方法返回对应的颜色
func Method(method string) string {
	switch method {
	case "GET":
		return Blue
	case "POST":
		return Green
	case "PUT":
		return Yellow
	case "DELETE":
		return Red
	case "PATCH":
		return Magenta
	case "HEAD":
		return Cyan
	default:
		return Reset
	}
}
