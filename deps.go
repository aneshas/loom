package loom

import (
	"fmt"
	"reflect"
	"sync"
)

// Deps represents the dependency registry
type Deps struct {
	mu       sync.RWMutex
	services map[reflect.Type]map[string]interface{}
}

// NewDeps creates a new dependency registry
func NewDeps() *Deps {
	return &Deps{
		services: make(map[reflect.Type]map[string]interface{}),
	}
}

// getType returns the reflect.Type for a generic type T, handling both concrete and interface types
func getType[T any]() reflect.Type {
	var zero T
	t := reflect.TypeOf(zero)
	if t == nil {
		// For interface types, use reflect.TypeOf((*T)(nil)).Elem()
		return reflect.TypeOf((*T)(nil)).Elem()
	}
	return t
}

// Add registers a dependency with its type
// Usage: Add(deps, &MyService{})
func Add[T any](d *Deps, service T) {
	AddWithLabel(d, service, "")
}

// AddWithLabel registers a dependency with a specific label
// Usage: AddWithLabel(deps, &MyService{}, "primary")
func AddWithLabel[T any](d *Deps, service T, label string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// For interface types, we want to store under the interface type, not the concrete type
	serviceType := getType[T]()

	if d.services[serviceType] == nil {
		d.services[serviceType] = make(map[string]interface{})
	}
	d.services[serviceType][label] = service
}

// Get retrieves a dependency by type
// Usage: service, err := Get[MyService](deps)
func Get[T any](d *Deps) (T, error) {
	return GetWithLabel[T](d, "")
}

// GetWithLabel retrieves a dependency by type and label
// Usage: service, err := GetWithLabel[MyService](deps, "primary")
func GetWithLabel[T any](d *Deps, label string) (T, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var zero T
	serviceType := getType[T]()

	services, exists := d.services[serviceType]
	if !exists {
		return zero, fmt.Errorf("no service registered for type %v", serviceType)
	}

	service, exists := services[label]
	if !exists {
		return zero, fmt.Errorf("no service registered for type %v with label '%s'", serviceType, label)
	}

	// Type assertion
	if result, ok := service.(T); ok {
		return result, nil
	}

	return zero, fmt.Errorf("type assertion failed for type %v", serviceType)
}

// MustGet retrieves a dependency by type, panicking if not found
// Usage: service := MustGet[MyService](deps)
func MustGet[T any](d *Deps) T {
	service, err := Get[T](d)
	if err != nil {
		panic(err)
	}
	return service
}

// MustGetWithLabel retrieves a dependency by type and label, panicking if not found
// Usage: service := MustGetWithLabel[MyService](deps, "primary")
func MustGetWithLabel[T any](d *Deps, label string) T {
	service, err := GetWithLabel[T](d, label)
	if err != nil {
		panic(err)
	}
	return service
}

// Has checks if a dependency of the given type is registered
// Usage: if Has[MyService](deps) { ... }
func Has[T any](d *Deps) bool {
	return HasWithLabel[T](d, "")
}

// HasWithLabel checks if a dependency of the given type with the specified label is registered
// Usage: if HasWithLabel[MyService](deps, "primary") { ... }
func HasWithLabel[T any](d *Deps, label string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	serviceType := getType[T]()

	services, exists := d.services[serviceType]
	if !exists {
		return false
	}

	_, exists = services[label]
	return exists
}

// Remove removes a dependency by type
// Usage: Remove[MyService](deps)
func Remove[T any](d *Deps) {
	RemoveWithLabel[T](d, "")
}

// RemoveWithLabel removes a dependency by type and label
// Usage: RemoveWithLabel[MyService](deps, "primary")
func RemoveWithLabel[T any](d *Deps, label string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	serviceType := getType[T]()

	if services, exists := d.services[serviceType]; exists {
		delete(services, label)
		if len(services) == 0 {
			delete(d.services, serviceType)
		}
	}
}

// GetAll returns all registered dependencies of a given type
// Usage: services := GetAll[MyService](deps)
func GetAll[T any](d *Deps) map[string]T {
	d.mu.RLock()
	defer d.mu.RUnlock()

	serviceType := getType[T]()

	services, exists := d.services[serviceType]
	if !exists {
		return make(map[string]T)
	}

	result := make(map[string]T)
	for label, service := range services {
		if typedService, ok := service.(T); ok {
			result[label] = typedService
		}
	}

	return result
}

// Count returns the number of registered dependencies of a given type
// Usage: count := Count[MyService](deps)
func Count[T any](d *Deps) int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	serviceType := getType[T]()

	services, exists := d.services[serviceType]
	if !exists {
		return 0
	}

	return len(services)
}

// Clear removes all registered dependencies
func (d *Deps) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.services = make(map[reflect.Type]map[string]interface{})
}

// GetRegisteredTypes returns all registered types
func (d *Deps) GetRegisteredTypes() []reflect.Type {
	d.mu.RLock()
	defer d.mu.RUnlock()

	types := make([]reflect.Type, 0, len(d.services))
	for t := range d.services {
		types = append(types, t)
	}
	return types
}
