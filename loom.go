package loom

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/form"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func New(deps *Deps) *Loom {
	return &Loom{
		E:                  echo.New(),
		Deps:               deps,
		controllerRegistry: make(map[string]*controller),
		methodRegistry:     make(map[string]*methodCall),
	}
}

type Loom struct {
	E *echo.Echo
	*Deps

	controllerRegistry map[string]*controller
	methodRegistry     map[string]*methodCall
}

type controller struct {
	Type     reflect.Type
	Instance any
}

type methodCall struct {
	controllerInstance any
	method             reflect.Method
}

func (g *Loom) Start(addr string) error {
	return g.E.Start(addr)
}

// Register registers a controller
// Usage: loom.Register[*UsersController](l)
func Register[T any](l *Loom) {
	var zero T

	controllerType := reflect.TypeOf(zero)
	typeName := controllerType.Elem().Name()

	if l.controllerRegistry == nil {
		l.controllerRegistry = make(map[string]*controller)
	}

	structType := controllerType.Elem()
	controllerInstance := reflect.New(structType).Interface()
	controllerValue := reflect.ValueOf(controllerInstance).Elem()

	// TODO: set individual fields on controllerValue based on any field type and which is present in l.Deps.Get
	// only exported and maybe with a tag auto for example

	if field, found := structType.FieldByName("Deps"); found {
		ctrl := Controller{
			Deps:        l.Deps,
			FormDecoder: form.NewDecoder(),
			Validator:   validator.New(validator.WithRequiredStructEnabled()),
		}

		ctrl.Validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
			return strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]
		})

		gField := controllerValue.Field(field.Index[0])
		if gField.CanSet() {
			gField.Set(reflect.ValueOf(ctrl))
		}
	}

	if initMethod, hasInit := controllerType.MethodByName("Init"); hasInit {
		results := initMethod.Func.Call([]reflect.Value{
			reflect.ValueOf(controllerInstance),
		})

		if len(results) > 0 && !results[0].IsNil() {
			panic(results[0].Interface().(error))
		}
	}

	l.controllerRegistry[typeName] = &controller{
		Type:     controllerType,
		Instance: controllerInstance,
	}
}

// GET registers a new GET route for a path with a matching handler in the router
// with optional route-level middleware.
//
// ctrlAction is the name of the controller action pair to be called (lowercased and without Controller suffix):
// "users.index"
func (l *Loom) GET(path string, ctrlAction string, m ...echo.MiddlewareFunc) *echo.Route {
	return l.E.GET(path, l.handlerFor(ctrlAction), m...)
}

// POST registers a new POST route for a path with a matching handler in the router
// with optional route-level middleware.
//
// ctrlAction is the name of the controller action pair to be called (lowercased and without Controller suffix):
// "users.index"
func (l *Loom) POST(path string, ctrlAction string, m ...echo.MiddlewareFunc) *echo.Route {
	return l.E.POST(path, l.handlerFor(ctrlAction), m...)
}

func (l *Loom) handlerFor(ctrlAction string) echo.HandlerFunc {
	parts := strings.Split(ctrlAction, ".")

	if len(parts) != 2 {
		panic(fmt.Sprintf("Invalid controller action format: %s. Expected 'Type.Method'", ctrlAction))
	}

	controllerTypeName := cases.Title(language.English).String(strings.ReplaceAll(parts[0], "_", " ")) + "Controller"
	methodName := cases.Title(language.English).String(strings.ReplaceAll(parts[1], "_", " "))

	controllerTypeName = strings.ReplaceAll(controllerTypeName, " ", "")
	methodName = strings.ReplaceAll(methodName, " ", "")

	methodCall := l.getOrCreateMethodCall(controllerTypeName, methodName)

	controller, exists := l.controllerRegistry[controllerTypeName]
	if !exists {
		panic(fmt.Sprintf("Controller type %s not found in registry", controllerTypeName))
	}

	return func(c echo.Context) error {
		// Call the target method
		results := methodCall.method.Func.Call([]reflect.Value{
			reflect.ValueOf(controller.Instance),
			reflect.ValueOf(c),
		})

		// Return the result (should be error)
		if len(results) > 0 && !results[0].IsNil() {
			return results[0].Interface().(error)
		}

		return nil
	}
}

func (l *Loom) getOrCreateMethodCall(controllerTypeName, methodName string) *methodCall {
	if l.methodRegistry == nil {
		l.methodRegistry = make(map[string]*methodCall)
	}

	methodKey := fmt.Sprintf("%s.%s", controllerTypeName, methodName)

	if methodCall, exists := l.methodRegistry[methodKey]; exists {
		return methodCall
	}

	controllerType, exists := l.controllerRegistry[controllerTypeName]
	if !exists {
		panic(fmt.Sprintf("Controller type %s not found in registry", controllerTypeName))
	}

	method, found := controllerType.Type.MethodByName(methodName)
	if !found {
		panic(fmt.Sprintf("Method %s not found on controller type %s", methodName, controllerTypeName))
	}

	methodCall := &methodCall{
		controllerInstance: controllerType.Instance,
		method:             method,
	}

	l.methodRegistry[methodKey] = methodCall

	return methodCall
}
