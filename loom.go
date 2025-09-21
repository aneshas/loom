package loom

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/labstack/echo/v4"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func New(deps *Deps) *Loom {
	return &Loom{
		E:                  echo.New(),
		Deps:               deps,
		controllerRegistry: make(map[string]reflect.Type),
		methodRegistry:     make(map[string]*MethodCall),
	}
}

type Loom struct {
	E *echo.Echo
	*Deps

	// Registry for controller types
	controllerRegistry map[string]reflect.Type
	// Registry for method calls
	methodRegistry map[string]*MethodCall
}

// MethodCall represents a pre-computed method call
type MethodCall struct {
	controllerType  reflect.Type
	method          reflect.Method
	gFieldIndex     int
	initMethodIndex int
	initMethod      reflect.Method
	hasInit         bool
}

func (g *Loom) Start(addr string) error {
	return g.E.Start(addr)
}

// GET registers a new GET route for a path with a matching handler in the router
// with optional route-level middleware.
//
// ctrlAction is the name of the controller action pair to be called (lowercased and without Controller suffix):
// "users.index"
func (g *Loom) GET(path string, ctrlAction string, m ...echo.MiddlewareFunc) *echo.Route {
	// Parse controller action in format "Type.Method"
	parts := strings.Split(ctrlAction, ".")
	if len(parts) != 2 {
		panic(fmt.Sprintf("Invalid controller action format: %s. Expected 'Type.Method'", ctrlAction))
	}

	ctrl := Controller{
		Deps: g.Deps,
	}

	// Convert from snake_case to PascalCase using golang.org/x/text/cases
	controllerTypeName := cases.Title(language.English).String(strings.ReplaceAll(parts[0], "_", " ")) + "Controller"
	methodName := cases.Title(language.English).String(strings.ReplaceAll(parts[1], "_", " "))

	controllerTypeName = strings.ReplaceAll(controllerTypeName, " ", "")
	methodName = strings.ReplaceAll(methodName, " ", "")

	// Get or create the method call from registry
	methodCall := g.getOrCreateMethodCall(controllerTypeName, methodName)

	// Create a handler function that will call the pre-registered method
	handler := func(c echo.Context) error {
		// Create a new instance of the controller
		var controllerInstance interface{}
		if methodCall.controllerType.Kind() == reflect.Ptr {
			// If it's a pointer type, create a new instance of the element type
			controllerInstance = reflect.New(methodCall.controllerType.Elem()).Interface()
		} else {
			// If it's a struct type, create a new instance
			controllerInstance = reflect.New(methodCall.controllerType).Interface()
		}

		controllerValue := reflect.ValueOf(controllerInstance).Elem()

		// Set the Loom dependency if the controller has a G field
		if methodCall.gFieldIndex >= 0 {
			gField := controllerValue.Field(methodCall.gFieldIndex)
			if gField.CanSet() {
				gField.Set(reflect.ValueOf(ctrl))
			}
		}

		// Call Init method if it exists
		if methodCall.hasInit {
			// Call the Init method directly using the stored method
			results := methodCall.initMethod.Func.Call([]reflect.Value{
				reflect.ValueOf(controllerInstance),
			})
			if len(results) > 0 && !results[0].IsNil() {
				return results[0].Interface().(error)
			}
		}

		// Call the target method
		results := methodCall.method.Func.Call([]reflect.Value{
			reflect.ValueOf(controllerInstance),
			reflect.ValueOf(c),
		})

		// Return the result (should be error)
		if len(results) > 0 && !results[0].IsNil() {
			return results[0].Interface().(error)
		}

		return nil
	}

	return g.E.GET(path, handler, m...)
}

// getOrCreateMethodCall gets or creates a method call from the registry
func (g *Loom) getOrCreateMethodCall(controllerTypeName, methodName string) *MethodCall {
	// Initialize method registry if needed
	if g.methodRegistry == nil {
		g.methodRegistry = make(map[string]*MethodCall)
	}

	// Create key for method call
	methodKey := fmt.Sprintf("%s.%s", controllerTypeName, methodName)

	// Check if method call already exists
	if methodCall, exists := g.methodRegistry[methodKey]; exists {
		return methodCall
	}

	// Get the controller type from the registry
	controllerType, exists := g.controllerRegistry[controllerTypeName]
	if !exists {
		panic(fmt.Sprintf("Controller type %s not found in registry", controllerTypeName))
	}

	// Find the method on the controller type
	method, found := controllerType.MethodByName(methodName)
	if !found {
		panic(fmt.Sprintf("Method %s not found on controller type %s", methodName, controllerTypeName))
	}

	// Get field and method information
	var gFieldIndex int = -1
	var initMethod reflect.Method
	var hasInit bool

	// Get the struct type for field lookup (if it's a pointer type, get the element type)
	structType := controllerType
	if controllerType.Kind() == reflect.Ptr {
		structType = controllerType.Elem()
	}

	// Check for G field
	if field, found := structType.FieldByName("Deps"); found {
		gFieldIndex = field.Index[0]
	}

	// Check for Init method
	if initMethod, hasInit = controllerType.MethodByName("Init"); hasInit {
		// Method found, keep it
	} else {
		hasInit = false
	}

	// Create and store the method call
	methodCall := &MethodCall{
		controllerType:  controllerType,
		method:          method,
		gFieldIndex:     gFieldIndex,
		initMethodIndex: -1, // We'll store the actual method instead of index
		hasInit:         hasInit,
	}

	// Store the Init method if it exists
	if hasInit {
		methodCall.initMethod = initMethod
	}

	g.methodRegistry[methodKey] = methodCall
	return methodCall
}

// RegisterController registers a controller type with the Loom instance
func (g *Loom) RegisterController(name string, controllerType reflect.Type) {
	if g.controllerRegistry == nil {
		g.controllerRegistry = make(map[string]reflect.Type)
	}
	g.controllerRegistry[name] = controllerType
}

// RegisterControllersFromPackage automatically registers all controller types from a package
// This is a helper function that can be used to register multiple controllers at once
func (g *Loom) RegisterControllersFromPackage(controllers map[string]reflect.Type) {
	if g.controllerRegistry == nil {
		g.controllerRegistry = make(map[string]reflect.Type)
	}
	for name, controllerType := range controllers {
		g.controllerRegistry[name] = controllerType
	}
}

// Register registers a controller type using its type name as the key
// Usage: Loom.Register[*controller.UsersController](g)
func Register[T any](g *Loom) {
	var zero T
	controllerType := reflect.TypeOf(zero)
	typeName := controllerType.Elem().Name()

	if g.controllerRegistry == nil {
		g.controllerRegistry = make(map[string]reflect.Type)
	}
	g.controllerRegistry[typeName] = controllerType
}
