package server

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"

	"fiber_web/pkg/color"

	"github.com/gofiber/fiber/v2"
)

type FiberServer struct {
	app    *fiber.App
	config *Config
}

type Config struct {
	ReadTimeout               time.Duration
	WriteTimeout              time.Duration
	IdleTimeout               time.Duration
	Env                       string
	PreFork                   bool
	ServerHeader              string
	StrictRouting             bool
	CaseSensitive             bool
	BodyLimit                 int
	Concurrency               int
	Views                     fiber.Views
	DisableKeepalive          bool
	DisableDefaultDate        bool
	DisableDefaultContentType bool
	DisableStartupMessage     bool
	AppName                   string
}

// Option 定义配置选项的函数类型
type Option func(*Config)

// 默认配置
func defaultConfig() *Config {
	return &Config{
		ReadTimeout:   time.Second * 30,
		WriteTimeout:  time.Second * 30,
		IdleTimeout:   time.Second * 30,
		Env:           "development",
		PreFork:       false,
		ServerHeader:  "Fiber",
		StrictRouting: false,
		CaseSensitive: false,
		BodyLimit:     4 * 1024 * 1024, // 4MB
		Concurrency:   256 * 1024,      // 256k
	}
}

// WithReadTimeout 配置选项函数
func WithReadTimeout(t time.Duration) Option {
	return func(c *Config) {
		c.ReadTimeout = t
	}
}

func WithWriteTimeout(t time.Duration) Option {
	return func(c *Config) {
		c.WriteTimeout = t
	}
}

func WithIdleTimeout(t time.Duration) Option {
	return func(c *Config) {
		c.IdleTimeout = t
	}
}

func WithEnv(env string) Option {
	return func(c *Config) {
		c.Env = env
	}
}

func WithPreFork(enable bool) Option {
	return func(c *Config) {
		c.PreFork = enable
	}
}

func WithServerHeader(header string) Option {
	return func(c *Config) {
		c.ServerHeader = header
	}
}

func WithStrictRouting(enable bool) Option {
	return func(c *Config) {
		c.StrictRouting = enable
	}
}

func WithCaseSensitive(enable bool) Option {
	return func(c *Config) {
		c.CaseSensitive = enable
	}
}

func WithBodyLimit(limit int) Option {
	return func(c *Config) {
		c.BodyLimit = limit
	}
}

func WithConcurrency(concurrency int) Option {
	return func(c *Config) {
		c.Concurrency = concurrency
	}
}

func WithViews(views fiber.Views) Option {
	return func(c *Config) {
		c.Views = views
	}
}

func WithDisableKeepalive(disable bool) Option {
	return func(c *Config) {
		c.DisableKeepalive = disable
	}
}

func WithDisableStartupMessage(disable bool) Option {
	return func(c *Config) {
		c.DisableStartupMessage = disable
	}
}

func WithAppName(name string) Option {
	return func(c *Config) {
		c.AppName = name
	}
}

// NewFiberServer 创建一个新的 Fiber 服务器实例
func NewFiberServer(opts ...Option) *FiberServer {
	// 使用默认配置
	config := defaultConfig()

	// 应用所有选项
	for _, opt := range opts {
		opt(config)
	}

	// 创建 Fiber 实例
	app := fiber.New(fiber.Config{
		ReadTimeout:               config.ReadTimeout,
		WriteTimeout:              config.WriteTimeout,
		IdleTimeout:               config.IdleTimeout,
		Prefork:                   config.PreFork,
		ServerHeader:              config.ServerHeader,
		StrictRouting:             config.StrictRouting,
		CaseSensitive:             config.CaseSensitive,
		BodyLimit:                 config.BodyLimit,
		Concurrency:               config.Concurrency,
		Views:                     config.Views,
		DisableKeepalive:          config.DisableKeepalive,
		DisableDefaultDate:        config.DisableDefaultDate,
		DisableDefaultContentType: config.DisableDefaultContentType,
		DisableStartupMessage:     config.DisableStartupMessage,
		AppName:                   config.AppName,
	})

	return &FiberServer{
		app:    app,
		config: config,
	}
}

func (s *FiberServer) App() *fiber.App {
	return s.app
}

// printRoute 打印单个路由信息
func (s *FiberServer) printRoute(route *fiber.Route) bool {
	// 跳过自动生成的 HEAD 方法
	if route.Method == fiber.MethodHead {
		return false
	}

	// 获取处理函数信息
	var handlerName string
	if len(route.Handlers) > 0 {
		handler := route.Handlers[len(route.Handlers)-1]
		fullFuncName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()

		// 如果最后一个处理函数是匿名函数，说明可能是路由组或中间件，跳过
		if strings.Contains(fullFuncName, ".func") {
			return false
		}

		if idx := strings.LastIndex(fullFuncName, "."); idx != -1 {
			handlerName = fullFuncName[idx+1:]
		}
		// 美化处理函数名称
		handlerName = strings.TrimSuffix(handlerName, "-fm")
		handlerName = strings.TrimSuffix(handlerName, ".func1")
	}

	// 获取中间件信息
	middlewares := s.getRouteMiddlewares(route)

	// Gin 风格的输出
	middlewareStr := ""
	if len(middlewares) > 0 {
		middlewareStr = color.Colorize(" ("+strings.Join(middlewares, ",")+")", color.Magenta)
	}

	fmt.Printf("[%s] %s --> %s%s\n",
		color.Colorize(fmt.Sprintf("%-7s", route.Method), color.Method(route.Method)),
		color.Colorize(fmt.Sprintf("%-50s", route.Path), color.Blue),
		color.Colorize(handlerName, color.Cyan),
		middlewareStr,
	)

	return true
}

// getRouteMiddlewares 获取路由的所有中间件
func (s *FiberServer) getRouteMiddlewares(route *fiber.Route) []string {
	var middlewares []string

	// 查找路由组的中间件
	for _, r := range s.app.GetRoutes() {
		if strings.HasPrefix(route.Path, r.Path+"/") && len(r.Handlers) > 0 {
			// 收集路由组的中间件
			for _, handler := range r.Handlers {
				if name := s.getHandlerName(handler); name != "" {
					middlewares = append(middlewares, name)
				}
			}
		}
	}

	// 收集路由自己的中间件
	if len(route.Handlers) > 1 {
		for _, handler := range route.Handlers[:len(route.Handlers)-1] {
			if name := s.getHandlerName(handler); name != "" {
				middlewares = append(middlewares, name)
			}
		}
	}

	return middlewares
}

// getHandlerName 获取处理函数的名称
func (s *FiberServer) getHandlerName(handler fiber.Handler) string {
	name := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	if idx := strings.LastIndex(name, "/"); idx != -1 {
		name = name[idx+1:]
	}
	name = strings.TrimSuffix(name, ".func1")
	name = strings.TrimSuffix(name, ".New.func1")
	if name != "" && !strings.HasPrefix(name, "func") {
		return name
	}
	return ""
}

func (s *FiberServer) Start(addr string) error {
	// 如果是开发环境，打印路由信息
	if s.config.Env == "development" {
		fmt.Printf("\n%s\n", color.Colorize("[Fiber]", color.Green))

		for key := range s.app.GetRoutes() {
			s.printRoute(&s.app.GetRoutes()[key])
		}

		fmt.Printf("\n%s %s\n\n",
			color.Colorize("[Fiber]", color.Green),
			color.Colorize("Server listening on "+addr, color.Yellow),
		)
	}

	return s.app.Listen(addr)
}

func (s *FiberServer) Shutdown(ctx context.Context) error {
	// 使用 context 控制关闭超时
	done := make(chan error, 1)
	go func() {
		done <- s.app.Shutdown()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
