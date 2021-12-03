package v1

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
)

// HelmValues is typedef for Helm .Values
// +kubebuilder:validation:Type=object
// +kubebuilder:validation:XPreserveUnknownFields
type HelmValues struct {
	data map[string]interface{} `json:"-"`
}

func NewHelmValues(values map[string]interface{}) *HelmValues {
	if values == nil {
		values = make(map[string]interface{})
	}
	return &HelmValues{data: values}
}

func (h *HelmValues) GetContent() map[string]interface{} {
	if h == nil {
		return nil
	}
	return h.data
}

func (h *HelmValues) GetFieldNoCopy(path string) (interface{}, bool, error) {
	if h == nil || h.data == nil {
		return nil, false, nil
	}
	return unstructured.NestedFieldNoCopy(h.data, strings.Split(path, ".")...)
}

func (h *HelmValues) GetBool(path string) (bool, bool, error) {
	if h == nil || h.data == nil {
		return false, false, nil
	}
	val, found, err := unstructured.NestedFieldNoCopy(h.data, strings.Split(path, ".")...)
	if !found || err != nil {
		return false, found, err
	}
	b, ok := val.(bool)
	if !ok {
		if val == nil {
			return false, false, nil
		}
		return false, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected bool", path, val, val)
	}
	return b, true, nil
}

func (h *HelmValues) GetAndRemoveBool(path string) (bool, bool, error) {
	value, ok, err := h.GetBool(path)
	if err == nil {
		h.RemoveField(path)
	}
	return value, ok, err
}

func (h *HelmValues) GetString(path string) (string, bool, error) {
	if h == nil || h.data == nil {
		return "", false, nil
	}
	val, found, err := unstructured.NestedFieldNoCopy(h.data, strings.Split(path, ".")...)
	if !found || err != nil {
		return "", found, err
	}
	s, ok := val.(string)
	if !ok {
		if val == nil {
			return "", false, nil
		}
		return "", false, fmt.Errorf("%v accessor error: %v is of the type %T, expected string", path, val, val)
	}
	return s, true, nil
}

func (h *HelmValues) GetForceNumberToString(path string) (string, bool, error) {
	if h == nil || h.data == nil {
		return "", false, nil
	}
	value, ok, err := unstructured.NestedFieldNoCopy(h.data, strings.Split(path, ".")...)
	if err != nil {
		return "", false, err
	} else if !ok {
		return "", false, nil
	}
	switch typeValue := value.(type) {
	case int64:
		return strconv.FormatInt(typeValue, 10), ok, nil
	case float64:
		return strconv.FormatFloat(typeValue, 'f', -1, 64), ok, nil
	case string:
		return typeValue, ok, nil
	case nil:
		return "", false, nil
	}
	return "", false, fmt.Errorf("could not convert type to string: %T=%s", value, value)
}

func (h *HelmValues) GetAndRemoveString(path string) (string, bool, error) {
	value, ok, err := h.GetString(path)
	if err == nil {
		h.RemoveField(path)
	}
	return value, ok, err
}

func (h *HelmValues) GetAndRemoveForceNumberToString(path string) (string, bool, error) {
	value, ok, err := h.GetForceNumberToString(path)
	if err == nil {
		h.RemoveField(path)
	}
	return value, ok, err
}

func (h *HelmValues) GetInt64(path string) (int64, bool, error) {
	if h == nil || h.data == nil {
		return 0, false, nil
	}
	val, found, err := unstructured.NestedFieldNoCopy(h.data, strings.Split(path, ".")...)
	if !found || err != nil {
		return 0, found, err
	}
	i, ok := val.(int64)
	if !ok {
		if val == nil {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected int64", path, val, val)
	}
	return i, true, nil
}

func (h *HelmValues) GetAndRemoveInt64(path string) (int64, bool, error) {
	value, ok, err := h.GetInt64(path)
	if err == nil {
		h.RemoveField(path)
	}
	return value, ok, err
}

func (h *HelmValues) GetFloat64(path string) (float64, bool, error) {
	if h == nil || h.data == nil {
		return 0, false, nil
	}
	val, found, err := unstructured.NestedFieldNoCopy(h.data, strings.Split(path, ".")...)
	if !found || err != nil {
		return 0, found, err
	}
	f, ok := val.(float64)
	if !ok {
		if val == nil {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected float64", path, val, val)
	}
	return f, true, nil
}

func (h *HelmValues) GetAndRemoveFloat64(path string) (float64, bool, error) {
	value, ok, err := h.GetFloat64(path)
	if err == nil {
		h.RemoveField(path)
	}
	return value, ok, err
}

func (h *HelmValues) GetStringSlice(path string) ([]string, bool, error) {
	if h == nil || h.data == nil {
		return nil, false, nil
	}
	slice, ok, err := unstructured.NestedStringSlice(h.data, strings.Split(path, ".")...)
	if err != nil {
		if val, _, _ := h.GetFieldNoCopy(path); val == nil {
			return nil, false, nil
		}
	}
	return slice, ok, err
}

func (h *HelmValues) GetAndRemoveStringSlice(path string) ([]string, bool, error) {
	value, ok, err := h.GetStringSlice(path)
	if err == nil {
		h.RemoveField(path)
	}
	return value, ok, err
}

func (h *HelmValues) GetSlice(path string) ([]interface{}, bool, error) {
	if h == nil || h.data == nil {
		return nil, false, nil
	}
	slice, ok, err := unstructured.NestedSlice(h.data, strings.Split(path, ".")...)
	if err != nil {
		if val, _, _ := h.GetFieldNoCopy(path); val == nil {
			return nil, false, nil
		}
	}
	return slice, ok, err
}

func (h *HelmValues) GetAndRemoveSlice(path string) ([]interface{}, bool, error) {
	value, ok, err := h.GetSlice(path)
	if err == nil {
		h.RemoveField(path)
	}
	return value, ok, err
}

func (h *HelmValues) GetMap(path string) (map[string]interface{}, bool, error) {
	if h == nil || h.data == nil {
		return nil, false, nil
	}
	rawval, ok, err := unstructured.NestedFieldCopy(h.data, strings.Split(path, ".")...)
	if ok {
		if rawval == nil {
			return nil, ok, err
		}
		if mapval, ok := rawval.(map[string]interface{}); ok {
			return mapval, ok, err
		}
		return nil, false, fmt.Errorf("%v accessor error: %v is of the type %T, expected map[string]interface{}", path, rawval, rawval)
	}
	return nil, ok, err
}

func (h *HelmValues) SetField(path string, value interface{}) error {
	if h == nil {
		panic("Tried to invoke SetField on nil *HelmValues")
	}
	if h.data == nil {
		h.data = map[string]interface{}{}
	}
	return unstructured.SetNestedField(h.data, value, strings.Split(path, ".")...)
}

func (h *HelmValues) SetStringSlice(path string, value []string) error {
	if h == nil {
		panic("Tried to invoke SetField on nil *HelmValues")
	}
	if h.data == nil {
		h.data = map[string]interface{}{}
	}
	return unstructured.SetNestedStringSlice(h.data, value, strings.Split(path, ".")...)
}

func (h *HelmValues) RemoveField(path string) {
	if h == nil || h.data == nil {
		return
	}
	unstructured.RemoveNestedField(h.data, strings.Split(path, ".")...)
}

func (h *HelmValues) UnmarshalJSON(in []byte) error {
	err := json.Unmarshal(in, &h.data)
	if err != nil {
		return err
	}
	return nil
}

func (h *HelmValues) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.data)
}

func (in *HelmValues) DeepCopyInto(out *HelmValues) {
	*out = HelmValues{}

	data, err := json.Marshal(in)
	if err != nil {
		// panic ???
		return
	}
	err = json.Unmarshal(data, out)
	if err != nil {
		// panic ???
		return
	}
	return
}
